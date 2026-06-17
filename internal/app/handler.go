package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/razad/razad/internal/api"
	"github.com/razad/razad/internal/auth"
	"github.com/razad/razad/internal/domain"
)

// Handler exposes app-related HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates an app HTTP handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListAll handles GET /api/v1/apps.
func (h *Handler) ListAll(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	apps, err := h.svc.ListAll(userID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "list_failed", "could not list apps")
		return
	}

	if apps == nil {
		apps = []domain.App{}
	}

	api.WriteJSON(w, http.StatusOK, apps)
}

// Create handles POST /api/v1/apps.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	var req CreateAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	app, err := h.svc.Create(userID, req)
	if err != nil {
		api.WriteError(w, http.StatusForbidden, "create_failed", err.Error())
		return
	}

	api.WriteJSON(w, http.StatusCreated, app)
}

// Get handles GET /api/v1/apps/{id}.
// Supports path matching by checking the last path segment.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	id := extractID(r.URL.Path, "/api/v1/apps/")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "invalid_path", "missing app id")
		return
	}

	app, err := h.svc.Get(userID, id)
	if err != nil {
		api.WriteError(w, http.StatusNotFound, "not_found", "app not found")
		return
	}

	api.WriteJSON(w, http.StatusOK, app)
}

// List handles GET /api/v1/projects/{projectId}/apps.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	projectID := extractProjectID(r.URL.Path)
	if projectID == "" {
		api.WriteError(w, http.StatusBadRequest, "invalid_path", "missing project id")
		return
	}

	apps, err := h.svc.ListByProject(userID, projectID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "list_failed", "could not list apps")
		return
	}

	if apps == nil {
		apps = []domain.App{}
	}

	api.WriteJSON(w, http.StatusOK, apps)
}

// Deploy handles POST /api/v1/apps/{id}/deploy.
func (h *Handler) Deploy(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	id := extractID(r.URL.Path, "/api/v1/apps/")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "invalid_path", "missing app id")
		return
	}

	app, err := h.svc.Deploy(r.Context(), userID, id)
	if err != nil {
		api.WriteError(w, http.StatusForbidden, "deploy_failed", err.Error())
		return
	}

	api.WriteJSON(w, http.StatusOK, app)
}

// Stop handles POST /api/v1/apps/{id}/stop.
func (h *Handler) Stop(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	id := extractID(r.URL.Path, "/api/v1/apps/")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "invalid_path", "missing app id")
		return
	}

	app, err := h.svc.Stop(r.Context(), userID, id)
	if err != nil {
		api.WriteError(w, http.StatusForbidden, "stop_failed", err.Error())
		return
	}

	api.WriteJSON(w, http.StatusOK, app)
}

// Restart handles POST /api/v1/apps/{id}/restart.
func (h *Handler) Restart(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	id := extractID(r.URL.Path, "/api/v1/apps/")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "invalid_path", "missing app id")
		return
	}

	app, err := h.svc.Restart(r.Context(), userID, id)
	if err != nil {
		api.WriteError(w, http.StatusForbidden, "restart_failed", err.Error())
		return
	}

	api.WriteJSON(w, http.StatusOK, app)
}

// Delete handles DELETE /api/v1/apps/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	id := extractID(r.URL.Path, "/api/v1/apps/")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "invalid_path", "missing app id")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		api.WriteError(w, http.StatusForbidden, "delete_failed", err.Error())
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Env handles GET/PUT /api/v1/apps/{id}/env.
func (h *Handler) Env(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	id := extractID(r.URL.Path, "/api/v1/apps/")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "invalid_path", "missing app id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		vars, err := h.svc.GetEnvVarKeys(userID, id)
		if err != nil {
			api.WriteError(w, http.StatusForbidden, "env_failed", err.Error())
			return
		}
		api.WriteJSON(w, http.StatusOK, vars)

	case http.MethodPut:
		var inputs []EnvVarInput
		if err := json.NewDecoder(r.Body).Decode(&inputs); err != nil {
			api.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
			return
		}
		if err := h.svc.SetEnvVars(userID, id, inputs); err != nil {
			api.WriteError(w, http.StatusForbidden, "env_failed", err.Error())
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]string{"status": "saved"})

	default:
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET or PUT")
	}
}

func extractID(path, prefix string) string {
	rest := strings.TrimPrefix(path, prefix)
	if rest == path || rest == "" {
		return ""
	}
	parts := strings.Split(rest, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func extractProjectID(path string) string {
	// Extract project ID from /api/v1/projects/{id}/apps
	parts := strings.Split(path, "/")
	if len(parts) < 6 {
		return ""
	}
	// /api/v1/projects/{id}/apps
	if parts[1] == "api" && parts[2] == "v1" && parts[3] == "projects" {
		return parts[4]
	}
	return ""
}
