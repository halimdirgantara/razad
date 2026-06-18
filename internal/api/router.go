// Package api provides a structured HTTP router with middleware chaining,
// consistent error envelopes, and request logging for the Razad daemon.
package api

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Router is a lightweight HTTP router with middleware support.
type Router struct {
	mux         *http.ServeMux
	middlewares []Middleware
}

// NewRouter creates a new API router.
func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
		middlewares: []Middleware{
			RecoveryMiddleware,
			RequestLogMiddleware,
		},
	}
}

// Use adds a global middleware to the router.
func (r *Router) Use(mw Middleware) {
	r.middlewares = append(r.middlewares, mw)
}

// Handle registers a handler at the given path with global middlewares.
func (r *Router) Handle(path string, handler http.Handler) {
	// Wrap with middlewares (innermost first, applied outermost first)
	wrapped := handler
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		wrapped = r.middlewares[i](wrapped)
	}
	r.mux.Handle(path, wrapped)
}

// HandleFunc registers a handler function at the given path.
func (r *Router) HandleFunc(path string, handler http.HandlerFunc) {
	r.Handle(path, handler)
}

// Group creates a route group with additional middleware.
func (r *Router) Group(mw Middleware) *RouteGroup {
	return &RouteGroup{
		parent: r,
		mw:     mw,
	}
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// RouteGroup is a set of routes that share additional middleware.
type RouteGroup struct {
	parent *Router
	mw     Middleware
}

// Handle registers a handler in the group with the group's additional middleware.
func (g *RouteGroup) Handle(path string, handler http.Handler) {
	// Wrap with group middleware first, then global middleware
	wrapped := handler
	wrapped = g.mw(wrapped)

	for i := len(g.parent.middlewares) - 1; i >= 0; i-- {
		wrapped = g.parent.middlewares[i](wrapped)
	}
	g.parent.mux.Handle(path, wrapped)
}

// HandleFunc registers a handler function in the group.
func (g *RouteGroup) HandleFunc(path string, handler http.HandlerFunc) {
	g.Handle(path, handler)
}

// ---------------------------------------------------------------------------
// Middleware Implementations
// ---------------------------------------------------------------------------

// RecoveryMiddleware recovers from panics and returns a 500 error.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered",
					"path", r.URL.Path,
					"error", rec,
				)
				WriteError(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequestLogMiddleware logs each HTTP request with duration.
func RequestLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" || r.Header.Get("Upgrade") == "websocket" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		// Wrap response writer to capture status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)

		slog.Debug("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", lrw.statusCode,
			"duration", time.Since(start).String(),
		)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture the status code.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := lrw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hj.Hijack()
}

func (lrw *loggingResponseWriter) Flush() {
	if f, ok := lrw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (lrw *loggingResponseWriter) CloseNotify() <-chan bool {
	if cn, ok := lrw.ResponseWriter.(http.CloseNotifier); ok {
		return cn.CloseNotify()
	}
	ch := make(chan bool, 1)
	return ch
}

func (lrw *loggingResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := lrw.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

// ---------------------------------------------------------------------------
// Response Helpers
// ---------------------------------------------------------------------------

// ErrorResponse is the standard API error envelope.
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// WriteError sends a JSON error response with the given status and message.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := ErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = message

	json.NewEncoder(w).Encode(resp)
}

// WriteJSON sends a JSON success response.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
