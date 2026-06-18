package database

import (
	"database/sql"
	"fmt"
)

// Repository handles database-management records.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a database-management repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) createForUser(ownerUserID string, inst *Instance) (*Instance, error) {
	created := &Instance{}
	err := r.db.QueryRow(
		`INSERT INTO database_instances (
			id, owner_user_id, name, engine, version, host, port, username, password, database_name, status, connection_string, created_at, updated_at
		) VALUES (
			lower(hex(randomblob(16))), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now')
		)
		RETURNING id, owner_user_id, name, engine, version, host, port, username, password, database_name, status, connection_string, created_at, updated_at`,
		ownerUserID, inst.Name, inst.Engine, inst.Version, inst.Host, inst.Port, inst.Username, inst.Password, inst.DatabaseName, inst.Status, inst.ConnectionString,
	).Scan(
		&created.ID, &created.OwnerUserID, &created.Name, &created.Engine, &created.Version, &created.Host, &created.Port,
		&created.Username, &created.Password, &created.DatabaseName, &created.Status, &created.ConnectionString,
		&created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("database: create instance: %w", err)
	}
	return created, nil
}

func (r *Repository) listForUser(ownerUserID string) ([]Instance, error) {
	rows, err := r.db.Query(
		`SELECT id, owner_user_id, name, engine, version, host, port, username, password, database_name, status, connection_string, created_at, updated_at
		 FROM database_instances
		 WHERE owner_user_id = ?
		 ORDER BY created_at DESC`,
		ownerUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("database: list instances: %w", err)
	}
	defer rows.Close()

	var instances []Instance
	for rows.Next() {
		var inst Instance
		if err := rows.Scan(
			&inst.ID, &inst.OwnerUserID, &inst.Name, &inst.Engine, &inst.Version, &inst.Host, &inst.Port,
			&inst.Username, &inst.Password, &inst.DatabaseName, &inst.Status, &inst.ConnectionString,
			&inst.CreatedAt, &inst.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("database: list scan: %w", err)
		}
		instances = append(instances, inst)
	}
	return instances, nil
}

func (r *Repository) findByIDForUser(ownerUserID, id string) (*Instance, error) {
	inst := &Instance{}
	err := r.db.QueryRow(
		`SELECT id, owner_user_id, name, engine, version, host, port, username, password, database_name, status, connection_string, created_at, updated_at
		 FROM database_instances
		 WHERE id = ? AND owner_user_id = ?`,
		id, ownerUserID,
	).Scan(
		&inst.ID, &inst.OwnerUserID, &inst.Name, &inst.Engine, &inst.Version, &inst.Host, &inst.Port,
		&inst.Username, &inst.Password, &inst.DatabaseName, &inst.Status, &inst.ConnectionString,
		&inst.CreatedAt, &inst.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("database: find instance: %w", err)
	}
	return inst, nil
}

func (r *Repository) deleteForUser(ownerUserID, id string) error {
	res, err := r.db.Exec(`DELETE FROM database_instances WHERE id = ? AND owner_user_id = ?`, id, ownerUserID)
	if err != nil {
		return fmt.Errorf("database: delete instance: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("database: delete instance: not found")
	}
	return nil
}
