package process

import "context"

// systemdRunner wraps systemctl for production process management.
// It generates .service unit files and controls them via systemctl commands.
type systemdRunner struct{}

func (r *systemdRunner) Start(ctx context.Context, name, command string, env []string, workDir string) error {
	// TODO: generate unit file, systemctl daemon-reload, systemctl start
	return nil
}

func (r *systemdRunner) Stop(ctx context.Context, name string) error {
	// TODO: systemctl stop
	return nil
}

func (r *systemdRunner) Restart(ctx context.Context, name string) error {
	// TODO: systemctl restart
	return nil
}

func (r *systemdRunner) Status(ctx context.Context, name string) (ProcessState, error) {
	// TODO: systemctl status --output=json
	return StateUnknown, nil
}

var _ Runner = (*systemdRunner)(nil)
