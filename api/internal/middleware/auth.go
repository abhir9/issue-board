package middleware

import (
	"encoding/json"
	"net/http"
)

// APIKeyAuth creates a middleware that checks for a valid API key in the header
func APIKeyAuth(validAPIKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" || apiKey != validAPIKey {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Unauthorized: Invalid or missing API key",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
