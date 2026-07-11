// Package install implements the Razad daemon installer. It is designed to
// be idempotent so that re-running it on a partially-installed machine is
// safe: existing files are detected, unchanged files are left in place,
// only changed or missing files are written, and any failure aborts before
// mutating system state.
//
// The installer is also deliberately split into small, individually
// callable steps (CheckPrereqs, EnsureDirs, WriteUnitFile, ...). The top
// level Run() orchestrator chains them in order; tests and recovery
// scripts can call the steps directly to recover from a failed run
// without re-doing successful work.
package install

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Options configures an Installer. Zero values are replaced with safe
// defaults by New(); callers normally only need to override DataDir and
// BinaryPath.
type Options struct {
	// DataDir is the base directory for state. Default: /var/lib/razad.
	DataDir string

	// BinaryPath is the path of the razad-daemon binary. Default: /usr/local/bin/razad-daemon.
	BinaryPath string

	// UnitPath is where the systemd unit file is written. Default: /etc/systemd/system/razad-daemon.service.
	UnitPath string

	// User and Group are the system user/group the daemon runs as. Default: razad / razad.
	User  string
	Group string

	// SkipSystemd disables systemctl calls. Tests and non-systemd environments
	// must set this to true; production installers should leave it false.
	SkipSystemd bool

	// DryRun reports what would happen without writing or executing anything.
	// When true, the returned Result contains the actions that would have been
	// performed but nothing is mutated.
	DryRun bool
}

// Result is returned by Run() and describes what was changed.
type Result struct {
	// CreatedDirs lists directories created by this run (empty on a no-op).
	CreatedDirs []string
	// ExistingDirs lists directories that already existed and were left in place.
	ExistingDirs []string
	// WroteUnit is true if the unit file was written (or would be, under DryRun).
	WroteUnit bool
	// UnitPath is the absolute path of the unit file.
	UnitPath string
	// EnabledService is true if systemctl enable was invoked successfully.
	EnabledService bool
	// Warnings collects non-fatal issues (e.g., systemd not detected but
	// unit file still written).
	Warnings []string
}

// Installer performs an idempotent install.
type Installer struct {
	opts Options
}

// New returns an Installer with default values filled in.
func New(opts Options) *Installer {
	if opts.DataDir == "" {
		opts.DataDir = "/var/lib/razad"
	}
	if opts.BinaryPath == "" {
		opts.BinaryPath = "/usr/local/bin/razad-daemon"
	}
	if opts.UnitPath == "" {
		opts.UnitPath = "/etc/systemd/system/razad-daemon.service"
	}
	if opts.User == "" {
		opts.User = "razad"
	}
	if opts.Group == "" {
		opts.Group = "razad"
	}
	return &Installer{opts: opts}
}

// Run performs the full install sequence and returns a Result. On a failure
// during any step, the error is returned and the caller can inspect the
// partial Result (when non-nil) to understand what was done before the
// failure.
func (i *Installer) Run() (*Result, error) {
	if err := i.CheckPrereqs(); err != nil {
		return nil, err
	}
	res := &Result{UnitPath: i.opts.UnitPath}

	created, existing, err := i.EnsureDirs()
	if err != nil {
		return res, fmt.Errorf("install: ensure dirs: %w", err)
	}
	res.CreatedDirs = created
	res.ExistingDirs = existing

	if err := i.WriteUnitFile(res); err != nil {
		return res, err
	}

	if !i.opts.SkipSystemd {
		if err := i.EnableService(res); err != nil {
			return res, err
		}
	}
	return res, nil
}

// CheckPrereqs validates that the host can run Razad. The checks are
// deliberately conservative: it is better to refuse than to half-install.
func (i *Installer) CheckPrereqs() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("install: razad requires linux, got %s", runtime.GOOS)
	}
	if i.opts.SkipSystemd {
		return nil
	}
	if !systemdAvailable() {
		return errors.New("install: systemd is required (no /run/systemd/system and no systemctl binary)")
	}
	return nil
}

// EnsureDirs creates every directory the daemon needs at runtime and
// returns the lists of directories created vs already-existing.
// Idempotent: re-running on a populated DataDir returns empty CreatedDirs.
func (i *Installer) EnsureDirs() (created []string, existing []string, err error) {
	dirs := []string{
		i.opts.DataDir,
		filepath.Join(i.opts.DataDir, "nginx"),
		filepath.Join(i.opts.DataDir, "nginx", "sites-available"),
		filepath.Join(i.opts.DataDir, "nginx", "sites-enabled"),
		filepath.Join(i.opts.DataDir, "nginx", "backups"),
		filepath.Join(i.opts.DataDir, "letsencrypt"),
		filepath.Join(i.opts.DataDir, "logs"),
		filepath.Join(i.opts.DataDir, "apps"),
		filepath.Join(i.opts.DataDir, "databases"),
		filepath.Join(i.opts.DataDir, "health"),
		filepath.Join(i.opts.DataDir, "backups"),
		filepath.Join(i.opts.DataDir, "audit"),
	}
	for _, d := range dirs {
		info, statErr := os.Stat(d)
		if statErr == nil {
			if !info.IsDir() {
				return created, existing, fmt.Errorf("install: %s exists but is not a directory", d)
			}
			existing = append(existing, d)
			continue
		}
		if !os.IsNotExist(statErr) {
			return created, existing, fmt.Errorf("install: stat %s: %w", d, statErr)
		}
		if i.opts.DryRun {
			created = append(created, d)
			continue
		}
		if err := os.MkdirAll(d, 0o755); err != nil {
			return created, existing, fmt.Errorf("install: mkdir %s: %w", d, err)
		}
		created = append(created, d)
	}
	return created, existing, nil
}

// WriteUnitFile renders the systemd unit and writes it to opts.UnitPath
// unless the existing file already has the expected content. Sets
// res.WroteUnit to indicate whether a write happened (or would have,
// under DryRun).
func (i *Installer) WriteUnitFile(res *Result) error {
	content := renderUnitFile(i.opts)
	if i.opts.DryRun {
		res.WroteUnit = true
		return nil
	}
	existing, err := os.ReadFile(i.opts.UnitPath)
	if err == nil && strings.TrimSpace(string(existing)) == strings.TrimSpace(content) {
		// Already up to date.
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("install: read unit %s: %w", i.opts.UnitPath, err)
	}
	if err := os.MkdirAll(filepath.Dir(i.opts.UnitPath), 0o755); err != nil {
		return fmt.Errorf("install: mkdir %s: %w", filepath.Dir(i.opts.UnitPath), err)
	}
	if err := os.WriteFile(i.opts.UnitPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("install: write unit %s: %w", i.opts.UnitPath, err)
	}
	res.WroteUnit = true
	return nil
}

// EnableService runs `systemctl enable` and (if DryRun) records the action
// without executing it. When SkipSystemd is true this is a no-op (the
// caller is expected to manage the service themselves or run outside
// systemd).
func (i *Installer) EnableService(res *Result) error {
	if i.opts.SkipSystemd {
		return nil
	}
	if i.opts.DryRun {
		res.EnabledService = true
		return nil
	}
	cmd := exec.Command("systemctl", "enable", "razad-daemon.service")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("install: systemctl enable: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	res.EnabledService = true
	return nil
}

// systemdAvailable reports whether systemd is detectable on this host.
func systemdAvailable() bool {
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return true
	}
	if _, err := exec.LookPath("systemctl"); err == nil {
		return true
	}
	return false
}
