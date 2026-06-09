package auth

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/razad/razad/internal/domain"
)

// Repository handles auth-related database operations.
type Repository struct {
	db *sql.DB
}

// NewRepository creates an auth repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser inserts a new user.
func (r *Repository) CreateUser(name, email, passwordHash string) (*domain.User, error) {
	user := &domain.User{}

	err := r.db.QueryRow(
		`INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
		 VALUES (lower(hex(randomblob(16))), ?, ?, ?, datetime('now'), datetime('now'))
		 RETURNING id, name, email, created_at, updated_at`,
		name, email, passwordHash,
	).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("auth: create user: %w", err)
	}

	return user, nil
}

// FindByEmail looks up a user by email.
func (r *Repository) FindByEmail(email string) (*domain.User, error) {
	user := &domain.User{}

	err := r.db.QueryRow(
		`SELECT id, name, email, password_hash, created_at, updated_at
		 FROM users WHERE email = ?`, email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("auth: find by email: %w", err)
	}

	return user, nil
}

// FindByID looks up a user by ID.
func (r *Repository) FindByID(id string) (*domain.User, error) {
	user := &domain.User{}

	err := r.db.QueryRow(
		`SELECT id, name, email, password_hash, created_at, updated_at
		 FROM users WHERE id = ?`, id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("auth: find by id: %w", err)
	}

	return user, nil
}

// CreateSession inserts a new session.
func (r *Repository) CreateSession(userID, token string, expiresAt time.Time) (*Session, error) {
	session := &Session{}

	err := r.db.QueryRow(
		`INSERT INTO sessions (id, user_id, token, expires_at, created_at)
		 VALUES (lower(hex(randomblob(16))), ?, ?, ?, datetime('now'))
		 RETURNING id, user_id, token, expires_at, created_at`,
		userID, token, expiresAt,
	).Scan(&session.ID, &session.UserID, &session.Token, &session.ExpiresAt, &session.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("auth: create session: %w", err)
	}

	return session, nil
}

// FindSession looks up a session by token.
func (r *Repository) FindSession(token string) (*Session, error) {
	session := &Session{}

	err := r.db.QueryRow(
		`SELECT id, user_id, token, expires_at, created_at
		 FROM sessions WHERE token = ?`, token,
	).Scan(&session.ID, &session.UserID, &session.Token, &session.ExpiresAt, &session.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("auth: find session: %w", err)
	}

	return session, nil
}

// DeleteSession removes a session.
func (r *Repository) DeleteSession(token string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	return err
}
