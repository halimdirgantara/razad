package app

import (
	"database/sql"
	"fmt"

	"github.com/razad/razad/internal/domain"
)

// Repository handles app-related database operations.
type Repository struct {
	db *sql.DB
}

// NewRepository creates an app repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ---------------------------------------------------------------------------
// Access helpers
// ---------------------------------------------------------------------------

func (r *Repository) canAccessProject(userID, projectID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		 FROM projects p
		 JOIN organization_members om ON om.organization_id = p.organization_id
		 WHERE p.id = ? AND om.user_id = ?`,
		projectID, userID,
	).Scan(&count)
	return count > 0, err
}

func (r *Repository) canAccessApp(userID, appID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		 FROM apps a
		 JOIN projects p ON p.id = a.project_id
		 JOIN organization_members om ON om.organization_id = p.organization_id
		 WHERE a.id = ? AND om.user_id = ?`,
		appID, userID,
	).Scan(&count)
	return count > 0, err
}

func scanApp(row scanner) (*domain.App, error) {
	app := &domain.App{}
	if err := row.Scan(&app.ID, &app.ProjectID, &app.Name, &app.GitURL, &app.Runtime, &app.StartCmd, &app.Status, &app.CreatedAt, &app.UpdatedAt); err != nil {
		return nil, err
	}
	return app, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanApps(rows *sql.Rows) ([]domain.App, error) {
	var apps []domain.App
	for rows.Next() {
		var a domain.App
		if err := rows.Scan(&a.ID, &a.ProjectID, &a.Name, &a.GitURL, &a.Runtime, &a.StartCmd, &a.Status, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("app: list scan: %w", err)
		}
		apps = append(apps, a)
	}
	return apps, nil
}

// ---------------------------------------------------------------------------
// Apps
// ---------------------------------------------------------------------------

// CreateForUser inserts a new app if the user can access the project.
func (r *Repository) CreateForUser(userID, projectID, name, gitURL, runtime, startCmd string) (*domain.App, error) {
	app := &domain.App{}
	err := r.db.QueryRow(
		`INSERT INTO apps (id, project_id, name, git_url, runtime, start_cmd, status, created_at, updated_at)
		 SELECT lower(hex(randomblob(16))), p.id, ?, ?, ?, ?, 'created', datetime('now'), datetime('now')
		 FROM projects p
		 JOIN organization_members om ON om.organization_id = p.organization_id
		 WHERE p.id = ? AND om.user_id = ?
		 RETURNING id, project_id, name, git_url, runtime, start_cmd, status, created_at, updated_at`,
		name, gitURL, runtime, startCmd, projectID, userID,
	).Scan(&app.ID, &app.ProjectID, &app.Name, &app.GitURL, &app.Runtime, &app.StartCmd, &app.Status, &app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("app: create: %w", err)
	}
	return app, nil
}

// FindByIDForUser retrieves an app by ID for an authorized user.
func (r *Repository) FindByIDForUser(userID, id string) (*domain.App, error) {
	app := &domain.App{}
	err := r.db.QueryRow(
		`SELECT a.id, a.project_id, a.name, a.git_url, a.runtime, a.start_cmd, a.status, a.created_at, a.updated_at
		 FROM apps a
		 JOIN projects p ON p.id = a.project_id
		 JOIN organization_members om ON om.organization_id = p.organization_id
		 WHERE a.id = ? AND om.user_id = ?`,
		id, userID,
	).Scan(&app.ID, &app.ProjectID, &app.Name, &app.GitURL, &app.Runtime, &app.StartCmd, &app.Status, &app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("app: find by id: %w", err)
	}
	return app, nil
}

// ListAllForUser returns all apps the user is allowed to see.
func (r *Repository) ListAllForUser(userID string) ([]domain.App, error) {
	rows, err := r.db.Query(
		`SELECT a.id, a.project_id, a.name, a.git_url, a.runtime, a.start_cmd, a.status, a.created_at, a.updated_at
		 FROM apps a
		 JOIN projects p ON p.id = a.project_id
		 JOIN organization_members om ON om.organization_id = p.organization_id
		 WHERE om.user_id = ?
		 ORDER BY a.name`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("app: list all: %w", err)
	}
	defer rows.Close()
	return scanApps(rows)
}

// ListByProjectForUser returns all apps in a project if the user can access it.
func (r *Repository) ListByProjectForUser(userID, projectID string) ([]domain.App, error) {
	rows, err := r.db.Query(
		`SELECT a.id, a.project_id, a.name, a.git_url, a.runtime, a.start_cmd, a.status, a.created_at, a.updated_at
		 FROM apps a
		 JOIN projects p ON p.id = a.project_id
		 JOIN organization_members om ON om.organization_id = p.organization_id
		 WHERE p.id = ? AND om.user_id = ?
		 ORDER BY a.name`,
		projectID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("app: list by project: %w", err)
	}
	defer rows.Close()
	return scanApps(rows)
}

// UpdateStatus updates the status of an app.
func (r *Repository) UpdateStatus(id, status string) error {
	_, err := r.db.Exec(
		`UPDATE apps SET status = ?, updated_at = datetime('now') WHERE id = ?`,
		status, id,
	)
	return err
}

// Delete removes an app by ID.
func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM apps WHERE id = ?`, id)
	return err
}

// ---------------------------------------------------------------------------
// Deployments
// ---------------------------------------------------------------------------

// CreateDeployment inserts a deployment record.
func (r *Repository) CreateDeployment(appID, version string) (*domain.AppDeployment, error) {
	d := &domain.AppDeployment{}

	err := r.db.QueryRow(
		`INSERT INTO app_deployments (id, app_id, version, status, created_at, updated_at)
		 VALUES (lower(hex(randomblob(16))), ?, ?, 'pending', datetime('now'), datetime('now'))
		 RETURNING id, app_id, version, status, created_at, updated_at`,
		appID, version,
	).Scan(&d.ID, &d.AppID, &d.Version, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("app: create deployment: %w", err)
	}
	return d, nil
}

// UpdateDeploymentStatus updates the status and log of a deployment.
func (r *Repository) UpdateDeploymentStatus(id, status, log string) error {
	_, err := r.db.Exec(
		`UPDATE app_deployments SET status = ?, log = ?, updated_at = datetime('now') WHERE id = ?`,
		status, log, id,
	)
	return err
}

// ListDeployments returns deployments for an app, most recent first.
func (r *Repository) ListDeployments(appID string) ([]domain.AppDeployment, error) {
	rows, err := r.db.Query(
		`SELECT id, app_id, version, status, log, created_at, updated_at
		 FROM app_deployments WHERE app_id = ? ORDER BY created_at DESC LIMIT 20`, appID,
	)
	if err != nil {
		return nil, fmt.Errorf("app: list deployments: %w", err)
	}
	defer rows.Close()

	var deploys []domain.AppDeployment
	for rows.Next() {
		var d domain.AppDeployment
		if err := rows.Scan(&d.ID, &d.AppID, &d.Version, &d.Status, &d.Log, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("app: list deployments scan: %w", err)
		}
		deploys = append(deploys, d)
	}
	return deploys, nil
}

// ---------------------------------------------------------------------------
// Env Vars
// ---------------------------------------------------------------------------

// UpsertEnvVar inserts or updates an environment variable.
func (r *Repository) UpsertEnvVar(appID, key, encryptedValue string) error {
	_, err := r.db.Exec(
		`INSERT INTO app_env_vars (id, app_id, key, value, created_at, updated_at)
		 VALUES (lower(hex(randomblob(16))), ?, ?, ?, datetime('now'), datetime('now'))
		 ON CONFLICT(app_id, key) DO UPDATE SET value = excluded.value, updated_at = datetime('now')`,
		appID, key, encryptedValue,
	)
	return err
}

// ListEnvVars returns all env var keys (values are encrypted) for an app.
func (r *Repository) ListEnvVars(appID string) ([]domain.AppEnvVar, error) {
	rows, err := r.db.Query(
		`SELECT id, app_id, key, value, created_at, updated_at
		 FROM app_env_vars WHERE app_id = ? ORDER BY key`, appID,
	)
	if err != nil {
		return nil, fmt.Errorf("app: list env vars: %w", err)
	}
	defer rows.Close()

	var vars []domain.AppEnvVar
	for rows.Next() {
		var v domain.AppEnvVar
		if err := rows.Scan(&v.ID, &v.AppID, &v.Key, &v.Value, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, fmt.Errorf("app: list env vars scan: %w", err)
		}
		v.Value = "" // never expose encrypted value
		vars = append(vars, v)
	}
	return vars, nil
}
