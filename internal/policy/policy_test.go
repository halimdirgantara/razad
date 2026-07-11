package policy

import (
	"context"
	"errors"
	"sync"
	"testing"
)

// fakeAuditor captures every Record call so tests can assert on audit side
// effects without a real database.
type fakeAuditor struct {
	mu     sync.Mutex
	events []recordedEvent
}

type recordedEvent struct {
	Actor       string
	Action      string
	EntityType  string
	EntityID    string
	Metadata    map[string]any
}

func (f *fakeAuditor) Record(_ context.Context, actor, action, entityType, entityID string, metadata map[string]any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make(map[string]any, len(metadata))
	for k, v := range metadata {
		cp[k] = v
	}
	f.events = append(f.events, recordedEvent{
		Actor:      actor,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   cp,
	})
	return nil
}

func (f *fakeAuditor) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.events)
}

func TestUnauthenticatedDenied(t *testing.T) {
	e := New(nil)
	d := e.Check(context.Background(), Actor{}, ActionAppDeploy, Resource{Type: "app", ID: "x"})
	if d.Allow {
		t.Fatalf("expected deny for empty actor, got allow: %+v", d)
	}
	if d.Reason != "unauthenticated" {
		t.Fatalf("expected reason=unauthenticated, got %q", d.Reason)
	}
}

func TestUnknownActionDenied(t *testing.T) {
	e := New(nil)
	e.Register("custom.action", ActionPolicy{AllowOwner: true})
	d := e.Check(context.Background(), Actor{UserID: "u1"}, "unknown.action", Resource{Type: "x"})
	if d.Allow {
		t.Fatalf("expected deny for unregistered action")
	}
	if d.Reason != "no_policy_for_action" {
		t.Fatalf("expected reason=no_policy_for_action, got %q", d.Reason)
	}
}

func TestAdminCanDoAdminOnly(t *testing.T) {
	e := New(nil)
	d := e.Check(context.Background(), Actor{UserID: "u1", IsAdmin: true}, ActionAppDelete, Resource{Type: "app", ID: "x"})
	if !d.Allow {
		t.Fatalf("expected admin allow, got deny: %+v", d)
	}
}

func TestNonAdminDeniedForAdminOnly(t *testing.T) {
	e := New(nil)
	d := e.Check(context.Background(), Actor{UserID: "u1", IsAdmin: false}, ActionAppDelete, Resource{Type: "app", ID: "x"})
	if d.Allow {
		t.Fatalf("expected deny, got allow: %+v", d)
	}
	if d.Reason != "requires_admin" {
		t.Fatalf("expected reason=requires_admin, got %q", d.Reason)
	}
}

func TestOwnerCanDoOwnerOrAdmin(t *testing.T) {
	e := New(nil)
	d := e.Check(context.Background(),
		Actor{UserID: "u1"},
		ActionAppDeploy,
		Resource{Type: "app", ID: "a1", OwnerUserID: "u1"},
	)
	if !d.Allow {
		t.Fatalf("expected owner allow, got deny: %+v", d)
	}
	if d.Reason != "owner_allowed" {
		t.Fatalf("expected reason=owner_allowed, got %q", d.Reason)
	}
}

func TestNonOwnerNonAdminDeniedForOwnerOrAdmin(t *testing.T) {
	e := New(nil)
	d := e.Check(context.Background(),
		Actor{UserID: "u1"},
		ActionAppDeploy,
		Resource{Type: "app", ID: "a1", OwnerUserID: "u2"},
	)
	if d.Allow {
		t.Fatalf("expected deny, got allow: %+v", d)
	}
	if d.Reason != "owner_or_admin_required" {
		t.Fatalf("expected reason=owner_or_admin_required, got %q", d.Reason)
	}
}

func TestAdminCanDoOwnerOrAdminEvenWithoutOwnership(t *testing.T) {
	e := New(nil)
	d := e.Check(context.Background(),
		Actor{UserID: "u1", IsAdmin: true},
		ActionAppDeploy,
		Resource{Type: "app", ID: "a1", OwnerUserID: "u2"},
	)
	if !d.Allow {
		t.Fatalf("expected admin allow, got deny: %+v", d)
	}
}

func TestMissingOwnerUserIDDeniesNonAdmin(t *testing.T) {
	e := New(nil)
	d := e.Check(context.Background(),
		Actor{UserID: "u1"},
		ActionAppDeploy,
		Resource{Type: "app", ID: "a1"}, // no OwnerUserID
	)
	if d.Allow {
		t.Fatalf("expected deny when ownership is unknown, got allow")
	}
}

func TestEveryCheckIsAudited(t *testing.T) {
	a := &fakeAuditor{}
	e := New(a)

	// Allow path.
	_ = e.Check(context.Background(), Actor{UserID: "u1", IsAdmin: true}, ActionAppDelete, Resource{Type: "app", ID: "a1"})
	// Deny path.
	_ = e.Check(context.Background(), Actor{UserID: "u2"}, ActionAppDelete, Resource{Type: "app", ID: "a1"})

	if got, want := a.count(), 2; got != want {
		t.Fatalf("audited events: got %d, want %d", got, want)
	}
	// Verify both events were policy.check with the right action.
	for _, ev := range a.events {
		if ev.Action != "policy.check" {
			t.Errorf("unexpected action %q", ev.Action)
		}
		if ev.EntityType != "app.delete" {
			t.Errorf("unexpected entity type %q", ev.EntityType)
		}
	}
}

func TestMustCheckReturnsErrorOnDeny(t *testing.T) {
	e := New(nil)
	err := e.MustCheck(context.Background(), Actor{UserID: "u1"}, ActionAppDelete, Resource{Type: "app", ID: "a1"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !contains(err.Error(), "policy: denied") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestMustCheckReturnsNilOnAllow(t *testing.T) {
	e := New(nil)
	err := e.MustCheck(context.Background(), Actor{UserID: "u1", IsAdmin: true}, ActionAppDelete, Resource{Type: "app", ID: "a1"})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRegisterOverridesDefault(t *testing.T) {
	e := New(nil)
	// Loosen user.register: even unauthenticated users... actually we can't
	// bypass unauthenticated. Use a policy that allows admins only.
	e.Register(ActionUserRegister, ActionPolicy{RequireAdmin: true})

	d := e.Check(context.Background(), Actor{UserID: "u1"}, ActionUserRegister, Resource{Type: "user"})
	if d.Allow {
		t.Fatal("expected deny for non-admin after override")
	}

	d = e.Check(context.Background(), Actor{UserID: "u1", IsAdmin: true}, ActionUserRegister, Resource{Type: "user"})
	if !d.Allow {
		t.Fatal("expected admin allow after override")
	}
}

func TestLookupReturnsRegisteredPolicy(t *testing.T) {
	e := New(nil)
	got := e.Lookup(ActionAppDelete)
	if !got.RequireAdmin {
		t.Fatal("expected ActionAppDelete to default to RequireAdmin")
	}
	got = e.Lookup("not.registered")
	if got.RequireAdmin || got.AllowOwner {
		t.Fatal("expected zero value for unknown action")
	}
}

func TestNilAuditorDoesNotPanic(t *testing.T) {
	e := New(nil)
	// Just verify Check returns without panicking when auditor is nil.
	_ = e.Check(context.Background(), Actor{UserID: "u1", IsAdmin: true}, ActionAppDelete, Resource{Type: "app"})
}

func TestAuditErrorIsIgnored(t *testing.T) {
	// If the auditor returns an error, Check should still return a decision
	// and not propagate the error (audit failures must not block privileged
	// checks at the engine boundary; the caller may choose to react).
	failingAuditor := auditorFunc(func(_ context.Context, _, _, _, _ string, _ map[string]any) error {
		return errors.New("audit down")
	})
	e := New(failingAuditor)
	d := e.Check(context.Background(), Actor{UserID: "u1", IsAdmin: true}, ActionAppDelete, Resource{Type: "app"})
	if !d.Allow {
		t.Fatal("expected allow despite audit error")
	}
}

type auditorFunc func(ctx context.Context, actor, action, entityType, entityID string, metadata map[string]any) error

func (f auditorFunc) Record(ctx context.Context, actor, action, entityType, entityID string, metadata map[string]any) error {
	return f(ctx, actor, action, entityType, entityID, metadata)
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
