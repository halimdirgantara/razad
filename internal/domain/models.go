// Package domain defines the core entity types for the Razad platform.
// These models are shared across all internal modules.
package domain

import "time"

// ---------------------------------------------------------------------------
// Identity & Tenancy
// ---------------------------------------------------------------------------

// User represents an authenticated human account.
type User struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	PasswordHash string `json:"-" db:"password_hash"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Organization represents a tenant boundary.
type Organization struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// OrganizationMember links a user to an organization with a role.
type OrganizationMember struct {
	ID             string    `json:"id" db:"id"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	UserID         string    `json:"user_id" db:"user_id"`
	Role           string    `json:"role" db:"role"` // admin, member
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// Project groups apps, databases, and domains within an organization.
type Project struct {
	ID             string    `json:"id" db:"id"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	Name           string    `json:"name" db:"name"`
	Slug           string    `json:"slug" db:"slug"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// ---------------------------------------------------------------------------
// Infrastructure & Nodes
// ---------------------------------------------------------------------------

// Server represents a physical or virtual machine managed by Razad.
type Server struct {
	ID             string    `json:"id" db:"id"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	Hostname       string    `json:"hostname" db:"hostname"`
	OS             string    `json:"os" db:"os"`
	Arch           string    `json:"arch" db:"arch"`
	Status         string    `json:"status" db:"status"` // active, offline, decommissioned
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// NodeAgent represents the software identity running on a server.
type NodeAgent struct {
	ID        string    `json:"id" db:"id"`
	ServerID  string    `json:"server_id" db:"server_id"`
	Version   string    `json:"version" db:"version"`
	Status    string    `json:"status" db:"status"` // enrolled, disconnected, revoked
	LastSeen  time.Time `json:"last_seen" db:"last_seen"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// NodeHeartbeat is a periodic liveness signal from a node agent.
type NodeHeartbeat struct {
	ID        string    `json:"id" db:"id"`
	NodeID    string    `json:"node_id" db:"node_id"`
	CPUUsage  float64   `json:"cpu_usage" db:"cpu_usage"`
	RAMUsage  float64   `json:"ram_usage" db:"ram_usage"`
	DiskUsage float64   `json:"disk_usage" db:"disk_usage"`
	LoadAvg   float64   `json:"load_avg" db:"load_avg"`
	RecordedAt time.Time `json:"recorded_at" db:"recorded_at"`
}

// HealthSnapshot captures server health at a point in time.
type HealthSnapshot struct {
	ID         string    `json:"id" db:"id"`
	ServerID   string    `json:"server_id" db:"server_id"`
	CPUUsage   float64   `json:"cpu_usage" db:"cpu_usage"`
	RAMUsage   float64   `json:"ram_usage" db:"ram_usage"`
	DiskUsage  float64   `json:"disk_usage" db:"disk_usage"`
	LoadAvg    float64   `json:"load_avg" db:"load_avg"`
	Uptime     int64     `json:"uptime" db:"uptime"`
	RecordedAt time.Time `json:"recorded_at" db:"recorded_at"`
}

// SystemService represents a system-level service managed by Razad.
type SystemService struct {
	ID        string    `json:"id" db:"id"`
	ServerID  string    `json:"server_id" db:"server_id"`
	Name      string    `json:"name" db:"name"`
	Type      string    `json:"type" db:"type"` // systemd, nginx, postgres, certbot
	Status    string    `json:"status" db:"status"` // running, stopped, failed
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ---------------------------------------------------------------------------
// Applications & Deployments
// ---------------------------------------------------------------------------

// App represents a managed application.
type App struct {
	ID        string    `json:"id" db:"id"`
	ProjectID string    `json:"project_id" db:"project_id"`
	Name      string    `json:"name" db:"name"`
	GitURL    string    `json:"git_url,omitempty" db:"git_url"`
	Runtime   string    `json:"runtime" db:"runtime"`
	StartCmd  string    `json:"start_cmd,omitempty" db:"start_cmd"`
	Status    string    `json:"status" db:"status"` // created, deploying, running, stopped, failed, deleted
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// AppDeployment records a single deployment attempt.
type AppDeployment struct {
	ID        string    `json:"id" db:"id"`
	AppID     string    `json:"app_id" db:"app_id"`
	Version   string    `json:"version" db:"version"`
	Status    string    `json:"status" db:"status"` // pending, running, success, failed
	Log       string    `json:"log,omitempty" db:"log"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// AppEnvVar is an environment variable for an application, stored encrypted.
type AppEnvVar struct {
	ID        string    `json:"id" db:"id"`
	AppID     string    `json:"app_id" db:"app_id"`
	Key       string    `json:"key" db:"key"`
	Value     string    `json:"-" db:"value"` // encrypted at rest
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// AppLogStream represents a log output stream from an application.
type AppLogStream struct {
	ID        string    `json:"id" db:"id"`
	AppID     string    `json:"app_id" db:"app_id"`
	Source    string    `json:"source" db:"source"` // stdout, stderr
	Message   string    `json:"message" db:"message"`
	Level     string    `json:"level" db:"level"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

// ---------------------------------------------------------------------------
// Domains & SSL
// ---------------------------------------------------------------------------

// Domain is a DNS domain name that can be bound to an app.
type Domain struct {
	ID        string    `json:"id" db:"id"`
	OrgID     string    `json:"org_id" db:"org_id"`
	Domain    string    `json:"domain" db:"domain"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DomainBinding links a domain to an application.
type DomainBinding struct {
	ID        string    `json:"id" db:"id"`
	DomainID  string    `json:"domain_id" db:"domain_id"`
	AppID     string    `json:"app_id" db:"app_id"`
	Port      int       `json:"port" db:"port"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// SSLCertificate tracks an SSL certificate for a domain.
type SSLCertificate struct {
	ID           string    `json:"id" db:"id"`
	DomainID     string    `json:"domain_id" db:"domain_id"`
	Status       string    `json:"status" db:"status"` // issued, renewing, failed, expired
	IssuedAt     time.Time `json:"issued_at" db:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	RenewedAt    time.Time `json:"renewed_at,omitempty" db:"renewed_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// ---------------------------------------------------------------------------
// Databases
// ---------------------------------------------------------------------------

// DatabaseInstance is a managed database service.
type DatabaseInstance struct {
	ID        string    `json:"id" db:"id"`
	ServerID  string    `json:"server_id" db:"server_id"`
	ProjectID string    `json:"project_id" db:"project_id"`
	Type      string    `json:"type" db:"type"` // postgresql (later: mysql, mariadb)
	Status    string    `json:"status" db:"status"` // provisioning, running, failed, deleting
	Port      int       `json:"port" db:"port"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DatabaseCredential holds connection credentials for a database.
type DatabaseCredential struct {
	ID         string    `json:"id" db:"id"`
	InstanceID string    `json:"instance_id" db:"instance_id"`
	Username   string    `json:"username" db:"username"`
	Password   string    `json:"-" db:"password"` // encrypted at rest
	Database   string    `json:"database" db:"database"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// DatabaseBackup records a database backup.
type DatabaseBackup struct {
	ID         string    `json:"id" db:"id"`
	InstanceID string    `json:"instance_id" db:"instance_id"`
	FilePath   string    `json:"file_path" db:"file_path"`
	SizeBytes  int64     `json:"size_bytes" db:"size_bytes"`
	Status     string    `json:"status" db:"status"` // created, uploaded, failed
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// ---------------------------------------------------------------------------
// AI Operations & Policy
// ---------------------------------------------------------------------------

// AIPolicy defines allowed AI actions.
type AIPolicy struct {
	ID        string    `json:"id" db:"id"`
	OrgID     string    `json:"org_id" db:"org_id"`
	Name      string    `json:"name" db:"name"`
	Enabled   bool      `json:"enabled" db:"enabled"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// AIActionTemplate defines a whitelisted AI action with parameters.
type AIActionTemplate struct {
	ID         string    `json:"id" db:"id"`
	PolicyID   string    `json:"policy_id" db:"policy_id"`
	Name       string    `json:"name" db:"name"`
	Destructive bool     `json:"destructive" db:"destructive"` // must be false for AI
	ParamSchema string    `json:"param_schema" db:"param_schema"` // JSON schema
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// AIAction records a single AI action proposal and its outcome.
type AIAction struct {
	ID          string    `json:"id" db:"id"`
	AppID       string    `json:"app_id" db:"app_id"`
	TemplateID  string    `json:"template_id" db:"template_id"`
	Proposal    string    `json:"proposal" db:"proposal"`
	PolicyResult string   `json:"policy_result" db:"policy_result"` // allowed, denied
	Executed    bool      `json:"executed" db:"executed"`
	Result      string    `json:"result,omitempty" db:"result"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ---------------------------------------------------------------------------
// Audit
// ---------------------------------------------------------------------------

// AuditEvent records a privileged action for the immutable audit trail.
type AuditEvent struct {
	ID        string    `json:"id" db:"id"`
	ActorID   string    `json:"actor_id" db:"actor_id"`
	Action    string    `json:"action" db:"action"`
	Target    string    `json:"target" db:"target"`
	TargetID  string    `json:"target_id" db:"target_id"`
	Success   bool      `json:"success" db:"success"`
	Metadata  string    `json:"metadata,omitempty" db:"metadata"` // JSON blob
	IP        string    `json:"ip,omitempty" db:"ip"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ---------------------------------------------------------------------------
// Logs & Metrics
// ---------------------------------------------------------------------------

// LogSource identifies a log producer on a server.
type LogSource struct {
	ID        string    `json:"id" db:"id"`
	ServerID  string    `json:"server_id" db:"server_id"`
	Name      string    `json:"name" db:"name"`
	Type      string    `json:"type" db:"type"` // app, systemd, nginx, database
	Path      string    `json:"path" db:"path"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ---------------------------------------------------------------------------
// Billing & Provisioning (Cloud Mode)
// ---------------------------------------------------------------------------

// BillingAccount links billing info to an organization.
type BillingAccount struct {
	ID            string    `json:"id" db:"id"`
	OrganizationID string   `json:"organization_id" db:"organization_id"`
	Provider      string    `json:"provider" db:"provider"`
	ProviderID    string    `json:"provider_id" db:"provider_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// Subscription represents a billing subscription.
type Subscription struct {
	ID               string    `json:"id" db:"id"`
	BillingAccountID string    `json:"billing_account_id" db:"billing_account_id"`
	Plan             string    `json:"plan" db:"plan"`
	Status           string    `json:"status" db:"status"` // active, canceled, past_due
	CurrentPeriodStart time.Time `json:"current_period_start" db:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end" db:"current_period_end"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// Invoice represents a billing invoice.
type Invoice struct {
	ID               string    `json:"id" db:"id"`
	SubscriptionID   string    `json:"subscription_id" db:"subscription_id"`
	Amount           int64     `json:"amount" db:"amount"` // cents
	Currency         string    `json:"currency" db:"currency"`
	Status           string    `json:"status" db:"status"` // pending, paid, failed
	PaidAt           time.Time `json:"paid_at,omitempty" db:"paid_at"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// UsageRecord tracks resource usage for billing.
type UsageRecord struct {
	ID             string    `json:"id" db:"id"`
	SubscriptionID string    `json:"subscription_id" db:"subscription_id"`
	Resource       string    `json:"resource" db:"resource"`
	Quantity       float64   `json:"quantity" db:"quantity"`
	RecordedAt     time.Time `json:"recorded_at" db:"recorded_at"`
}

// ProvisioningJob tracks infrastructure provisioning operations.
type ProvisioningJob struct {
	ID        string    `json:"id" db:"id"`
	OrgID     string    `json:"org_id" db:"org_id"`
	Type      string    `json:"type" db:"type"` // create_vps, install_agent, etc.
	Status    string    `json:"status" db:"status"` // pending, running, success, failed
	Provider  string    `json:"provider" db:"provider"`
	ProviderID string   `json:"provider_id,omitempty" db:"provider_id"`
	Error     string    `json:"error,omitempty" db:"error"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
