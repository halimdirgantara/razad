package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Event is an immutable audit record.
type Event struct {
	ID           string    `json:"id"`
	ActorUserID  string    `json:"actor_user_id"`
	Action       string    `json:"action"`
	EntityType   string    `json:"entity_type"`
	EntityID     string    `json:"entity_id"`
	MetadataJSON string    `json:"metadata_json"`
	CreatedAt    time.Time `json:"created_at"`
}

// Service persists audit events.
type Service struct {
	db *sql.DB
}

// NewService creates an audit service.
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// Record stores a new immutable audit event.
func (s *Service) Record(ctx context.Context, actorUserID, action, entityType, entityID string, metadata map[string]any) error {
	payload := map[string]any{}
	for k, v := range metadata {
		payload[k] = v
	}
	if payload == nil {
		payload = map[string]any{}
	}
	meta, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("audit: encode metadata: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO audit_events (id, actor_user_id, action, entity_type, entity_id, metadata_json, created_at)
		 VALUES (lower(hex(randomblob(16))), ?, ?, ?, ?, ?, datetime('now'))`,
		actorUserID, action, entityType, entityID, string(meta),
	)
	if err != nil {
		return fmt.Errorf("audit: insert event: %w", err)
	}
	return nil
}

// ListRecent returns the most recent audit events.
func (s *Service) ListRecent(limit int) ([]Event, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(
		`SELECT id, actor_user_id, action, entity_type, entity_id, metadata_json, created_at
		 FROM audit_events ORDER BY created_at DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("audit: list recent: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.ActorUserID, &e.Action, &e.EntityType, &e.EntityID, &e.MetadataJSON, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("audit: scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}
