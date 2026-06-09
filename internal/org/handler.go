package org

import (
	"encoding/json"
	"net/http"

	"github.com/razad/razad/internal/api"
	"github.com/razad/razad/internal/auth"
)

// Handler exposes org-related HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates an org HTTP handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Create handles POST /api/v1/orgs.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}

	userID := auth.GetUserID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	org, err := h.svc.Create(req.Name, req.Slug, userID)
	if err != nil {
		code := "create_failed"
		msg := err.Error()
		switch err {
		case ErrInvalidSlug:
			code = "validation_error"
		case ErrSlugTaken:
			code = "slug_taken"
		}
		api.WriteError(w, http.StatusConflict, code, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(org)
}

// List handles GET /api/v1/orgs.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET")
		return
	}

	userID := auth.GetUserID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	orgs, err := h.svc.List(userID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "list_failed", "could not list organizations")
		return
	}

	if orgs == nil {
		orgs = []Organization{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgs)
}
