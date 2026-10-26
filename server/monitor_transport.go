package server

import (
	"github.com/komari-monitor/komari-agent/dnsresolver"
	"net/http"
	"time"
)

// Reuse the pooled transport without mutating the shared client's redirect
// policy. Monitoring credentials never follow a redirect to a different URL.
func monitorHTTPClient(timeout time.Duration) *http.Client {
	client := *dnsresolver.GetHTTPClientWithPreference(timeout, flags.PreferIPVersion)
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &client
}
