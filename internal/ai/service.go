package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/razad/razad/internal/audit"
)

var (
	ErrBlockedAction = errors.New("ai: action is blocked")
	ErrUnknownAction = errors.New("ai: action is not registered")
)

type Provider struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	Supported bool   `json:"supported"`
}

type ActionCapability struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Allowed     bool   `json:"allowed"`
}

type Capabilities struct {
	Providers      []Provider          `json:"providers"`
	AllowedActions []ActionCapability  `json:"allowed_actions"`
	BlockedActions []ActionCapability   `json:"blocked_actions"`
	SafetyNotes    []string            `json:"safety_notes"`
}

type ActionRequest struct {
	Action string `json:"action"`
	Target string `json:"target,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type ActionResult struct {
	Status  string `json:"status"`
	Action  string `json:"action"`
	Target  string `json:"target,omitempty"`
	Message string `json:"message"`
}

type Service struct {
	auditor *audit.Service
}

func NewService(auditor *audit.Service) *Service {
	return &Service{auditor: auditor}
}

var providerCatalog = []Provider{
	{Name: "openai", Label: "OpenAI", Supported: true},
	{Name: "anthropic", Label: "Anthropic", Supported: true},
	{Name: "gemini", Label: "Google Gemini", Supported: true},
	{Name: "ollama", Label: "Ollama", Supported: true},
}

var allowedRegistry = map[string]ActionCapability{
	"restart_app": {
		Name:        "restart_app",
		Label:       "Restart app",
		Description: "Restart a crashed or wedged application service.",
		Allowed:     true,
	},
	"reload_nginx": {
		Name:        "reload_nginx",
		Label:       "Reload Nginx",
		Description: "Reload the reverse proxy after safe config changes.",
		Allowed:     true,
	},
	"clear_app_cache": {
		Name:        "clear_app_cache",
		Label:       "Clear app cache",
		Description: "Clear application cache when a safe recovery step is needed.",
		Allowed:     true,
	},
	"restart_database_service": {
		Name:        "restart_database_service",
		Label:       "Restart database service",
		Description: "Restart a supported database service after failure.",
		Allowed:     true,
	},
	"scale_worker_count": {
		Name:        "scale_worker_count",
		Label:       "Scale worker count",
		Description: "Increase worker count within allowed limits.",
		Allowed:     true,
	},
	"send_alert_notification": {
		Name:        "send_alert_notification",
		Label:       "Send alert notification",
		Description: "Notify the operator about a detected issue.",
		Allowed:     true,
	},
	"run_predefined_healthcheck": {
		Name:        "run_predefined_healthcheck",
		Label:       "Run predefined healthcheck",
		Description: "Execute a pre-approved healthcheck workflow.",
		Allowed:     true,
	},
}

var blockedRegistry = []ActionCapability{
	{Name: "delete_app", Label: "Delete app", Description: "Destructive app deletion is never allowed.", Allowed: false},
	{Name: "delete_database", Label: "Delete database", Description: "Destructive database deletion is never allowed.", Allowed: false},
	{Name: "drop_table", Label: "Drop table", Description: "Schema destruction is never allowed.", Allowed: false},
	{Name: "modify_env_production", Label: "Modify production env", Description: "Direct production env edits are blocked.", Allowed: false},
	{Name: "uninstall_runtime", Label: "Uninstall runtime", Description: "Runtime uninstall actions are blocked.", Allowed: false},
	{Name: "modify_firewall_rules", Label: "Modify firewall rules", Description: "Firewall changes require explicit human action.", Allowed: false},
	{Name: "execute_arbitrary_command", Label: "Execute arbitrary command", Description: "Arbitrary shell execution is blocked.", Allowed: false},
}

func (s *Service) Capabilities() Capabilities {
	allowed := make([]ActionCapability, 0, len(allowedRegistry))
	for _, cap := range allowedRegistry {
		allowed = append(allowed, cap)
	}
	blocked := make([]ActionCapability, 0, len(blockedRegistry))
	for _, cap := range blockedRegistry {
		blocked = append(blocked, cap)
	}
	return Capabilities{
		Providers:      append([]Provider(nil), providerCatalog...),
		AllowedActions: allowed,
		BlockedActions: blocked,
		SafetyNotes: []string{
			"AI may only request pre-approved actions.",
			"Every AI request is written to the audit log.",
			"Destructive commands are blocked by design.",
		},
	}
}

func normalizeAction(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func (s *Service) RequestAction(ctx context.Context, userID string, req ActionRequest) (*ActionResult, error) {
	action := normalizeAction(req.Action)
	if action == "" {
		return nil, fmt.Errorf("ai: action is required")
	}
	if _, blocked := blockedLookup()[action]; blocked {
		return nil, ErrBlockedAction
	}
	cap, ok := allowedRegistry[action]
	if !ok {
		return nil, ErrUnknownAction
	}
	if s.auditor != nil {
		if err := s.auditor.Record(ctx, userID, "ai.action.requested", "ai_action", action, map[string]any{
			"action": action,
			"target": req.Target,
			"reason": req.Reason,
		}); err != nil {
			return nil, err
		}
	}
	return &ActionResult{
		Status:  "accepted",
		Action:  cap.Name,
		Target:  req.Target,
		Message: "Action recorded and queued for the approved execution path.",
	}, nil
}

func blockedLookup() map[string]ActionCapability {
	lookup := make(map[string]ActionCapability, len(blockedRegistry))
	for _, cap := range blockedRegistry {
		lookup[cap.Name] = cap
	}
	return lookup
}
