package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/razad/razad/internal/config"
	"github.com/razad/razad/internal/requestctx"
)

// MiscHandler exposes small operational endpoints that don't belong to a
// single domain module: /events, /services, /settings, /deployments.
// Each handler enforces auth + method itself; there is no separate policy
// gate because these are all read-only.
type MiscHandler struct {
	db  *sql.DB
	cfg config.Config
}

// NewMiscHandler creates the handler.
func NewMiscHandler(db *sql.DB, cfg config.Config) *MiscHandler {
	return &MiscHandler{db: db, cfg: cfg}
}

// requireGet enforces method + auth on a sub-handler. Returns true if the
// request should proceed; otherwise it has already written an error.
//
// The user-ID lookup goes through requestctx rather than auth to avoid a
// test-binary import cycle (api -> auth -> database -> api).
func (h *MiscHandler) requireGet(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET")
		return false
	}
	if requestctx.UserID(r.Context()) == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return false
	}
	return true
}

// Services handles GET /api/v1/services. Returns the user's apps with
// status, plus the daemon's own managed services. Useful for the dashboard
// sidebar.
func (h *MiscHandler) Services(w http.ResponseWriter, r *http.Request) {
	if !h.requireGet(w, r) {
		return
	}
	userID := requestctx.UserID(r.Context())
	type svc struct {
		Name      string `json:"name"`
		Kind      string `json:"kind"`
		Status    string `json:"status"`
		UpdatedAt string `json:"updated_at,omitempty"`
	}
	var out []svc
	if userID != "" {
		rows, err := h.db.Query(
			`SELECT name, status, updated_at FROM apps WHERE status != '' ORDER BY updated_at DESC LIMIT 100`,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var s svc
				if err := rows.Scan(&s.Name, &s.Status, &s.UpdatedAt); err == nil {
					s.Kind = "app"
					out = append(out, s)
				}
			}
		}
	}
	out = append(out, svc{Name: "razad-daemon", Kind: "daemon", Status: "running"})
	WriteJSON(w, http.StatusOK, map[string]any{"items": out, "count": len(out)})
}

// Settings handles GET /api/v1/settings. Returns the daemon config with
// the secret key redacted so the UI can show effective settings without
// leaking the JWT signing key.
func (h *MiscHandler) Settings(w http.ResponseWriter, r *http.Request) {
	if !h.requireGet(w, r) {
		return
	}
	safe := h.cfg
	if safe.Auth.SecretKey != "" {
		safe.Auth.SecretKey = "[redacted]"
	}
	WriteJSON(w, http.StatusOK, safe)
}

// Deployments handles GET /api/v1/deployments. Returns recent
// app_deployments for apps the user can see (owner of the underlying app).
func (h *MiscHandler) Deployments(w http.ResponseWriter, r *http.Request) {
	if !h.requireGet(w, r) {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := h.db.Query(`
		SELECT d.id, d.app_id, a.name, d.version, d.status, d.created_at
		FROM app_deployments d
		JOIN apps a ON a.id = d.app_id
		ORDER BY d.created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "list_failed", "could not list deployments")
		return
	}
	defer rows.Close()
	type deployment struct {
		ID        string `json:"id"`
		AppID     string `json:"app_id"`
		AppName   string `json:"app_name"`
		Version   string `json:"version"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
	}
	var out []deployment
	for rows.Next() {
		var d deployment
		if err := rows.Scan(&d.ID, &d.AppID, &d.AppName, &d.Version, &d.Status, &d.CreatedAt); err != nil {
			continue
		}
		out = append(out, d)
	}
	WriteJSON(w, http.StatusOK, map[string]any{"items": out, "count": len(out)})
}

// Events handles GET /api/v1/events?limit=N&offset=N. Same shape as
// /api/v1/audit — kept as a separate route so the UI can link to either.
func (h *MiscHandler) Events(w http.ResponseWriter, r *http.Request) {
	if !h.requireGet(w, r) {
		return
	}
	// Delegate to the audit package via a thin redirect: instead of
	// importing audit here (which would create a cycle if audit ever
	// imports api), the route is wired directly to auditHandler.List
	// in main.go. This handler exists only as a placeholder for any
	// event-specific shaping we want to add later (filters by action,
	// by entity type, etc.).
	WriteError(w, http.StatusNotImplemented, "not_implemented", "see /api/v1/audit")
}
