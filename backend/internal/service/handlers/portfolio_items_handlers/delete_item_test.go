package p_handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pashagolub/pgxmock/v5"

	"portfolio-tracker/internal/adapter/db/repos"
	"portfolio-tracker/internal/service/middlewares"
)

func TestDeletePortfolioItemSuccess(t *testing.T) {
	mock := usePortfolioItemMockDB(t)
	ticker := "BTCUSDT"
	portfolioID := int32(4)
	portfolioCreatedAt := time.Now().UTC()
	mock.ExpectQuery("SELECT id, portfolio_id, ticker, risk_free_rate_id, total_amount").
		WithArgs(int32(9)).
		WillReturnRows(mock.NewRows([]string{"id", "portfolio_id", "ticker", "risk_free_rate_id", "total_amount"}).
			AddRow(int32(9), portfolioID, &ticker, nil, float32(1.25)))
	mock.ExpectQuery("SELECT id, user_id, risk_free_rate_id, name, created_at FROM portfolios").
		WithArgs(portfolioID).
		WillReturnRows(mock.NewRows([]string{"id", "user_id", "risk_free_rate_id", "name", "created_at"}).
			AddRow(&portfolioID, int32(42), int32(2), "Main", &portfolioCreatedAt))
	mock.ExpectExec("DELETE FROM portfolio_items").
		WithArgs(int32(9)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	router := chi.NewRouter()
	router.Delete("/{id}", DeletePortfolioItem)
	request := httptest.NewRequest(http.MethodDelete, "/9", nil)
	request = request.WithContext(context.WithValue(request.Context(), middlewares.UserIDKey, int32(42)))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status: want %d, got %d; body: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	var response APIResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if response.Status != "success" {
		t.Fatalf("unexpected response: %+v", response)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet database expectations: %v", err)
	}
}

func usePortfolioItemMockDB(t *testing.T) pgxmock.PgxConnIface {
	t.Helper()
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("create pgx mock: %v", err)
	}
	oldGetDB := getDBConnection
	getDBConnection = func() (repos.DBConnection, error) { return mock, nil }
	t.Cleanup(func() {
		getDBConnection = oldGetDB
		_ = mock.Close(context.Background())
	})
	return mock
}
