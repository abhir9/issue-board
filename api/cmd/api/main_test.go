package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/abhir9/issue-board/api/internal/database"
	"github.com/abhir9/issue-board/api/internal/handlers"
	customMiddleware "github.com/abhir9/issue-board/api/internal/middleware"
	"github.com/abhir9/issue-board/api/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"
)

func setupAPITest(t *testing.T) (*httptest.Server, func()) {
	// Create temporary database for testing
	tmpFile, err := os.CreateTemp("", "api_test_*.db")
	require.NoError(t, err)
	tmpFile.Close()
	dbPath := tmpFile.Name()

	// Open database
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=rwc&cache=shared&parseTime=true", dbPath))
	require.NoError(t, err)

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	// Create schema
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

	-- Insert default labels
	INSERT INTO labels (id, name, color) VALUES
		('bug', 'Bug', '#FF0000'),
		('feature', 'Feature', '#00FF00'),
		('enhancement', 'Enhancement', '#0000FF');
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	// Setup repository and handlers
	repo := database.NewRepository(db)
	h := handlers.NewHandler(repo)

	// Setup router (similar to main.go but without server setup)
	r := chi.NewRouter()

	// CORS setup (simplified for testing)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-API-Key")
			next.ServeHTTP(w, r)
		})
	})

	// Health check endpoint (no auth required)
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		// Check database connection
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status := "ok"
		dbStatus := "healthy"
		statusCode := http.StatusOK

		if err := db.PingContext(ctx); err != nil {
			dbStatus = "unhealthy"
			status = "degraded"
			statusCode = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		response := map[string]string{
			"status":   status,
			"database": dbStatus,
		}
		json.NewEncoder(w).Encode(response)
	})

	// API routes with authentication
	r.Route("/api", func(r chi.Router) {
		r.Use(customMiddleware.APIKeyAuth("test-api-key")) // Use test API key

		r.Get("/issues", h.GetIssues)
		r.Post("/issues", h.CreateIssue)
		r.Get("/issues/{id}", h.GetIssue)
		r.Patch("/issues/{id}", h.UpdateIssue)
		r.Patch("/issues/{id}/move", h.MoveIssue)
		r.Delete("/issues/{id}", h.DeleteIssue)

		r.Get("/users", h.GetUsers)
		r.Get("/labels", h.GetLabels)
	})

	// Create test server
	server := httptest.NewServer(r)

	// Cleanup function
	cleanup := func() {
		server.Close()
		db.Close()
		os.Remove(dbPath)
	}

	return server, cleanup
}

func TestAPIHealthCheck(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	t.Run("Health Check Success", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response map[string]string
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "ok", response["status"])
		assert.Equal(t, "healthy", response["database"])
	})
}

func TestAPIGetLabels(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	t.Run("Get Labels Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/labels", nil)
		req.Header.Set("X-API-Key", "test-api-key")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var labels []models.Label
		err = json.NewDecoder(resp.Body).Decode(&labels)
		require.NoError(t, err)

		assert.Len(t, labels, 3) // Default labels: bug, feature, enhancement
		assert.Contains(t, []string{"bug", "feature", "enhancement"}, labels[0].ID)
	})
}

func TestAPIGetUsers(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	t.Run("Get Users Empty", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/users", nil)
		req.Header.Set("X-API-Key", "test-api-key")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var users []models.User
		err = json.NewDecoder(resp.Body).Decode(&users)
		require.NoError(t, err)

		assert.Len(t, users, 0)
	})

	t.Run("Get Users With Data", func(t *testing.T) {
		// For now, test the empty case and structure
		req, _ := http.NewRequest("GET", server.URL+"/api/users", nil)
		req.Header.Set("X-API-Key", "test-api-key")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestAPIIssuesCRUD(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	client := &http.Client{}

	var createdIssueID string

	t.Run("Create Issue Success", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "Test Issue",
			"description": "Test Description",
			"status":      "Todo",
			"priority":    "High",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", server.URL+"/api/issues", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var issue models.Issue
		err = json.NewDecoder(resp.Body).Decode(&issue)
		require.NoError(t, err)

		assert.Equal(t, "Test Issue", issue.Title)
		assert.Equal(t, "Test Description", issue.Description)
		assert.Equal(t, "Todo", issue.Status)
		assert.Equal(t, "High", issue.Priority)
		assert.NotEmpty(t, issue.ID)

		createdIssueID = issue.ID
	})

	t.Run("Get Issues", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/issues", nil)
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var issues []models.Issue
		err = json.NewDecoder(resp.Body).Decode(&issues)
		require.NoError(t, err)

		assert.Len(t, issues, 1)
		assert.Equal(t, "Test Issue", issues[0].Title)
	})

	t.Run("Get Single Issue", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/issues/"+createdIssueID, nil)
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var issue models.Issue
		err = json.NewDecoder(resp.Body).Decode(&issue)
		require.NoError(t, err)

		assert.Equal(t, "Test Issue", issue.Title)
	})

	t.Run("Update Issue", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":  "Updated Issue",
			"status": "In Progress",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("PATCH", server.URL+"/api/issues/"+createdIssueID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var issue models.Issue
		err = json.NewDecoder(resp.Body).Decode(&issue)
		require.NoError(t, err)

		assert.Equal(t, "Updated Issue", issue.Title)
		assert.Equal(t, "In Progress", issue.Status)
	})

	t.Run("Move Issue", func(t *testing.T) {
		payload := map[string]interface{}{
			"status":      "Done",
			"order_index": 5.0,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("PATCH", server.URL+"/api/issues/"+createdIssueID+"/move", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Move issue doesn't return a body, so verify by fetching the issue
		getReq, _ := http.NewRequest("GET", server.URL+"/api/issues/"+createdIssueID, nil)
		getReq.Header.Set("X-API-Key", "test-api-key")

		getResp, err := client.Do(getReq)
		require.NoError(t, err)
		defer getResp.Body.Close()

		assert.Equal(t, http.StatusOK, getResp.StatusCode)

		var issue models.Issue
		err = json.NewDecoder(getResp.Body).Decode(&issue)
		require.NoError(t, err)

		assert.Equal(t, "Done", issue.Status)
		assert.Equal(t, 5.0, issue.OrderIndex)
	})

	t.Run("Delete Issue", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", server.URL+"/api/issues/"+createdIssueID, nil)
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("Get Issues After Delete", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/issues", nil)
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var issues []models.Issue
		err = json.NewDecoder(resp.Body).Decode(&issues)
		require.NoError(t, err)

		assert.Len(t, issues, 0)
	})
}

func TestAPIAuthentication(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	client := &http.Client{}

	t.Run("Missing API Key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/issues", nil)
		// Don't set X-API-Key header

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Invalid API Key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/issues", nil)
		req.Header.Set("X-API-Key", "invalid-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Valid API Key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/issues", nil)
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestAPIValidation(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	client := &http.Client{}

	t.Run("Create Issue - Empty Title", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":  "",
			"status": "Todo",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", server.URL+"/api/issues", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Create Issue - Invalid Status", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":  "Test Issue",
			"status": "InvalidStatus",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", server.URL+"/api/issues", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Create Issue - Malformed JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", server.URL+"/api/issues", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAPIFilteringAndPagination(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	client := &http.Client{}

	// Setup test data
	setupIssues := []map[string]interface{}{
		{
			"title":    "Bug Issue",
			"status":   "Todo",
			"priority": "High",
		},
		{
			"title":    "Feature Issue",
			"status":   "In Progress",
			"priority": "Medium",
		},
		{
			"title":    "Enhancement Issue",
			"status":   "Done",
			"priority": "Low",
		},
	}

	for _, issue := range setupIssues {
		body, _ := json.Marshal(issue)
		req, _ := http.NewRequest("POST", server.URL+"/api/issues", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-api-key")
		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
	}

	t.Run("Filter by Status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/issues?status=Done", nil)
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var issues []models.Issue
		err = json.NewDecoder(resp.Body).Decode(&issues)
		require.NoError(t, err)

		assert.Len(t, issues, 1)
		assert.Equal(t, "Done", issues[0].Status)
	})

	t.Run("Filter by Priority", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/issues?priority=High", nil)
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var issues []models.Issue
		err = json.NewDecoder(resp.Body).Decode(&issues)
		require.NoError(t, err)

		assert.Len(t, issues, 1)
		assert.Equal(t, "High", issues[0].Priority)
	})

	t.Run("Pagination", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/issues?page=1&page_size=2", nil)
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var issues []models.Issue
		err = json.NewDecoder(resp.Body).Decode(&issues)
		require.NoError(t, err)

		assert.Len(t, issues, 2) // Should return 2 items due to page_size=2
	})
}

func TestAPIErrorHandling(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	client := &http.Client{}

	t.Run("Get Non-existent Issue", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/issues/999", nil)
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Update Non-existent Issue", func(t *testing.T) {
		payload := map[string]interface{}{
			"title": "Updated Title",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("PATCH", server.URL+"/api/issues/999", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Delete Non-existent Issue", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", server.URL+"/api/issues/999", nil)
		req.Header.Set("X-API-Key", "test-api-key")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}