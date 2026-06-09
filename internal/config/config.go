// Package config provides typed application configuration loading with env var
// overrides. Config is loaded once at startup and validated before use.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config is the top-level runtime configuration for the razad daemon.
type Config struct {
	// Debug enables debug-level logging.
	Debug bool `json:"debug" yaml:"debug"`

	// Version is the build version injected at compile time.
	Version string `json:"version" yaml:"version"`

	// Mode is the operating mode: "self-hosted", "byo-vps", or "managed".
	Mode string `json:"mode" yaml:"mode"`

	// Port is the HTTP listen port.
	Port int `json:"port" yaml:"port"`

	// Host is the HTTP listen address.
	Host string `json:"host" yaml:"host"`

	// DataDir is the base data directory (typically /var/lib/razad).
	DataDir string `json:"data_dir" yaml:"data_dir"`

	// Database holds database connection configuration.
	Database DatabaseConfig `json:"database" yaml:"database"`

	// Auth holds authentication configuration.
	Auth AuthConfig `json:"auth" yaml:"auth"`

	// AutoMigrate runs database migrations on startup if true.
	AutoMigrate bool `json:"auto_migrate" yaml:"auto_migrate"`
}

// DatabaseConfig holds database connection parameters.
type DatabaseConfig struct {
	// Driver is "postgres" or "sqlite".
	Driver string `json:"driver" yaml:"driver"`

	// DSN is the connection string.
	DSN string `json:"dsn" yaml:"dsn"`

	// MaxOpenConns is the max number of open connections.
	MaxOpenConns int `json:"max_open_conns" yaml:"max_open_conns"`

	// MaxIdleConns is the max number of idle connections.
	MaxIdleConns int `json:"max_idle_conns" yaml:"max_idle_conns"`

	// ConnMaxLifetimeSeconds is the maximum lifetime of a connection.
	ConnMaxLifetimeSeconds int `json:"conn_max_lifetime_seconds" yaml:"conn_max_lifetime_seconds"`
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	// SessionTTLMinutes is the session duration in minutes.
	SessionTTLMinutes int `json:"session_ttl_minutes" yaml:"session_ttl_minutes"`

	// TokenLength is the length of generated API tokens.
	TokenLength int `json:"token_length" yaml:"token_length"`

	// SecretKey is the key used for session signing and encryption.
	SecretKey string `json:"secret_key" yaml:"secret_key"`
}

// Defaults returns a Config with sensible defaults for self-hosted mode.
func Defaults() Config {
	return Config{
		Debug:  false,
		Mode:   "self-hosted",
		Port:   8080,
		Host:   "127.0.0.1",
		DataDir: "/var/lib/razad",
		Database: DatabaseConfig{
			Driver:                 "sqlite3",
			DSN:                    "",
			MaxOpenConns:          10,
			MaxIdleConns:          5,
			ConnMaxLifetimeSeconds: 300,
		},
		Auth: AuthConfig{
			SessionTTLMinutes: 1440, // 24 hours
			TokenLength:       32,
		},
		AutoMigrate: true,
	}
}

// Load reads configuration from environment variables, falling back to defaults.
// Config file support (YAML) will be added when the installer module is built.
func Load() (Config, error) {
	cfg := Defaults()
	cfg.applyEnvOverrides()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate checks that the configuration is valid and resolves defaults.
func (c *Config) Validate() error {
	var errs []string

	if c.Port < 1 || c.Port > 65535 {
		errs = append(errs, "port must be between 1 and 65535")
	}
	if c.Host == "" {
		errs = append(errs, "host must not be empty")
	}
	if c.DataDir == "" {
		errs = append(errs, "data_dir must not be empty")
	}
	switch c.Mode {
	case "self-hosted", "byo-vps", "managed":
		// valid
	default:
		errs = append(errs, "mode must be one of: self-hosted, byo-vps, managed")
	}
	if c.Database.Driver != "postgres" && c.Database.Driver != "sqlite3" {
		errs = append(errs, "database.driver must be postgres or sqlite3")
	}

	// Derive default DSN from data dir for SQLite
	if c.Database.DSN == "" && c.Database.Driver == "sqlite3" {
		c.Database.DSN = c.DataDir + "/razad.db"
	}
	if c.Database.DSN == "" {
		errs = append(errs, "database.dsn must not be empty")
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

// applyEnvOverrides reads RAZAD_* environment variables and overrides config fields.
func (c *Config) applyEnvOverrides() {
	if v := os.Getenv("RAZAD_DEBUG"); v != "" {
		c.Debug = v == "true" || v == "1"
	}
	if v := os.Getenv("RAZAD_MODE"); v != "" {
		c.Mode = v
	}
	if v := os.Getenv("RAZAD_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			c.Port = p
		}
	}
	if v := os.Getenv("RAZAD_HOST"); v != "" {
		c.Host = v
	}
	if v := os.Getenv("RAZAD_DATA_DIR"); v != "" {
		c.DataDir = v
	}
	if v := os.Getenv("RAZAD_DB_DRIVER"); v != "" {
		c.Database.Driver = v
	}
	if v := os.Getenv("RAZAD_DB_DSN"); v != "" {
		c.Database.DSN = v
	}
	if v := os.Getenv("RAZAD_DB_MAX_OPEN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Database.MaxOpenConns = n
		}
	}
	if v := os.Getenv("RAZAD_DB_MAX_IDLE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Database.MaxIdleConns = n
		}
	}
	if v := os.Getenv("RAZAD_SECRET_KEY"); v != "" {
		c.Auth.SecretKey = v
	}
	if v := os.Getenv("RAZAD_AUTO_MIGRATE"); v != "" {
		c.AutoMigrate = v == "true" || v == "1"
	}
}
