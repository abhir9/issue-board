package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Save original env vars
	originalAPIKey := os.Getenv("API_KEY")
	originalPort := os.Getenv("PORT")
	originalDBPath := os.Getenv("DATABASE_PATH")

	// Restore after test
	defer func() {
		os.Setenv("API_KEY", originalAPIKey)
		os.Setenv("PORT", originalPort)
		os.Setenv("DATABASE_PATH", originalDBPath)
	}()

	t.Run("Load with required API_KEY", func(t *testing.T) {
		os.Setenv("API_KEY", "test-api-key-123")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if cfg.Auth.APIKey != "test-api-key-123" {
			t.Errorf("Expected API key 'test-api-key-123', got '%s'", cfg.Auth.APIKey)
		}
	})

	t.Run("Load fails without API_KEY", func(t *testing.T) {
		os.Unsetenv("API_KEY")

		_, err := Load()
		if err == nil {
			t.Error("Expected error when API_KEY is missing, got nil")
		}
	})

	t.Run("Load with custom port", func(t *testing.T) {
		os.Setenv("API_KEY", "test-key")
		os.Setenv("PORT", "9000")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if cfg.Server.Port != "9000" {
			t.Errorf("Expected port '9000', got '%s'", cfg.Server.Port)
		}
	})

	t.Run("Load with default values", func(t *testing.T) {
		os.Setenv("API_KEY", "test-key")
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_PATH")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if cfg.Server.Port != "8080" {
			t.Errorf("Expected default port '8080', got '%s'", cfg.Server.Port)
		}

		if cfg.Database.Path != "./issues.db" {
			t.Errorf("Expected default db path './issues.db', got '%s'", cfg.Database.Path)
		}
	})

	t.Run("Load with custom database settings", func(t *testing.T) {
		os.Setenv("API_KEY", "test-key")
		os.Setenv("DATABASE_PATH", "/custom/path/db.sqlite")
		os.Setenv("DB_MAX_OPEN_CONNS", "50")
		os.Setenv("DB_MAX_IDLE_CONNS", "10")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if cfg.Database.Path != "/custom/path/db.sqlite" {
			t.Errorf("Expected db path '/custom/path/db.sqlite', got '%s'", cfg.Database.Path)
		}

		if cfg.Database.MaxOpenConns != 50 {
			t.Errorf("Expected max open conns 50, got %d", cfg.Database.MaxOpenConns)
		}

		if cfg.Database.MaxIdleConns != 10 {
			t.Errorf("Expected max idle conns 10, got %d", cfg.Database.MaxIdleConns)
		}
	})
}

func TestGetEnv(t *testing.T) {
	t.Run("Get existing env var", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test-value")
		defer os.Unsetenv("TEST_VAR")

		result := getEnv("TEST_VAR", "default")
		if result != "test-value" {
			t.Errorf("Expected 'test-value', got '%s'", result)
		}
	})

	t.Run("Get default value", func(t *testing.T) {
		result := getEnv("NON_EXISTING_VAR", "default-value")
		if result != "default-value" {
			t.Errorf("Expected 'default-value', got '%s'", result)
		}
	})
}

func TestGetInt(t *testing.T) {
	t.Run("Get valid integer", func(t *testing.T) {
		os.Setenv("TEST_INT", "42")
		defer os.Unsetenv("TEST_INT")

		result := getInt("TEST_INT", 10)
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("Get default for invalid integer", func(t *testing.T) {
		os.Setenv("TEST_INT", "not-a-number")
		defer os.Unsetenv("TEST_INT")

		result := getInt("TEST_INT", 10)
		if result != 10 {
			t.Errorf("Expected default 10, got %d", result)
		}
	})

	t.Run("Get default for missing var", func(t *testing.T) {
		result := getInt("NON_EXISTING_INT", 99)
		if result != 99 {
			t.Errorf("Expected default 99, got %d", result)
		}
	})
}

func TestGetDuration(t *testing.T) {
	t.Run("Get valid duration", func(t *testing.T) {
		os.Setenv("TEST_DURATION", "30s")
		defer os.Unsetenv("TEST_DURATION")

		result := getDuration("TEST_DURATION", 10*time.Second)
		if result != 30*time.Second {
			t.Errorf("Expected 30s, got %v", result)
		}
	})

	t.Run("Get default for invalid duration", func(t *testing.T) {
		os.Setenv("TEST_DURATION", "invalid")
		defer os.Unsetenv("TEST_DURATION")

		result := getDuration("TEST_DURATION", 10*time.Second)
		if result != 10*time.Second {
			t.Errorf("Expected default 10s, got %v", result)
		}
	})

	t.Run("Get default for missing var", func(t *testing.T) {
		result := getDuration("NON_EXISTING_DURATION", 5*time.Minute)
		if result != 5*time.Minute {
			t.Errorf("Expected default 5m, got %v", result)
		}
	})
}

func TestGetKeepAliveURL(t *testing.T) {
	// Save original env vars
	originalRender := os.Getenv("RENDER_EXTERNAL_URL")
	originalApp := os.Getenv("APP_URL")

	defer func() {
		os.Setenv("RENDER_EXTERNAL_URL", originalRender)
		os.Setenv("APP_URL", originalApp)
	}()

	t.Run("Get RENDER_EXTERNAL_URL", func(t *testing.T) {
		os.Setenv("RENDER_EXTERNAL_URL", "https://render.example.com")
		os.Unsetenv("APP_URL")

		result := getKeepAliveURL()
		if result != "https://render.example.com" {
			t.Errorf("Expected 'https://render.example.com', got '%s'", result)
		}
	})

	t.Run("Get APP_URL when RENDER not set", func(t *testing.T) {
		os.Unsetenv("RENDER_EXTERNAL_URL")
		os.Setenv("APP_URL", "https://app.example.com")

		result := getKeepAliveURL()
		if result != "https://app.example.com" {
			t.Errorf("Expected 'https://app.example.com', got '%s'", result)
		}
	})

	t.Run("Return empty when both not set", func(t *testing.T) {
		os.Unsetenv("RENDER_EXTERNAL_URL")
		os.Unsetenv("APP_URL")

		result := getKeepAliveURL()
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})

	t.Run("Prefer RENDER_EXTERNAL_URL", func(t *testing.T) {
		os.Setenv("RENDER_EXTERNAL_URL", "https://render.example.com")
		os.Setenv("APP_URL", "https://app.example.com")

		result := getKeepAliveURL()
		if result != "https://render.example.com" {
			t.Errorf("Expected RENDER_EXTERNAL_URL to take precedence, got '%s'", result)
		}
	})
}

func TestGetAllowedOrigins(t *testing.T) {
	originalOrigins := os.Getenv("ALLOWED_ORIGINS")
	defer os.Setenv("ALLOWED_ORIGINS", originalOrigins)

	t.Run("Get default origins", func(t *testing.T) {
		os.Unsetenv("ALLOWED_ORIGINS")

		result := getAllowedOrigins()
		if len(result) != 2 {
			t.Errorf("Expected 2 default origins, got %d", len(result))
		}

		expectedOrigins := []string{
			"http://localhost:3000",
			"https://issue-board-front.netlify.app",
		}

		for i, expected := range expectedOrigins {
			if result[i] != expected {
				t.Errorf("Expected origin '%s', got '%s'", expected, result[i])
			}
		}
	})

	t.Run("Get custom origin", func(t *testing.T) {
		os.Setenv("ALLOWED_ORIGINS", "https://custom.example.com")

		result := getAllowedOrigins()
		if len(result) != 1 {
			t.Errorf("Expected 1 custom origin, got %d", len(result))
		}

		if result[0] != "https://custom.example.com" {
			t.Errorf("Expected 'https://custom.example.com', got '%s'", result[0])
		}
	})
}
