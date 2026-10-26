package server

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestICMPPing(t *testing.T) {
	result := measureProbe(context.Background(), ProbeTask{ID: 1, Type: "icmp", Target: "127.0.0.1", TimeoutMs: 1000})
	if result.Status == "permission_denied" {
		t.Skip("loopback ICMP requires raw-socket permission")
	}
	if result.Status != "ok" || result.Value < 0 {
		t.Fatalf("loopback ICMP: %#v", result)
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
	result := measureProbe(context.Background(), ProbeTask{ID: 1, Type: "tcp", Target: listener.Addr().String(), TimeoutMs: 1000})
	if result.Status != "ok" || result.Value < 0 {
		t.Fatalf("loopback TCP: %#v", result)
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
	if result := measureProbe(context.Background(), ProbeTask{ID: 1, Type: "http", Target: server.URL, TimeoutMs: 1000}); result.Status != "ok" || result.Value < 0 {
		t.Fatalf("loopback HTTP: %#v", result)
	}
	if result := measureProbe(context.Background(), ProbeTask{ID: 1, Type: "http", Target: server.URL + "/unavailable", TimeoutMs: 1000}); result.Status != "http_status" || result.HTTPStatus != 503 {
		t.Fatal("HTTP error status accepted")
	}
}
