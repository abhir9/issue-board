package main

import (
	"fmt"
	"issue-board-backend/internal/database"
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

	// Seed Issues - One per status
	statuses := []string{"Backlog", "Todo", "In Progress", "Done", "Canceled"}
	priorities := []string{"Low", "Medium", "High", "Critical", "Medium"}
	titles := []string{
		"Setup project infrastructure",
		"Implement user authentication",
		"Design database schema",
		"Deploy to production",
		"Old feature request",
	}
	descriptions := []string{
		"Initialize the project with proper folder structure and dependencies",
		"Add login and registration functionality with JWT tokens",
		"Create database tables and relationships for the application",
		"Configure CI/CD pipeline and deploy to production server",
		"This feature is no longer needed and has been canceled",
	}

	for i := 0; i < len(statuses); i++ {
		id := uuid.New().String()
		title := titles[i]
		desc := descriptions[i]
		status := statuses[i]
		priority := priorities[i]
		assigneeID := userIDs[i%len(userIDs)]

		_, err := database.DB.Exec(`
			INSERT INTO issues (id, title, description, status, priority, assignee_id, order_index)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, id, title, desc, status, priority, assigneeID, float64(i))
		if err != nil {
			log.Fatalf("Failed to insert issue: %v", err)
		}

		// Add a label
		labelID := labelIDs[i%len(labelIDs)]
		_, err = database.DB.Exec("INSERT INTO issue_labels (issue_id, label_id) VALUES (?, ?)", id, labelID)
		if err != nil {
			log.Fatalf("Failed to insert issue label: %v", err)
		}

		fmt.Printf("Inserted issue: %s (Status: %s)\\n", title, status)
	}
	fmt.Println("Seeding complete! Created 5 issues (1 per status)")
}
