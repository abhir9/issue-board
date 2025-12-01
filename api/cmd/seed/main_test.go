package main

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/abhir9/issue-board/api/internal/database"
	"github.com/abhir9/issue-board/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"
)

// Use database package to avoid unused import warning
var _ = database.DB

var idCounter int

// Helper function to generate deterministic IDs for testing
func generateID() string {
	idCounter++
	return fmt.Sprintf("test-id-%d", idCounter)
}

func setupSeedTest(t *testing.T) (*sql.DB, func()) {
	// Create temporary database for testing
	tmpFile, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Enable foreign keys
	_, err = tmpFile.Exec("PRAGMA foreign_keys = ON")
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
	`
	_, err = tmpFile.Exec(schema)
	require.NoError(t, err)

	cleanup := func() {
		tmpFile.Close()
	}

	return tmpFile, cleanup
}

func TestClearExistingData(t *testing.T) {
	db, cleanup := setupSeedTest(t)
	defer cleanup()

	// Insert some test data
	_, err := db.Exec("INSERT INTO users (id, name) VALUES (?, ?)", generateID(), "Test User")
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO labels (id, name, color) VALUES (?, ?, ?)", generateID(), "Test Label", "#000000")
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO issues (id, title, status, priority, order_index) VALUES (?, ?, ?, ?, ?)",
		generateID(), "Test Issue", "Todo", "Low", 0.0)
	require.NoError(t, err)

	// Override database.DB for testing
	originalDB := database.DB
	database.DB = db
	defer func() { database.DB = originalDB }()

	// Verify data exists
	var count int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	assert.Equal(t, 1, count)

	// Clear data
	err = clearExistingData()
	require.NoError(t, err)

	// Verify data was cleared
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	assert.Equal(t, 0, count)
	db.QueryRow("SELECT COUNT(*) FROM labels").Scan(&count)
	assert.Equal(t, 0, count)
	db.QueryRow("SELECT COUNT(*) FROM issues").Scan(&count)
	assert.Equal(t, 0, count)
	db.QueryRow("SELECT COUNT(*) FROM issue_labels").Scan(&count)
	assert.Equal(t, 0, count)
}

func TestSeedUsers(t *testing.T) {
	db, cleanup := setupSeedTest(t)
	defer cleanup()

	// Override database.DB for testing
	originalDB := database.DB
	database.DB = db
	defer func() { database.DB = originalDB }()

	err := seedUsers()
	require.NoError(t, err)

	// Verify users were inserted
	rows, err := db.Query("SELECT id, name, avatar_url FROM users ORDER BY name")
	require.NoError(t, err)
	defer rows.Close()

	var insertedUsers []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.AvatarURL)
		require.NoError(t, err)
		insertedUsers = append(insertedUsers, user)
	}

	assert.Len(t, insertedUsers, 3)
	assert.Equal(t, "Alice", insertedUsers[0].Name)
	assert.Equal(t, "Bob", insertedUsers[1].Name)
	assert.Equal(t, "Charlie", insertedUsers[2].Name)
	assert.Contains(t, insertedUsers[0].AvatarURL, "Alice")
	assert.Contains(t, insertedUsers[1].AvatarURL, "Bob")
	assert.Contains(t, insertedUsers[2].AvatarURL, "Charlie")
}

func TestSeedLabels(t *testing.T) {
	db, cleanup := setupSeedTest(t)
	defer cleanup()

	// Override database.DB for testing
	originalDB := database.DB
	database.DB = db
	defer func() { database.DB = originalDB }()

	labelIDs, err := seedLabels()
	require.NoError(t, err)
	assert.Len(t, labelIDs, 4)

	// Verify labels were inserted
	rows, err := db.Query("SELECT id, name, color FROM labels ORDER BY name")
	require.NoError(t, err)
	defer rows.Close()

	var insertedLabels []models.Label
	for rows.Next() {
		var label models.Label
		err := rows.Scan(&label.ID, &label.Name, &label.Color)
		require.NoError(t, err)
		insertedLabels = append(insertedLabels, label)
	}

	assert.Len(t, insertedLabels, 4)
	assert.Equal(t, "Bug", insertedLabels[0].Name)
	assert.Equal(t, "#ef4444", insertedLabels[0].Color)
	assert.Equal(t, "Documentation", insertedLabels[1].Name)
	assert.Equal(t, "#f59e0b", insertedLabels[1].Color)
	assert.Equal(t, "Enhancement", insertedLabels[2].Name)
	assert.Equal(t, "#10b981", insertedLabels[2].Color)
	assert.Equal(t, "Feature", insertedLabels[3].Name)
	assert.Equal(t, "#3b82f6", insertedLabels[3].Color)
}

func TestGetUserIDs(t *testing.T) {
	db, cleanup := setupSeedTest(t)
	defer cleanup()

	// Override database.DB for testing
	originalDB := database.DB
	database.DB = db
	defer func() { database.DB = originalDB }()

	// Insert test users
	userIDs := []string{generateID(), generateID(), generateID()}
	_, err := db.Exec("INSERT INTO users (id, name) VALUES (?, ?), (?, ?), (?, ?)",
		userIDs[0], "Alice", userIDs[1], "Bob", userIDs[2], "Charlie")
	require.NoError(t, err)

	retrievedIDs, err := getUserIDs()
	require.NoError(t, err)
	assert.Len(t, retrievedIDs, 3)
}

func TestSeedIssues(t *testing.T) {
	// This test is covered by TestSeedDatabase which tests the full seeding process
	t.Skip("Covered by TestSeedDatabase")
}

func TestSeedDatabase(t *testing.T) {
	db, cleanup := setupSeedTest(t)
	defer cleanup()

	// Override database.DB for testing
	originalDB := database.DB
	database.DB = db
	defer func() { database.DB = originalDB }()

	err := seedDatabase()
	require.NoError(t, err)

	// Verify all data was seeded
	var userCount, labelCount, issueCount, relationshipCount int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	db.QueryRow("SELECT COUNT(*) FROM labels").Scan(&labelCount)
	db.QueryRow("SELECT COUNT(*) FROM issues").Scan(&issueCount)
	db.QueryRow("SELECT COUNT(*) FROM issue_labels").Scan(&relationshipCount)

	assert.Equal(t, 3, userCount)
	assert.Equal(t, 4, labelCount)
	assert.Equal(t, 20, issueCount)
	assert.True(t, relationshipCount >= 20)
}