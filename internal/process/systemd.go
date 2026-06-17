package process

import (
	"context"
	"fmt"
)

// systemdRunner wraps systemctl for production process management.
// It generates .service unit files and controls them via systemctl commands.
type systemdRunner struct{}

func (r *systemdRunner) Start(ctx context.Context, name, command string, env []string, workDir string) error {
	return fmt.Errorf("%w: systemd support is not implemented yet", ErrUnsupportedRunner)
}

func (r *systemdRunner) Stop(ctx context.Context, name string) error {
	return fmt.Errorf("%w: systemd support is not implemented yet", ErrUnsupportedRunner)
}

func (r *systemdRunner) Restart(ctx context.Context, name string) error {
	return fmt.Errorf("%w: systemd support is not implemented yet", ErrUnsupportedRunner)
}

func (r *systemdRunner) Status(ctx context.Context, name string) (ProcessState, error) {
	return StateUnknown, fmt.Errorf("%w: systemd support is not implemented yet", ErrUnsupportedRunner)
}

var _ Runner = (*systemdRunner)(nil)
