// Package storage provides typed, configurable filesystem paths for the
// Razad daemon. Centralizing path computation here removes the hardcoded
// `filepath.Join(cfg.DataDir, ...)` calls that were previously scattered
// across main.go and gives future features (per-tenant layouts, alternative
// storage backends) a single integration point.
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Paths derives well-known directory and file locations from a base directory.
type Paths struct {
	base string
}

// New returns a Paths rooted at base. base must be an absolute path; if a
// relative path is passed it is resolved against the current working directory.
func New(base string) *Paths {
	if base == "" {
		base = "."
	}
	if !filepath.IsAbs(base) {
		if abs, err := filepath.Abs(base); err == nil {
			base = abs
		}
	}
	return &Paths{base: filepath.Clean(base)}
}

// Base returns the absolute base directory.
func (p *Paths) Base() string { return p.base }

// --- Top-level directories --------------------------------------------------

// Nginx is the base directory for Nginx configuration.
func (p *Paths) Nginx() string { return filepath.Join(p.base, "nginx") }

// NginxAvailable is where candidate site configs live before being enabled.
func (p *Paths) NginxAvailable() string { return filepath.Join(p.Nginx(), "sites-available") }

// NginxEnabled is where active site configs (symlinks or copies) live.
func (p *Paths) NginxEnabled() string { return filepath.Join(p.Nginx(), "sites-enabled") }

// NginxBackups is where last-known-good site configs are stored for rollback.
func (p *Paths) NginxBackups() string { return filepath.Join(p.Nginx(), "backups") }

// LetsEncrypt is the base directory for certbot-managed certificates.
func (p *Paths) LetsEncrypt() string { return filepath.Join(p.base, "letsencrypt") }

// Logs is the base directory for daemon-managed log files.
func (p *Paths) Logs() string { return filepath.Join(p.base, "logs") }

// Apps is the base directory for app workspaces.
func (p *Paths) Apps() string { return filepath.Join(p.base, "apps") }

// AppWorkspace is the working directory for a single app's source, build
// artifacts, and runtime files.
func (p *Paths) AppWorkspace(appID string) string {
	return filepath.Join(p.Apps(), sanitizeID(appID))
}

// Databases is the base directory for database instance data dirs.
func (p *Paths) Databases() string { return filepath.Join(p.base, "databases") }

// DatabaseData is the data directory for a single database instance.
func (p *Paths) DatabaseData(dbID string) string {
	return filepath.Join(p.Databases(), sanitizeID(dbID))
}

// Health is where health snapshots are persisted.
func (p *Paths) Health() string { return filepath.Join(p.base, "health") }

// Backups is the base directory for backups (DB, app data).
func (p *Paths) Backups() string { return filepath.Join(p.base, "backups") }

// BackupsDatabase is the directory for a single database instance's backups.
func (p *Paths) BackupsDatabase(dbID string) string {
	return filepath.Join(p.Backups(), "databases", sanitizeID(dbID))
}

// Audit is the directory where audit log exports are written.
func (p *Paths) Audit() string { return filepath.Join(p.base, "audit") }

// --- EnsureDirs -------------------------------------------------------------

// EnsureDirs creates every directory the daemon expects to find at startup.
// It is idempotent: existing directories are left untouched. Returns the
// first error encountered so the daemon can fail closed.
func (p *Paths) EnsureDirs() error {
	dirs := []string{
		p.Nginx(),
		p.NginxAvailable(),
		p.NginxEnabled(),
		p.NginxBackups(),
		p.LetsEncrypt(),
		p.Logs(),
		p.Apps(),
		p.Databases(),
		p.Health(),
		p.Backups(),
		p.Audit(),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("storage: create %s: %w", d, err)
		}
	}
	return nil
}

// sanitizeID defensively rejects path-traversal characters in identifiers
// passed to AppWorkspace/DatabaseData. Callers should pass UUIDs or similar
// opaque tokens, but this guard ensures a malformed ID cannot escape the
// data directory.
func sanitizeID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return "_"
	}
	cleaned := make([]rune, 0, len(id))
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-', r == '_':
			cleaned = append(cleaned, r)
		default:
			cleaned = append(cleaned, '_')
		}
	}
	out := string(cleaned)
	if out == "" || out == "." || out == ".." {
		return "_"
	}
	return out
}
