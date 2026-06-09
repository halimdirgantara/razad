// Package process provides a pluggable ProcessRunner interface for managing
// application processes. Two implementations exist:
//   - localRunner: forks processes with PID tracking (dev/testing)
//   - systemdRunner: wraps systemctl for production Linux deployments
package process

import (
	"context"
	"errors"
)

// Common errors returned by process runners.
var (
	ErrNotFound    = errors.New("process: service not found")
	ErrNotRunning  = errors.New("process: service not running")
	ErrStartFailed = errors.New("process: failed to start")
	ErrStopFailed  = errors.New("process: failed to stop")
)

// ProcessState represents the current state of a managed process.
type ProcessState string

const (
	StateRunning ProcessState = "running"
	StateStopped ProcessState = "stopped"
	StateFailed  ProcessState = "failed"
	StateUnknown ProcessState = "unknown"
)

// Runner defines the interface for process lifecycle management.
type Runner interface {
	// Start launches a process with the given name, command, environment,
	// and working directory. Returns an error if start fails.
	Start(ctx context.Context, name string, command string, env []string, workDir string) error

	// Stop terminates a process gracefully (SIGTERM), then forcefully (SIGKILL)
	// after a timeout if needed.
	Stop(ctx context.Context, name string) error

	// Restart stops then starts the process.
	Restart(ctx context.Context, name string) error

	// Status returns the current state of the named process.
	Status(ctx context.Context, name string) (ProcessState, error)
}

// Config holds process runner configuration.
type Config struct {
	// Runner is the backend to use: "local" or "systemd".
	Runner string

	// DataDir is where PID files and app data are stored (local runner).
	DataDir string
}

// New creates a Runner based on the provided config.
// Returns a local runner by default.
func New(cfg Config) Runner {
	if cfg.Runner == "systemd" {
		return &systemdRunner{}
	}
	return newLocalRunner(cfg.DataDir)
}
