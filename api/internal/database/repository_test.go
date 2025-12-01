package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/abhir9/issue-board/api/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *Repository {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory db: %v", err)
	}

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

	return NewRepository(db)
}

func seedTestData(t *testing.T, repo *Repository) (string, string, string) {
	ctx := context.Background()

	// Create users
	_, err := repo.DB.ExecContext(ctx, "INSERT INTO users (id, name, avatar_url) VALUES (?, ?, ?)", 
		"user1", "Alice", "https://example.com/alice.jpg")
	if err != nil {
		t.Fatalf("Failed to seed user: %v", err)
	}

	// Create labels
	_, err = repo.DB.ExecContext(ctx, "INSERT INTO labels (id, name, color) VALUES (?, ?, ?)", 
		"label1", "Bug", "#FF0000")
	if err != nil {
		t.Fatalf("Failed to seed label: %v", err)
	}

	_, err = repo.DB.ExecContext(ctx, "INSERT INTO labels (id, name, color) VALUES (?, ?, ?)", 
		"label2", "Feature", "#00FF00")
	if err != nil {
		t.Fatalf("Failed to seed label: %v", err)
	}

	return "user1", "label1", "label2"
}

func TestCreateIssue(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	issue := models.Issue{
		ID:          "test-issue-1",
		Title:       "Test Issue",
		Description: "Test Description",
		Status:      "Todo",
		Priority:    "High",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		OrderIndex:  1.0,
	}

	err := repo.CreateIssue(ctx, issue)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	// Verify issue was created
	createdIssue, err := repo.GetIssue(ctx, "test-issue-1")
	if err != nil {
		t.Fatalf("Failed to get created issue: %v", err)
	}

	if createdIssue == nil {
		t.Fatal("Expected issue to be created, got nil")
	}

	if createdIssue.Title != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got '%s'", createdIssue.Title)
	}
}

func TestGetIssue(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Create an issue
	issue := models.Issue{
		ID:          "test-issue-1",
		Title:       "Test Issue",
		Description: "Test Description",
		Status:      "Todo",
		Priority:    "High",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		OrderIndex:  1.0,
	}
	repo.CreateIssue(ctx, issue)

	t.Run("Existing Issue", func(t *testing.T) {
		result, err := repo.GetIssue(ctx, "test-issue-1")
		if err != nil {
			t.Fatalf("Failed to get issue: %v", err)
		}

		if result == nil {
			t.Fatal("Expected issue, got nil")
		}

		if result.Title != "Test Issue" {
			t.Errorf("Expected title 'Test Issue', got '%s'", result.Title)
		}
	})

	t.Run("Non-Existing Issue", func(t *testing.T) {
		result, err := repo.GetIssue(ctx, "non-existing")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result != nil {
			t.Error("Expected nil for non-existing issue")
		}
	})
}

func TestGetIssues(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()
	userID, label1, label2 := seedTestData(t, repo)

	// Create multiple issues
	issues := []models.Issue{
		{
			ID:          "issue-1",
			Title:       "Issue 1",
			Description: "Description 1",
			Status:      "Todo",
			Priority:    "High",
			AssigneeID:  &userID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			OrderIndex:  1.0,
		},
		{
			ID:          "issue-2",
			Title:       "Issue 2",
			Description: "Description 2",
			Status:      "In Progress",
			Priority:    "Low",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			OrderIndex:  2.0,
		},
		{
			ID:          "issue-3",
			Title:       "Issue 3",
			Description: "Description 3",
			Status:      "Done",
			Priority:    "High",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			OrderIndex:  3.0,
		},
	}

	for _, issue := range issues {
		if err := repo.CreateIssue(ctx, issue); err != nil {
			t.Fatalf("Failed to create issue: %v", err)
		}
	}

	// Add labels to issue-1
	repo.UpdateIssueLabels(ctx, "issue-1", []string{label1, label2})

	t.Run("Get All Issues", func(t *testing.T) {
		results, err := repo.GetIssues(ctx, nil, "", nil, nil, 1, 0)
		if err != nil {
			t.Fatalf("Failed to get issues: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 issues, got %d", len(results))
		}
	})

	t.Run("Filter by Status", func(t *testing.T) {
		results, err := repo.GetIssues(ctx, []string{"Todo"}, "", nil, nil, 1, 0)
		if err != nil {
			t.Fatalf("Failed to get issues: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 issue, got %d", len(results))
		}

		if results[0].Status != "Todo" {
			t.Errorf("Expected status 'Todo', got '%s'", results[0].Status)
		}
	})

	t.Run("Filter by Priority", func(t *testing.T) {
		results, err := repo.GetIssues(ctx, nil, "", []string{"High"}, nil, 1, 0)
		if err != nil {
			t.Fatalf("Failed to get issues: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 issues, got %d", len(results))
		}
	})

	t.Run("Filter by Assignee", func(t *testing.T) {
		results, err := repo.GetIssues(ctx, nil, userID, nil, nil, 1, 0)
		if err != nil {
			t.Fatalf("Failed to get issues: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 issue, got %d", len(results))
		}
	})

	t.Run("Filter by Label", func(t *testing.T) {
		results, err := repo.GetIssues(ctx, nil, "", nil, []string{"Bug"}, 1, 0)
		if err != nil {
			t.Fatalf("Failed to get issues: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 issue with Bug label, got %d", len(results))
		}

		if len(results[0].Labels) != 2 {
			t.Errorf("Expected 2 labels, got %d", len(results[0].Labels))
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		results, err := repo.GetIssues(ctx, nil, "", nil, nil, 1, 2)
		if err != nil {
			t.Fatalf("Failed to get issues: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 issues (page size), got %d", len(results))
		}
	})
}

func TestUpdateIssue(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Create an issue
	issue := models.Issue{
		ID:          "test-issue-1",
		Title:       "Original Title",
		Description: "Original Description",
		Status:      "Todo",
		Priority:    "Low",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		OrderIndex:  1.0,
	}
	repo.CreateIssue(ctx, issue)

	t.Run("Update Title", func(t *testing.T) {
		updates := map[string]interface{}{
			"title":      "Updated Title",
			"updated_at": time.Now(),
		}

		err := repo.UpdateIssue(ctx, "test-issue-1", updates)
		if err != nil {
			t.Fatalf("Failed to update issue: %v", err)
		}

		updated, _ := repo.GetIssue(ctx, "test-issue-1")
		if updated.Title != "Updated Title" {
			t.Errorf("Expected title 'Updated Title', got '%s'", updated.Title)
		}
	})

	t.Run("Update Multiple Fields", func(t *testing.T) {
		updates := map[string]interface{}{
			"status":     "In Progress",
			"priority":   "High",
			"updated_at": time.Now(),
		}

		err := repo.UpdateIssue(ctx, "test-issue-1", updates)
		if err != nil {
			t.Fatalf("Failed to update issue: %v", err)
		}

		updated, _ := repo.GetIssue(ctx, "test-issue-1")
		if updated.Status != "In Progress" {
			t.Errorf("Expected status 'In Progress', got '%s'", updated.Status)
		}
		if updated.Priority != "High" {
			t.Errorf("Expected priority 'High', got '%s'", updated.Priority)
		}
	})

	t.Run("Update Non-Existing Issue", func(t *testing.T) {
		updates := map[string]interface{}{
			"title": "New Title",
		}

		err := repo.UpdateIssue(ctx, "non-existing", updates)
		if err == nil {
			t.Error("Expected error for non-existing issue, got nil")
		}
	})
}

func TestUpdateIssueLabels(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()
	_, label1, label2 := seedTestData(t, repo)

	// Create an issue
	issue := models.Issue{
		ID:          "test-issue-1",
		Title:       "Test Issue",
		Status:      "Todo",
		Priority:    "Low",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		OrderIndex:  1.0,
	}
	repo.CreateIssue(ctx, issue)

	t.Run("Add Labels", func(t *testing.T) {
		err := repo.UpdateIssueLabels(ctx, "test-issue-1", []string{label1, label2})
		if err != nil {
			t.Fatalf("Failed to update labels: %v", err)
		}

		labels, err := repo.GetLabelsForIssue(ctx, "test-issue-1")
		if err != nil {
			t.Fatalf("Failed to get labels: %v", err)
		}

		if len(labels) != 2 {
			t.Errorf("Expected 2 labels, got %d", len(labels))
		}
	})

	t.Run("Replace Labels", func(t *testing.T) {
		err := repo.UpdateIssueLabels(ctx, "test-issue-1", []string{label1})
		if err != nil {
			t.Fatalf("Failed to update labels: %v", err)
		}

		labels, err := repo.GetLabelsForIssue(ctx, "test-issue-1")
		if err != nil {
			t.Fatalf("Failed to get labels: %v", err)
		}

		if len(labels) != 1 {
			t.Errorf("Expected 1 label after replacement, got %d", len(labels))
		}
	})

	t.Run("Remove All Labels", func(t *testing.T) {
		err := repo.UpdateIssueLabels(ctx, "test-issue-1", []string{})
		if err != nil {
			t.Fatalf("Failed to update labels: %v", err)
		}

		labels, err := repo.GetLabelsForIssue(ctx, "test-issue-1")
		if err != nil {
			t.Fatalf("Failed to get labels: %v", err)
		}

		if len(labels) != 0 {
			t.Errorf("Expected 0 labels, got %d", len(labels))
		}
	})
}

func TestDeleteIssue(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Create an issue
	issue := models.Issue{
		ID:          "test-issue-1",
		Title:       "Test Issue",
		Status:      "Todo",
		Priority:    "Low",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		OrderIndex:  1.0,
	}
	repo.CreateIssue(ctx, issue)

	t.Run("Delete Existing Issue", func(t *testing.T) {
		err := repo.DeleteIssue(ctx, "test-issue-1")
		if err != nil {
			t.Fatalf("Failed to delete issue: %v", err)
		}

		// Verify deletion
		deleted, _ := repo.GetIssue(ctx, "test-issue-1")
		if deleted != nil {
			t.Error("Expected issue to be deleted")
		}
	})

	t.Run("Delete Non-Existing Issue", func(t *testing.T) {
		err := repo.DeleteIssue(ctx, "non-existing")
		if err == nil {
			t.Error("Expected error for non-existing issue, got nil")
		}
	})
}

func TestGetUsers(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Seed users
	users := []struct {
		id     string
		name   string
		avatar string
	}{
		{"user1", "Alice", "https://example.com/alice.jpg"},
		{"user2", "Bob", "https://example.com/bob.jpg"},
		{"user3", "Charlie", ""},
	}

	for _, u := range users {
		_, err := repo.DB.ExecContext(ctx, 
			"INSERT INTO users (id, name, avatar_url) VALUES (?, ?, ?)", 
			u.id, u.name, u.avatar)
		if err != nil {
			t.Fatalf("Failed to seed user: %v", err)
		}
	}

	results, err := repo.GetUsers(ctx)
	if err != nil {
		t.Fatalf("Failed to get users: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 users, got %d", len(results))
	}

	// Verify user with no avatar
	hasEmptyAvatar := false
	for _, user := range results {
		if user.AvatarURL == "" {
			hasEmptyAvatar = true
			break
		}
	}
	if !hasEmptyAvatar {
		t.Error("Expected at least one user with empty avatar")
	}
}

func TestGetLabels(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	// Seed labels
	labels := []struct {
		id    string
		name  string
		color string
	}{
		{"label1", "Bug", "#FF0000"},
		{"label2", "Feature", "#00FF00"},
		{"label3", "Enhancement", "#0000FF"},
	}

	for _, l := range labels {
		_, err := repo.DB.ExecContext(ctx, 
			"INSERT INTO labels (id, name, color) VALUES (?, ?, ?)", 
			l.id, l.name, l.color)
		if err != nil {
			t.Fatalf("Failed to seed label: %v", err)
		}
	}

	results, err := repo.GetLabels(ctx)
	if err != nil {
		t.Fatalf("Failed to get labels: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 labels, got %d", len(results))
	}

	// Verify label data
	for _, label := range results {
		if label.Name == "" || label.Color == "" {
			t.Error("Expected all labels to have name and color")
		}
	}
}

func TestGetLabelsForIssues(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()
	_, label1, label2 := seedTestData(t, repo)

	// Create issues
	issue1 := models.Issue{
		ID:          "issue-1",
		Title:       "Issue 1",
		Status:      "Todo",
		Priority:    "Low",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		OrderIndex:  1.0,
	}
	issue2 := models.Issue{
		ID:          "issue-2",
		Title:       "Issue 2",
		Status:      "Todo",
		Priority:    "Low",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		OrderIndex:  2.0,
	}

	repo.CreateIssue(ctx, issue1)
	repo.CreateIssue(ctx, issue2)

	// Add labels
	repo.UpdateIssueLabels(ctx, "issue-1", []string{label1, label2})
	repo.UpdateIssueLabels(ctx, "issue-2", []string{label1})

	t.Run("Batch Fetch Labels", func(t *testing.T) {
		labelMap, err := repo.GetLabelsForIssues(ctx, []string{"issue-1", "issue-2"})
		if err != nil {
			t.Fatalf("Failed to get labels for issues: %v", err)
		}

		if len(labelMap) != 2 {
			t.Errorf("Expected labels for 2 issues, got %d", len(labelMap))
		}

		if len(labelMap["issue-1"]) != 2 {
			t.Errorf("Expected 2 labels for issue-1, got %d", len(labelMap["issue-1"]))
		}

		if len(labelMap["issue-2"]) != 1 {
			t.Errorf("Expected 1 label for issue-2, got %d", len(labelMap["issue-2"]))
		}
	})

	t.Run("Empty Issue List", func(t *testing.T) {
		labelMap, err := repo.GetLabelsForIssues(ctx, []string{})
		if err != nil {
			t.Fatalf("Failed to get labels for empty issue list: %v", err)
		}

		if len(labelMap) != 0 {
			t.Errorf("Expected empty map, got %d entries", len(labelMap))
		}
	})
}
