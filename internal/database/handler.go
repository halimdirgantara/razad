package database

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/razad/razad/internal/api"
	"github.com/razad/razad/internal/requestctx"
)

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
	svc *Service
}

// NewHandler creates a database HTTP handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List handles GET /api/v1/databases.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET")
		return
	}

	userID := requestctx.UserID(r.Context())
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	instances, err := h.svc.List(userID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "list_failed", "could not list databases")
		return
	}
	if instances == nil {
		instances = []Instance{}
	}
	api.WriteJSON(w, http.StatusOK, instances)
}

// Create handles POST /api/v1/databases.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}

	userID := requestctx.UserID(r.Context())
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

	userID := requestctx.UserID(r.Context())
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

// Delete handles DELETE /api/v1/databases/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use DELETE")
		return
	}

	userID := requestctx.UserID(r.Context())
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
