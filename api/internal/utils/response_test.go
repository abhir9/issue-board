package utils

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	t.Run("Write valid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]interface{}{
			"message": "Hello",
			"count":   42,
		}

		WriteJSON(w, 200, data)

		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["message"] != "Hello" {
			t.Errorf("Expected message 'Hello', got '%v'", response["message"])
		}

		if response["count"] != float64(42) {
			t.Errorf("Expected count 42, got %v", response["count"])
		}
	})

	t.Run("Write array JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := []string{"item1", "item2", "item3"}

		WriteJSON(w, 200, data)

		var response []string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if len(response) != 3 {
			t.Errorf("Expected 3 items, got %d", len(response))
		}
	})

	t.Run("Write null JSON", func(t *testing.T) {
		w := httptest.NewRecorder()

		WriteJSON(w, 200, nil)

		if w.Body.String() != "null\n" {
			t.Errorf("Expected 'null', got '%s'", w.Body.String())
		}
	})
}

func TestWriteError(t *testing.T) {
	t.Run("Write error with details", func(t *testing.T) {
		w := httptest.NewRecorder()
		details := map[string]interface{}{
			"field": "email",
			"issue": "invalid format",
		}

		WriteError(w, 400, "Validation failed", details)

		if w.Code != 400 {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		var response ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Error != "Validation failed" {
			t.Errorf("Expected error 'Validation failed', got '%s'", response.Error)
		}

		if response.Details["field"] != "email" {
			t.Errorf("Expected field 'email', got '%v'", response.Details["field"])
		}
	})

	t.Run("Write error without details", func(t *testing.T) {
		w := httptest.NewRecorder()

		WriteError(w, 500, "Internal server error", nil)

		if w.Code != 500 {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		var response ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Error != "Internal server error" {
			t.Errorf("Expected error 'Internal server error', got '%s'", response.Error)
		}

		if response.Details != nil {
			t.Errorf("Expected nil details, got %v", response.Details)
		}
	})

	t.Run("Write 404 error", func(t *testing.T) {
		w := httptest.NewRecorder()

		WriteError(w, 404, "Resource not found", nil)

		if w.Code != 404 {
			t.Errorf("Expected status 404, got %d", w.Code)
		}

		var response ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Error != "Resource not found" {
			t.Errorf("Expected error 'Resource not found', got '%s'", response.Error)
		}
	})

	t.Run("Write unauthorized error", func(t *testing.T) {
		w := httptest.NewRecorder()

		WriteError(w, 401, "Unauthorized", nil)

		if w.Code != 401 {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})
}

func TestErrorResponse(t *testing.T) {
	t.Run("ErrorResponse structure", func(t *testing.T) {
		details := map[string]interface{}{
			"code":    "ERR_001",
			"message": "Something went wrong",
		}

		errorResp := ErrorResponse{
			Error:   "Test error",
			Details: details,
		}

		if errorResp.Error != "Test error" {
			t.Errorf("Expected error 'Test error', got '%s'", errorResp.Error)
		}

		if errorResp.Details["code"] != "ERR_001" {
			t.Error("Expected details to contain code")
		}
	})
}
