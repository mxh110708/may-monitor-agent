package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gorilla/websocket"
)

func TestMonitoringCredentialsStayInHeadersAndDoNotRedirect(t *testing.T) {
	previousEndpoint, previousToken := flags.Endpoint, flags.Token
	defer func() { flags.Endpoint = previousEndpoint; flags.Token = previousToken }()
	flags.Token = "synthetic-monitor-token"
	var destinationCalls atomic.Int32
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { destinationCalls.Add(1) }))
	defer destination.Close()
	redirect := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" || r.Header.Get("X-May-Agent-Token") != flags.Token || r.Header.Get("Authorization") != "" {
			t.Error("agent credential escaped its dedicated header")
		}
		if redirect {
			http.Redirect(w, r, destination.URL, 302)
			return
		}
		if websocket.IsWebSocketUpgrade(r) {
			connection, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
			if err == nil {
				connection.Close()
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"status":"success"}}`))
	}))
	defer server.Close()
	flags.Endpoint = server.URL
	if strings.Contains(buildWebSocketEndpoint(), flags.Token) {
		t.Fatal("token remains in websocket URL")
	}
	connection, err := connectWebSocket(buildWebSocketEndpoint())
	if err != nil {
		t.Fatal(err)
	}
	connection.Close()
	if _, err := postV2Request([]byte(`{"jsonrpc":"2.0","id":1,"method":"agent.pull"}`)); err != nil {
		t.Fatal(err)
	}
	redirect = true
	if _, err := postV2Request([]byte(`{}`)); err == nil || destinationCalls.Load() != 0 {
		t.Fatal("monitoring token followed a redirect")
	}
}
