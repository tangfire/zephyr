package main

import (
	"testing"
	"time"
)

func TestShouldWriteHTTPAccessLogAttentionMode(t *testing.T) {
	slow := 3 * time.Second
	tests := []struct {
		name    string
		status  int
		latency time.Duration
		path    string
		want    bool
	}{
		{name: "fast success is quiet", status: 200, latency: 20 * time.Millisecond, path: "/api/state", want: false},
		{name: "slow success is logged", status: 200, latency: 4 * time.Second, path: "/api/logs/query", want: true},
		{name: "client error is logged", status: 404, latency: 20 * time.Millisecond, path: "/missing", want: true},
		{name: "server error is logged", status: 502, latency: 20 * time.Millisecond, path: "/api/tasks/x/run", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldWriteHTTPAccessLog("attention", tt.status, tt.latency, slow, tt.path)
			if got != tt.want {
				t.Fatalf("shouldWriteHTTPAccessLog()=%v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldWriteHTTPAccessLogModes(t *testing.T) {
	if shouldWriteHTTPAccessLog("all", 200, time.Millisecond, 3*time.Second, "/healthz") {
		t.Fatal("all mode should still suppress healthy health checks")
	}
	if !shouldWriteHTTPAccessLog("all", 200, time.Millisecond, 3*time.Second, "/api/state") {
		t.Fatal("all mode should log non-health success")
	}
	if shouldWriteHTTPAccessLog("off", 500, 10*time.Second, 3*time.Second, "/api/state") {
		t.Fatal("off mode should suppress all access logs")
	}
}
