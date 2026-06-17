package org

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/razad/razad/internal/audit"
)

var (
	ErrNotFound      = errors.New("org: not found")
	ErrSlugTaken     = errors.New("org: slug already taken")
	ErrAlreadyMember = errors.New("org: user is already a member")
	ErrInvalidSlug   = errors.New("org: slug must be 3-50 characters, lowercase alphanumeric with hyphens")
)

var slugRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{1,48})[a-z0-9]$`)

// Service handles organization business logic.
type Service struct {
	repo   *Repository
	audit  *audit.Service
}

// NewService creates an org service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// SetAuditor attaches an audit recorder to the service.
func (s *Service) SetAuditor(auditor *audit.Service) {
	s.audit = auditor
}

func (s *Service) record(actorID, action, entityType, entityID string, metadata map[string]any) {
	if s.audit == nil {
		return
	}
	if err := s.audit.Record(context.Background(), actorID, action, entityType, entityID, metadata); err != nil {
		// Intentionally ignore audit failures to keep primary workflows available.
	}
}

// Create creates a new organization and adds the creator as admin.
func (s *Service) Create(name, slug, creatorUserID string) (*Organization, error) {
	slug = strings.TrimSpace(strings.ToLower(slug))

	if !slugRegex.MatchString(slug) {
		return nil, ErrInvalidSlug
	}

	// Check slug uniqueness
	existing, _ := s.repo.FindBySlug(slug)
	if existing != nil {
		return nil, ErrSlugTaken
	}

	org, err := s.repo.Create(name, slug)
	if err != nil {
		return nil, fmt.Errorf("org: create: %w", err)
	}

	// Add creator as admin
	if _, err := s.repo.AddMember(org.ID, creatorUserID, "admin"); err != nil {
		return nil, fmt.Errorf("org: add creator: %w", err)
	}

	s.record(creatorUserID, "org.create", "organization", org.ID, map[string]any{"slug": org.Slug, "name": org.Name})
	return org, nil
}

// Get retrieves an organization by ID, checking membership.
func (s *Service) Get(id, userID string) (*Organization, error) {
	org, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}

	member, err := s.repo.IsMember(org.ID, userID)
	if err != nil || !member {
		return nil, ErrNotFound
	}

	return org, nil
}

// List returns organizations the user belongs to.
func (s *Service) List(userID string) ([]Organization, error) {
	return s.repo.List(userID)
}

// AddMember adds a user to an organization.
func (s *Service) AddMember(orgID, actorID, targetUserID, role string) error {
	// Verify actor is admin
	actorIsMember, err := s.repo.IsMember(orgID, actorID)
	if err != nil || !actorIsMember {
		return ErrNotFound
	}

	// Check if target already exists
	targetIsMember, _ := s.repo.IsMember(orgID, targetUserID)
	if targetIsMember {
		return ErrAlreadyMember
	}

	_, err = s.repo.AddMember(orgID, targetUserID, role)
	if err == nil {
		s.record(actorID, "org.member.add", "organization", orgID, map[string]any{"target_user_id": targetUserID, "role": role})
	}
	return err
}
