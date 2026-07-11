package install

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewAppliesDefaults(t *testing.T) {
	i := New(Options{})
	if i.opts.DataDir == "" {
		t.Error("DataDir default not applied")
	}
	if i.opts.BinaryPath == "" {
		t.Error("BinaryPath default not applied")
	}
	if i.opts.UnitPath == "" {
		t.Error("UnitPath default not applied")
	}
	if i.opts.User == "" {
		t.Error("User default not applied")
	}
	if i.opts.Group == "" {
		t.Error("Group default not applied")
	}
}

func TestCheckPrereqsRejectsNonLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("test only meaningful on non-linux hosts")
	}
	i := New(Options{SkipSystemd: true})
	if err := i.CheckPrereqs(); err == nil {
		t.Error("expected error on non-linux host")
	}
}

func TestCheckPrereqsAllowsLinuxWithSkipSystemd(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("test only meaningful on linux")
	}
	i := New(Options{SkipSystemd: true})
	if err := i.CheckPrereqs(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestEnsureDirsCreatesMissingAndReportsExisting(t *testing.T) {
	base := t.TempDir()
	pre := filepath.Join(base, "health")
	if err := os.MkdirAll(pre, 0o755); err != nil {
		t.Fatalf("pre-create: %v", err)
	}

	i := New(Options{DataDir: base, SkipSystemd: true})
	created, existing, err := i.EnsureDirs()
	if err != nil {
		t.Fatalf("EnsureDirs: %v", err)
	}
	// base (from t.TempDir) and pre (created above) should both be reported
	// as already existing; the other directories should be in `created`.
	foundPre := false
	for _, d := range existing {
		if d == pre {
			foundPre = true
		}
	}
	if !foundPre {
		t.Errorf("expected existing to contain %s, got %v", pre, existing)
	}
	for _, c := range created {
		if c == pre {
			t.Errorf("pre-created directory %s should not be in created list: %v", pre, created)
		}
	}
	for _, d := range []string{
		filepath.Join(base, "logs"),
		filepath.Join(base, "nginx"),
		filepath.Join(base, "apps"),
	} {
		found := false
		for _, c := range created {
			if c == d {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected created to include %s, got %v", d, created)
		}
		if _, err := os.Stat(d); err != nil {
			t.Errorf("expected %s to exist on disk: %v", d, err)
		}
	}
}

func TestEnsureDirsIsIdempotent(t *testing.T) {
	base := t.TempDir()
	i := New(Options{DataDir: base, SkipSystemd: true})

	_, _, err := i.EnsureDirs()
	if err != nil {
		t.Fatalf("first EnsureDirs: %v", err)
	}
	created, existing, err := i.EnsureDirs()
	if err != nil {
		t.Fatalf("second EnsureDirs: %v", err)
	}
	if len(created) != 0 {
		t.Errorf("second run should create nothing, got %v", created)
	}
	if len(existing) == 0 {
		t.Error("second run should report existing dirs")
	}
}

func TestEnsureDirsRejectsPathCollision(t *testing.T) {
	base := t.TempDir()
	// Create a regular file where a directory is expected.
	bad := filepath.Join(base, "logs")
	if err := os.WriteFile(bad, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	i := New(Options{DataDir: base, SkipSystemd: true})
	if _, _, err := i.EnsureDirs(); err == nil {
		t.Error("expected error when path is a regular file")
	}
}

func TestWriteUnitFileIsIdempotent(t *testing.T) {
	base := t.TempDir()
	unitPath := filepath.Join(base, "razad-daemon.service")
	i := New(Options{
		DataDir:     base,
		BinaryPath:  "/usr/local/bin/razad-daemon",
		UnitPath:    unitPath,
		SkipSystemd: true,
	})
	res := &Result{UnitPath: unitPath}

	if err := i.WriteUnitFile(res); err != nil {
		t.Fatalf("first WriteUnitFile: %v", err)
	}
	if !res.WroteUnit {
		t.Error("first call should report WroteUnit=true")
	}
	if _, err := os.ReadFile(unitPath); err != nil {
		t.Fatalf("read unit: %v", err)
	}

	// Second call: content unchanged → no write.
	res2 := &Result{UnitPath: unitPath}
	if err := i.WriteUnitFile(res2); err != nil {
		t.Fatalf("second WriteUnitFile: %v", err)
	}
	if res2.WroteUnit {
		t.Error("second call should report WroteUnit=false (content unchanged)")
	}
}

func TestWriteUnitFileRewritesWhenContentDiffers(t *testing.T) {
	base := t.TempDir()
	unitPath := filepath.Join(base, "razad-daemon.service")
	if err := os.WriteFile(unitPath, []byte("old different content"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	i := New(Options{
		DataDir:     base,
		BinaryPath:  "/usr/local/bin/razad-daemon",
		UnitPath:    unitPath,
		SkipSystemd: true,
	})
	res := &Result{UnitPath: unitPath}
	if err := i.WriteUnitFile(res); err != nil {
		t.Fatalf("WriteUnitFile: %v", err)
	}
	if !res.WroteUnit {
		t.Error("should report WroteUnit=true when content differs")
	}
	data, err := os.ReadFile(unitPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if strings.Contains(string(data), "old different content") {
		t.Error("unit file was not rewritten")
	}
}

func TestUnitFileContent(t *testing.T) {
	i := New(Options{
		DataDir:    "/var/lib/razad",
		BinaryPath: "/usr/local/bin/razad-daemon",
		User:       "razad",
		Group:      "razad",
	})
	out := renderUnitFile(i.opts)
	for _, want := range []string{
		"[Unit]",
		"[Service]",
		"[Install]",
		"User=razad",
		"Group=razad",
		"ExecStart=/usr/local/bin/razad-daemon",
		"WorkingDirectory=/var/lib/razad",
		"Restart=on-failure",
		"ReadWritePaths=/var/lib/razad",
		"WantedBy=multi-user.target",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("unit file missing %q", want)
		}
	}
}
