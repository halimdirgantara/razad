package process

import (
	"testing"
	"context"
	"os"
	"path/filepath"
	"time"
)

func TestLocalRunner_StartStop(t *testing.T) {
	dir := t.TempDir()
	r := New(Config{Runner: "local", DataDir: dir})

	ctx := context.Background()

	// Start a simple sleep command
	err := r.Start(ctx, "test-svc", "sleep 2", nil, dir)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Check status
	state, err := r.Status(ctx, "test-svc")
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if state != StateRunning {
		t.Errorf("expected running, got %s", state)
	}

	// Stop
	err = r.Stop(ctx, "test-svc")
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Should be stopped now
	state, _ = r.Status(ctx, "test-svc")
	if state == StateRunning {
		t.Error("expected service to be stopped")
	}
}

func TestLocalRunner_IdempotentStart(t *testing.T) {
	dir := t.TempDir()
	r := New(Config{Runner: "local", DataDir: dir})

	ctx := context.Background()

	err := r.Start(ctx, "idempotent", "sleep 5", nil, dir)
	if err != nil {
		t.Fatalf("first start failed: %v", err)
	}

	// Starting again should be a no-op
	err = r.Start(ctx, "idempotent", "sleep 5", nil, dir)
	if err != nil {
		t.Errorf("second start should be idempotent: %v", err)
	}

	r.Stop(ctx, "idempotent")
}

func TestLocalRunner_StatusStopped(t *testing.T) {
	dir := t.TempDir()
	r := New(Config{Runner: "local", DataDir: dir})
	ctx := context.Background()

	state, err := r.Status(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Status for nonexistent should not error: %v", err)
	}
	if state != StateStopped {
		t.Errorf("expected stopped for nonexistent, got %s", state)
	}
}

func TestLocalRunner_Restart(t *testing.T) {
	dir := t.TempDir()
	r := New(Config{Runner: "local", DataDir: dir})
	ctx := context.Background()

	r.Start(ctx, "restart-test", "sleep 3", nil, dir)

	state, _ := r.Status(ctx, "restart-test")
	if state != StateRunning {
		t.Fatalf("expected running before restart")
	}

	r.Restart(ctx, "restart-test")

	state, _ = r.Status(ctx, "restart-test")
	if state != StateStopped {
		t.Errorf("expected stopped after restart (caller must Start), got %s", state)
	}
}

func TestLocalRunner_PIDFile(t *testing.T) {
	dir := t.TempDir()
	r := New(Config{Runner: "local", DataDir: dir})
	ctx := context.Background()

	r.Start(ctx, "pid-test", "sleep 3", nil, dir)
	defer r.Stop(ctx, "pid-test")

	// PID file should exist
	pidFile := filepath.Join(dir, "pids", "pid-test.pid")
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		t.Error("expected PID file to exist")
	}

	// Log file should exist
	logFile := filepath.Join(dir, "logs", "pid-test", "output.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("expected log file to exist")
	}
}

func TestLocalRunner_StartInvalidCommand(t *testing.T) {
	dir := t.TempDir()
	r := New(Config{Runner: "local", DataDir: dir})
	ctx := context.Background()

	err := r.Start(ctx, "bad", "", nil, dir)
	if err == nil {
		t.Error("expected error for empty command")
	}
}

func TestSystemdRunner_Stub(t *testing.T) {
	r := New(Config{Runner: "systemd", DataDir: t.TempDir()})
	ctx := context.Background()

	// Systemd runner is a stub — these just shouldn't panic
	_ = r.Start(ctx, "test", "echo hi", nil, "/tmp")
	_ = r.Stop(ctx, "test")
	state, _ := r.Status(ctx, "test")
	if state != StateUnknown {
		t.Log("systemd runner is a stub, expected unknown status")
	}
}

func TestNewRunner_DefaultLocal(t *testing.T) {
	r := New(Config{Runner: "", DataDir: t.TempDir()})
	_ = r.Start(context.Background(), "default-test", "sleep 1", nil, t.TempDir())
	_ = r.Stop(context.Background(), "default-test")
}

func TestLocalRunner_LogFileCreated(t *testing.T) {
	dir := t.TempDir()
	r := New(Config{Runner: "local", DataDir: dir})
	ctx := context.Background()

	r.Start(ctx, "logger", "echo hello", nil, dir)
	defer r.Stop(ctx, "logger")

	logFile := filepath.Join(dir, "logs", "logger", "output.log")
	// Give it a moment to write
	time.Sleep(200 * time.Millisecond)

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected log content, got empty file")
	}

	r.Stop(ctx, "logger")
}
