package main

import (
	"fmt"
	"log"

	"github.com/abhir9/issue-board/api/internal/database"
	"github.com/google/uuid"
)

func main() {
	if err := database.InitDB("./issues.db"); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}
	defer database.DB.Close()

	if err := database.RunMigrations("./migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if err := seedDatabase(); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	fmt.Println("Seeding complete! Created 20 issues spread across statuses")
}

func seedDatabase() error {
	// Clear existing data
	if err := clearExistingData(); err != nil {
		return fmt.Errorf("failed to clear existing data: %w", err)
	}

	// Seed data
	if err := seedUsers(); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	labelIDs, err := seedLabels()
	if err != nil {
		return fmt.Errorf("failed to seed labels: %w", err)
	}

	userIDs, err := getUserIDs()
	if err != nil {
		return fmt.Errorf("failed to get user IDs: %w", err)
	}

	if err := seedIssues(userIDs, labelIDs); err != nil {
		return fmt.Errorf("failed to seed issues: %w", err)
	}

	return nil
}

func clearExistingData() error {
	_, err := database.DB.Exec("DELETE FROM issue_labels")
	if err != nil {
		return err
	}
	_, err = database.DB.Exec("DELETE FROM issues")
	if err != nil {
		return err
	}
	_, err = database.DB.Exec("DELETE FROM labels")
	if err != nil {
		return err
	}
	_, err = database.DB.Exec("DELETE FROM users")
	return err
}

func seedUsers() error {
	users := []struct {
		Name      string
		AvatarURL string
	}{
		{"Alice", "https://api.dicebear.com/7.x/avataaars/svg?seed=Alice"},
		{"Bob", "https://api.dicebear.com/7.x/avataaars/svg?seed=Bob"},
		{"Charlie", "https://api.dicebear.com/7.x/avataaars/svg?seed=Charlie"},
	}

	for _, u := range users {
		id := uuid.New().String()
		_, err := database.DB.Exec("INSERT INTO users (id, name, avatar_url) VALUES (?, ?, ?)", id, u.Name, u.AvatarURL)
		if err != nil {
			return err
		}
		fmt.Printf("Inserted user: %s\n", u.Name)
	}
	return nil
}

func seedLabels() ([]string, error) {
	labels := []struct {
		Name  string
		Color string
	}{
		{"Bug", "#ef4444"},
		{"Feature", "#3b82f6"},
		{"Enhancement", "#10b981"},
		{"Documentation", "#f59e0b"},
	}

	var labelIDs []string
	for _, l := range labels {
		id := uuid.New().String()
		_, err := database.DB.Exec("INSERT INTO labels (id, name, color) VALUES (?, ?, ?)", id, l.Name, l.Color)
		if err != nil {
			return nil, err
		}
		labelIDs = append(labelIDs, id)
		fmt.Printf("Inserted label: %s\n", l.Name)
	}
	return labelIDs, nil
}

func getUserIDs() ([]string, error) {
	rows, err := database.DB.Query("SELECT id FROM users ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, id)
	}
	return userIDs, rows.Err()
}

func seedIssues(userIDs, labelIDs []string) error {
	type IssueData struct {
		Title       string
		Description string
		Status      string
		Priority    string
	}

	issues := []IssueData{
		// Backlog (4 issues)
		{"Setup CI/CD pipeline", "Configure automated testing and deployment workflows", "Backlog", "Medium"},
		{"Add email notifications", "Send email alerts for issue assignments and updates", "Backlog", "Low"},
		{"Implement dark mode", "Add dark theme support across the application", "Backlog", "Low"},
		{"Performance optimization", "Optimize database queries and frontend rendering", "Backlog", "High"},

		// Todo (5 issues)
		{"Setup project infrastructure", "Initialize the project with proper folder structure and dependencies", "Todo", "High"},
		{"Create user profile page", "Design and implement user profile with settings", "Todo", "Medium"},
		{"Add search functionality", "Implement full-text search for issues", "Todo", "Medium"},
		{"Write API documentation", "Document all API endpoints with examples", "Todo", "Low"},
		{"Setup error tracking", "Integrate Sentry or similar error tracking service", "Todo", "High"},

		// In Progress (4 issues)
		{"Implement user authentication", "Add login and registration functionality with JWT tokens", "In Progress", "Critical"},
		{"Design database schema", "Create database tables and relationships for the application", "In Progress", "High"},
		{"Add file upload feature", "Allow users to attach files to issues", "In Progress", "Medium"},
		{"Implement real-time updates", "Add WebSocket support for live issue updates", "In Progress", "High"},

		// Done (5 issues)
		{"Deploy to production", "Configure CI/CD pipeline and deploy to production server", "Done", "Critical"},
		{"Setup development environment", "Configure local development setup with Docker", "Done", "Medium"},
		{"Create landing page", "Design and implement the main landing page", "Done", "Medium"},
		{"Add basic CRUD operations", "Implement create, read, update, delete for issues", "Done", "High"},
		{"Setup database migrations", "Configure database migration system", "Done", "Medium"},

		// Canceled (2 issues)
		{"Old feature request", "This feature is no longer needed and has been canceled", "Canceled", "Low"},
		{"Legacy API support", "Support for old API version - no longer required", "Canceled", "Low"},
	}

	for i, issue := range issues {
		id := uuid.New().String()
		assigneeID := userIDs[i%len(userIDs)]

		_, err := database.DB.Exec(`
			INSERT INTO issues (id, title, description, status, priority, assignee_id, order_index)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, id, issue.Title, issue.Description, issue.Status, issue.Priority, assigneeID, float64(i))
		if err != nil {
			return err
		}

		// Add 1-2 labels per issue
		labelID := labelIDs[i%len(labelIDs)]
		_, err = database.DB.Exec("INSERT INTO issue_labels (issue_id, label_id) VALUES (?, ?)", id, labelID)
		if err != nil {
			return err
		}

		// Add a second label to some issues
		if i%3 == 0 && len(labelIDs) > 1 {
			secondLabelID := labelIDs[(i+1)%len(labelIDs)]
			_, err = database.DB.Exec("INSERT INTO issue_labels (issue_id, label_id) VALUES (?, ?)", id, secondLabelID)
			if err != nil {
				return err
			}
		}

		fmt.Printf("Inserted issue: %s (Status: %s, Priority: %s)\n", issue.Title, issue.Status, issue.Priority)
	}
	return nil
}
