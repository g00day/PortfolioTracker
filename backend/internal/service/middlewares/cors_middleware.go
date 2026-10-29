package middlewares

import (
	"net/http"
	"os"
	"strings"
)

// CORSMiddleware allows the browser client to call the API directly when the
// frontend and backend are served from different local origins.
func CORSMiddleware(next http.Handler) http.Handler {
	defaults := "http://localhost:5173,http://127.0.0.1:5173,http://localhost:4173,http://127.0.0.1:4173"
	configured := os.Getenv("CORS_ALLOWED_ORIGINS")
	if configured == "" {
		configured = defaults
	}

	allowed := make(map[string]bool)
	for _, origin := range strings.Split(configured, ",") {
		allowed[strings.TrimSpace(origin)] = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
