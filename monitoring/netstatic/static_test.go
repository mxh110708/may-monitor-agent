package netstatic

import (
	"math"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestConcurrentLifecycleAndReadsKeepOneWorkerGeneration(t *testing.T) {
	SaveFilePath = filepath.Join(t.TempDir(), "net.json")
	before := runtime.NumGoroutine()
	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 60; i++ {
				if err := StartOrContinue(); err != nil {
					t.Error(err)
					return
				}
				if err := SetNewConfig(NetStaticConfig{DetectInterval: .002, SaveInterval: .007}); err != nil {
					t.Error(err)
					return
				}
				GetNetStatic()
				GetNetStaticBetween(0, 0)
				GetTotalTraffic()
				GetTotalTrafficBetween(0, 0)
				if err := Stop(); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	wg.Wait()
	Stop()
	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > before+2 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if got := runtime.NumGoroutine(); got > before+2 {
		t.Fatalf("workers leaked: before=%d after=%d", before, got)
	}
}

func TestRetentionRestartCounterResetAndInvalidIntervals(t *testing.T) {
	SaveFilePath = filepath.Join(t.TempDir(), "net.json")
	Stop()
	Clear()
	if err := SetNewConfig(NetStaticConfig{DataPreserveDay: 1, DetectInterval: 2, SaveInterval: 600}); err != nil {
		t.Fatal(err)
	}
	now := nowUnix()
	if err := ForceReplaceRecord(map[string][]TrafficData{"test": {{Timestamp: now - 48*3600, Tx: 99, Rx: 99}, {Timestamp: now - 10, Tx: 3, Rx: 7}}}); err != nil {
		t.Fatal(err)
	}
	if err := SetNewConfig(NetStaticConfig{DataPreserveDay: 1}); err != nil {
		t.Fatal(err)
	}
	if err := StartOrContinue(); err != nil {
		t.Fatal(err)
	}
	totals, _ := GetTotalTraffic()
	if totals["test"].Tx != 3 || totals["test"].Rx != 7 {
		t.Fatal("expired history retained or totals lost on restart")
	}
	if safeDelta(2, 99) != 0 || safeDelta(99, 2) != 97 {
		t.Fatal("counter reset produced fictitious traffic")
	}
	for _, value := range []float64{-1, math.NaN(), math.Inf(1), 1e30} {
		if SetNewConfig(NetStaticConfig{DetectInterval: value}) == nil {
			t.Fatal("invalid interval accepted")
		}
	}
	current, _ := GetNetStatic()
	if current.Config.DetectInterval != 2 {
		t.Fatal("invalid configuration changed live sampling")
	}
	if err := Stop(); err != nil {
		t.Fatal(err)
	}
}
