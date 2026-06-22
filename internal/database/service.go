package database

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/razad/razad/internal/process"
)

var (
	ErrNotFound       = errors.New("database: not found")
	ErrInvalidName    = errors.New("database: name is required")
	ErrInvalidEngine  = errors.New("database: engine must be postgresql, mysql, redis, mongodb, or sqlite")
	ErrInvalidVersion = errors.New("database: version is invalid")
)

var safeNamePattern = regexp.MustCompile(`[^a-z0-9]+`)

// Service manages database instance provisioning records and runtime lifecycles.
type Service struct {
	repo    *Repository
	runner  process.Runner
	dataDir string
}

// NewService creates a database service.
func NewService(repo *Repository, runner process.Runner, dataDir string) *Service {
	return &Service{repo: repo, runner: runner, dataDir: dataDir}
}

type dbDefaults struct {
	host       string
	port       int
	version    string
	deployable bool
}

func defaultSpec(engine string) (dbDefaults, bool) {
	switch normalizeEngine(engine) {
	case "postgresql":
		return dbDefaults{host: "127.0.0.1", port: 5432, version: "16", deployable: true}, true
	case "mysql":
		return dbDefaults{host: "127.0.0.1", port: 3306, version: "8.0", deployable: true}, true
	case "redis":
		return dbDefaults{host: "127.0.0.1", port: 6379, version: "7", deployable: true}, true
	case "mongodb":
		return dbDefaults{host: "127.0.0.1", port: 27017, version: "7", deployable: true}, true
	case "sqlite":
		return dbDefaults{host: "127.0.0.1", port: 0, version: "3", deployable: false}, true
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
	case "mongo", "mongodb":
		return "mongodb"
	case "sqlite", "sqlite3":
		return "sqlite"
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

func (s *Service) baseDir() string {
	if strings.TrimSpace(s.dataDir) == "" {
		return os.TempDir()
	}
	return s.dataDir
}

func (s *Service) instanceDir(id string) string {
	return filepath.Join(s.baseDir(), "databases", id)
}

func (s *Service) instanceDataDir(id string) string {
	return filepath.Join(s.instanceDir(id), "data")
}

func (s *Service) sqlitePath(id string) string {
	return filepath.Join(s.instanceDir(id), "database.db")
}

func allocatePort(host string) (int, error) {
	ln, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("database: unexpected address type %T", ln.Addr())
	}
	return addr.Port, nil
}

func connectionStringFor(inst *Instance, runtimeDir string) string {
	switch inst.Engine {
	case "postgresql":
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", inst.Username, inst.Password, inst.Host, inst.Port, inst.DatabaseName)

		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", inst.Username, inst.Password, inst.Host, inst.Port, inst.DatabaseName)
	case "redis":
		return fmt.Sprintf("redis://:%s@%s:%d/0", inst.Password, inst.Host, inst.Port)
	case "mongodb":
		return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=admin", inst.Username, inst.Password, inst.Host, inst.Port, inst.DatabaseName)
	case "sqlite":
		return fmt.Sprintf("file:%s", filepath.Clean(filepath.Join(runtimeDir, "database.db")))
	default:
		return ""
	}
}

func engineCommand(inst *Instance, runtimeDir string) (string, error) {
	dataDir := filepath.Join(runtimeDir, "data")
	switch inst.Engine {
	case "postgresql":
		return fmt.Sprintf("postgres -D %s -p %d -h %s", dataDir, inst.Port, inst.Host), nil
	case "mysql":
		return fmt.Sprintf("mysqld --datadir=%s --port=%d --bind-address=%s", dataDir, inst.Port, inst.Host), nil
	case "redis":
		return fmt.Sprintf("redis-server --port %d --bind %s --dir %s --dbfilename dump.rdb", inst.Port, inst.Host, dataDir), nil
	case "mongodb":
		return fmt.Sprintf("mongod --dbpath=%s --port=%d --bind_ip=%s", dataDir, inst.Port, inst.Host), nil
	case "sqlite":
		return "", nil
	default:
		return "", ErrInvalidEngine
	}
}

func (s *Service) prepareRuntimeDirs(inst *Instance) error {
	if err := os.MkdirAll(s.instanceDataDir(inst.ID), 0o755); err != nil {
		return fmt.Errorf("database: create data dir: %w", err)
	}
	if inst.Engine == "sqlite" {
		path := s.sqlitePath(inst.ID)
		f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
		if err != nil {
			return fmt.Errorf("database: create sqlite file: %w", err)
		}
		_ = f.Close()
	}
	return nil
}

func (s *Service) persistCreated(userID string, inst *Instance) (*Instance, error) {
	created, err := s.repo.createForUser(userID, inst)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) deploy(ctx context.Context, inst *Instance) (*Instance, error) {
	spec, ok := defaultSpec(inst.Engine)
	if !ok {
		return nil, ErrInvalidEngine
	}
	if !spec.deployable {
		if err := s.repo.updateStatus(inst.ID, inst.OwnerUserID, "provisioned"); err != nil {
			return nil, err
		}
		updated, err := s.repo.findByIDForUser(inst.OwnerUserID, inst.ID)
		if err != nil {
			return nil, err
		}
		return updated, nil
	}
	if s.runner == nil {
		return nil, fmt.Errorf("database: no process runner configured")
	}

	runtimeDir := s.instanceDir(inst.ID)
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		return nil, fmt.Errorf("database: create runtime dir: %w", err)
	}
	if err := os.MkdirAll(s.instanceDataDir(inst.ID), 0o755); err != nil {
		return nil, fmt.Errorf("database: create data dir: %w", err)
	}

	command, err := engineCommand(inst, runtimeDir)
	if err != nil {
		return nil, err
	}
	if err := s.runner.Start(ctx, inst.ID, command, nil, runtimeDir); err != nil {
		_ = s.repo.updateStatus(inst.ID, inst.OwnerUserID, "failed")
		return nil, fmt.Errorf("database: start %s: %w", inst.Engine, err)
	}
	if err := s.repo.updateStatus(inst.ID, inst.OwnerUserID, "running"); err != nil {
		return nil, err
	}
	return s.repo.findByIDForUser(inst.OwnerUserID, inst.ID)
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

// Create provisions a new database instance record and starts the backing service when supported.
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
	if engine == "sqlite" && version == "" {
		version = "3"
	}

	port := defaults.port
	if defaults.deployable {
		allocated, err := allocatePort(defaults.host)
		if err != nil {
			return nil, fmt.Errorf("database: allocate port: %w", err)
		}
		port = allocated
	}

	inst := &Instance{
		ID:           randomHexString(16),
		OwnerUserID:  userID,
		Name:         name,
		Engine:       engine,
		Version:      version,
		Host:         defaults.host,
		Port:         port,
		Username:     makeUsername(name),
		Password:     randomHexString(16),
		DatabaseName: makeDatabaseName(name),
		Status:       "provisioning",
	}
	if engine == "sqlite" {
		inst.Status = "provisioned"
	}
	inst.ConnectionString = connectionStringFor(inst, s.instanceDir(inst.ID))

	if err := s.prepareRuntimeDirs(inst); err != nil {
		return nil, err
	}

	created, err := s.persistCreated(userID, inst)
	if err != nil {
		return nil, err
	}

	if engine == "sqlite" {
		return created, nil
	}

	deployed, err := s.deploy(context.Background(), created)
	if err != nil {
		return nil, err
	}
	return deployed, nil
}

// Get retrieves a database instance for the current user.
func (s *Service) Get(userID, id string) (*Instance, error) {
	inst, err := s.repo.findByIDForUser(userID, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return inst, nil
}

// Deploy starts or re-provisions a database instance.
func (s *Service) Deploy(userID, id string) (*Instance, error) {
	inst, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}
	if err := s.prepareRuntimeDirs(inst); err != nil {
		return nil, err
	}
	return s.deploy(context.Background(), inst)
}

// Stop stops a running database instance.
func (s *Service) Stop(userID, id string) (*Instance, error) {
	inst, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}
	if inst.Engine == "sqlite" {
		return inst, nil
	}
	if s.runner == nil {
		return nil, fmt.Errorf("database: no process runner configured")
	}
	if err := s.runner.Stop(context.Background(), id); err != nil {
		return nil, fmt.Errorf("database: stop %s: %w", inst.Engine, err)
	}
	if err := s.repo.updateStatus(inst.ID, inst.OwnerUserID, "stopped"); err != nil {
		return nil, err
	}
	return s.repo.findByIDForUser(inst.OwnerUserID, inst.ID)
}

// Restart restarts a database instance.
func (s *Service) Restart(userID, id string) (*Instance, error) {
	inst, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}
	if inst.Engine == "sqlite" {
		return inst, nil
	}
	if s.runner == nil {
		return nil, fmt.Errorf("database: no process runner configured")
	}
	command, err := engineCommand(inst, s.instanceDir(inst.ID))
	if err != nil {
		return nil, err
	}
	_ = s.runner.Stop(context.Background(), id)
	if err := s.runner.Start(context.Background(), id, command, nil, s.instanceDir(inst.ID)); err != nil {
		_ = s.repo.updateStatus(inst.ID, inst.OwnerUserID, "failed")
		return nil, fmt.Errorf("database: restart %s: %w", inst.Engine, err)
	}
	if err := s.repo.updateStatus(inst.ID, inst.OwnerUserID, "running"); err != nil {
		return nil, err
	}
	return s.repo.findByIDForUser(inst.OwnerUserID, inst.ID)
}

// Status refreshes the database instance status from the runner.
func (s *Service) Status(userID, id string) (*Instance, error) {
	inst, err := s.Get(userID, id)
	if err != nil {
		return nil, err
	}
	if inst.Engine == "sqlite" {
		if err := s.repo.updateStatus(inst.ID, inst.OwnerUserID, "provisioned"); err != nil {
			return nil, err
		}
		return s.repo.findByIDForUser(inst.OwnerUserID, inst.ID)
	}
	if s.runner == nil {
		return nil, fmt.Errorf("database: no process runner configured")
	}
	state, err := s.runner.Status(context.Background(), id)
	if err != nil {
		return nil, err
	}
	status := string(state)
	if status == "" {
		status = "unknown"
	}
	if err := s.repo.updateStatus(inst.ID, inst.OwnerUserID, status); err != nil {
		return nil, err
	}
	return s.repo.findByIDForUser(inst.OwnerUserID, inst.ID)
}

// Delete removes a database instance owned by the user.
func (s *Service) Delete(userID, id string) error {
	inst, err := s.Get(userID, id)
	if err != nil {
		return err
	}
	if inst.Engine != "sqlite" && s.runner != nil {
		_ = s.runner.Stop(context.Background(), id)
	}
	if err := s.repo.deleteForUser(userID, id); err != nil {
		return ErrNotFound
	}
	_ = os.RemoveAll(s.instanceDir(id))
	return nil
}
