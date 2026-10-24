package cmd

import (
	"fmt"
	"math"
	"net"
	"net/url"

	pkg_flags "github.com/komari-monitor/komari-agent/cmd/flags"
)

func validateConnectionPolicy(cfg *pkg_flags.Config) error {
	u, err := url.Parse(cfg.Endpoint)
	if err != nil || u.Host == "" || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("invalid monitor endpoint")
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("monitor endpoint must not contain query or fragment")
	}
	ip := net.ParseIP(u.Hostname())
	if u.Scheme != "https" && u.Hostname() != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return fmt.Errorf("HTTPS is required for non-loopback monitor endpoints")
	}
	if cfg.Interval < 1 || math.IsNaN(cfg.Interval) || math.IsInf(cfg.Interval, 0) {
		return fmt.Errorf("report interval must be at least one second")
	}
	if cfg.ReconnectInterval < 1 || cfg.InfoReportInterval < 1 {
		return fmt.Errorf("reconnect and inventory intervals must be positive")
	}
	return nil
}
