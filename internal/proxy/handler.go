package proxy

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/razad/razad/internal/api"
	"github.com/razad/razad/internal/audit"
	"github.com/razad/razad/internal/auth"
	"github.com/razad/razad/internal/policy"
)

// Handler exposes proxy configuration endpoints.
type Handler struct {
	svc     *Service
	auditor *audit.Service
	policy  *policy.Engine
}

// NewHandler creates a proxy HTTP handler.
func NewHandler(svc *Service, auditor *audit.Service, pol *policy.Engine) *Handler {
	return &Handler{svc: svc, auditor: auditor, policy: pol}
}

// gate runs a policy check for a proxy action. Returns true if the request is
// allowed. On deny, it writes the error response and returns false.
func (h *Handler) gate(w http.ResponseWriter, r *http.Request, action policy.Action, resource policy.Resource) bool {
	actor := auth.GetActor(r)
	if h.policy == nil {
		return true
	}
	if err := h.policy.MustCheck(r.Context(), policy.Actor{UserID: actor.UserID, IsAdmin: actor.IsAdmin}, action, resource); err != nil {
		api.WriteError(w, http.StatusForbidden, "forbidden", err.Error())
		return false
	}
	return true
}

type bindingRequest struct {
	Name         string `json:"name"`
	Domain       string `json:"domain"`
	UpstreamHost string `json:"upstream_host"`
	UpstreamPort int    `json:"upstream_port"`
	TLS          bool   `json:"tls"`
	BodyLimitMB  int    `json:"body_limit_mb"`
}

// Render handles POST /api/v1/proxy/render.
func (h *Handler) Render(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}
	userID := auth.GetUserID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	var req bindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if !h.gate(w, r, policy.ActionProxyRender, policy.Resource{Type: "proxy", ID: req.Name}) {
		return
	}

	binding := Binding{
		Name:         req.Name,
		Domain:       req.Domain,
		UpstreamHost: req.UpstreamHost,
		UpstreamPort: req.UpstreamPort,
		TLS:          req.TLS,
		BodyLimitMB:  req.BodyLimitMB,
	}
	cfg, err := h.svc.Render(binding)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "render_failed", err.Error())
		return
	}
	if h.auditor != nil {
		_ = h.auditor.Record(r.Context(), userID, "proxy.render", "proxy", req.Name, map[string]any{"domain": req.Domain, "tls": req.TLS})
	}
	api.WriteJSON(w, http.StatusOK, map[string]any{"config": cfg})
}

// Apply handles POST /api/v1/proxy/apply.
func (h *Handler) Apply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}
	userID := auth.GetUserID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	var req bindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if !h.gate(w, r, policy.ActionProxyApply, policy.Resource{Type: "proxy", ID: req.Name}) {
		return
	}

	binding := Binding{
		Name:         req.Name,
		Domain:       req.Domain,
		UpstreamHost: req.UpstreamHost,
		UpstreamPort: req.UpstreamPort,
		TLS:          req.TLS,
		BodyLimitMB:  req.BodyLimitMB,
	}
	if err := h.svc.Apply(binding); err != nil {
		api.WriteError(w, http.StatusBadRequest, "apply_failed", err.Error())
		return
	}
	candidate, _ := h.svc.CandidatePath(binding)
	enabled, _ := h.svc.EnabledPath(binding)
	backup, _ := h.svc.BackupPath(binding)
	if h.auditor != nil {
		_ = h.auditor.Record(r.Context(), userID, "proxy.apply", "proxy", req.Name, map[string]any{"domain": req.Domain, "candidate": candidate, "enabled": enabled})
	}
	api.WriteJSON(w, http.StatusOK, map[string]any{
		"status":    "applied",
		"candidate": candidate,
		"enabled":   enabled,
		"backup":    backup,
	})
}

// Rollback handles POST /api/v1/proxy/rollback.
func (h *Handler) Rollback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}
	userID := auth.GetUserID(r)
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}

	var req bindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if !h.gate(w, r, policy.ActionProxyRollback, policy.Resource{Type: "proxy", ID: req.Name}) {
		return
	}

	binding := Binding{
		Name:         req.Name,
		Domain:       req.Domain,
		UpstreamHost: req.UpstreamHost,
		UpstreamPort: req.UpstreamPort,
		TLS:          req.TLS,
		BodyLimitMB:  req.BodyLimitMB,
	}
	if err := h.svc.Rollback(binding); err != nil {
		api.WriteError(w, http.StatusBadRequest, "rollback_failed", err.Error())
		return
	}
	if h.auditor != nil {
		_ = h.auditor.Record(r.Context(), userID, "proxy.rollback", "proxy", req.Name, map[string]any{"domain": req.Domain})
	}
	api.WriteJSON(w, http.StatusOK, map[string]any{"status": "rolled_back"})
}

// helper for compatibility in tests/clients that want a stringified port.
func stringifyPort(port int) string { return strconv.Itoa(port) }
