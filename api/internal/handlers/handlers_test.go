package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abhir9/issue-board/api/internal/models"
)

func TestCreateIssue(t *testing.T) {
	repo := setupTestDB(t)
	r := setupRouter(repo)

	t.Run("Success", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "Test Issue",
			"description": "Test Description",
			"status":      "Todo",
			"priority":    "High",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/issues", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
		}

		var issue models.Issue
		json.Unmarshal(w.Body.Bytes(), &issue)
		if issue.Title != "Test Issue" {
			t.Errorf("Expected title 'Test Issue', got '%s'", issue.Title)
		}
		if issue.Status != "Todo" {
			t.Errorf("Expected status 'Todo', got '%s'", issue.Status)
		}
	})

	t.Run("Success with all fields", func(t *testing.T) {
		// Create a user first to satisfy foreign key constraint
		repo.DB.Exec("INSERT INTO users (id, name) VALUES ('user1', 'Test User')")
		
		assigneeID := "user1"
		payload := map[string]interface{}{
			"title":       "Full Issue",
			"description": "Complete Description",
			"status":      "In Progress",
			"priority":    "Critical",
			"assignee_id": assigneeID,
			"label_ids":   []string{},
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/issues", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Invalid Input - Malformed JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/issues", bytes.NewBuffer([]byte("invalid")))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Validation - Empty Title", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "",
			"description": "Test Description",
			"status":      "Todo",
			"priority":    "High",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/issues", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for empty title, got %d", w.Code)
		}
	})

	t.Run("Validation - Invalid Status", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "Test Issue",
			"description": "Test Description",
			"status":      "InvalidStatus",
			"priority":    "High",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/issues", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid status, got %d", w.Code)
		}
	})

	t.Run("Validation - Invalid Priority", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "Test Issue",
			"description": "Test Description",
			"status":      "Todo",
			"priority":    "InvalidPriority",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/issues", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid priority, got %d", w.Code)
		}
	})

	t.Run("Validation - Title Too Long", func(t *testing.T) {
		longTitle := string(make([]byte, 201))
		for i := range longTitle {
			longTitle = string(append([]byte(longTitle[:i]), 'a'))
		}
		payload := map[string]interface{}{
			"title":       longTitle,
			"description": "Test Description",
			"status":      "Todo",
			"priority":    "High",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/issues", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for title too long, got %d", w.Code)
		}
	})
}

func TestGetIssues(t *testing.T) {
	repo := setupTestDB(t)
	r := setupRouter(repo)

	// Seed multiple issues
	ctx := context.Background()
	// Create user first to satisfy foreign key constraint
	repo.DB.Exec("INSERT INTO users (id, name) VALUES ('user1', 'Test User')")
	
	assigneeID := "user1"
	repo.CreateIssue(ctx, models.Issue{
		ID:          "1",
		Title:       "Issue 1",
		Status:      "Todo",
		Priority:    "Low",
		AssigneeID:  &assigneeID,
		Description: "Description 1",
		OrderIndex:  1.0,
	})
	repo.CreateIssue(ctx, models.Issue{
		ID:          "2",
		Title:       "Issue 2",
		Status:      "In Progress",
		Priority:    "High",
		Description: "Description 2",
		OrderIndex:  2.0,
	})
	repo.CreateIssue(ctx, models.Issue{
		ID:          "3",
		Title:       "Issue 3",
		Status:      "Done",
		Priority:    "Medium",
		Description: "Description 3",
		OrderIndex:  3.0,
	})

	t.Run("List All", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/issues", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var issues []models.Issue
		json.Unmarshal(w.Body.Bytes(), &issues)
		if len(issues) != 3 {
			t.Errorf("Expected 3 issues, got %d", len(issues))
		}
	})

	t.Run("Filter by Status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/issues?status=Done", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var issues []models.Issue
		json.Unmarshal(w.Body.Bytes(), &issues)
		if len(issues) != 1 {
			t.Errorf("Expected 1 issue with status Done, got %d", len(issues))
		}
		if len(issues) > 0 && issues[0].Status != "Done" {
			t.Errorf("Expected status 'Done', got '%s'", issues[0].Status)
		}
	})

	t.Run("Filter by Multiple Statuses", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/issues?status=Todo&status=In Progress", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var issues []models.Issue
		json.Unmarshal(w.Body.Bytes(), &issues)
		if len(issues) != 2 {
			t.Errorf("Expected 2 issues, got %d", len(issues))
		}
	})

	t.Run("Filter by Priority", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/issues?priority=High", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var issues []models.Issue
		json.Unmarshal(w.Body.Bytes(), &issues)
		if len(issues) != 1 {
			t.Errorf("Expected 1 issue with priority High, got %d", len(issues))
		}
	})

	t.Run("Filter by Assignee", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/issues?assignee=user1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var issues []models.Issue
		json.Unmarshal(w.Body.Bytes(), &issues)
		if len(issues) != 1 {
			t.Errorf("Expected 1 issue with assignee, got %d", len(issues))
		}
	})

	t.Run("Pagination - Page 1", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/issues?page=1&page_size=2", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var issues []models.Issue
		json.Unmarshal(w.Body.Bytes(), &issues)
		if len(issues) != 2 {
			t.Errorf("Expected 2 issues on page 1, got %d", len(issues))
		}
	})

	t.Run("Pagination - Page 2", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/issues?page=2&page_size=2", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var issues []models.Issue
		json.Unmarshal(w.Body.Bytes(), &issues)
		if len(issues) != 1 {
			t.Errorf("Expected 1 issue on page 2, got %d", len(issues))
		}
	})

	t.Run("Combined Filters", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/issues?status=Todo&priority=Low", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var issues []models.Issue
		json.Unmarshal(w.Body.Bytes(), &issues)
		if len(issues) != 1 {
			t.Errorf("Expected 1 issue with combined filters, got %d", len(issues))
		}
	})

	t.Run("No Results", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/issues?status=Canceled", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var issues []models.Issue
		json.Unmarshal(w.Body.Bytes(), &issues)
		if len(issues) != 0 {
			t.Errorf("Expected 0 issues, got %d", len(issues))
		}
	})
}

func TestUpdateIssue(t *testing.T) {
	repo := setupTestDB(t)
	r := setupRouter(repo)

	ctx := context.Background()
	repo.CreateIssue(ctx, models.Issue{
		ID:          "1",
		Title:       "Old Title",
		Description: "Old Description",
		Status:      "Todo",
		Priority:    "Low",
		OrderIndex:  1.0,
	})

	t.Run("Update Title", func(t *testing.T) {
		newTitle := "New Title"
		payload := map[string]interface{}{
			"title": newTitle,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH", "/issues/1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var issue models.Issue
		json.Unmarshal(w.Body.Bytes(), &issue)
		if issue.Title != "New Title" {
			t.Errorf("Expected title 'New Title', got '%s'", issue.Title)
		}
	})

	t.Run("Update Description", func(t *testing.T) {
		newDesc := "New Description"
		payload := map[string]interface{}{
			"description": newDesc,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH", "/issues/1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Update Status", func(t *testing.T) {
		newStatus := "In Progress"
		payload := map[string]interface{}{
			"status": newStatus,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH", "/issues/1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var issue models.Issue
		json.Unmarshal(w.Body.Bytes(), &issue)
		if issue.Status != "In Progress" {
			t.Errorf("Expected status 'In Progress', got '%s'", issue.Status)
		}
	})

	t.Run("Update Priority", func(t *testing.T) {
		newPriority := "High"
		payload := map[string]interface{}{
			"priority": newPriority,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH", "/issues/1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var issue models.Issue
		json.Unmarshal(w.Body.Bytes(), &issue)
		if issue.Priority != "High" {
			t.Errorf("Expected priority 'High', got '%s'", issue.Priority)
		}
	})

	t.Run("Update Multiple Fields", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":    "Updated Title",
			"status":   "Done",
			"priority": "Critical",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH", "/issues/1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var issue models.Issue
		json.Unmarshal(w.Body.Bytes(), &issue)
		if issue.Title != "Updated Title" {
			t.Errorf("Expected title 'Updated Title', got '%s'", issue.Title)
		}
		if issue.Status != "Done" {
			t.Errorf("Expected status 'Done', got '%s'", issue.Status)
		}
	})

	t.Run("Update with Assignee", func(t *testing.T) {
		// Create user first to satisfy foreign key constraint
		repo.DB.Exec("INSERT INTO users (id, name) VALUES ('user1', 'Test User')")
		
		assigneeID := "user1"
		payload := map[string]interface{}{
			"assignee_id": assigneeID,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH", "/issues/1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", "/issues/1", bytes.NewBuffer([]byte("invalid")))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Validation - Invalid Status", func(t *testing.T) {
		payload := map[string]interface{}{
			"status": "InvalidStatus",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH", "/issues/1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid status, got %d", w.Code)
		}
	})

	t.Run("Validation - Invalid Priority", func(t *testing.T) {
		payload := map[string]interface{}{
			"priority": "SuperHigh",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH", "/issues/1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid priority, got %d", w.Code)
		}
	})

	t.Run("Validation - Empty Title", func(t *testing.T) {
		payload := map[string]interface{}{
			"title": "",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH", "/issues/1", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for empty title, got %d", w.Code)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		payload := map[string]interface{}{
			"title": "New Title",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH", "/issues/999", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500 for non-existing issue, got %d", w.Code)
		}
	})
}

func TestDeleteIssue(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		ctx := context.Background()
		repo.CreateIssue(ctx, models.Issue{
			ID:       "1",
			Title:    "To Delete",
			Status:   "Todo",
			Priority: "Low",
		})

		req, _ := http.NewRequest("DELETE", "/issues/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", w.Code)
		}

		// Verify it's gone
		issue, _ := repo.GetIssue(ctx, "1")
		if issue != nil {
			t.Error("Expected issue to be deleted")
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		req, _ := http.NewRequest("DELETE", "/issues/nonexistent", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500 for non-existing issue, got %d", w.Code)
		}
	})

	t.Run("Delete Issue With Labels", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		ctx := context.Background()
		// Create issue with label
		repo.CreateIssue(ctx, models.Issue{
			ID:       "1",
			Title:    "Issue with Labels",
			Status:   "Todo",
			Priority: "High",
		})

		// Add label to issue
		repo.DB.Exec("INSERT INTO issue_labels (issue_id, label_id) VALUES ('1', 'bug')")

		req, _ := http.NewRequest("DELETE", "/issues/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", w.Code)
		}

		// Verify issue is deleted
		issue, _ := repo.GetIssue(ctx, "1")
		if issue != nil {
			t.Error("Expected issue to be deleted")
		}

		// Verify labels association is also deleted (CASCADE should handle this)
		var count int
		repo.DB.QueryRow("SELECT COUNT(*) FROM issue_labels WHERE issue_id = '1'").Scan(&count)
		if count != 0 {
			t.Error("Expected issue labels to be deleted via CASCADE")
		}
	})
}

func TestMoveIssue(t *testing.T) {
	t.Run("Success - Move to Different Status", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		ctx := context.Background()
		repo.CreateIssue(ctx, models.Issue{
			ID:         "1",
			Title:      "Move Me",
			Status:     "Todo",
			OrderIndex: 0,
			Priority:   "Low",
		})

		payload := map[string]interface{}{
			"status":      "Done",
			"order_index": 5.5,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH", "/issues/1/move", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		issue, _ := repo.GetIssue(ctx, "1")
		if issue.Status != "Done" {
			t.Errorf("Expected status 'Done', got '%s'", issue.Status)
		}
		if issue.OrderIndex != 5.5 {
			t.Errorf("Expected order_index 5.5, got %f", issue.OrderIndex)
		}
	})

	t.Run("Move Within Same Status", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		ctx := context.Background()
		repo.CreateIssue(ctx, models.Issue{
			ID:         "1",
			Title:      "Issue 1",
			Status:     "Todo",
			OrderIndex: 0,
			Priority:   "Low",
		})

		payload := map[string]interface{}{
			"status":      "Todo",
			"order_index": 3.0,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH", "/issues/1/move", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		issue, _ := repo.GetIssue(ctx, "1")
		if issue.OrderIndex != 3.0 {
			t.Errorf("Expected order_index 3.0, got %f", issue.OrderIndex)
		}
	})

	t.Run("Malformed JSON", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		ctx := context.Background()
		repo.CreateIssue(ctx, models.Issue{
			ID:       "1",
			Title:    "Issue",
			Status:   "Todo",
			Priority: "Low",
		})

		req, _ := http.NewRequest("PATCH", "/issues/1/move", bytes.NewBuffer([]byte("invalid-json")))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for malformed JSON, got %d", w.Code)
		}
	})

	t.Run("Invalid Status", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		ctx := context.Background()
		repo.CreateIssue(ctx, models.Issue{
			ID:       "1",
			Title:    "Issue",
			Status:   "Todo",
			Priority: "Low",
		})

		payload := map[string]interface{}{
			"status":      "InvalidStatus",
			"order_index": 1.0,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH", "/issues/1/move", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Note: MoveIssue currently doesn't validate status, so this might succeed
		// This test documents the current behavior
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 (no validation), got %d", w.Code)
		}
	})

	t.Run("Negative Order Index", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		ctx := context.Background()
		repo.CreateIssue(ctx, models.Issue{
			ID:       "1",
			Title:    "Issue",
			Status:   "Todo",
			Priority: "Low",
		})

		payload := map[string]interface{}{
			"status":      "Todo",
			"order_index": -1.0,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH", "/issues/1/move", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// This should still succeed as negative index might be valid for reordering
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		payload := map[string]interface{}{
			"status":      "Done",
			"order_index": 1.0,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PATCH", "/issues/999/move", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500 for non-existing issue, got %d", w.Code)
		}
	})
}

func TestGetIssue(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		ctx := context.Background()
		repo.CreateIssue(ctx, models.Issue{
			ID:       "1",
			Title:    "Get Single",
			Status:   "Todo",
			Priority: "Low",
		})

		req, _ := http.NewRequest("GET", "/issues/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var issue models.Issue
		json.Unmarshal(w.Body.Bytes(), &issue)
		if issue.ID != "1" {
			t.Errorf("Expected ID '1', got '%s'", issue.ID)
		}
		if issue.Title != "Get Single" {
			t.Errorf("Expected title 'Get Single', got '%s'", issue.Title)
		}
	})

	t.Run("Success With All Fields", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		ctx := context.Background()
		// Create user first to satisfy foreign key constraint
		repo.DB.Exec("INSERT INTO users (id, name) VALUES ('user1', 'Test User')")
		
		repo.CreateIssue(ctx, models.Issue{
			ID:          "2",
			Title:       "Full Issue",
			Description: "Complete description",
			Status:      "In Progress",
			Priority:    "High",
			AssigneeID:  ptr("user1"),
		})

		req, _ := http.NewRequest("GET", "/issues/2", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var issue models.Issue
		json.Unmarshal(w.Body.Bytes(), &issue)
		if issue.Description != "Complete description" {
			t.Errorf("Expected description 'Complete description', got '%s'", issue.Description)
		}
		if issue.Priority != "High" {
			t.Errorf("Expected priority 'High', got '%s'", issue.Priority)
		}
		if issue.AssigneeID == nil || *issue.AssigneeID != "user1" {
			t.Error("Expected assignee_id 'user1'")
		}
	})

	t.Run("Success With Labels", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		ctx := context.Background()
		// First create a label
		repo.DB.Exec("INSERT INTO labels (id, name, color) VALUES ('bug', 'Bug', '#FF0000')")
		
		repo.CreateIssue(ctx, models.Issue{
			ID:       "3",
			Title:    "Issue with Labels",
			Status:   "Todo",
			Priority: "Medium",
		})

		// Add label to issue
		repo.DB.Exec("INSERT INTO issue_labels (issue_id, label_id) VALUES ('3', 'bug')")

		req, _ := http.NewRequest("GET", "/issues/3", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var issue models.Issue
		json.Unmarshal(w.Body.Bytes(), &issue)
		if len(issue.Labels) == 0 {
			t.Error("Expected at least one label")
		}
		if issue.Labels[0].ID != "bug" {
			t.Errorf("Expected label ID 'bug', got '%s'", issue.Labels[0].ID)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		req, _ := http.NewRequest("GET", "/issues/999", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("Empty ID", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		req, _ := http.NewRequest("GET", "/issues/", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 404 or match the issues list endpoint
		if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 200 or 404, got %d", w.Code)
		}
	})
}

func TestGetUsers(t *testing.T) {
	t.Run("Success - Single User", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		// Seed user manually
		repo.DB.Exec("INSERT INTO users (id, name) VALUES ('u1', 'User 1')")

		req, _ := http.NewRequest("GET", "/users", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var users []models.User
		json.Unmarshal(w.Body.Bytes(), &users)
		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}
		if users[0].ID != "u1" {
			t.Errorf("Expected user ID 'u1', got '%s'", users[0].ID)
		}
		if users[0].Name != "User 1" {
			t.Errorf("Expected user name 'User 1', got '%s'", users[0].Name)
		}
	})

	t.Run("Success - Multiple Users", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		// Seed multiple users
		repo.DB.Exec("INSERT INTO users (id, name) VALUES ('u1', 'Alice'), ('u2', 'Bob'), ('u3', 'Charlie')")

		req, _ := http.NewRequest("GET", "/users", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var users []models.User
		json.Unmarshal(w.Body.Bytes(), &users)
		if len(users) != 3 {
			t.Errorf("Expected 3 users, got %d", len(users))
		}
	})

	t.Run("Empty Users List", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		// Don't seed any users

		req, _ := http.NewRequest("GET", "/users", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var users []models.User
		json.Unmarshal(w.Body.Bytes(), &users)
		if len(users) != 0 {
			t.Errorf("Expected 0 users, got %d", len(users))
		}
	})

	t.Run("Users With Avatar", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		// Seed user with avatar
		repo.DB.Exec("INSERT INTO users (id, name, avatar_url) VALUES ('u1', 'User With Avatar', 'https://example.com/avatar.jpg')")

		req, _ := http.NewRequest("GET", "/users", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var users []models.User
		json.Unmarshal(w.Body.Bytes(), &users)
		if len(users) != 1 {
			t.Errorf("Expected 1 user, got %d", len(users))
		}
		if users[0].AvatarURL != "https://example.com/avatar.jpg" {
			t.Errorf("Expected avatar URL, got '%s'", users[0].AvatarURL)
		}
	})
}

func TestGetLabels(t *testing.T) {
	t.Run("Success - Single Label", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		// Seed label manually
		repo.DB.Exec("INSERT INTO labels (id, name, color) VALUES ('l1', 'Label 1', '#000000')")

		req, _ := http.NewRequest("GET", "/labels", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var labels []models.Label
		json.Unmarshal(w.Body.Bytes(), &labels)
		if len(labels) != 1 {
			t.Errorf("Expected 1 label, got %d", len(labels))
		}
		if labels[0].ID != "l1" {
			t.Errorf("Expected label ID 'l1', got '%s'", labels[0].ID)
		}
		if labels[0].Name != "Label 1" {
			t.Errorf("Expected label name 'Label 1', got '%s'", labels[0].Name)
		}
		if labels[0].Color != "#000000" {
			t.Errorf("Expected label color '#000000', got '%s'", labels[0].Color)
		}
	})

	t.Run("Success - Multiple Labels", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		// Seed multiple labels manually (test DB doesn't run seed data)
		repo.DB.Exec("INSERT INTO labels (id, name, color) VALUES ('bug', 'Bug', '#FF0000'), ('feature', 'Feature', '#00FF00'), ('custom1', 'Custom Label', '#FF5733')")

		req, _ := http.NewRequest("GET", "/labels", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var labels []models.Label
		json.Unmarshal(w.Body.Bytes(), &labels)
		if len(labels) != 3 {
			t.Errorf("Expected 3 labels, got %d", len(labels))
		}
	})

	t.Run("Check Label Retrieval", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		// Manually seed labels for test
		repo.DB.Exec("INSERT INTO labels (id, name, color) VALUES ('bug', 'Bug', '#FF0000'), ('feature', 'Feature', '#00FF00'), ('enhancement', 'Enhancement', '#0000FF')")

		req, _ := http.NewRequest("GET", "/labels", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var labels []models.Label
		json.Unmarshal(w.Body.Bytes(), &labels)
		
		// Verify labels are present
		labelMap := make(map[string]bool)
		for _, label := range labels {
			labelMap[label.ID] = true
		}

		if !labelMap["bug"] {
			t.Error("Expected 'bug' label to be present")
		}
		if !labelMap["feature"] {
			t.Error("Expected 'feature' label to be present")
		}
		if !labelMap["enhancement"] {
			t.Error("Expected 'enhancement' label to be present")
		}
	})

	t.Run("Label Color Format", func(t *testing.T) {
		repo := setupTestDB(t)
		r := setupRouter(repo)

		// Insert label with specific color format
		repo.DB.Exec("INSERT INTO labels (id, name, color) VALUES ('color_test', 'Color Test', '#AABBCC')")

		req, _ := http.NewRequest("GET", "/labels", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var labels []models.Label
		json.Unmarshal(w.Body.Bytes(), &labels)
		
		// Find our test label
		var colorTestLabel *models.Label
		for i := range labels {
			if labels[i].ID == "color_test" {
				colorTestLabel = &labels[i]
				break
			}
		}

		if colorTestLabel == nil {
			t.Error("Expected to find 'color_test' label")
		} else if colorTestLabel.Color != "#AABBCC" {
			t.Errorf("Expected color '#AABBCC', got '%s'", colorTestLabel.Color)
		}
	})
}
