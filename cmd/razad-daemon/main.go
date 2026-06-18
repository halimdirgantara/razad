package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/razad/razad/internal/api"
	"github.com/razad/razad/internal/ai"
	"github.com/razad/razad/internal/app"
	"github.com/razad/razad/internal/audit"
	"github.com/razad/razad/internal/auth"
	"github.com/razad/razad/internal/config"
	"github.com/razad/razad/internal/crypto"
	"github.com/razad/razad/internal/database"
	"github.com/razad/razad/internal/org"
	"github.com/razad/razad/internal/process"
	"github.com/razad/razad/internal/proxy"
	"github.com/razad/razad/internal/server"
	"github.com/razad/razad/internal/ssl"
	"github.com/razad/razad/internal/observability"
	websocketpkg "github.com/razad/razad/internal/websocket"
	"github.com/razad/razad/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logLevel := slog.LevelInfo
	if cfg.Debug {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})))

	slog.Info("starting razad daemon", "version", cfg.Version, "mode", cfg.Mode)

	db, err := database.Open(cfg.Database)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if cfg.AutoMigrate {
		if err := database.Migrate(db); err != nil {
			slog.Error("failed to run migrations", "error", err)
			os.Exit(1)
		}
	}

	// --- Audit ---
	auditSvc := audit.NewService(db)

	// --- Auth ---
	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo, cfg.Auth.SessionTTLMinutes)
	authHandler := auth.NewHandler(authSvc)
	authMiddleware := auth.Middleware(authSvc)

	// --- Org ---
	orgRepo := org.NewRepository(db)
	orgSvc := org.NewService(orgRepo)
	orgSvc.SetAuditor(auditSvc)
	orgHandler := org.NewHandler(orgSvc)

	// --- Crypto ---
	enc, err := crypto.New(cfg.Auth.SecretKey)
	if err != nil {
		slog.Error("failed to initialize crypto", "error", err)
		os.Exit(1)
	}

	// --- Process Runner ---
	procRunnerCfg := process.Config{
		Runner:  "local",
		DataDir: cfg.DataDir,
	}
	if v := os.Getenv("RAZAD_PROCESS_RUNNER"); v == "systemd" {
		procRunnerCfg.Runner = "systemd"
	}
	procRunner := process.New(procRunnerCfg)

	// --- Health Collector ---
	healthCollector := server.NewCollector(cfg.DataDir)

	// --- WebSocket Hub ---
	wsHub := websocketpkg.NewHub()
	logStreamer := observability.NewLogStreamer(wsHub, cfg.DataDir)

	// --- App ---
	appRepo := app.NewRepository(db)
	appSvc := app.NewService(appRepo, procRunner, enc, cfg.DataDir)
	appSvc.SetAuditor(auditSvc)
	appSvc.SetLogStreamer(logStreamer)
	appHandler := app.NewHandler(appSvc)

	// --- Database ---
	dbRepo := database.NewRepository(db)
	dbSvc := database.NewService(dbRepo)
	dbHandler := database.NewHandler(dbSvc)

	// --- AI ---
	aiSvc := ai.NewService(auditSvc)
	aiHandler := ai.NewHandler(aiSvc)

	// --- Proxy ---
	proxySvc := proxy.NewService(filepath.Join(cfg.DataDir, "nginx"))
	proxyHandler := proxy.NewHandler(proxySvc, auditSvc)

	// --- SSL ---
	sslSvc := ssl.NewService(filepath.Join(cfg.DataDir, "letsencrypt"))
	sslHandler := ssl.NewHandler(sslSvc, auditSvc)

	// --- Seed admin ---
	seedAdminIfNeeded(authSvc)

	// --- Router ---
	router := api.NewRouter()

	router.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	router.HandleFunc("/api/v1/auth/login", authHandler.Login)
	router.HandleFunc("/api/v1/auth/logout", authHandler.Logout)
	router.HandleFunc("/api/v1/auth/register", authHandler.Register)

	protected := router.Group(authMiddleware)
	protected.HandleFunc("/api/v1/health/stats", func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, healthCollector.Collect())
	})
	protected.HandleFunc("/ws", wsHub.HandleConnection)
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

	// --- App Routes ---
	protected.Handle("/api/v1/apps", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			appHandler.ListAll(w, r)
		case http.MethodPost:
			appHandler.Create(w, r)
		default:
			api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		}
	}))

	protected.Handle("/api/v1/apps/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route based on path suffix
		path := r.URL.Path

		switch {
		case hasSuffix(path, "/deployments") && r.Method == http.MethodGet:
			appHandler.Deployments(w, r)
		case hasSuffix(path, "/deploy") && r.Method == http.MethodPost:
			appHandler.Deploy(w, r)
		case hasSuffix(path, "/stop") && r.Method == http.MethodPost:
			appHandler.Stop(w, r)
		case hasSuffix(path, "/restart") && r.Method == http.MethodPost:
			appHandler.Restart(w, r)
		case hasSuffix(path, "/env"):
			appHandler.Env(w, r)
		case r.Method == http.MethodGet:
			appHandler.Get(w, r)
		case r.Method == http.MethodDelete:
			appHandler.Delete(w, r)
		default:
			api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		}
	}))

	protected.Handle("/api/v1/projects/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hasSuffix(r.URL.Path, "/apps") && r.Method == http.MethodGet {
			appHandler.List(w, r)
		} else {
			api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		}
	}))

	protected.Handle("/api/v1/databases", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			dbHandler.List(w, r)
		case http.MethodPost:
			dbHandler.Create(w, r)
		default:
			api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		}
	}))

	protected.Handle("/api/v1/databases/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			dbHandler.Get(w, r)
		case http.MethodDelete:
			dbHandler.Delete(w, r)
		default:
			api.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		}
	}))

	protected.HandleFunc("/api/v1/ai", aiHandler.Index)
	protected.HandleFunc("/api/v1/ai/actions", aiHandler.Action)

	protected.HandleFunc("/api/v1/proxy/render", proxyHandler.Render)
	protected.HandleFunc("/api/v1/proxy/apply", proxyHandler.Apply)
	protected.HandleFunc("/api/v1/proxy/rollback", proxyHandler.Rollback)
	protected.HandleFunc("/api/v1/ssl/issue", sslHandler.Issue)
	protected.HandleFunc("/api/v1/ssl/renew", sslHandler.Renew)

	// --- UI ---
	router.Handle("/", spaFallbackHandler(web.UI))

	// --- Server ---
	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

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

func seedAdminIfNeeded(svc *auth.Service) {
	_, err := svc.Register("admin", "admin@razad.local", "razadadmin")
	if err != nil {
		return
	}
	slog.Info("created default admin user (admin@razad.local / razadadmin)")
}

func hasSuffix(path, suffix string) bool {
	n := len(path) - len(suffix)
	return n >= 0 && path[n:] == suffix
}

func spaFallbackHandler(embedFS fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// API routes
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			http.NotFound(w, r)
			return
		}

		// Build path within the embed: "build" + request path
		reqPath := r.URL.Path
		if reqPath == "/" {
			reqPath = "/index.html"
		}
		embedPath := "build" + reqPath

		// Try to open the file directly in the embedded FS
		f, err := embedFS.Open(embedPath)
		if err == nil {
			f.Close()
			// Serve with correct content type
			data, readErr := fs.ReadFile(embedFS, embedPath)
			if readErr == nil {
				ctype := mimeType(embedPath)
				w.Header().Set("Content-Type", ctype)
				w.Header().Set("Cache-Control", "public, max-age=3600")
				w.Write(data)
				return
			}
		}

		// SPA fallback: serve index.html
		data, fallbackErr := fs.ReadFile(embedFS, "build/index.html")
		if fallbackErr != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(data)
	})
}

func mimeType(path string) string {
	switch {
	case hasSuffix(path, ".html"):
		return "text/html; charset=utf-8"
	case hasSuffix(path, ".js"):
		return "application/javascript"
	case hasSuffix(path, ".css"):
		return "text/css"
	case hasSuffix(path, ".json"):
		return "application/json"
	case hasSuffix(path, ".png"):
		return "image/png"
	case hasSuffix(path, ".svg"):
		return "image/svg+xml"
	case hasSuffix(path, ".woff2"):
		return "font/woff2"
	case hasSuffix(path, ".woff"):
		return "font/woff"
	default:
		return "application/octet-stream"
	}
}
