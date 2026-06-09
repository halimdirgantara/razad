package database

import (
	"testing"

	"github.com/razad/razad/internal/config"
)

func TestOpen_InvalidDriver(t *testing.T) {
	cfg := config.DatabaseConfig{
		Driver: "nonexistent",
		DSN:    ":memory:",
	}
	_, err := Open(cfg)
	if err == nil {
		t.Error("expected error for nonexistent driver")
	}
}

func TestMigrate_NoDatabase(t *testing.T) {
	err := Migrate(nil)
	if err == nil {
		t.Error("expected error when db is nil")
	}
}
