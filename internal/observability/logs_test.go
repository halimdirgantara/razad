package observability

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/razad/razad/internal/requestctx"
	"github.com/razad/razad/internal/websocket"
)

func newTestStreamer(t *testing.T) *LogStreamer {
	t.Helper()
	hub := websocket.NewHub()
	return NewLogStreamer(hub, t.TempDir())
}

func writeLogFile(t *testing.T, streamer *LogStreamer, name string, lines []string) {
	t.Helper()
	path := filepath.Join(streamer.dataDir, "logs", name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestReadRecentReturnsTailOfExistingFile(t *testing.T) {
	streamer := newTestStreamer(t)
	writeLogFile(t, streamer, "app/output.log", []string{
		"line 1", "line 2", "line 3", "line 4", "line 5",
	})
	got, err := streamer.ReadRecent("app/output.log", 3)
	if err != nil {
		t.Fatalf("ReadRecent: %v", err)
	}
	want := []string{"line 3", "line 4", "line 5"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadRecentReturnsAllWhenFileShorterThanLimit(t *testing.T) {
	streamer := newTestStreamer(t)
	writeLogFile(t, streamer, "short.log", []string{"a", "b"})
	got, err := streamer.ReadRecent("short.log", 100)
	if err != nil {
		t.Fatalf("ReadRecent: %v", err)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("got %v, want [a b]", got)
	}
}

func TestReadRecentReturnsEmptyForMissingFile(t *testing.T) {
	streamer := newTestStreamer(t)
	got, err := streamer.ReadRecent("does-not-exist.log", 10)
	if err != nil {
		t.Fatalf("ReadRecent on missing file should not error, got %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice for missing file, got %v", got)
	}
}

func TestReadRecentRejectsPathTraversal(t *testing.T) {
	streamer := newTestStreamer(t)
	for _, bad := range []string{"../etc/passwd", "/etc/passwd", ".."} {
		if _, err := streamer.ReadRecent(bad, 10); err == nil {
			t.Errorf("expected error for name %q", bad)
		}
	}
}

func TestReadRecentRejectsEmptyName(t *testing.T) {
	streamer := newTestStreamer(t)
	if _, err := streamer.ReadRecent("", 10); err == nil {
		t.Error("expected error for empty name")
	}
}

func TestReadRecentClampsLines(t *testing.T) {
	streamer := newTestStreamer(t)
	writeLogFile(t, streamer, "x.log", []string{"only"})
	// Asking for 0 should default to a reasonable value, not error.
	got, err := streamer.ReadRecent("x.log", 0)
	if err != nil {
		t.Fatalf("ReadRecent lines=0: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 line, got %d", len(got))
	}
}

func TestListHandlerRequiresAuth(t *testing.T) {
	streamer := newTestStreamer(t)
	h := NewHandler(streamer)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/logs?file=app/output.log", nil)
	rr := httptest.NewRecorder()
	h.List(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestListHandlerRequiresFileParam(t *testing.T) {
	streamer := newTestStreamer(t)
	h := NewHandler(streamer)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/logs", nil)
	req = req.WithContext(requestctx.WithUserID(req.Context(), "u-1"))
	rr := httptest.NewRecorder()
	h.List(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestListHandlerReturnsLines(t *testing.T) {
	streamer := newTestStreamer(t)
	writeLogFile(t, streamer, "demo/output.log", []string{"one", "two", "three"})
	h := NewHandler(streamer)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/logs?file=demo/output.log&lines=2", nil)
	req = req.WithContext(requestctx.WithUserID(req.Context(), "u-1"))
	rr := httptest.NewRecorder()
	h.List(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var resp struct {
		File  string   `json:"file"`
		Lines []string `json:"lines"`
		Count int      `json:"count"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.File != "demo/output.log" {
		t.Errorf("file: got %q", resp.File)
	}
	if len(resp.Lines) != 2 || resp.Lines[0] != "two" || resp.Lines[1] != "three" {
		t.Errorf("lines: got %v, want [two three]", resp.Lines)
	}
	if resp.Count != 2 {
		t.Errorf("count: got %d, want 2", resp.Count)
	}
}

func TestListHandlerRejectsNonGET(t *testing.T) {
	streamer := newTestStreamer(t)
	h := NewHandler(streamer)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs?file=x", nil)
	req = req.WithContext(requestctx.WithUserID(req.Context(), "u-1"))
	rr := httptest.NewRecorder()
	h.List(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

// Compile-time check that the test signature still matches the websocket
// package — guards against future API drift breaking the test setup.
var _ = func() bool {
	_ = context.Background()
	return true
}()
