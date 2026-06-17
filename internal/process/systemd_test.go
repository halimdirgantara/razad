package process

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSystemdRunner_RenderAndStart(t *testing.T) {
	unitDir := t.TempDir()
	oldUnitDir := systemdUnitDir
	oldExec := systemctlExec
	defer func() {
		systemdUnitDir = oldUnitDir
		systemctlExec = oldExec
	}()

	systemdUnitDir = func() string { return unitDir }
	calls := make([][]string, 0, 3)
	systemctlExec = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		calls = append(calls, append([]string{name}, args...))
		return exec.CommandContext(ctx, "sh", "-c", "exit 0")
	}

	r := New(Config{Runner: "systemd", DataDir: t.TempDir()})
	if err := r.Start(context.Background(), "my-app", "python app.py", []string{"PORT=8080"}, "/srv/my-app"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	unitPath := filepath.Join(unitDir, "razad-my-app.service")
	data, err := os.ReadFile(unitPath)
	if err != nil {
		t.Fatalf("expected unit file: %v", err)
	}
	content := string(data)
	for _, want := range []string{"Description=Razad app my-app", "WorkingDirectory=/srv/my-app", "Environment=\"PORT=8080\"", "ExecStart=/bin/sh -lc \"python app.py\""} {
		if !strings.Contains(content, want) {
			t.Fatalf("unit file missing %q\n%s", want, content)
		}
	}
	if len(calls) != 2 {
		t.Fatalf("expected 2 systemctl calls, got %d", len(calls))
	}
	if got := strings.Join(calls[1], " "); !strings.Contains(got, "enable --now razad-my-app.service") {
		t.Fatalf("unexpected enable call: %s", got)
	}
}
