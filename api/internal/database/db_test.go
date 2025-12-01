package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitDB(t *testing.T) {
	t.Run("Initialize in-memory database", func(t *testing.T) {
		err := InitDB(":memory:")
		if err != nil {
			t.Fatalf("Failed to init in-memory database: %v", err)
		}

		if DB == nil {
			t.Fatal("Expected DB to be initialized")
		}

		// Test connection
		err = DB.Ping()
		if err != nil {
			t.Fatalf("Failed to ping database: %v", err)
		}

		DB.Close()
	})

	t.Run("Initialize file database", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "test.db")

		err := InitDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to init file database: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Error("Expected database file to be created")
		}

		DB.Close()
	})

	t.Run("Invalid database path", func(t *testing.T) {
		err := InitDB("/invalid/path/that/does/not/exist/db.sqlite")
		if err == nil {
			t.Error("Expected error for invalid database path")
		}
	})
}

func TestRunMigrations(t *testing.T) {
	t.Run("Run valid migrations", func(t *testing.T) {
		// Create temp dir for migrations
		tempDir := t.TempDir()

		// Create a simple migration file
		migrationSQL := `
			CREATE TABLE IF NOT EXISTS test_table (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL
			);
		`
		migrationPath := filepath.Join(tempDir, "001_test.sql")
		err := os.WriteFile(migrationPath, []byte(migrationSQL), 0644)
		if err != nil {
			t.Fatalf("Failed to create migration file: %v", err)
		}

		// Initialize in-memory database
		err = InitDB(":memory:")
		if err != nil {
			t.Fatalf("Failed to init database: %v", err)
		}
		defer DB.Close()

		// Run migrations
		err = RunMigrations(tempDir)
		if err != nil {
			t.Fatalf("Failed to run migrations: %v", err)
		}

		// Verify table was created
		var tableName string
		err = DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='test_table'").Scan(&tableName)
		if err != nil {
			t.Fatalf("Failed to query table: %v", err)
		}

		if tableName != "test_table" {
			t.Errorf("Expected table 'test_table', got '%s'", tableName)
		}
	})

	t.Run("Run multiple migrations", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create multiple migration files
		migrations := []struct {
			name    string
			content string
		}{
			{
				name:    "001_users.sql",
				content: "CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY);",
			},
			{
				name:    "002_posts.sql",
				content: "CREATE TABLE IF NOT EXISTS posts (id TEXT PRIMARY KEY);",
			},
		}

		for _, m := range migrations {
			migrationPath := filepath.Join(tempDir, m.name)
			err := os.WriteFile(migrationPath, []byte(m.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create migration file: %v", err)
			}
		}

		// Initialize database and run migrations
		err := InitDB(":memory:")
		if err != nil {
			t.Fatalf("Failed to init database: %v", err)
		}
		defer DB.Close()

		err = RunMigrations(tempDir)
		if err != nil {
			t.Fatalf("Failed to run migrations: %v", err)
		}

		// Verify both tables were created
		var count int
		err = DB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('users', 'posts')").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query tables: %v", err)
		}

		if count != 2 {
			t.Errorf("Expected 2 tables, got %d", count)
		}
	})

	t.Run("Invalid migration directory", func(t *testing.T) {
		err := InitDB(":memory:")
		if err != nil {
			t.Fatalf("Failed to init database: %v", err)
		}
		defer DB.Close()

		err = RunMigrations("/invalid/migration/path")
		if err == nil {
			t.Error("Expected error for invalid migration directory")
		}
	})

	t.Run("Invalid SQL in migration", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create migration with invalid SQL
		migrationPath := filepath.Join(tempDir, "001_invalid.sql")
		err := os.WriteFile(migrationPath, []byte("INVALID SQL STATEMENT;"), 0644)
		if err != nil {
			t.Fatalf("Failed to create migration file: %v", err)
		}

		err = InitDB(":memory:")
		if err != nil {
			t.Fatalf("Failed to init database: %v", err)
		}
		defer DB.Close()

		err = RunMigrations(tempDir)
		if err == nil {
			t.Error("Expected error for invalid SQL in migration")
		}
	})

	t.Run("Skip non-SQL files", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a valid migration
		validMigration := filepath.Join(tempDir, "001_test.sql")
		err := os.WriteFile(validMigration, []byte("CREATE TABLE IF NOT EXISTS test (id TEXT);"), 0644)
		if err != nil {
			t.Fatalf("Failed to create migration file: %v", err)
		}

		// Create a non-SQL file
		nonSQLFile := filepath.Join(tempDir, "readme.txt")
		err = os.WriteFile(nonSQLFile, []byte("This is not a migration"), 0644)
		if err != nil {
			t.Fatalf("Failed to create non-SQL file: %v", err)
		}

		err = InitDB(":memory:")
		if err != nil {
			t.Fatalf("Failed to init database: %v", err)
		}
		defer DB.Close()

		err = RunMigrations(tempDir)
		if err != nil {
			t.Fatalf("Failed to run migrations: %v", err)
		}

		// Should only have applied the SQL file
		var tableName string
		err = DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='test'").Scan(&tableName)
		if err != nil {
			t.Fatalf("Failed to query table: %v", err)
		}

		if tableName != "test" {
			t.Errorf("Expected table 'test' to be created")
		}
	})
}

func TestNewRepository(t *testing.T) {
	err := InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	defer DB.Close()

	repo := NewRepository(DB)

	if repo == nil {
		t.Fatal("Expected repository to be created")
	}

	if repo.DB == nil {
		t.Fatal("Expected repository DB to be initialized")
	}
}
