package observability

import (
	"net/http"
	"strconv"

	"github.com/razad/razad/internal/api"
	"github.com/razad/razad/internal/auth"
)

// Handler exposes HTTP endpoints for the observability subsystem.
type Handler struct {
	streamer *LogStreamer
}

// NewHandler creates an observability HTTP handler.
func NewHandler(streamer *LogStreamer) *Handler {
	return &Handler{streamer: streamer}
}

// logsResponse is the shape returned by List.
type logsResponse struct {
	File  string   `json:"file"`
	Lines []string `json:"lines"`
	Count int      `json:"count"`
}

// List handles GET /api/v1/logs?file=NAME&lines=N. Returns the last N
// lines of the named log file under paths.Logs() (dataDir/logs). For an
// unknown file returns an empty lines array rather than 404 — a freshly
// provisioned app has no output yet.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET")
		return
	}
	if auth.GetUserID(r) == "" {
		api.WriteError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
		return
	}
	q := r.URL.Query()
	name := q.Get("file")
	if name == "" {
		api.WriteError(w, http.StatusBadRequest, "missing_param", "file query parameter is required")
		return
	}
	lines, _ := strconv.Atoi(q.Get("lines"))

	entries, err := h.streamer.ReadRecent(name, lines)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "read_failed", err.Error())
		return
	}
	api.WriteJSON(w, http.StatusOK, logsResponse{
		File:  name,
		Lines: entries,
		Count: len(entries),
	})
}
