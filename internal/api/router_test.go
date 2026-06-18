package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)

	if body["status"] != "ok" {
		t.Errorf("expected status 'ok', got %s", body["status"])
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, http.StatusNotFound, "not_found", "resource not found")

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)

	errObj, ok := body["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object")
	}
	if errObj["code"] != "not_found" {
		t.Errorf("expected code 'not_found', got %s", errObj["code"])
	}
	if errObj["message"] != "resource not found" {
		t.Errorf("expected message 'resource not found', got %s", errObj["message"])
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/panic")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 after panic, got %d", resp.StatusCode)
	}
}

func TestWebSocketRoutePreservesHijacker(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(http.Hijacker); !ok {
			t.Fatalf("expected websocket route writer to implement http.Hijacker")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/ws", nil)
	if err != nil {
		t.Fatalf("request build failed: %v", err)
	}
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected websocket-capable response, got %d", resp.StatusCode)
	}
}

func TestRouteGroup(t *testing.T) {
	router := NewRouter()

	// Auth middleware that sets a header
	authMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Auth", "passed")
			next.ServeHTTP(w, r)
		})
	}

	group := router.Group(authMW)
	group.HandleFunc("/api/v1/protected", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/protected")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("X-Auth") != "passed" {
		t.Error("expected X-Auth header from group middleware")
	}
}
