package org

import (
	"database/sql"
	"fmt"
)

// Repository handles org-related database operations.
type Repository struct {
	db *sql.DB
}

// NewRepository creates an org repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new organization.
func (r *Repository) Create(name, slug string) (*Organization, error) {
	org := &Organization{}

	err := r.db.QueryRow(
		`INSERT INTO organizations (id, name, slug, created_at, updated_at)
		 VALUES (lower(hex(randomblob(16))), ?, ?, datetime('now'), datetime('now'))
		 RETURNING id, name, slug, created_at, updated_at`,
		name, slug,
	).Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt, &org.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("org: create: %w", err)
	}

	return org, nil
}

// FindByID retrieves an organization by ID.
func (r *Repository) FindByID(id string) (*Organization, error) {
	org := &Organization{}

	err := r.db.QueryRow(
		`SELECT id, name, slug, created_at, updated_at FROM organizations WHERE id = ?`, id,
	).Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt, &org.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("org: find by id: %w", err)
	}

	return org, nil
}

// FindBySlug retrieves an organization by slug.
func (r *Repository) FindBySlug(slug string) (*Organization, error) {
	org := &Organization{}

	err := r.db.QueryRow(
		`SELECT id, name, slug, created_at, updated_at FROM organizations WHERE slug = ?`, slug,
	).Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt, &org.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("org: find by slug: %w", err)
	}

	return org, nil
}

// List returns all organizations for the given user.
func (r *Repository) List(userID string) ([]Organization, error) {
	rows, err := r.db.Query(
		`SELECT o.id, o.name, o.slug, o.created_at, o.updated_at
		 FROM organizations o
		 JOIN organization_members om ON om.organization_id = o.id
		 WHERE om.user_id = ?
		 ORDER BY o.name`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("org: list: %w", err)
	}
	defer rows.Close()

	var orgs []Organization
	for rows.Next() {
		var org Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, fmt.Errorf("org: list scan: %w", err)
		}
		orgs = append(orgs, org)
	}

	return orgs, nil
}

// Delete removes an organization by ID.
func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM organizations WHERE id = ?`, id)
	return err
}

// AddMember adds a user to an organization with the given role.
func (r *Repository) AddMember(orgID, userID, role string) (*Member, error) {
	member := &Member{}

	err := r.db.QueryRow(
		`INSERT INTO organization_members (id, organization_id, user_id, role, created_at)
		 VALUES (lower(hex(randomblob(16))), ?, ?, ?, datetime('now'))
		 RETURNING id, organization_id, user_id, role, created_at`,
		orgID, userID, role,
	).Scan(&member.ID, &member.OrganizationID, &member.UserID, &member.Role, &member.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("org: add member: %w", err)
	}

	return member, nil
}

// RemoveMember removes a user from an organization.
func (r *Repository) RemoveMember(orgID, userID string) error {
	_, err := r.db.Exec(
		`DELETE FROM organization_members WHERE organization_id = ? AND user_id = ?`,
		orgID, userID,
	)
	return err
}

// IsMember checks if a user is a member of the organization.
func (r *Repository) IsMember(orgID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM organization_members WHERE organization_id = ? AND user_id = ?`,
		orgID, userID,
	).Scan(&count)

	return count > 0, err
}
