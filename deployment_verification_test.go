package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestApplyDeploymentVerificationMarkerAndHealthOK(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "current-source-sha")
	if err := os.WriteFile(marker, []byte("abcdef1234567890\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	status := &DeploymentStatus{CurrentCommit: "abcdef12"}
	applyDeploymentVerification(status, deploymentVerifyConfig{MarkerPath: marker, HealthURL: server.URL, Timeout: time.Second})

	if !status.DeployVerified {
		t.Fatalf("DeployVerified = false, message=%q", status.DeployVerifyMessage)
	}
	if status.DeployVerifyStatus != "verified" {
		t.Fatalf("DeployVerifyStatus = %q", status.DeployVerifyStatus)
	}
	if status.ActualCommit != "abcdef1234567890" {
		t.Fatalf("ActualCommit = %q", status.ActualCommit)
	}
}

func TestApplyDeploymentVerificationDetectsMarkerMismatch(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "current-source-sha")
	if err := os.WriteFile(marker, []byte("bbbbbbbb\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	status := &DeploymentStatus{CurrentCommit: "aaaaaaaa"}
	applyDeploymentVerification(status, deploymentVerifyConfig{MarkerPath: marker})

	if status.DeployVerified {
		t.Fatal("DeployVerified = true, want false")
	}
	if status.DeployVerifyStatus != "marker_mismatch" {
		t.Fatalf("DeployVerifyStatus = %q", status.DeployVerifyStatus)
	}
	if !strings.Contains(status.DeployVerifyMessage, "实际版本") {
		t.Fatalf("message = %q", status.DeployVerifyMessage)
	}
}

func TestApplyDeploymentVerificationPipelineOnly(t *testing.T) {
	status := &DeploymentStatus{CurrentCommit: "aaaaaaaa"}
	applyDeploymentVerification(status, deploymentVerifyConfig{})

	if status.DeployVerified {
		t.Fatal("DeployVerified = true, want false")
	}
	if status.DeployVerifyStatus != "pipeline_only" {
		t.Fatalf("DeployVerifyStatus = %q", status.DeployVerifyStatus)
	}
}
