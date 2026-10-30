package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	db "portfolio-tracker/internal/adapter/db"
)

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

func checkDatabase(ctx context.Context) error {
	dbConn, err := db.GetDBConnection()
	if err != nil {
		return err
	}
	defer dbConn.Close(ctx)

	return dbConn.Ping(ctx)
}

func newHealthCheckHandler(databaseCheck func(context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response := HealthResponse{
			Status:    "ok",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Checks:    make(map[string]string),
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := databaseCheck(ctx); err != nil {
			response.Status = "DOWN"
			response.Checks["database"] = "unhealthy: " + err.Error()

			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(response)
			return
		}

		response.Checks["database"] = "healthy"

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	newHealthCheckHandler(checkDatabase).ServeHTTP(w, r)
}
