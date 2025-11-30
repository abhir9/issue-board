package handlers

import (
	"api/internal/database"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates an in-memory SQLite database and runs migrations
func setupTestDB(t *testing.T) *database.Repository {
	// Enable parseTime=true and shared cache for in-memory DB
	// Use sanitized t.Name() to ensure unique DB per test and valid filename
	dbName := strings.ReplaceAll(t.Name(), "/", "_")
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&parseTime=true", dbName)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatalf("Failed to open in-memory db: %v", err)
	}

	// Create tables manually or read migration files.
	// For testing, it's often easier to just execute the schema directly if it's small,
	// or ensure the migration path is correct.
	// Here we'll execute the schema directly to avoid path issues during test execution.
	schema := `
	CREATE TABLE users (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		avatar_url TEXT
	);

	CREATE TABLE labels (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		color TEXT NOT NULL
	);

	CREATE TABLE issues (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT NOT NULL,
		priority TEXT NOT NULL,
		assignee_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		order_index REAL NOT NULL DEFAULT 0,
		FOREIGN KEY (assignee_id) REFERENCES users(id)
	);

	CREATE TABLE issue_labels (
		issue_id TEXT NOT NULL,
		label_id TEXT NOT NULL,
		PRIMARY KEY (issue_id, label_id),
		FOREIGN KEY (issue_id) REFERENCES issues(id) ON DELETE CASCADE,
		FOREIGN KEY (label_id) REFERENCES labels(id) ON DELETE CASCADE
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return database.NewRepository(db)
}

func setupRouter(repo *database.Repository) *chi.Mux {
	h := NewHandler(repo)
	r := chi.NewRouter()
	r.Get("/issues", h.GetIssues)
	r.Post("/issues", h.CreateIssue)
	r.Get("/issues/{id}", h.GetIssue)
	r.Patch("/issues/{id}", h.UpdateIssue)
	r.Patch("/issues/{id}/move", h.MoveIssue)
	r.Delete("/issues/{id}", h.DeleteIssue)
	r.Get("/users", h.GetUsers)
	r.Get("/labels", h.GetLabels)
	return r
}
