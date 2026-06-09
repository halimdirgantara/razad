package config

import (
	"os"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
	if cfg.Mode != "self-hosted" {
		t.Errorf("expected mode self-hosted, got %s", cfg.Mode)
	}
	if cfg.Database.Driver != "sqlite3" {
		t.Errorf("expected sqlite3 driver, got %s", cfg.Database.Driver)
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := Defaults()
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	cfg := Defaults()
	cfg.Port = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid port")
	}

	cfg.Port = 70000
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for out-of-range port")
	}
}

func TestValidate_InvalidMode(t *testing.T) {
	cfg := Defaults()
	cfg.Mode = "invalid"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid mode")
	}
}

func TestValidate_InvalidDriver(t *testing.T) {
	cfg := Defaults()
	cfg.Database.Driver = "mysql"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid database driver")
	}
}

func TestValidate_EmptyHost(t *testing.T) {
	cfg := Defaults()
	cfg.Host = ""
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty host")
	}
}

func TestValidate_EmptyDSN(t *testing.T) {
	cfg := Defaults()
	cfg.Database.DSN = ""
	cfg.Database.Driver = "postgres" // SQLite derives DSN from DataDir, so use postgres
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty DSN with postgres driver")
	}
}

func TestValidate_DerivedSQLiteDSN(t *testing.T) {
	cfg := Defaults()
	cfg.Database.DSN = ""
	cfg.Database.Driver = "sqlite3"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error for sqlite3 with empty DSN, got: %v", err)
	}
	if cfg.Database.DSN != "/var/lib/razad/razad.db" {
		t.Errorf("expected DSN to be derived from DataDir, got %s", cfg.Database.DSN)
	}
}

func TestEnvOverrides(t *testing.T) {
	os.Setenv("RAZAD_DEBUG", "true")
	os.Setenv("RAZAD_PORT", "9090")
	os.Setenv("RAZAD_MODE", "byo-vps")
	os.Setenv("RAZAD_DB_DRIVER", "postgres")
	os.Setenv("RAZAD_DB_DSN", "postgres://localhost:5432/razad")
	defer os.Unsetenv("RAZAD_DEBUG")
	defer os.Unsetenv("RAZAD_PORT")
	defer os.Unsetenv("RAZAD_MODE")
	defer os.Unsetenv("RAZAD_DB_DRIVER")
	defer os.Unsetenv("RAZAD_DB_DSN")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !cfg.Debug {
		t.Error("expected debug=true from env")
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.Mode != "byo-vps" {
		t.Errorf("expected mode byo-vps, got %s", cfg.Mode)
	}
	if cfg.Database.Driver != "postgres" {
		t.Errorf("expected postgres driver, got %s", cfg.Database.Driver)
	}
	if cfg.Database.DSN != "postgres://localhost:5432/razad" {
		t.Errorf("expected custom DSN, got %s", cfg.Database.DSN)
	}
}
