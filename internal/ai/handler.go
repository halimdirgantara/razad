package ai

import (
	"encoding/json"
	"net/http"

	"github.com/razad/razad/internal/api"
	"github.com/razad/razad/internal/requestctx"
)

// Handler exposes AI orchestration endpoints.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	api.WriteJSON(w, http.StatusOK, h.svc.Capabilities())
}

func (h *Handler) Action(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	userID := requestctx.UserID(r.Context())
	if userID == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}
	var req ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	res, err := h.svc.RequestAction(r.Context(), userID, req)
	if err != nil {
		switch err {
		case ErrBlockedAction:
			api.WriteError(w, http.StatusForbidden, "action_blocked", "action is blocked by policy")
		case ErrUnknownAction:
			api.WriteError(w, http.StatusBadRequest, "unknown_action", "action is not in the approved registry")
		default:
			api.WriteError(w, http.StatusInternalServerError, "ai_request_failed", err.Error())
		}
		return
	}
	api.WriteJSON(w, http.StatusAccepted, res)
}
