package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	systemdUnitDir = func() string {
		if v := os.Getenv("RAZAD_SYSTEMD_UNIT_DIR"); v != "" {
			return v
		}
		return "/etc/systemd/system"
	}
	systemctlExec = exec.CommandContext
)

// systemdRunner manages services through systemctl and unit files.
type systemdRunner struct{}

func (r *systemdRunner) Start(ctx context.Context, name, command string, env []string, workDir string) error {
	unitName := unitFileName(name)
	unitPath := filepath.Join(systemdUnitDir(), unitName)
	if err := os.MkdirAll(filepath.Dir(unitPath), 0o755); err != nil {
		return fmt.Errorf("process: create unit dir: %w", err)
	}
	if err := os.WriteFile(unitPath, []byte(renderUnit(name, command, env, workDir)), 0o644); err != nil {
		return fmt.Errorf("process: write unit file: %w", err)
	}
	if _, err := runSystemctl(ctx, "daemon-reload"); err != nil {
		return fmt.Errorf("process: systemctl daemon-reload: %w", err)
	}
	if _, err := runSystemctl(ctx, "enable", "--now", unitName); err != nil {
		return fmt.Errorf("process: systemctl enable --now: %w", err)
	}
	return nil
}

func (r *systemdRunner) Stop(ctx context.Context, name string) error {
	unitName := unitFileName(name)
	if _, err := runSystemctl(ctx, "stop", unitName); err != nil {
		return fmt.Errorf("process: systemctl stop: %w", err)
	}
	return nil
}

func (r *systemdRunner) Restart(ctx context.Context, name string) error {
	unitName := unitFileName(name)
	if _, err := runSystemctl(ctx, "restart", unitName); err != nil {
		return fmt.Errorf("process: systemctl restart: %w", err)
	}
	return nil
}

func (r *systemdRunner) Status(ctx context.Context, name string) (ProcessState, error) {
	unitName := unitFileName(name)
	out, err := runSystemctl(ctx, "is-active", unitName)
	state := strings.TrimSpace(out)
	switch state {
	case "active":
		return StateRunning, nil
	case "inactive", "deactivating":
		return StateStopped, nil
	case "failed":
		return StateFailed, nil
	case "activating":
		return StateRunning, nil
	default:
		if err != nil {
			return StateUnknown, err
		}
		return StateUnknown, nil
	}
}

func runSystemctl(ctx context.Context, args ...string) (string, error) {
	cmd := systemctlExec(ctx, "systemctl", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func unitFileName(name string) string {
	clean := unitNameSanitizer.ReplaceAllString(strings.TrimSpace(name), "-")
	clean = strings.Trim(clean, "-")
	if clean == "" {
		clean = "app"
	}
	return "razad-" + clean + ".service"
}

func renderUnit(name, command string, env []string, workDir string) string {
	var b strings.Builder
	b.WriteString("[Unit]\n")
	b.WriteString(fmt.Sprintf("Description=Razad app %s\n", name))
	b.WriteString("After=network.target\n\n")
	b.WriteString("[Service]\n")
	b.WriteString("Type=simple\n")
	if workDir != "" {
		b.WriteString(fmt.Sprintf("WorkingDirectory=%s\n", workDir))
	}
	for _, item := range env {
		if item == "" || !strings.Contains(item, "=") {
			continue
		}
		b.WriteString(fmt.Sprintf("Environment=%q\n", item))
	}
	b.WriteString(fmt.Sprintf("ExecStart=/bin/sh -lc %q\n", command))
	b.WriteString("Restart=always\n")
	b.WriteString("RestartSec=5\n\n")
	b.WriteString("[Install]\nWantedBy=multi-user.target\n")
	return b.String()
}

var unitNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9_.@-]+`)

var _ Runner = (*systemdRunner)(nil)
