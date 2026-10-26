package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestProbePreservesHighLatencyWithoutRetriesOrFalseLoss(t *testing.T) {
	var calls atomic.Int32
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		time.Sleep(1100 * time.Millisecond)
		w.WriteHeader(204)
	}))
	defer endpoint.Close()
	result := measureProbe(context.Background(), ProbeTask{ID: 1, Type: "http", Target: endpoint.URL, TimeoutMs: 2000})
	if result.Status != "ok" || result.Value < 1000 || calls.Load() != 1 {
		t.Fatalf("high latency was hidden or retried: %#v calls=%d", result, calls.Load())
	}
}

func TestProbeTLSRedirectTimeoutAndInvalidTargets(t *testing.T) {
	var redirected atomic.Int32
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { redirected.Add(1) }))
	defer destination.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, destination.URL, 302) }))
	defer redirect.Close()
	if result := measureProbe(context.Background(), ProbeTask{ID: 1, Type: "http", Target: redirect.URL}); result.Status != "ok" || result.HTTPStatus != 302 || redirected.Load() != 0 {
		t.Fatal("probe followed a target's redirect")
	}
	tls := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer tls.Close()
	if result := measureProbe(context.Background(), ProbeTask{ID: 1, Type: "http", Target: tls.URL}); result.Status != "tls_error" {
		t.Fatalf("TLS check relaxed: %#v", result)
	}
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer slow.Close()
	start := time.Now()
	result := measureProbe(context.Background(), ProbeTask{ID: 1, Type: "http", Target: slow.URL, TimeoutMs: 100})
	if result.Status != "timeout" || time.Since(start) > time.Second {
		t.Fatalf("probe deadline exceeded: %#v", result)
	}
	for _, task := range []ProbeTask{{ID: 1, Type: "file", Target: "/etc/passwd"}, {ID: 1, Type: "http", Target: "https://user:password@localhost/"}, {ID: 1, Type: "tcp", Target: "localhost:999999"}, {ID: 1, Type: "icmp", Target: "bad host"}, {ID: 1, Type: "http", Target: slow.URL, TimeoutMs: 20000}} {
		if result := measureProbe(context.Background(), task); result.Status != "invalid_target" {
			t.Fatalf("invalid probe accepted: %#v", task)
		}
	}
	if probeFailure(syscall.EPERM).Status != "permission_denied" {
		t.Fatal("missing ICMP permission reported as packet loss")
	}
}

func TestProbeQueueConcurrencyDedupAndShutdownAreBounded(t *testing.T) {
	gate := make(chan struct{})
	var active, peak atomic.Int32
	runner := newProbeRunner(2, 3, func(ctx context.Context, task ProbeTask) ProbeResult {
		n := active.Add(1)
		defer active.Add(-1)
		for {
			old := peak.Load()
			if n <= old || peak.CompareAndSwap(old, n) {
				break
			}
		}
		select {
		case <-gate:
		case <-ctx.Done():
		}
		return ProbeResult{Status: "ok", FinishedAt: time.Now()}
	})
	defer runner.stop()
	delivered := make(chan struct{}, 8)
	emit := func(ProbeTask, ProbeResult) { delivered <- struct{}{} }
	if !runner.submit(ProbeTask{ID: 1}, emit) || runner.submit(ProbeTask{ID: 1}, emit) {
		t.Fatal("pending task was duplicated")
	}
	accepted := 1
	for n := uint(2); n < 200; n++ {
		if runner.submit(ProbeTask{ID: n}, emit) {
			accepted++
		}
	}
	if accepted > 5 {
		t.Fatalf("queue grew beyond active workers plus capacity: %d", accepted)
	}
	close(gate)
	for n := 0; n < accepted; n++ {
		select {
		case <-delivered:
		case <-time.After(time.Second):
			t.Fatal("bounded probe did not finish")
		}
	}
	if peak.Load() > 2 {
		t.Fatal("probe exceeded worker limit")
	}
}
