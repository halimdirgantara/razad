package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewResolvesRelativePath(t *testing.T) {
	p := New("./relative/data")
	if !filepath.IsAbs(p.Base()) {
		t.Fatalf("expected absolute base, got %q", p.Base())
	}
}

func TestPathsAreStable(t *testing.T) {
	base := "/var/lib/razad"
	p := New(base)

	cases := []struct {
		name string
		got  string
		want string
	}{
		{"Nginx", p.Nginx(), filepath.Join(base, "nginx")},
		{"NginxAvailable", p.NginxAvailable(), filepath.Join(base, "nginx", "sites-available")},
		{"NginxEnabled", p.NginxEnabled(), filepath.Join(base, "nginx", "sites-enabled")},
		{"NginxBackups", p.NginxBackups(), filepath.Join(base, "nginx", "backups")},
		{"LetsEncrypt", p.LetsEncrypt(), filepath.Join(base, "letsencrypt")},
		{"Logs", p.Logs(), filepath.Join(base, "logs")},
		{"Apps", p.Apps(), filepath.Join(base, "apps")},
		{"AppWorkspace", p.AppWorkspace("abc-123"), filepath.Join(base, "apps", "abc-123")},
		{"Databases", p.Databases(), filepath.Join(base, "databases")},
		{"DatabaseData", p.DatabaseData("db-1"), filepath.Join(base, "databases", "db-1")},
		{"Health", p.Health(), filepath.Join(base, "health")},
		{"Backups", p.Backups(), filepath.Join(base, "backups")},
		{"BackupsDatabase", p.BackupsDatabase("db-1"), filepath.Join(base, "backups", "databases", "db-1")},
		{"Audit", p.Audit(), filepath.Join(base, "audit")},
	}

	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

func TestAppWorkspaceBlocksTraversal(t *testing.T) {
	p := New("/var/lib/razad")
	bad := []string{"../../etc/passwd", "/etc/passwd", "..", "."}
	for _, id := range bad {
		got := p.AppWorkspace(id)
		if strings.Contains(got, "..") || strings.HasPrefix(got, "/etc") {
			t.Errorf("AppWorkspace(%q) leaked outside data dir: %q", id, got)
		}
	}
}

func TestEnsureDirsIsIdempotent(t *testing.T) {
	tmp := t.TempDir()
	p := New(tmp)

	if err := p.EnsureDirs(); err != nil {
		t.Fatalf("first EnsureDirs: %v", err)
	}
	if err := p.EnsureDirs(); err != nil {
		t.Fatalf("second EnsureDirs: %v", err)
	}
	for _, d := range []string{p.Nginx(), p.LetsEncrypt(), p.Logs(), p.Apps()} {
		info, err := os.Stat(d)
		if err != nil {
			t.Errorf("missing directory %s: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", d)
		}
	}
}
