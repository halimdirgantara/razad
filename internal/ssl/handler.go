package ssl

import (
	"encoding/json"
	"net/http"

	"github.com/razad/razad/internal/api"
	"github.com/razad/razad/internal/audit"
	"github.com/razad/razad/internal/auth"
)

// Handler exposes SSL/certbot endpoints.
type Handler struct {
	svc     *Service
	auditor *audit.Service
}

// NewHandler creates an SSL HTTP handler.
func NewHandler(svc *Service, auditor *audit.Service) *Handler {
	return &Handler{svc: svc, auditor: auditor}
}

type issueRequest struct {
	Domain  string `json:"domain"`
	Email   string `json:"email"`
	Webroot string `json:"webroot"`
}

type renewRequest struct {
	Domain string `json:"domain"`
}

// Issue handles POST /api/v1/ssl/issue.
func (h *Handler) Issue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}
	userID := auth.GetUserID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	var req issueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	cmd, err := h.svc.IssueCommand(Request{Domain: req.Domain, Email: req.Email, Webroot: req.Webroot})
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "issue_failed", err.Error())
		return
	}
	certPath, keyPath, _ := h.svc.Paths(req.Domain)
	if h.auditor != nil {
		_ = h.auditor.Record(r.Context(), userID, "ssl.issue", "ssl", req.Domain, map[string]any{"email": req.Email, "webroot": req.Webroot})
	}
	api.WriteJSON(w, http.StatusOK, map[string]any{"command": cmd, "cert_path": certPath, "key_path": keyPath})
}

// Renew handles POST /api/v1/ssl/renew.
func (h *Handler) Renew(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}
	userID := auth.GetUserID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	var req renewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	cmd, err := h.svc.RenewCommand(req.Domain)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "renew_failed", err.Error())
		return
	}
	if h.auditor != nil {
		_ = h.auditor.Record(r.Context(), userID, "ssl.renew", "ssl", req.Domain, nil)
	}
	api.WriteJSON(w, http.StatusOK, map[string]any{"command": cmd})
}
