package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

func TestRouterRegistersEveryEndpoint(t *testing.T) {
	t.Parallel()

	want := []string{
		"DELETE /portfolio-items/{id}",
		"DELETE /portfolios/{id}",
		"GET /health",
		"GET /portfolios/",
		"GET /portfolios/{id}/overview",
		"GET /portfolios/{id}/risks",
		"GET /portfolios/{id}/structure",
		"GET /portfolios/{id}/transactions",
		"GET /portfolios/{id}/yield",
		"GET /risk-free-rates/",
		"GET /risk-free-rates/{id}",
		"POST /auth/login",
		"POST /auth/register",
		"POST /auth/request-password-reset",
		"POST /auth/reset-password",
		"POST /auth/verify-otc",
		"POST /portfolios/",
		"POST /portfolios/{id}/items",
		"POST /portfolios/{id}/transactions",
		"POST /risk-free-rates/",
		"PUT /portfolios/{id}",
	}

	var got []string
	err := chi.Walk(newRouter(), func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		got = append(got, method+" "+route)
		return nil
	})
	if err != nil {
		t.Fatalf("walk routes: %v", err)
	}

	sort.Strings(want)
	sort.Strings(got)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("registered routes differ\nwant:\n%s\n\ngot:\n%s", strings.Join(want, "\n"), strings.Join(got, "\n"))
	}
}

func TestPublicEndpointsRejectInvalidRequests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
	}{
		{name: "register", path: "/auth/register", body: "{", wantStatus: http.StatusBadRequest},
		{name: "login", path: "/auth/login", body: "{", wantStatus: http.StatusBadRequest},
		{name: "request password reset", path: "/auth/request-password-reset", body: "{", wantStatus: http.StatusBadRequest},
		{name: "verify one-time code", path: "/auth/verify-otc", body: "{", wantStatus: http.StatusBadRequest},
		{name: "reset password", path: "/auth/reset-password", body: `{}`, wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()
			newRouter().ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status: want %d, got %d; body: %s", tt.wantStatus, recorder.Code, recorder.Body.String())
			}
			assertJSONResponse(t, recorder)
		})
	}
}

func TestEveryProtectedEndpointRequiresAuthentication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "create risk-free rate", method: http.MethodPost, path: "/risk-free-rates/"},
		{name: "list risk-free rates", method: http.MethodGet, path: "/risk-free-rates/"},
		{name: "read risk-free rate", method: http.MethodGet, path: "/risk-free-rates/1"},
		{name: "create portfolio", method: http.MethodPost, path: "/portfolios/"},
		{name: "list portfolios", method: http.MethodGet, path: "/portfolios/"},
		{name: "update portfolio", method: http.MethodPut, path: "/portfolios/1"},
		{name: "delete portfolio", method: http.MethodDelete, path: "/portfolios/1"},
		{name: "add portfolio item", method: http.MethodPost, path: "/portfolios/1/items"},
		{name: "list portfolio transactions", method: http.MethodGet, path: "/portfolios/1/transactions"},
		{name: "add portfolio transaction", method: http.MethodPost, path: "/portfolios/1/transactions"},
		{name: "portfolio structure", method: http.MethodGet, path: "/portfolios/1/structure"},
		{name: "portfolio overview", method: http.MethodGet, path: "/portfolios/1/overview"},
		{name: "portfolio yield", method: http.MethodGet, path: "/portfolios/1/yield"},
		{name: "portfolio risks", method: http.MethodGet, path: "/portfolios/1/risks"},
		{name: "delete portfolio item", method: http.MethodDelete, path: "/portfolio-items/1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{}`))
			recorder := httptest.NewRecorder()
			newRouter().ServeHTTP(recorder, req)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status: want %d, got %d; body: %s", http.StatusUnauthorized, recorder.Code, recorder.Body.String())
			}
			assertJSONResponse(t, recorder)
			if !strings.Contains(recorder.Body.String(), "Missing authorization header") {
				t.Fatalf("unexpected response body: %s", recorder.Body.String())
			}
		})
	}
}

func TestProtectedEndpointsValidateInputAfterAuthentication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "create risk-free rate", method: http.MethodPost, path: "/risk-free-rates/", body: `{}`},
		{name: "read risk-free rate", method: http.MethodGet, path: "/risk-free-rates/not-an-id"},
		{name: "create portfolio", method: http.MethodPost, path: "/portfolios/", body: `{}`},
		{name: "update portfolio", method: http.MethodPut, path: "/portfolios/not-an-id", body: `{}`},
		{name: "delete portfolio", method: http.MethodDelete, path: "/portfolios/not-an-id"},
		{name: "add portfolio item", method: http.MethodPost, path: "/portfolios/not-an-id/items", body: `{}`},
		{name: "list portfolio transactions", method: http.MethodGet, path: "/portfolios/not-an-id/transactions"},
		{name: "add portfolio transaction", method: http.MethodPost, path: "/portfolios/not-an-id/transactions", body: `{}`},
		{name: "portfolio structure", method: http.MethodGet, path: "/portfolios/not-an-id/structure"},
		{name: "portfolio overview", method: http.MethodGet, path: "/portfolios/not-an-id/overview"},
		{name: "portfolio yield", method: http.MethodGet, path: "/portfolios/not-an-id/yield"},
		{name: "portfolio risks", method: http.MethodGet, path: "/portfolios/not-an-id/risks"},
		{name: "delete portfolio item", method: http.MethodDelete, path: "/portfolio-items/not-an-id"},
	}

	token := signedAccessToken(t, 42)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Authorization", "Bearer "+token)
			recorder := httptest.NewRecorder()
			newRouter().ServeHTTP(recorder, req)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status: want %d, got %d; body: %s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
			}
			assertJSONResponse(t, recorder)
		})
	}
}

func signedAccessToken(t *testing.T, userID int32) string {
	t.Helper()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_fallback_secret_key"
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign access token: %v", err)
	}
	return signed
}

func assertJSONResponse(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()

	if got := recorder.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("Content-Type: want application/json, got %q", got)
	}
	if recorder.Body.Len() == 0 {
		t.Fatal("expected a JSON response body")
	}
}
