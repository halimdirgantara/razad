package auth

import "net/http"

// Actor is the authenticated principal plus role information needed by the
// policy engine to evaluate decisions. Handlers obtain an Actor via GetActor
// and pass it (after conversion to policy.Actor) to policy.MustCheck.
type Actor struct {
	UserID  string
	IsAdmin bool
}

// AdminRule decides whether a given user ID should be treated as an admin
// for the purposes of policy evaluation. It is installed once at startup by
// the daemon via SetAdminRule. The function must be safe for concurrent reads.
type AdminRule func(userID string) bool

// adminRule defaults to "nobody is admin". main.go must install a real rule
// before serving traffic.
var adminRule AdminRule = func(string) bool { return false }

// SetAdminRule installs the admin predicate. A nil fn is ignored so the
// default "nobody is admin" remains in place.
func SetAdminRule(fn AdminRule) {
	if fn != nil {
		adminRule = fn
	}
}

// GetActor returns the Actor for the current request, deriving IsAdmin from
// the configured admin rule. Unauthenticated requests receive an Actor with
// an empty UserID; the policy engine treats those as deny by default.
func GetActor(r *http.Request) Actor {
	userID := GetUserID(r)
	return Actor{
		UserID:  userID,
		IsAdmin: userID != "" && adminRule(userID),
	}
}

// IsAdmin reports whether the given user ID should be treated as an admin.
// It is the package-level form of Actor.IsAdmin, exposed so other packages
// (notably the database handler, which cannot import auth without creating a
// test-binary import cycle) can apply the same admin rule through a
// SetIsAdminFn callback wired up by main.go.
func IsAdmin(userID string) bool {
	return userID != "" && adminRule(userID)
}
