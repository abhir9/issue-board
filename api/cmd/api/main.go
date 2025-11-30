package main

// go run github.com/swaggo/swag/cmd/swag init -g cmd/api/main.go --parseDependency --parseInternal
import (
	"fmt"
	"issue-board-backend/internal/database"
	"issue-board-backend/internal/handlers"
	customMiddleware "issue-board-backend/internal/middleware"
	"log"
	"net/http"
	"os"

	_ "issue-board-backend/docs"

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
	dbPath := "./issues.db"
	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	defer database.DB.Close()

	if err := database.RunMigrations("./migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	repo := database.NewRepository(database.DB)
	h := handlers.NewHandler(repo)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS setup
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*", "X-API-Key"},
		AllowCredentials: true,
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(customMiddleware.APIKeyAuth) // Apply Auth middleware to /api routes

		r.Get("/issues", h.GetIssues)
		r.Post("/issues", h.CreateIssue)
		r.Get("/issues/{id}", h.GetIssue)
		r.Patch("/issues/{id}", h.UpdateIssue)
		r.Patch("/issues/{id}/move", h.MoveIssue)
		r.Delete("/issues/{id}", h.DeleteIssue)

		r.Get("/users", h.GetUsers)
		r.Get("/labels", h.GetLabels)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server running on port %s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
