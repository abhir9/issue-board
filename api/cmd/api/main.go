package main

// go run github.com/swaggo/swag/cmd/swag init -g cmd/api/main.go --parseDependency --parseInternal
import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/abhir9/issue-board/api/docs"
	"github.com/abhir9/issue-board/api/internal/config"
	"github.com/abhir9/issue-board/api/internal/database"
	"github.com/abhir9/issue-board/api/internal/handlers"
	customMiddleware "github.com/abhir9/issue-board/api/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title Issue Board API
// @version 1.0
// @description This is a simple issue board API.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /api
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
func main() {
	// Setup structured logging
	logger := setupLogger()
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting Issue Board API", "version", "1.0.0")

	// Setup database
	if err := setupDatabase(cfg); err != nil {
		slog.Error("Failed to setup database", "error", err)
		os.Exit(1)
	}
	defer database.DB.Close()

	// Setup router
	r := setupRouter(cfg)

	// Create and start server
	server := setupServer(cfg, r)
	startServer(server, cfg)
}

// keepAlive pings the health endpoint every 5 minutes to prevent Render free tier sleep
func keepAlive(baseURL string) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Initial delay to allow server to start
	time.Sleep(30 * time.Second)

	for range ticker.C {
		healthURL := baseURL + "/api/health"
		resp, err := client.Get(healthURL)
		if err != nil {
			slog.Warn("Keep-alive ping failed", "url", healthURL, "error", err)
			continue
		}
		resp.Body.Close()
		slog.Debug("Keep-alive ping successful", "url", healthURL, "status", resp.StatusCode)
	}
}

func setupLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func setupDatabase(cfg *config.Config) error {
	// Initialize database with connection pool settings
	if err := database.InitDB(cfg.Database.Path); err != nil {
		return err
	}

	// Configure connection pool
	database.DB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	database.DB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	database.DB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Run migrations
	return database.RunMigrations(cfg.Database.MigrationDir)
}

func setupRouter(cfg *config.Config) *chi.Mux {
	// Setup repository and handlers
	repo := database.NewRepository(database.DB)
	h := handlers.NewHandler(repo)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS setup with improved configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.Server.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(c.Handler)

	// Redirect /docs to /docs/index.html
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusMovedPermanently)
	})
	r.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
	))

	// Health check endpoint (no auth required)
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		// Check database connection
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status := "ok"
		dbStatus := "healthy"
		statusCode := http.StatusOK

		if err := database.DB.PingContext(ctx); err != nil {
			dbStatus = "unhealthy"
			status = "degraded"
			statusCode = http.StatusServiceUnavailable
			slog.Warn("Health check: database ping failed", "error", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		response := map[string]string{
			"status":   status,
			"database": dbStatus,
		}
		if err := handlers.WriteJSON(w, response); err != nil {
			slog.Error("Failed to write health check response", "error", err)
		}
	})

	// API routes with authentication
	r.Route("/api", func(r chi.Router) {
		r.Use(customMiddleware.APIKeyAuth(cfg.Auth.APIKey)) // Apply Auth middleware to /api routes

		r.Get("/issues", h.GetIssues)
		r.Post("/issues", h.CreateIssue)
		r.Get("/issues/{id}", h.GetIssue)
		r.Patch("/issues/{id}", h.UpdateIssue)
		r.Patch("/issues/{id}/move", h.MoveIssue)
		r.Delete("/issues/{id}", h.DeleteIssue)

		r.Get("/users", h.GetUsers)
		r.Get("/labels", h.GetLabels)
	})

	return r
}

func setupServer(cfg *config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}
}

func startServer(server *http.Server, cfg *config.Config) {
	// Start keep-alive pinger if enabled
	if cfg.Server.EnableKeepAlive && cfg.Server.KeepAliveURL != "" {
		go keepAlive(cfg.Server.KeepAliveURL)
		slog.Info("Keep-alive pinger started", "target", cfg.Server.KeepAliveURL+"/api/health")
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("Server starting", "port", cfg.Server.Port, "host", cfg.Server.Host)
		serverErrors <- server.ListenAndServe()
	}()

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		slog.Error("Server error", "error", err)
		os.Exit(1)

	case sig := <-shutdown:
		slog.Info("Shutdown signal received", "signal", sig.String())

		// Create context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Graceful shutdown failed, forcing shutdown", "error", err)
			if err := server.Close(); err != nil {
				slog.Error("Failed to close server", "error", err)
			}
			os.Exit(1)
		}

		slog.Info("Server shutdown complete")
	}
}
