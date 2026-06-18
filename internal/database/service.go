package database

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrNotFound       = errors.New("database: not found")
	ErrInvalidName    = errors.New("database: name is required")
	ErrInvalidEngine  = errors.New("database: engine must be postgresql, mysql, or redis")
	ErrInvalidVersion = errors.New("database: version is invalid")
)

var safeNamePattern = regexp.MustCompile(`[^a-z0-9]+`)

// Service manages database instance provisioning records.
type Service struct {
	repo *Repository
}

// NewService creates a database service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type dbDefaults struct {
	host    string
	port    int
	version string
}

func defaultSpec(engine string) (dbDefaults, bool) {
	switch normalizeEngine(engine) {
	case "postgresql":
		return dbDefaults{host: "127.0.0.1", port: 5432, version: "16"}, true
	case "mysql":
		return dbDefaults{host: "127.0.0.1", port: 3306, version: "8.0"}, true
	case "redis":
		return dbDefaults{host: "127.0.0.1", port: 6379, version: "7"}, true
	default:
		return dbDefaults{}, false
	}
}

func normalizeEngine(engine string) string {
	switch strings.ToLower(strings.TrimSpace(engine)) {
	case "postgres", "postgresql", "pg":
		return "postgresql"
	case "mysql":
		return "mysql"
	case "redis":
		return "redis"
	default:
		return ""
	}
}

func randomHexString(bytes int) string {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "deadbeef"
	}
	return hex.EncodeToString(buf)
}

func sanitizeDBName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = safeNamePattern.ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	if name == "" {
		name = "database"
	}
	return name
}

func makeDatabaseName(name string) string {
	return fmt.Sprintf("%s_%s", sanitizeDBName(name), randomHexString(3))
}

func makeUsername(name string) string {
	return fmt.Sprintf("db_%s", sanitizeDBName(name))
}

func makeConnectionString(inst *Instance) string {
	switch inst.Engine {
	case "postgresql":
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", inst.Username, inst.Password, inst.Host, inst.Port, inst.DatabaseName)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", inst.Username, inst.Password, inst.Host, inst.Port, inst.DatabaseName)
	case "redis":
		return fmt.Sprintf("redis://:%s@%s:%d/0", inst.Password, inst.Host, inst.Port)
	default:
		return ""
	}
}

// List returns all database instances visible to the user.
func (s *Service) List(userID string) ([]Instance, error) {
	instances, err := s.repo.listForUser(userID)
	if err != nil {
		return nil, err
	}
	if instances == nil {
		return []Instance{}, nil
	}
	return instances, nil
}

// Create provisions a new database instance record.
func (s *Service) Create(userID string, req CreateRequest) (*Instance, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrInvalidName
	}

	engine := normalizeEngine(req.Engine)
	defaults, ok := defaultSpec(engine)
	if !ok {
		return nil, ErrInvalidEngine
	}

	version := strings.TrimSpace(req.Version)
	if version == "" {
		version = defaults.version
	}

	inst := &Instance{
		OwnerUserID:      userID,
		Name:             name,
		Engine:           engine,
		Version:          version,
		Host:             defaults.host,
		Port:             defaults.port,
		Username:         makeUsername(name),
		Password:         randomHexString(16),
		DatabaseName:     makeDatabaseName(name),
		Status:           "provisioned",
		ConnectionString: "",
	}
	inst.ConnectionString = makeConnectionString(inst)
	created, err := s.repo.createForUser(userID, inst)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// Get retrieves a database instance for the current user.
func (s *Service) Get(userID, id string) (*Instance, error) {
	inst, err := s.repo.findByIDForUser(userID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return inst, nil
}

// Delete removes a database instance owned by the user.
func (s *Service) Delete(userID, id string) error {
	if err := s.repo.deleteForUser(userID, id); err != nil {
		return ErrNotFound
	}
	return nil
}
