package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/razad/razad/internal/api"
	"github.com/razad/razad/internal/auth"
	"github.com/razad/razad/internal/config"
	"github.com/razad/razad/internal/database"
	"github.com/razad/razad/internal/org"
	"github.com/razad/razad/web"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Set up structured logger
	logLevel := slog.LevelInfo
	if cfg.Debug {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})))

	slog.Info("starting razad daemon", "version", cfg.Version, "mode", cfg.Mode)

	// Initialize database
	db, err := database.Open(cfg.Database)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Run pending migrations
	if cfg.AutoMigrate {
		if err := database.Migrate(db); err != nil {
			slog.Error("failed to run migrations", "error", err)
			os.Exit(1)
		}
	}

	// Set up auth
	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo, cfg.Auth.SessionTTLMinutes)
	authHandler := auth.NewHandler(authSvc)
	authMiddleware := auth.Middleware(authSvc)

	// Set up org
	orgRepo := org.NewRepository(db)
	orgSvc := org.NewService(orgRepo)
	orgHandler := org.NewHandler(orgSvc)

	// Seed the first admin user if no users exist (self-hosted bootstrap)
	seedAdminIfNeeded(authSvc)

	// Set up API router
	router := api.NewRouter()

	// Public routes
	router.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	router.HandleFunc("/api/v1/auth/login", authHandler.Login)
	router.HandleFunc("/api/v1/auth/logout", authHandler.Logout)
	router.HandleFunc("/api/v1/auth/register", authHandler.Register)

	// Protected routes (require auth middleware)
	protected := router.Group(authMiddleware)
	protected.HandleFunc("/api/v1/auth/me", authHandler.Me)
	protected.Handle("/api/v1/orgs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			orgHandler.List(w, r)
		case http.MethodPost:
			orgHandler.Create(w, r)
		default:
			api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		}
	}))

	// Serve embedded UI (SvelteKit static build) with SPA fallback
	uiFS, err := fs.Sub(web.UI, "build")
	if err != nil {
		slog.Warn("no embedded UI found, UI will not be served", "error", err)
	} else {
		router.Handle("/", spaFallbackHandler(uiFS))
	}

	// Start HTTP server
	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
		os.Exit(1)
	}
	slog.Info("shutdown complete")
}

// seedAdminIfNeeded creates a default admin user if no users exist.
func seedAdminIfNeeded(svc *auth.Service) {
	_, err := svc.Register("admin", "admin@razad.local", "razadadmin")
	if err != nil {
		// User already exists — ignore
		return
	}
	slog.Info("created default admin user (admin@razad.local / razadadmin)")
}

// spaFallbackHandler serves static files with SPA fallback to index.html
// for client-side routing (SvelteKit). API paths are not handled here.
func spaFallbackHandler(uiFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(uiFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Let API routes pass through to the router
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			http.NotFound(w, r)
			return
		}

		// Check if the file exists in the embedded FS
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		f, err := uiFS.Open(path[1:]) // strip leading slash
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback: serve index.html for all non-file routes
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
