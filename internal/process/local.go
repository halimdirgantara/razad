package process

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// localRunner manages processes by forking exec.Cmd instances and tracking
// them via PID files. Used for development and testing where systemd is
// not available.
type localRunner struct {
	dataDir string
	pidDir  string
	procs   map[string]*exec.Cmd // in-memory tracker for active processes
}

func newLocalRunner(dataDir string) *localRunner {
	pidDir := filepath.Join(dataDir, "pids")
	return &localRunner{
		dataDir: dataDir,
		pidDir:  pidDir,
		procs:   make(map[string]*exec.Cmd),
	}
}

func (r *localRunner) Start(ctx context.Context, name, command string, env []string, workDir string) error {
	// Ensure PID directory exists
	if err := os.MkdirAll(r.pidDir, 0755); err != nil {
		return fmt.Errorf("process: create pid dir: %w", err)
	}

	// Check if already running
	state, _ := r.Status(ctx, name)
	if state == StateRunning {
		return nil // already running — idempotent
	}

	// Parse command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("%w: empty command for %s", ErrStartFailed, name)
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), env...)

	// Capture stdout/stderr to app log file
	logDir := filepath.Join(r.dataDir, "logs", name)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("process: create log dir: %w", err)
	}

	logFile := filepath.Join(logDir, "output.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("process: open log file: %w", err)
	}

	cmd.Stdout = f
	cmd.Stderr = f

	// Start the process
	if err := cmd.Start(); err != nil {
		f.Close()
		return fmt.Errorf("%w: %v", ErrStartFailed, err)
	}

	// Write PID file
	pidFile := filepath.Join(r.pidDir, name+".pid")
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644); err != nil {
		slog.Warn("process: failed to write pid file", "name", name, "error", err)
	}

	// Track in memory
	r.procs[name] = cmd

	slog.Info("process started", "name", name, "pid", cmd.Process.Pid, "workDir", workDir)
	return nil
}

func (r *localRunner) Stop(ctx context.Context, name string) error {
	pid, err := r.readPID(name)
	if err != nil {
		return err
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		r.cleanupPID(name)
		return nil // process already gone
	}

	// Send SIGTERM for graceful shutdown
	slog.Debug("process: sending SIGTERM", "name", name, "pid", pid)
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		// Process might have already exited
		r.cleanupPID(name)
		return nil
	}

	// Wait a moment, then SIGKILL if still running
	done := make(chan struct{}, 1)
	go func() {
		proc.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("process stopped gracefully", "name", name)
	case <-time.After(5 * time.Second):
		slog.Warn("process: force killing", "name", name, "pid", pid)
		proc.Signal(syscall.SIGKILL)
		<-done
	}

	r.cleanupPID(name)
	delete(r.procs, name)
	return nil
}

func (r *localRunner) Restart(ctx context.Context, name string) error {
	if err := r.Stop(ctx, name); err != nil && err != ErrNotRunning {
		return err
	}

	// The caller must call Start again with the stored config
	return nil
}

func (r *localRunner) Status(ctx context.Context, name string) (ProcessState, error) {
	// Check in-memory first
	if cmd, ok := r.procs[name]; ok {
		if cmd.Process == nil {
			delete(r.procs, name)
			return StateStopped, nil
		}
		// Signal 0 checks if process exists
		if err := cmd.Process.Signal(syscall.Signal(0)); err == nil {
			return StateRunning, nil
		}
		delete(r.procs, name)
		return StateStopped, nil
	}

	// Fall back to PID file
	pid, err := r.readPID(name)
	if err != nil {
		return StateStopped, nil // no PID file means stopped
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		r.cleanupPID(name)
		return StateStopped, nil
	}

	if err := proc.Signal(syscall.Signal(0)); err == nil {
		return StateRunning, nil
	}

	r.cleanupPID(name)
	return StateStopped, nil
}

// readPID reads the PID from the PID file for the named process.
func (r *localRunner) readPID(name string) (int, error) {
	pidFile := filepath.Join(r.pidDir, name+".pid")
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, ErrNotFound
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		r.cleanupPID(name)
		return 0, ErrNotFound
	}

	return pid, nil
}

// cleanupPID removes the PID file for the named process.
func (r *localRunner) cleanupPID(name string) {
	pidFile := filepath.Join(r.pidDir, name+".pid")
	os.Remove(pidFile)
}

// ensure interface compliance
var _ Runner = (*localRunner)(nil)
