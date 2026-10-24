package cmd

import (
	pkg_flags "github.com/komari-monitor/komari-agent/cmd/flags"
	"testing"
)

func TestMonitorConnectionPolicy(t *testing.T) {
	for _, tt := range []struct {
		endpoint string
		ok       bool
	}{
		{"https://monitor.example", true}, {"http://127.0.0.1:12345", true}, {"http://[::1]:12345", true},
		{"http://monitor.example", false}, {"https://user:synthetic@monitor.example", false}, {"file:///tmp/synthetic", false}, {"https://monitor.example/?token=synthetic", false},
	} {
		cfg := &pkg_flags.Config{Endpoint: tt.endpoint, Interval: 3, ReconnectInterval: 5, InfoReportInterval: 5}
		if err := validateConnectionPolicy(cfg); (err == nil) != tt.ok {
			t.Errorf("endpoint=%q err=%v", tt.endpoint, err)
		}
	}
	if flag := RootCmd.PersistentFlags().Lookup("disable-web-ssh"); flag == nil || flag.DefValue != "true" {
		t.Fatal("remote control default is not disabled")
	}
	if flag := RootCmd.PersistentFlags().Lookup("disable-auto-update"); flag == nil || flag.DefValue != "true" {
		t.Fatal("automatic update default is not disabled")
	}
}
