-- 0002_erd_parity.sql
-- Add the ERD-parity tables and the credentials split for database_instances.
--
-- This migration is purely additive from the perspective of existing tables.
-- The credentials columns (username, password, connection_string) have already
-- been removed from `database_instances` in 0001 so we do not need to backfill
-- any data; if you are upgrading from the pre-migration schema, drop those
-- columns manually before applying this migration.
--
-- Tables added (per docs/razad_erd_v_1_0.md):
--   servers, node_agents, node_heartbeats
--   domains, domain_bindings, ssl_certificates
--   ai_policies, ai_action_templates, ai_actions
--   system_services, database_credentials, database_backups
--   log_sources, provisioning_jobs, health_snapshots
--
-- All tables use TEXT primary keys (UUIDs) and TEXT timestamps in UTC ISO-8601.

-- ---------------------------------------------------------------------------
-- Servers and node agents (execution plane)
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS servers (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organizations(id),
    project_id TEXT REFERENCES projects(id),
    name TEXT NOT NULL,
    provider_type TEXT NOT NULL DEFAULT 'self_hosted',
    mode TEXT NOT NULL DEFAULT 'self_hosted',
    region TEXT NOT NULL DEFAULT '',
    hostname TEXT NOT NULL DEFAULT '',
    ip_address TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'unknown',
    last_seen_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS node_agents (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL REFERENCES servers(id),
    identity_key TEXT NOT NULL UNIQUE,
    version TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    last_heartbeat_at TIMESTAMP,
    enrolled_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS node_heartbeats (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL REFERENCES node_agents(id),
    captured_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    payload_json TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS health_snapshots (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL REFERENCES node_agents(id),
    captured_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    payload_json TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS system_services (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL REFERENCES servers(id),
    name TEXT NOT NULL,
    unit_name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'unknown',
    last_checked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(server_id, unit_name)
);

CREATE TABLE IF NOT EXISTS log_sources (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL REFERENCES servers(id),
    source_type TEXT NOT NULL,
    path TEXT NOT NULL,
    active INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS provisioning_jobs (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id),
    server_id TEXT REFERENCES servers(id),
    kind TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    payload_json TEXT NOT NULL DEFAULT '{}',
    error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ---------------------------------------------------------------------------
-- Domains and SSL
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS domains (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id),
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, name)
);

CREATE TABLE IF NOT EXISTS domain_bindings (
    id TEXT PRIMARY KEY,
    domain_id TEXT NOT NULL REFERENCES domains(id),
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    path_prefix TEXT NOT NULL DEFAULT '/',
    upstream_host TEXT NOT NULL,
    upstream_port INTEGER NOT NULL,
    body_limit_mb INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(domain_id, path_prefix)
);

CREATE TABLE IF NOT EXISTS ssl_certificates (
    id TEXT PRIMARY KEY,
    domain_id TEXT NOT NULL REFERENCES domains(id),
    issuer TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    cert_path TEXT NOT NULL DEFAULT '',
    key_path TEXT NOT NULL DEFAULT '',
    issued_at TIMESTAMP,
    expires_at TIMESTAMP,
    last_renewed_at TIMESTAMP,
    renewal_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(domain_id)
);

-- ---------------------------------------------------------------------------
-- AI orchestration persistence
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS ai_policies (
    id TEXT PRIMARY KEY,
    organization_id TEXT REFERENCES organizations(id),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    definition_json TEXT NOT NULL DEFAULT '{}',
    active INTEGER NOT NULL DEFAULT 1,
    built_in INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, name)
);

CREATE TABLE IF NOT EXISTS ai_action_templates (
    id TEXT PRIMARY KEY,
    policy_id TEXT REFERENCES ai_policies(id),
    name TEXT NOT NULL,
    label TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    allowed INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(policy_id, name)
);

CREATE TABLE IF NOT EXISTS ai_actions (
    id TEXT PRIMARY KEY,
    template_id TEXT REFERENCES ai_action_templates(id),
    actor_user_id TEXT NOT NULL REFERENCES users(id),
    target TEXT NOT NULL DEFAULT '',
    reason TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'accepted',
    result_json TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ---------------------------------------------------------------------------
-- Database credentials split + backups
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS database_credentials (
    id TEXT PRIMARY KEY,
    database_instance_id TEXT NOT NULL UNIQUE REFERENCES database_instances(id),
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    connection_string TEXT NOT NULL,
    rotated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS database_backups (
    id TEXT PRIMARY KEY,
    database_instance_id TEXT NOT NULL REFERENCES database_instances(id),
    artifact_path TEXT NOT NULL,
    size_bytes INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending',
    captured_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    restored_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ---------------------------------------------------------------------------
-- Indexes for the new tables
-- ---------------------------------------------------------------------------

CREATE INDEX IF NOT EXISTS idx_servers_org ON servers (organization_id);
CREATE INDEX IF NOT EXISTS idx_servers_status ON servers (status);
CREATE INDEX IF NOT EXISTS idx_node_agents_server ON node_agents (server_id);
CREATE INDEX IF NOT EXISTS idx_node_agents_status ON node_agents (status);
CREATE INDEX IF NOT EXISTS idx_node_heartbeats_agent ON node_heartbeats (agent_id, captured_at DESC);
CREATE INDEX IF NOT EXISTS idx_health_snapshots_agent ON health_snapshots (agent_id, captured_at DESC);
CREATE INDEX IF NOT EXISTS idx_system_services_server ON system_services (server_id);
CREATE INDEX IF NOT EXISTS idx_log_sources_server ON log_sources (server_id);
CREATE INDEX IF NOT EXISTS idx_provisioning_jobs_project ON provisioning_jobs (project_id, status);
CREATE INDEX IF NOT EXISTS idx_provisioning_jobs_server ON provisioning_jobs (server_id);

CREATE INDEX IF NOT EXISTS idx_domains_project ON domains (project_id);
CREATE INDEX IF NOT EXISTS idx_domain_bindings_app ON domain_bindings (app_id);
CREATE INDEX IF NOT EXISTS idx_domain_bindings_domain ON domain_bindings (domain_id);
CREATE INDEX IF NOT EXISTS idx_ssl_certificates_domain ON ssl_certificates (domain_id);
CREATE INDEX IF NOT EXISTS idx_ssl_certificates_expires ON ssl_certificates (expires_at);

CREATE INDEX IF NOT EXISTS idx_ai_policies_org ON ai_policies (organization_id);
CREATE INDEX IF NOT EXISTS idx_ai_action_templates_policy ON ai_action_templates (policy_id);
CREATE INDEX IF NOT EXISTS idx_ai_actions_actor_created ON ai_actions (actor_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ai_actions_template ON ai_actions (template_id);

CREATE INDEX IF NOT EXISTS idx_database_backups_instance ON database_backups (database_instance_id, captured_at DESC);
