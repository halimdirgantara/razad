// Package policy is the minimal validation/authorization gate that the design
// requires every privileged mutation to pass through. It evaluates a fixed
// allow/deny decision against (Actor, Action, Resource) and emits an audit
// event for every check, regardless of the outcome.
//
// The engine is deliberately small: no DSL, no external policy file, no
// rule reload. Per-action policies are baked in via addDefaults() and can be
// overridden per-action through Engine.Register(). Sufficient for the
// self-hosted MVP; cloud mode can later replace this with a richer engine
// behind the same Actor/Action/Resource/Decision contract.
package policy

import (
	"context"
	"fmt"
	"sync"
)

// Action is the symbolic name of a privileged operation.
type Action string

// All actions defined in the system. Handlers must use these constants when
// calling Engine.Check; free-form strings will hit the "no_policy_for_action"
// deny path.
const (
	ActionUserRegister  Action = "user.register"
	ActionOrgCreate     Action = "org.create"

	ActionAppCreate   Action = "app.create"
	ActionAppDeploy   Action = "app.deploy"
	ActionAppStop     Action = "app.stop"
	ActionAppRestart  Action = "app.restart"
	ActionAppEnvWrite Action = "app.env.write"
	ActionAppDelete   Action = "app.delete"

	ActionProxyRender   Action = "proxy.render"
	ActionProxyApply    Action = "proxy.apply"
	ActionProxyRollback Action = "proxy.rollback"

	ActionSSLIssue Action = "ssl.issue"
	ActionSSLRenew Action = "ssl.renew"

	ActionDBProvision Action = "database.provision"
	ActionDBDeploy    Action = "database.deploy"
	ActionDBStop      Action = "database.stop"
	ActionDBRestart   Action = "database.restart"
	ActionDBDelete    Action = "database.delete"
)

// Actor identifies the principal performing the action.
type Actor struct {
	UserID  string
	IsAdmin bool
}

// Resource identifies the target of an action. OwnerUserID, when non-empty,
// enables owner-based authorization (the actor must match it unless they
// are an admin).
type Resource struct {
	Type        string
	ID          string
	OwnerUserID string
}

// Decision is the outcome of a policy check.
type Decision struct {
	Allow  bool
	Reason string
}

// Auditor is the minimal interface policy needs from the audit package.
// Defined here so policy has no hard dependency on internal/audit.
type Auditor interface {
	Record(ctx context.Context, actorUserID, action, entityType, entityID string, metadata map[string]any) error
}

// ActionPolicy declares who is allowed to perform an action.
type ActionPolicy struct {
	// RequireAdmin denies everyone except admins.
	RequireAdmin bool
	// AllowOwner permits the resource owner (Actor.UserID == Resource.OwnerUserID)
	// in addition to admins.
	AllowOwner bool
}

// Engine evaluates decisions and records them in the audit log.
type Engine struct {
	mu       sync.RWMutex
	auditor  Auditor
	policies map[Action]ActionPolicy
}

// New returns an Engine populated with the built-in defaults suitable for
// self-hosted mode. auditor may be nil, in which case decisions are silently
// not recorded (useful for tests that don't care about audit side-effects).
func New(auditor Auditor) *Engine {
	e := &Engine{
		auditor:  auditor,
		policies: map[Action]ActionPolicy{},
	}
	e.addDefaults()
	return e
}

func (e *Engine) addDefaults() {
	adminOnly := []Action{
		ActionAppDelete,
		ActionDBDelete,
	}
	for _, a := range adminOnly {
		e.policies[a] = ActionPolicy{RequireAdmin: true}
	}

	ownerOrAdmin := []Action{
		ActionAppCreate,
		ActionAppDeploy,
		ActionAppStop,
		ActionAppRestart,
		ActionAppEnvWrite,
		ActionProxyRender,
		ActionProxyApply,
		ActionProxyRollback,
		ActionSSLIssue,
		ActionSSLRenew,
		ActionDBProvision,
		ActionDBDeploy,
		ActionDBStop,
		ActionDBRestart,
	}
	for _, a := range ownerOrAdmin {
		e.policies[a] = ActionPolicy{AllowOwner: true}
	}

	// Registration and org creation are open to any authenticated user.
	e.policies[ActionUserRegister] = ActionPolicy{AllowOwner: true}
	e.policies[ActionOrgCreate] = ActionPolicy{AllowOwner: true}
}

// Register sets or overrides the policy for a single action. Intended for
// tests and future per-organization customization. Safe for concurrent use.
func (e *Engine) Register(action Action, p ActionPolicy) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.policies == nil {
		e.policies = map[Action]ActionPolicy{}
	}
	e.policies[action] = p
}

// Lookup returns the registered policy for an action (or the zero value).
func (e *Engine) Lookup(action Action) ActionPolicy {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.policies[action]
}

// Check evaluates the policy for action against resource on behalf of actor
// and returns the decision. Default deny. Always audits the outcome.
func (e *Engine) Check(ctx context.Context, actor Actor, action Action, resource Resource) Decision {
	d := e.evaluate(actor, action, resource)
	if e.auditor != nil {
		_ = e.auditor.Record(ctx, actor.UserID, "policy.check", string(action), resource.ID, map[string]any{
			"decision":      d.Allow,
			"reason":        d.Reason,
			"resource_type": resource.Type,
			"is_admin":      actor.IsAdmin,
		})
	}
	return d
}

func (e *Engine) evaluate(actor Actor, action Action, resource Resource) Decision {
	if actor.UserID == "" {
		return Decision{Allow: false, Reason: "unauthenticated"}
	}
	e.mu.RLock()
	policy, ok := e.policies[action]
	e.mu.RUnlock()
	if !ok {
		return Decision{Allow: false, Reason: "no_policy_for_action"}
	}
	if policy.RequireAdmin {
		if actor.IsAdmin {
			return Decision{Allow: true, Reason: "admin_allowed"}
		}
		return Decision{Allow: false, Reason: "requires_admin"}
	}
	if policy.AllowOwner {
		if actor.IsAdmin {
			return Decision{Allow: true, Reason: "admin_allowed"}
		}
		if resource.OwnerUserID != "" && actor.UserID == resource.OwnerUserID {
			return Decision{Allow: true, Reason: "owner_allowed"}
		}
		return Decision{Allow: false, Reason: "owner_or_admin_required"}
	}
	return Decision{Allow: false, Reason: "policy_not_satisfied"}
}

// MustCheck is a convenience wrapper that returns a non-nil error when the
// decision is deny.
func (e *Engine) MustCheck(ctx context.Context, actor Actor, action Action, resource Resource) error {
	d := e.Check(ctx, actor, action, resource)
	if !d.Allow {
		return fmt.Errorf("policy: denied %s on %s/%s: %s", action, resource.Type, resource.ID, d.Reason)
	}
	return nil
}

// Denied constructs an error matching MustCheck's output without re-evaluating.
// Useful when a handler has already computed the decision and only needs the
// error string for a response.
func Denied(action Action, resource Resource, reason string) error {
	return fmt.Errorf("policy: denied %s on %s/%s: %s", action, resource.Type, resource.ID, reason)
}
