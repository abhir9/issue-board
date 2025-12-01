package main

import (
	"net/http"
	"testing"

	"github.com/abhir9/issue-board/api/internal/config"
)

func TestSetupLogger(t *testing.T) {
	logger := setupLogger()
	if logger == nil {
		t.Error("setupLogger() returned nil logger")
	}
}

func TestSetupDatabase(t *testing.T) {
	// Use a test database path
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Path:             ":memory:",
			MaxOpenConns:     1,
			MaxIdleConns:     1,
			ConnMaxLifetime:  0,
			MigrationDir:     "../../migrations",
		},
	}

	err := setupDatabase(cfg)
	if err != nil {
		t.Errorf("setupDatabase() failed: %v", err)
	}
}

func TestSetupRouter(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			AllowedOrigins: []string{"*"},
		},
		Auth: config.AuthConfig{
			APIKey: "test-key",
		},
	}

	router := setupRouter(cfg)
	if router == nil {
		t.Error("setupRouter() returned nil router")
	}
}

func TestSetupServer(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:         "localhost",
			Port:         "8080",
			ReadTimeout:  30,
			WriteTimeout: 30,
		},
	}

	// Create a dummy handler
	var handler http.Handler = http.NewServeMux()

	server := setupServer(cfg, handler)
	if server == nil {
		t.Error("setupServer() returned nil server")
	}

	expectedAddr := cfg.Server.Host + ":" + cfg.Server.Port
	if server.Addr != expectedAddr {
		t.Errorf("setupServer() Addr = %v, want %v", server.Addr, expectedAddr)
	}

	if server.Handler != handler {
		t.Error("setupServer() Handler not set correctly")
	}
}