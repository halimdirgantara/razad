// Package auth provides authentication, session management, and API token
// validation for the Razad daemon.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Common auth errors.
var (
	ErrInvalidCredentials = errors.New("auth: invalid email or password")
	ErrSessionExpired     = errors.New("auth: session expired")
	ErrSessionNotFound    = errors.New("auth: session not found")
	ErrEmailTaken         = errors.New("auth: email already registered")
	ErrWeakPassword       = errors.New("auth: password too weak (minimum 8 characters)")
)

// Service handles authentication business logic.
type Service struct {
	repo      *Repository
	sessionTTL time.Duration
}

// NewService creates an auth service with the given repository and config.
func NewService(repo *Repository, sessionTTLMinutes int) *Service {
	ttl := time.Duration(sessionTTLMinutes) * time.Minute
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &Service{repo: repo, sessionTTL: ttl}
}

// Register creates a new user account.
func (s *Service) Register(name, email, password string) (*UserInfo, error) {
	if len(password) < 8 {
		return nil, ErrWeakPassword
	}

	// Check for existing user
	existing, _ := s.repo.FindByEmail(email)
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to hash password: %w", err)
	}

	user, err := s.repo.CreateUser(name, email, string(hash))
	if err != nil {
		return nil, fmt.Errorf("auth: failed to create user: %w", err)
	}

	return &UserInfo{ID: user.ID, Name: user.Name, Email: user.Email}, nil
}

// Login validates credentials and creates a session.
func (s *Service) Login(email, password string) (*LoginResponse, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := generateToken(32)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to generate token: %w", err)
	}

	session, err := s.repo.CreateSession(user.ID, token, time.Now().Add(s.sessionTTL))
	if err != nil {
		return nil, fmt.Errorf("auth: failed to create session: %w", err)
	}

	return &LoginResponse{
		Token: session.Token,
		User:  UserInfo{ID: user.ID, Name: user.Name, Email: user.Email},
	}, nil
}

// ValidateSession checks if a session token is valid and not expired.
func (s *Service) ValidateSession(token string) (*UserInfo, error) {
	session, err := s.repo.FindSession(token)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	if time.Now().After(session.ExpiresAt) {
		s.repo.DeleteSession(token)
		return nil, ErrSessionExpired
	}

	user, err := s.repo.FindByID(session.UserID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	return &UserInfo{ID: user.ID, Name: user.Name, Email: user.Email}, nil
}

// Logout invalidates a session.
func (s *Service) Logout(token string) error {
	return s.repo.DeleteSession(token)
}

// LookupByEmail returns the public user record for an email address. It is
// used by the daemon at startup to resolve the seeded admin's user ID for
// the policy engine. Returns ErrSessionNotFound-style semantics via
// fmt.Errorf when the user is missing.
func (s *Service) LookupByEmail(email string) (*UserInfo, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	return &UserInfo{ID: user.ID, Name: user.Name, Email: user.Email}, nil
}

// generateToken generates a cryptographically random hex token.
func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
