package database

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/razad/razad/internal/api"
	"github.com/razad/razad/internal/policy"
	"github.com/razad/razad/internal/requestctx"
)

// isAdminFn is set once at startup by SetIsAdminFn. It is consulted by the
// gate helper to populate the policy.Actor's IsAdmin field without forcing
// this package to import internal/auth (which would create a test-binary
// import cycle through internal/auth/service_test.go).
var isAdminFn = func(string) bool { return false }

// SetIsAdminFn installs the admin predicate used by the gate helper. A nil
// fn is ignored so the default "nobody is admin" remains in place. main.go
// calls this with auth.IsAdmin after the admin rule has been installed.
func SetIsAdminFn(fn func(userID string) bool) {
	if fn != nil {
		isAdminFn = fn
	}
}

func extractID(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	id := strings.TrimPrefix(path, prefix)
	if i := strings.IndexByte(id, '/'); i >= 0 {
		id = id[:i]
	}
	return id
}

// Handler exposes database-management endpoints.
type Handler struct {
	svc    *Service
	policy *policy.Engine
}

// NewHandler creates a database HTTP handler.
func NewHandler(svc *Service, pol *policy.Engine) *Handler {
	return &Handler{svc: svc, policy: pol}
}

func (h *Handler) gate(w http.ResponseWriter, r *http.Request, action policy.Action, resource policy.Resource) bool {
	if h.policy == nil {
		return true
	}
	userID := requestctx.UserID(r.Context())
	actor := policy.Actor{UserID: userID, IsAdmin: isAdminFn(userID)}
	if err := h.policy.MustCheck(r.Context(), actor, action, resource); err != nil {
		api.WriteError(w, http.StatusForbidden, "forbidden", err.Error())
		return false
	}
	return true
}

func (h *Handler) userID(r *http.Request) string {
	return requestctx.UserID(r.Context())
}

// List handles GET /api/v1/databases.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET")
		return
	}
	userID := h.userID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}
	instances, err := h.svc.List(userID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "list_failed", "could not list databases")
		return
	}
	api.WriteJSON(w, http.StatusOK, instances)
}

// Create handles POST /api/v1/databases.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}
	userID := h.userID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}
	inst, err := h.svc.Create(userID, req)
	if err != nil {
		code := "create_failed"
		status := http.StatusBadRequest
		switch err {
		case ErrInvalidName, ErrInvalidEngine, ErrInvalidVersion:
			code = "validation_error"
		case ErrNotFound:
			status = http.StatusNotFound
			code = "not_found"
		}
		api.WriteError(w, status, code, err.Error())
		return
	}
	api.WriteJSON(w, http.StatusCreated, inst)
}

// Get handles GET /api/v1/databases/{id}.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET")
		return
	}
	userID := h.userID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}
	id := extractID(r.URL.Path, "/api/v1/databases/")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "invalid_path", "missing database id")
		return
	}
	inst, err := h.svc.Get(userID, id)
	if err != nil {
		api.WriteError(w, http.StatusNotFound, "not_found", "database not found")
		return
	}
	api.WriteJSON(w, http.StatusOK, inst)
}

// Deploy handles POST /api/v1/databases/{id}/deploy.
func (h *Handler) Deploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}
	userID := h.userID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}
	id := extractID(r.URL.Path, "/api/v1/databases/")
	id = strings.TrimSuffix(id, "/deploy")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "invalid_path", "missing database id")
		return
	}
	inst, err := h.svc.Deploy(userID, id)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "deploy_failed", err.Error())
		return
	}
	api.WriteJSON(w, http.StatusOK, inst)
}

// Stop handles POST /api/v1/databases/{id}/stop.
func (h *Handler) Stop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}
	userID := h.userID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}
	id := extractID(r.URL.Path, "/api/v1/databases/")
	id = strings.TrimSuffix(id, "/stop")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "invalid_path", "missing database id")
		return
	}
	inst, err := h.svc.Stop(userID, id)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "stop_failed", err.Error())
		return
	}
	api.WriteJSON(w, http.StatusOK, inst)
}

// Restart handles POST /api/v1/databases/{id}/restart.
func (h *Handler) Restart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}
	userID := h.userID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}
	id := extractID(r.URL.Path, "/api/v1/databases/")
	id = strings.TrimSuffix(id, "/restart")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "invalid_path", "missing database id")
		return
	}
	inst, err := h.svc.Restart(userID, id)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "restart_failed", err.Error())
		return
	}
	api.WriteJSON(w, http.StatusOK, inst)
}

// Status handles GET /api/v1/databases/{id}/status.
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET")
		return
	}
	userID := h.userID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}
	id := extractID(r.URL.Path, "/api/v1/databases/")
	id = strings.TrimSuffix(id, "/status")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "invalid_path", "missing database id")
		return
	}
	inst, err := h.svc.Status(userID, id)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "status_failed", err.Error())
		return
	}
	api.WriteJSON(w, http.StatusOK, inst)
}

// Delete handles DELETE /api/v1/databases/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use DELETE")
		return
	}
	userID := h.userID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}
	id := extractID(r.URL.Path, "/api/v1/databases/")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "invalid_path", "missing database id")
		return
	}
	if err := h.svc.Delete(userID, id); err != nil {
		api.WriteError(w, http.StatusNotFound, "not_found", "database not found")
		return
	}
	api.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
