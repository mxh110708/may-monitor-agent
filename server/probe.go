package server

import (
	"context"
	"crypto/x509"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	v2 "github.com/komari-monitor/komari-agent/protocol/v2"
	"github.com/komari-monitor/komari-agent/ws"
	ping "github.com/prometheus-community/pro-bing"
)

type ProbeTask struct {
	ID        uint
	Type      string
	Target    string
	Revision  string
	TimeoutMs int
}

type ProbeResult struct {
	Status     string
	Value      int
	HTTPStatus int
	FinishedAt time.Time
}

func probeFailure(err error) ProbeResult {
	status := "network_error"
	var network net.Error
	var authority x509.UnknownAuthorityError
	var hostname x509.HostnameError
	var certificate x509.CertificateInvalidError
	if errors.Is(err, syscall.EPERM) || errors.Is(err, syscall.EACCES) {
		status = "permission_denied"
	} else if errors.As(err, &authority) || errors.As(err, &hostname) || errors.As(err, &certificate) {
		status = "tls_error"
	} else if errors.Is(err, context.DeadlineExceeded) || errors.As(err, &network) && network.Timeout() {
		status = "timeout"
	} else {
		var dns *net.DNSError
		if errors.As(err, &dns) {
			status = "dns_error"
		}
	}
	return ProbeResult{Status: status, Value: -1, FinishedAt: time.Now().UTC()}
}

// A single observation is retained as measured. Confirmation belongs to the
// alert policy, never to hidden retries that turn high latency into packet loss.
func measureProbe(parent context.Context, task ProbeTask) ProbeResult {
	failure := ProbeResult{Status: "invalid_target", Value: -1, FinishedAt: time.Now().UTC()}
	if task.TimeoutMs == 0 {
		task.TimeoutMs = 3000
	}
	if task.TimeoutMs < 100 || task.TimeoutMs > 10000 || task.ID == 0 || len(task.Target) == 0 || len(task.Target) > 2048 || strings.ContainsAny(task.Target, "\r\n\x00") {
		return failure
	}
	ctx, cancel := context.WithTimeout(parent, time.Duration(task.TimeoutMs)*time.Millisecond)
	defer cancel()
	finish := func(at time.Time) ProbeResult {
		return ProbeResult{Status: "ok", Value: int(time.Since(at).Milliseconds()), FinishedAt: time.Now().UTC()}
	}
	resolve := func(host string) (string, error) {
		if ip := net.ParseIP(host); ip != nil {
			return ip.String(), nil
		}
		addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return "", err
		}
		if len(addresses) == 0 {
			return "", &net.DNSError{Err: "no addresses", Name: host}
		}
		return addresses[0].IP.String(), nil
	}
	switch task.Type {
	case "tcp", "icmp":
		host, port, err := net.SplitHostPort(task.Target)
		if err != nil {
			host = task.Target
			port = "80"
		}
		host = strings.Trim(host, "[]")
		if strings.ContainsAny(host, " /?#@") {
			return failure
		}
		number, err := strconv.Atoi(port)
		if err != nil || number < 1 || number > 65535 {
			return failure
		}
		ip, err := resolve(host)
		if err != nil {
			return probeFailure(err)
		}
		if task.Type == "tcp" {
			started := time.Now()
			conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", net.JoinHostPort(ip, port))
			if err != nil {
				return probeFailure(err)
			}
			conn.Close()
			return finish(started)
		}
		pinger, err := ping.NewPinger(ip)
		if err != nil {
			return probeFailure(err)
		}
		pinger.Count = 1
		pinger.Timeout = time.Duration(task.TimeoutMs) * time.Millisecond
		pinger.SetPrivileged(true)
		if err = pinger.RunWithContext(ctx); err != nil {
			return probeFailure(err)
		}
		stat := pinger.Statistics()
		if stat.PacketsRecv == 0 {
			return probeFailure(context.DeadlineExceeded)
		}
		return ProbeResult{Status: "ok", Value: int(stat.AvgRtt.Milliseconds()), FinishedAt: time.Now().UTC()}
	case "http":
		target := task.Target
		if ip := net.ParseIP(target); ip != nil && ip.To4() == nil {
			target = "[" + target + "]"
		}
		if !strings.Contains(target, "://") {
			target = "http://" + target
		}
		parsed, err := url.Parse(target)
		if err != nil || parsed.Hostname() == "" || parsed.User != nil || parsed.Fragment != "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return failure
		}
		transport := &http.Transport{Proxy: nil, DisableKeepAlives: true, ResponseHeaderTimeout: time.Duration(task.TimeoutMs) * time.Millisecond}
		defer transport.CloseIdleConnections()
		client := &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			return failure
		}
		started := time.Now()
		response, err := client.Do(request)
		if err != nil {
			return probeFailure(err)
		}
		response.Body.Close()
		result := finish(started)
		result.HTTPStatus = response.StatusCode
		if response.StatusCode < 200 || response.StatusCode >= 400 {
			result.Status = "http_status"
			result.Value = -1
		}
		return result
	default:
		return failure
	}
}

type probeJob struct {
	task   ProbeTask
	emit   func(ProbeTask, ProbeResult)
	queued time.Time
}
type probeCompletion struct {
	job    probeJob
	result ProbeResult
}
type probeRunner struct {
	ctx     context.Context
	cancel  context.CancelFunc
	work    chan probeJob
	results chan probeCompletion
	mu      sync.Mutex
	pending map[string]bool
	wg      sync.WaitGroup
	measure func(context.Context, ProbeTask) ProbeResult
}

func newProbeRunner(workers, capacity int, measure func(context.Context, ProbeTask) ProbeResult) *probeRunner {
	ctx, cancel := context.WithCancel(context.Background())
	r := &probeRunner{ctx: ctx, cancel: cancel, work: make(chan probeJob, capacity), results: make(chan probeCompletion, capacity+workers), pending: map[string]bool{}, measure: measure}
	for n := 0; n < workers; n++ {
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			for {
				select {
				case <-r.ctx.Done():
					return
				case job := <-r.work:
					if r.ctx.Err() != nil {
						return
					}
					result := ProbeResult{Status: "busy", Value: -1, FinishedAt: time.Now().UTC()}
					if time.Since(job.queued) < 30*time.Second {
						result = r.measure(r.ctx, job.task)
					}
					select {
					case r.results <- probeCompletion{job, result}:
					case <-r.ctx.Done():
						return
					}
				}
			}
		}()
	}
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		for {
			select {
			case <-r.ctx.Done():
				return
			case item := <-r.results:
				if r.ctx.Err() != nil {
					return
				}
				item.job.emit(item.job.task, item.result)
				r.mu.Lock()
				delete(r.pending, probeKey(item.job.task))
				r.mu.Unlock()
			}
		}
	}()
	return r
}

func probeKey(task ProbeTask) string {
	return strconv.FormatUint(uint64(task.ID), 10) + ":" + task.Revision
}
func (r *probeRunner) submit(task ProbeTask, emit func(ProbeTask, ProbeResult)) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ctx.Err() != nil || r.pending[probeKey(task)] {
		return false
	}
	select {
	case r.work <- probeJob{task: task, emit: emit, queued: time.Now()}:
		r.pending[probeKey(task)] = true
		return true
	default:
		return false
	}
}
func (r *probeRunner) stop() { r.cancel(); r.wg.Wait() }

var probeOnce sync.Once
var agentProbes *probeRunner

func submitAgentProbe(conn *ws.SafeConn, task ProbeTask) {
	if task.ID == 0 || len(task.Target) > 2048 || len(task.Revision) > 64 {
		return
	}
	probeOnce.Do(func() { agentProbes = newProbeRunner(4, 32, measureProbe) })
	if !agentProbes.submit(task, func(task ProbeTask, result ProbeResult) {
		request := v2.BuildPingResultPayload(task.ID, task.Type, result.Value, result.FinishedAt).(v2.Request)
		params := request.Params.(map[string]interface{})
		params["status"] = result.Status
		params["revision"] = task.Revision
		if result.HTTPStatus != 0 {
			params["http_status"] = result.HTTPStatus
		}
		var err error
		if conn == nil {
			err = postV2RPC(request)
		} else {
			err = conn.WriteJSON(request)
		}
		if err != nil {
			log.Printf("Probe result was not delivered for task %d", task.ID)
		}
	}) {
		log.Printf("Probe task %d skipped: already pending or queue full", task.ID)
	}
}
