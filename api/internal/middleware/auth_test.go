package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKeyAuth(t *testing.T) {
	validAPIKey := "test-api-key-123"

	// Create a test handler that returns 200 OK
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap with auth middleware
	authMiddleware := APIKeyAuth(validAPIKey)
	handler := authMiddleware(testHandler)

	t.Run("Valid API Key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", validAPIKey)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if w.Body.String() != "success" {
			t.Errorf("Expected 'success', got '%s'", w.Body.String())
		}
	})

	t.Run("Invalid API Key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "invalid-key")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["error"] == "" {
			t.Error("Expected error message in response")
		}
	})

	t.Run("Missing API Key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("Empty API Key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("Case Sensitive API Key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "TEST-API-KEY-123")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 (case sensitive), got %d", w.Code)
		}
	})

	t.Run("Response Content Type", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "wrong-key")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}
	})
}
