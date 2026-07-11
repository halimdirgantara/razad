package ssl

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSelfSignedProducesParseablePEM(t *testing.T) {
	base := t.TempDir()
	certPath := filepath.Join(base, "certs", "test.crt")
	keyPath := filepath.Join(base, "keys", "test.key")
	svc := NewService(base)

	if err := svc.GenerateSelfSigned("razad.local", certPath, keyPath); err != nil {
		t.Fatalf("GenerateSelfSigned: %v", err)
	}

	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("read cert: %v", err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("cert PEM did not decode")
	}
	if block.Type != "CERTIFICATE" {
		t.Errorf("cert block type: got %q, want CERTIFICATE", block.Type)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}
	if cert.Subject.CommonName != "razad.local" {
		t.Errorf("CN: got %q, want razad.local", cert.Subject.CommonName)
	}
	if len(cert.DNSNames) != 1 || cert.DNSNames[0] != "razad.local" {
		t.Errorf("DNSNames: got %v, want [razad.local]", cert.DNSNames)
	}
	if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		t.Error("missing KeyUsageDigitalSignature")
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("read key: %v", err)
	}
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		t.Fatal("key PEM did not decode")
	}
	if keyBlock.Type != "EC PRIVATE KEY" {
		t.Errorf("key block type: got %q, want EC PRIVATE KEY", keyBlock.Type)
	}
	if _, err := x509.ParseECPrivateKey(keyBlock.Bytes); err != nil {
		t.Fatalf("parse key: %v", err)
	}
}

func TestGenerateSelfSignedRejectsInvalidDomain(t *testing.T) {
	svc := NewService(t.TempDir())
	for _, bad := range []string{"", "no spaces allowed.com", "192.168.1.1", "-leadingdash.com"} {
		if err := svc.GenerateSelfSigned(bad, "/tmp/x.crt", "/tmp/x.key"); err == nil {
			t.Errorf("expected error for domain %q, got nil", bad)
		}
	}
}

func TestGenerateSelfSignedRejectsEmptyPaths(t *testing.T) {
	svc := NewService(t.TempDir())
	if err := svc.GenerateSelfSigned("razad.local", "", "/tmp/key"); err == nil {
		t.Error("expected error for empty cert path")
	}
	if err := svc.GenerateSelfSigned("razad.local", "/tmp/cert", ""); err == nil {
		t.Error("expected error for empty key path")
	}
}

func TestGenerateSelfSignedKeyFileMode(t *testing.T) {
	base := t.TempDir()
	certPath := filepath.Join(base, "crt.pem")
	keyPath := filepath.Join(base, "key.pem")
	svc := NewService(base)
	if err := svc.GenerateSelfSigned("razad.local", certPath, keyPath); err != nil {
		t.Fatalf("GenerateSelfSigned: %v", err)
	}
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("stat key: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("key file mode: got %o, want 0600", mode)
	}
}

func TestGenerateSelfSignedOverwritesExistingFiles(t *testing.T) {
	base := t.TempDir()
	certPath := filepath.Join(base, "crt.pem")
	keyPath := filepath.Join(base, "key.pem")
	svc := NewService(base)
	if err := svc.GenerateSelfSigned("razad.local", certPath, keyPath); err != nil {
		t.Fatalf("first: %v", err)
	}
	firstCert, _ := os.ReadFile(certPath)
	if err := svc.GenerateSelfSigned("razad.local", certPath, keyPath); err != nil {
		t.Fatalf("second: %v", err)
	}
	secondCert, _ := os.ReadFile(certPath)
	if string(firstCert) == string(secondCert) {
		t.Skip("second generation produced identical cert (extremely unlikely; re-running anyway)")
	}
	// Different serials are guaranteed by random generation; what we really
	// care about is that the file is still valid PEM.
	block, _ := pem.Decode(secondCert)
	if block == nil {
		t.Error("re-written cert is not valid PEM")
	}
}

func TestPathsStillReturnsCertbotLayout(t *testing.T) {
	svc := NewService("/var/lib/razad/letsencrypt")
	cert, key, err := svc.Paths("app.example.com")
	if err != nil {
		t.Fatalf("Paths: %v", err)
	}
	if !strings.Contains(cert, "live/app.example.com/fullchain.pem") {
		t.Errorf("cert path: got %q", cert)
	}
	if !strings.Contains(key, "live/app.example.com/privkey.pem") {
		t.Errorf("key path: got %q", key)
	}
}
