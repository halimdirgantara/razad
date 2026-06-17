package proxy

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Binding describes the minimum information needed to render an Nginx site.
type Binding struct {
	Name         string
	Domain       string
	UpstreamHost string
	UpstreamPort int
	TLS          bool
	BodyLimitMB  int
}

// Service renders and manages Nginx config scaffolding.
type Service struct {
	baseDir string
}

// NewService creates a proxy service rooted at baseDir.
func NewService(baseDir string) *Service {
	return &Service{baseDir: baseDir}
}

// Validate checks the binding for obvious mistakes before rendering.
func (s *Service) Validate(binding Binding) error {
	if strings.TrimSpace(binding.Name) == "" {
		return fmt.Errorf("proxy: name is required")
	}
	if err := validateHostname(binding.Domain); err != nil {
		return fmt.Errorf("proxy: invalid domain: %w", err)
	}
	if err := validateUpstreamHost(binding.UpstreamHost); err != nil {
		return fmt.Errorf("proxy: invalid upstream host: %w", err)
	}
	if binding.UpstreamPort <= 0 || binding.UpstreamPort > 65535 {
		return fmt.Errorf("proxy: invalid upstream port")
	}
	if binding.BodyLimitMB < 0 {
		return fmt.Errorf("proxy: body limit must be >= 0")
	}
	return nil
}

// Render returns a complete Nginx server block.
func (s *Service) Render(binding Binding) (string, error) {
	if err := s.Validate(binding); err != nil {
		return "", err
	}

	upstream := net.JoinHostPort(binding.UpstreamHost, fmt.Sprintf("%d", binding.UpstreamPort))
	clientMaxBody := ""
	if binding.BodyLimitMB > 0 {
		clientMaxBody = fmt.Sprintf("    client_max_body_size %dm;\n", binding.BodyLimitMB)
	}

	var b strings.Builder
	b.WriteString("server {\n")
	b.WriteString("    listen 80;\n")
	if binding.TLS {
		b.WriteString("    listen 443 ssl http2;\n")
		b.WriteString(fmt.Sprintf("    server_name %s;\n", binding.Domain))
		b.WriteString(fmt.Sprintf("    ssl_certificate %s;\n", filepath.Join("/etc/letsencrypt/live", binding.Domain, "fullchain.pem")))
		b.WriteString(fmt.Sprintf("    ssl_certificate_key %s;\n", filepath.Join("/etc/letsencrypt/live", binding.Domain, "privkey.pem")))
	} else {
		b.WriteString(fmt.Sprintf("    server_name %s;\n", binding.Domain))
	}
	if clientMaxBody != "" {
		b.WriteString(clientMaxBody)
	}
	b.WriteString("\n")
	if !binding.TLS {
		b.WriteString("    location / {\n")
		b.WriteString(fmt.Sprintf("        proxy_pass http://%s;\n", upstream))
		b.WriteString("        proxy_set_header Host $host;\n")
		b.WriteString("        proxy_set_header X-Real-IP $remote_addr;\n")
		b.WriteString("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
		b.WriteString("        proxy_set_header X-Forwarded-Proto $scheme;\n")
		b.WriteString("    }\n")
	} else {
		b.WriteString("    if ($scheme = http) {\n")
		b.WriteString("        return 301 https://$host$request_uri;\n")
		b.WriteString("    }\n\n")
		b.WriteString("    location / {\n")
		b.WriteString(fmt.Sprintf("        proxy_pass http://%s;\n", upstream))
		b.WriteString("        proxy_set_header Host $host;\n")
		b.WriteString("        proxy_set_header X-Real-IP $remote_addr;\n")
		b.WriteString("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
		b.WriteString("        proxy_set_header X-Forwarded-Proto $scheme;\n")
		b.WriteString("    }\n")
	}
	b.WriteString("}\n")
	return b.String(), nil
}

// CandidatePath returns the path for a candidate config file.
func (s *Service) CandidatePath(binding Binding) (string, error) {
	if err := s.Validate(binding); err != nil {
		return "", err
	}
	return filepath.Join(s.baseDir, "sites-available", binding.Name+".conf"), nil
}

// EnabledPath returns the path for the active site config.
func (s *Service) EnabledPath(binding Binding) (string, error) {
	if err := s.Validate(binding); err != nil {
		return "", err
	}
	return filepath.Join(s.baseDir, "sites-enabled", binding.Name+".conf"), nil
}

// WriteCandidate writes the rendered config into sites-available.
func (s *Service) WriteCandidate(binding Binding) (string, error) {
	cfg, err := s.Render(binding)
	if err != nil {
		return "", err
	}
	path, err := s.CandidatePath(binding)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(cfg), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// Apply promotes the candidate config to the enabled site path and keeps a backup.
func (s *Service) Apply(binding Binding) error {
	candidate, err := s.WriteCandidate(binding)
	if err != nil {
		return err
	}
	enabled, err := s.EnabledPath(binding)
	if err != nil {
		return err
	}
	backup, err := s.BackupPath(binding)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(enabled), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(enabled); err == nil {
		if err := copyFile(enabled, backup); err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		if err := copyFile(candidate, backup); err != nil {
			return err
		}
	}
	return copyFile(candidate, enabled)
}

// Rollback restores the most recent enabled config backup.
func (s *Service) Rollback(binding Binding) error {
	backup, err := s.BackupPath(binding)
	if err != nil {
		return err
	}
	enabled, err := s.EnabledPath(binding)
	if err != nil {
		return err
	}
	if _, err := os.Stat(backup); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(enabled), 0o755); err != nil {
		return err
	}
	return copyFile(backup, enabled)
}

// BackupPath returns the path of the last-known-good config snapshot.
func (s *Service) BackupPath(binding Binding) (string, error) {
	if err := s.Validate(binding); err != nil {
		return "", err
	}
	return filepath.Join(s.baseDir, "backups", binding.Name+".conf.bak"), nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

var hostnameRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)+$`)

func validateHostname(host string) error {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return fmt.Errorf("empty hostname")
	}
	if strings.Contains(host, "_") {
		return fmt.Errorf("hostname contains underscore")
	}
	if !hostnameRe.MatchString(host) {
		return fmt.Errorf("hostname %q is not valid", host)
	}
	return nil
}

func validateUpstreamHost(host string) error {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return fmt.Errorf("empty upstream host")
	}
	if strings.ContainsAny(host, " /	\n") {
		return fmt.Errorf("upstream host contains whitespace")
	}
	if net.ParseIP(host) != nil {
		return nil
	}
	if !hostnameRe.MatchString(host) {
		return fmt.Errorf("upstream host %q is not valid", host)
	}
	return nil
}
