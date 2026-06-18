package database

import "time"

// Instance represents a provisioned database instance record.
type Instance struct {
	ID               string    `json:"id"`
	OwnerUserID      string    `json:"owner_user_id"`
	Name             string    `json:"name"`
	Engine           string    `json:"engine"`
	Version          string    `json:"version"`
	Host             string    `json:"host"`
	Port             int       `json:"port"`
	Username         string    `json:"username"`
	Password         string    `json:"password"`
	DatabaseName     string    `json:"database_name"`
	Status           string    `json:"status"`
	ConnectionString string    `json:"connection_string"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CreateRequest is the payload for provisioning a database instance.
type CreateRequest struct {
	Name    string `json:"name"`
	Engine  string `json:"engine"`
	Version string `json:"version"`
}
