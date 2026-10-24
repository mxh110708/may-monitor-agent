package server

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestICMPPing(t *testing.T) {
	latency, err := icmpPing("127.0.0.1", time.Second)
	if err != nil && (errors.Is(err, os.ErrPermission) || strings.Contains(err.Error(), "operation not permitted")) {
		t.Skip("loopback ICMP requires raw-socket permission")
	}
	if err != nil || latency < 0 {
		t.Fatalf("loopback ICMP: latency=%d err=%v", latency, err)
	}
}
func TestTCPPing(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			conn.Close()
		}
	}()
	latency, err := tcpPing(listener.Addr().String(), time.Second)
	if err != nil || latency < 0 {
		t.Fatalf("loopback TCP: latency=%d err=%v", latency, err)
	}
}
func TestHTTPPing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/unavailable" {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	if latency, err := httpPing(server.URL, time.Second); err != nil || latency < 0 {
		t.Fatalf("loopback HTTP: latency=%d err=%v", latency, err)
	}
	if _, err := httpPing(server.URL+"/unavailable", time.Second); err == nil {
		t.Fatal("HTTP error status accepted")
	}
}
