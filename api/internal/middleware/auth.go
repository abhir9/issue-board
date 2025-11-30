package middleware

import (
	"net/http"
)

// APIKeyAuth is a middleware that checks for a valid API key in the header
func APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For simplicity, we hardcode the key here as requested.
		// In a real app, this should come from environment variables or a database.
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "J3yPAMuS0j5w4AWj6P0bh2l7prZKBSq6" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
