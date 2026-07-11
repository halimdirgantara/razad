package ssl

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Request describes a cert issuance/renewal request.
type Request struct {
	Domain  string
	Email   string
	Webroot string
}

// Service provides certbot command scaffolding and path helpers.
type Service struct {
	baseDir string
}

// NewService creates an SSL service rooted at baseDir.
func NewService(baseDir string) *Service {
	return &Service{baseDir: baseDir}
}

// ValidateRequest checks whether the request is suitable for certbot.
func (s *Service) ValidateRequest(req Request) error {
	if err := validateDomain(req.Domain); err != nil {
		return fmt.Errorf("ssl: invalid domain: %w", err)
	}
	if strings.TrimSpace(req.Email) == "" || !strings.Contains(req.Email, "@") {
		return fmt.Errorf("ssl: email is required")
	}
	if strings.TrimSpace(req.Webroot) == "" {
		return fmt.Errorf("ssl: webroot is required")
	}
	return nil
}

// IssueCommand renders a certbot issuance command.
func (s *Service) IssueCommand(req Request) (string, error) {
	if err := s.ValidateRequest(req); err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"certbot certonly --webroot -w %s -d %s --email %s --agree-tos --non-interactive --keep-until-expiring",
		req.Webroot, req.Domain, req.Email,
	), nil
}

// RenewCommand renders a certbot renewal command for a specific domain.
func (s *Service) RenewCommand(domain string) (string, error) {
	if err := validateDomain(domain); err != nil {
		return "", fmt.Errorf("ssl: invalid domain: %w", err)
	}
	return fmt.Sprintf("certbot renew --cert-name %s", domain), nil
}

// Paths returns the certificate file locations for a domain.
func (s *Service) Paths(domain string) (certPath, keyPath string, err error) {
	if err := validateDomain(domain); err != nil {
		return "", "", fmt.Errorf("ssl: invalid domain: %w", err)
	}
	live := filepath.Join(s.baseDir, "live", domain)
	return filepath.Join(live, "fullchain.pem"), filepath.Join(live, "privkey.pem"), nil
}

var domainRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)+$`)

// SelfSignedValidity is how long a generated self-signed cert is valid for.
// 365 days matches Let's Encrypt's free cert lifetime.
const SelfSignedValidity = 365 * 24 * time.Hour

// GenerateSelfSigned mints an ECDSA P-256 self-signed certificate for the
// given domain and writes the PEM-encoded leaf cert and private key to
// certPath and keyPath respectively. Existing files are overwritten.
//
// Self-signed certs are appropriate for local development and quick trials
// where a real CA-issued cert is overkill. They will produce browser
// warnings. Production deployments should use certbot-issued certs.
func (s *Service) GenerateSelfSigned(domain, certPath, keyPath string) error {
	if err := validateDomain(domain); err != nil {
		return fmt.Errorf("ssl: invalid domain: %w", err)
	}
	if strings.TrimSpace(certPath) == "" {
		return fmt.Errorf("ssl: cert path is required")
	}
	if strings.TrimSpace(keyPath) == "" {
		return fmt.Errorf("ssl: key path is required")
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("ssl: generate key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("ssl: generate serial: %w", err)
	}

	now := time.Now().UTC()
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: domain},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(SelfSignedValidity),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{domain},
		IPAddresses:  nil,
	}

	derCert, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("ssl: create certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derCert})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return fmt.Errorf("ssl: marshal key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	if err := os.MkdirAll(filepath.Dir(certPath), 0o755); err != nil {
		return fmt.Errorf("ssl: mkdir cert dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(keyPath), 0o755); err != nil {
		return fmt.Errorf("ssl: mkdir key dir: %w", err)
	}
	if err := os.WriteFile(certPath, certPEM, 0o644); err != nil {
		return fmt.Errorf("ssl: write cert: %w", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		return fmt.Errorf("ssl: write key: %w", err)
	}
	return nil
}

func validateDomain(host string) error {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return fmt.Errorf("empty domain")
	}
	if net.ParseIP(host) != nil {
		return fmt.Errorf("domain must not be an IP address")
	}
	if !domainRe.MatchString(host) {
		return fmt.Errorf("domain %q is not valid", host)
	}
	return nil
}
