package ssl

import (
	"strings"
	"testing"
)

func TestIssueCommand(t *testing.T) {
	svc := NewService("/etc/letsencrypt")
	cmd, err := svc.IssueCommand(Request{
		Domain:  "app.example.com",
		Email:   "ops@example.com",
		Webroot: "/var/www/html",
	})
	if err != nil {
		t.Fatalf("IssueCommand failed: %v", err)
	}
	for _, want := range []string{"certbot certonly", "-d app.example.com", "--email ops@example.com", "--webroot -w /var/www/html"} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("command missing %q: %s", want, cmd)
		}
	}
}

func TestPaths(t *testing.T) {
	svc := NewService("/etc/letsencrypt")
	cert, key, err := svc.Paths("app.example.com")
	if err != nil {
		t.Fatalf("Paths failed: %v", err)
	}
	if !strings.HasSuffix(cert, "/live/app.example.com/fullchain.pem") || !strings.HasSuffix(key, "/live/app.example.com/privkey.pem") {
		t.Fatalf("unexpected paths: %s %s", cert, key)
	}
}

func TestRejectInvalidDomain(t *testing.T) {
	svc := NewService("/etc/letsencrypt")
	if _, err := svc.IssueCommand(Request{Domain: "bad_domain", Email: "ops@example.com", Webroot: "/var/www/html"}); err == nil {
		t.Fatal("expected validation error")
	}
}
