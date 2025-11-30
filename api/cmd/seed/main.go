package main

import (
	"api/internal/database"
	"fmt"
	"log"

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

	// Clear existing data
	database.DB.Exec("DELETE FROM issue_labels")
	database.DB.Exec("DELETE FROM issues")
	database.DB.Exec("DELETE FROM labels")
	database.DB.Exec("DELETE FROM users")

	// Seed Users
	users := []struct {
		Name      string
		AvatarURL string
	}{
		{"Alice", "https://api.dicebear.com/7.x/avataaars/svg?seed=Alice"},
		{"Bob", "https://api.dicebear.com/7.x/avataaars/svg?seed=Bob"},
		{"Charlie", "https://api.dicebear.com/7.x/avataaars/svg?seed=Charlie"},
	}

	var userIDs []string
	for _, u := range users {
		id := uuid.New().String()
		_, err := database.DB.Exec("INSERT INTO users (id, name, avatar_url) VALUES (?, ?, ?)", id, u.Name, u.AvatarURL)
		if err != nil {
			log.Fatalf("Failed to insert user: %v", err)
		}
		userIDs = append(userIDs, id)
		fmt.Printf("Inserted user: %s\n", u.Name)
	}

	// Seed Labels
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
			log.Fatalf("Failed to insert label: %v", err)
		}
		labelIDs = append(labelIDs, id)
		fmt.Printf("Inserted label: %s\n", l.Name)
	}

	// Seed Issues - 20 issues spread across statuses
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
			log.Fatalf("Failed to insert issue: %v", err)
		}

		// Add 1-2 labels per issue
		labelID := labelIDs[i%len(labelIDs)]
		_, err = database.DB.Exec("INSERT INTO issue_labels (issue_id, label_id) VALUES (?, ?)", id, labelID)
		if err != nil {
			log.Fatalf("Failed to insert issue label: %v", err)
		}

		// Add a second label to some issues
		if i%3 == 0 && len(labelIDs) > 1 {
			secondLabelID := labelIDs[(i+1)%len(labelIDs)]
			database.DB.Exec("INSERT INTO issue_labels (issue_id, label_id) VALUES (?, ?)", id, secondLabelID)
		}

		fmt.Printf("Inserted issue: %s (Status: %s, Priority: %s)\n", issue.Title, issue.Status, issue.Priority)
	}
	fmt.Println("Seeding complete! Created 20 issues spread across statuses")
}
