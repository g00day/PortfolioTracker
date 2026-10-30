package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthCheckHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		databaseCheck  func(context.Context) error
		wantStatusCode int
		wantStatus     string
		wantCheck      string
	}{
		{
			name: "healthy database",
			databaseCheck: func(context.Context) error {
				return nil
			},
			wantStatusCode: http.StatusOK,
			wantStatus:     "ok",
			wantCheck:      "healthy",
		},
		{
			name: "database unavailable",
			databaseCheck: func(context.Context) error {
				return errors.New("connection refused")
			},
			wantStatusCode: http.StatusServiceUnavailable,
			wantStatus:     "DOWN",
			wantCheck:      "unhealthy: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/health", nil)
			newHealthCheckHandler(tt.databaseCheck).ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatusCode {
				t.Fatalf("status: want %d, got %d", tt.wantStatusCode, recorder.Code)
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type: want application/json, got %q", got)
			}

			var response HealthResponse
			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if response.Status != tt.wantStatus {
				t.Errorf("status body: want %q, got %q", tt.wantStatus, response.Status)
			}
			if response.Checks["database"] != tt.wantCheck {
				t.Errorf("database check: want %q, got %q", tt.wantCheck, response.Checks["database"])
			}
			if _, err := time.Parse(time.RFC3339, response.Timestamp); err != nil {
				t.Errorf("timestamp is not RFC3339: %q", response.Timestamp)
			}
		})
	}
}

func TestHealthCheckHandlerUsesRequestContext(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	requestContext, cancel := context.WithCancel(request.Context())
	cancel()
	request = request.WithContext(requestContext)

	recorder := httptest.NewRecorder()
	newHealthCheckHandler(func(ctx context.Context) error {
		return ctx.Err()
	}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status: want %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), context.Canceled.Error()) {
		t.Fatalf("response does not include context error: %s", recorder.Body.String())
	}
}
