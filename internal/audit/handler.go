package audit

import (
	"net/http"
	"strconv"

	"github.com/razad/razad/internal/api"
	"github.com/razad/razad/internal/auth"
)

// Handler exposes HTTP endpoints for reading the audit log.
type Handler struct {
	svc *Service
}

// NewHandler creates an audit HTTP handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// auditResponse wraps a page of events with pagination metadata so clients
// can render "page N of M" without doing an extra COUNT round-trip.
type auditResponse struct {
	Items  []Event `json:"items"`
	Total  int     `json:"total"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
}

// List handles GET /api/v1/audit?limit=N&offset=N.
// Defaults: limit=50, offset=0. Caps: limit<=200.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET")
		return
	}
	if auth.GetUserID(r) == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	events, total, err := h.svc.ListPage(limit, offset)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "list_failed", "could not list audit events")
		return
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	api.WriteJSON(w, http.StatusOK, auditResponse{
		Items:  events,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}
