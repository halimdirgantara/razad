package ssl

import (
	"fmt"
	"net"
	"path/filepath"
	"regexp"
	"strings"
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
