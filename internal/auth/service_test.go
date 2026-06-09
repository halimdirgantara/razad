package auth

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/razad/razad/internal/database"
)

// setupTestDB creates an in-memory SQLite database with the auth schema.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tmpFile := "/tmp/razad-auth-test-" + t.Name() + ".db"
	os.Remove(tmpFile)

	db, err := sql.Open("sqlite3", tmpFile)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpFile)
	})

	if err := database.Migrate(db); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	db := setupTestDB(t)
	repo := NewRepository(db)
	return NewService(repo, 60) // 60 minute TTL
}

func TestRegister_Success(t *testing.T) {
	svc := newTestService(t)

	user, err := svc.Register("Test User", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if user.Name != "Test User" {
		t.Errorf("expected name 'Test User', got %s", user.Name)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %s", user.Email)
	}
	if user.ID == "" {
		t.Error("expected non-empty user ID")
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.Register("Test", "test@example.com", "short")
	if err != ErrWeakPassword {
		t.Errorf("expected ErrWeakPassword, got %v", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.Register("User 1", "dup@example.com", "password123")
	if err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	_, err = svc.Register("User 2", "dup@example.com", "password456")
	if err != ErrEmailTaken {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	svc := newTestService(t)

	svc.Register("Test User", "login@example.com", "password123")

	resp, err := svc.Login("login@example.com", "password123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.User.Email != "login@example.com" {
		t.Errorf("expected email 'login@example.com', got %s", resp.User.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := newTestService(t)

	svc.Register("Test User", "wrongpw@example.com", "password123")

	_, err := svc.Login("wrongpw@example.com", "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_WrongEmail(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.Login("nonexistent@example.com", "password123")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestValidateSession_Success(t *testing.T) {
	svc := newTestService(t)

	svc.Register("Session User", "session@example.com", "password123")
	resp, _ := svc.Login("session@example.com", "password123")

	user, err := svc.ValidateSession(resp.Token)
	if err != nil {
		t.Fatalf("ValidateSession failed: %v", err)
	}

	if user.Email != "session@example.com" {
		t.Errorf("expected email 'session@example.com', got %s", user.Email)
	}
}

func TestValidateSession_InvalidToken(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.ValidateSession("nonexistent-token")
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestLogout(t *testing.T) {
	svc := newTestService(t)

	svc.Register("Logout User", "logout@example.com", "password123")
	resp, _ := svc.Login("logout@example.com", "password123")

	// Logout should succeed
	if err := svc.Logout(resp.Token); err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Session should be invalid after logout
	_, err := svc.ValidateSession(resp.Token)
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound after logout, got %v", err)
	}
}
