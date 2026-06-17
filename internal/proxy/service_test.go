package proxy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderHTTP(t *testing.T) {
	svc := NewService("/etc/nginx")
	cfg, err := svc.Render(Binding{
		Name:         "app-one",
		Domain:       "app.example.com",
		UpstreamHost: "127.0.0.1",
		UpstreamPort: 8080,
		TLS:          false,
		BodyLimitMB:  20,
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	for _, want := range []string{"server_name app.example.com;", "proxy_pass http://127.0.0.1:8080;", "client_max_body_size 20m;"} {
		if !strings.Contains(cfg, want) {
			t.Fatalf("config missing %q:\n%s", want, cfg)
		}
	}
}

func TestRenderHTTPS(t *testing.T) {
	svc := NewService("/etc/nginx")
	cfg, err := svc.Render(Binding{
		Name:         "app-two",
		Domain:       "app2.example.com",
		UpstreamHost: "10.0.0.8",
		UpstreamPort: 3000,
		TLS:          true,
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	for _, want := range []string{"listen 443 ssl http2;", "ssl_certificate /etc/letsencrypt/live/app2.example.com/fullchain.pem;", "return 301 https://$host$request_uri;"} {
		if !strings.Contains(cfg, want) {
			t.Fatalf("config missing %q:\n%s", want, cfg)
		}
	}
}

func TestApplyAndRollback(t *testing.T) {
	base := t.TempDir()
	svc := NewService(base)
	binding := Binding{
		Name:         "app-one",
		Domain:       "app.example.com",
		UpstreamHost: "127.0.0.1",
		UpstreamPort: 8080,
	}

	candidate, err := svc.WriteCandidate(binding)
	if err != nil {
		t.Fatalf("WriteCandidate failed: %v", err)
	}
	if _, err := os.Stat(candidate); err != nil {
		t.Fatalf("candidate missing: %v", err)
	}

	if err := svc.Apply(binding); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	enabled, _ := svc.EnabledPath(binding)
	data, err := os.ReadFile(enabled)
	if err != nil {
		t.Fatalf("enabled config missing: %v", err)
	}
	if !strings.Contains(string(data), "server_name app.example.com;") {
		t.Fatalf("enabled config not rendered correctly:\n%s", string(data))
	}

	backup, _ := svc.BackupPath(binding)
	if _, err := os.Stat(backup); err != nil {
		t.Fatalf("backup not created: %v", err)
	}

	if err := os.WriteFile(enabled, []byte("changed"), 0o644); err != nil {
		t.Fatalf("manual edit failed: %v", err)
	}
	if err := svc.Rollback(binding); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
	rolledBack, err := os.ReadFile(enabled)
	if err != nil {
		t.Fatalf("read rolled back config failed: %v", err)
	}
	if !strings.Contains(string(rolledBack), "server_name app.example.com;") {
		t.Fatalf("rollback did not restore config:\n%s", string(rolledBack))
	}
}

func TestValidateRejectsBadDomain(t *testing.T) {
	svc := NewService("/etc/nginx")
	if err := svc.Validate(Binding{Name: "x", Domain: "bad_domain", UpstreamHost: "127.0.0.1", UpstreamPort: 1}); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestPaths(t *testing.T) {
	svc := NewService("/tmp/nginx")
	binding := Binding{Name: "app", Domain: "app.example.com", UpstreamHost: "127.0.0.1", UpstreamPort: 8080}
	candidate, err := svc.CandidatePath(binding)
	if err != nil {
		t.Fatalf("CandidatePath failed: %v", err)
	}
	if got := filepath.Clean(candidate); !strings.HasSuffix(got, filepath.Join("sites-available", "app.conf")) {
		t.Fatalf("unexpected candidate path: %s", got)
	}
}
