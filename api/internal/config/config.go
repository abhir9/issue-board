package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
}

type ServerConfig struct {
	Port              string
	Host              string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	ShutdownTimeout   time.Duration
	EnableKeepAlive   bool
	KeepAliveURL      string
	AllowedOrigins    []string
}

type DatabaseConfig struct {
	Path            string
	MigrationDir    string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type AuthConfig struct {
	APIKey string
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", "8080"),
			Host:            getEnv("HOST", "0.0.0.0"),
			ReadTimeout:     getDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: getDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
			EnableKeepAlive: getEnv("ENABLE_KEEP_ALIVE", "false") == "true" || getEnv("RENDER", "") != "",
			KeepAliveURL:    getKeepAliveURL(),
			AllowedOrigins:  getAllowedOrigins(),
		},
		Database: DatabaseConfig{
			Path:            getEnv("DATABASE_PATH", "./issues.db"),
			MigrationDir:    getEnv("MIGRATION_DIR", "./migrations"),
			MaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Auth: AuthConfig{
			APIKey: getEnv("API_KEY", ""),
		},
	}

	// Validate required fields
	if cfg.Auth.APIKey == "" {
		return nil, fmt.Errorf("API_KEY environment variable is required")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}

func getDuration(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	duration, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return duration
}

func getKeepAliveURL() string {
	if url := os.Getenv("RENDER_EXTERNAL_URL"); url != "" {
		return url
	}
	if url := os.Getenv("APP_URL"); url != "" {
		return url
	}
	return ""
}

func getAllowedOrigins() []string {
	origins := os.Getenv("ALLOWED_ORIGINS")
	if origins == "" {
		// Default allowed origins
		return []string{"http://localhost:3000", "https://issue-board-front.netlify.app"}
	}
	// Split by comma if multiple origins
	return []string{origins}
}
