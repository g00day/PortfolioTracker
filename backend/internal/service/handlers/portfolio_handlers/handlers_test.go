package portfoliohandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pashagolub/pgxmock/v5"

	"portfolio-tracker/internal/adapter/db/repos"
	"portfolio-tracker/internal/service/middlewares"
)

const testUserID int32 = 42

func TestCreatePortfolioSuccess(t *testing.T) {
	mock := usePortfolioMockDB(t)
	mock.ExpectQuery("INSERT INTO portfolios").
		WithArgs(testUserID, int32(2), "Main").
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(int32(5)))

	recorder := performPortfolioRequest(http.MethodPost, "/portfolios/", "/portfolios/", CreatePortfolio, `{"name":"Main","risk_free_rate_id":2}`)

	assertPortfolioStatus(t, recorder, http.StatusCreated)
	var response APIResponse
	decodePortfolioJSON(t, recorder, &response)
	if response.Status != "success" || !strings.Contains(response.Message, "5") {
		t.Fatalf("unexpected response: %+v", response)
	}
	assertPortfolioExpectations(t, mock)
}

func TestReadPortfoliosSuccess(t *testing.T) {
	mock := usePortfolioMockDB(t)
	id := int32(5)
	createdAt := time.Now().UTC().Truncate(time.Second)
	mock.ExpectQuery("SELECT id, user_id, risk_free_rate_id, name, created_at").
		WithArgs(testUserID).
		WillReturnRows(mock.NewRows([]string{"id", "user_id", "risk_free_rate_id", "name", "created_at"}).
			AddRow(&id, testUserID, int32(2), "Main", &createdAt))

	recorder := performPortfolioRequest(http.MethodGet, "/portfolios/", "/portfolios/", ReadPortfolios, "")

	assertPortfolioStatus(t, recorder, http.StatusOK)
	var response []PortfolioResponse
	decodePortfolioJSON(t, recorder, &response)
	if len(response) != 1 || response[0].ID != id || response[0].Name != "Main" || !response[0].CreatedAt.Equal(createdAt) {
		t.Fatalf("unexpected response: %+v", response)
	}
	assertPortfolioExpectations(t, mock)
}

func TestUpdatePortfolioSuccess(t *testing.T) {
	mock := usePortfolioMockDB(t)
	expectOwnedPortfolio(mock, 5)
	mock.ExpectExec("UPDATE portfolios").
		WithArgs("Long term", int32(3), int32(5)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	recorder := performPortfolioRequest(http.MethodPut, "/portfolios/5", "/portfolios/{id}", UpdatePortfolio, `{"name":"Long term","risk_free_rate_id":3}`)

	assertSuccessAPIResponse(t, recorder, http.StatusOK)
	assertPortfolioExpectations(t, mock)
}

func TestDeletePortfolioSuccess(t *testing.T) {
	mock := usePortfolioMockDB(t)
	expectOwnedPortfolio(mock, 5)
	mock.ExpectExec("DELETE FROM portfolios").
		WithArgs(int32(5)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	recorder := performPortfolioRequest(http.MethodDelete, "/portfolios/5", "/portfolios/{id}", DeletePortfolio, "")

	assertSuccessAPIResponse(t, recorder, http.StatusOK)
	assertPortfolioExpectations(t, mock)
}

func TestAddPortfolioItemSuccess(t *testing.T) {
	mock := usePortfolioMockDB(t)
	expectOwnedPortfolio(mock, 5)
	mock.ExpectQuery("INSERT INTO portfolio_items").
		WithArgs(int32(5), pgxmock.AnyArg(), pgxmock.AnyArg(), float32(100)).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(int32(9)))
	mock.ExpectQuery("INSERT INTO transactions").
		WithArgs(int32(5), int32(9), "deposit", float64(100), float64(1), pgxmock.AnyArg()).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(int32(17)))

	recorder := performPortfolioRequest(http.MethodPost, "/portfolios/5/items", "/portfolios/{id}/items", AddPortfolioItem, `{"risk_free_rate_id":2,"total_amount":100}`)

	assertPortfolioStatus(t, recorder, http.StatusCreated)
	var response APIResponse
	decodePortfolioJSON(t, recorder, &response)
	if response.Status != "success" || !strings.Contains(response.Message, "9") {
		t.Fatalf("unexpected response: %+v", response)
	}
	assertPortfolioExpectations(t, mock)
}

func TestReadPortfolioTransactionsSuccess(t *testing.T) {
	mock := usePortfolioMockDB(t)
	expectOwnedPortfolio(mock, 5)
	transactionID := int32(17)
	processedAt := time.Now().UTC().Truncate(time.Second)
	mock.ExpectQuery("SELECT id, portfolio_id, portfolio_item_id, type, amount, price_usd, processed_at").
		WithArgs(int32(5)).
		WillReturnRows(mock.NewRows([]string{"id", "portfolio_id", "portfolio_item_id", "type", "amount", "price_usd", "processed_at"}).
			AddRow(&transactionID, int32(5), int32(9), "deposit", float64(0.5), float64(81000), &processedAt))

	recorder := performPortfolioRequest(http.MethodGet, "/portfolios/5/transactions", "/portfolios/{id}/transactions", ReadPortfolioTransactions, "")

	assertPortfolioStatus(t, recorder, http.StatusOK)
	var response []TransactionResponse
	decodePortfolioJSON(t, recorder, &response)
	if len(response) != 1 || response[0].ID != transactionID || response[0].PortfolioItemID != 9 || !response[0].ProcessedAt.Equal(processedAt) {
		t.Fatalf("unexpected response: %+v", response)
	}
	assertPortfolioExpectations(t, mock)
}

func TestAddPortfolioTransactionSuccess(t *testing.T) {
	mock := usePortfolioMockDB(t)
	expectOwnedPortfolio(mock, 5)
	ticker := "BTCUSDT"
	mock.ExpectQuery("SELECT id, portfolio_id, ticker, risk_free_rate_id, total_amount").
		WithArgs(int32(9)).
		WillReturnRows(mock.NewRows([]string{"id", "portfolio_id", "ticker", "risk_free_rate_id", "total_amount"}).
			AddRow(int32(9), int32(5), &ticker, nil, float32(1.25)))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE portfolio_items SET total_amount").
		WithArgs(float32(1.75), int32(9)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectQuery("INSERT INTO transactions").
		WithArgs(int32(5), int32(9), "deposit", float64(0.5), float64(81000), pgxmock.AnyArg()).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(int32(18)))
	mock.ExpectCommit()

	recorder := performPortfolioRequest(http.MethodPost, "/portfolios/5/transactions", "/portfolios/{id}/transactions", AddPortfolioTransaction, `{"portfolio_item_id":9,"type":"deposit","amount":0.5,"price_usd":81000}`)

	assertPortfolioStatus(t, recorder, http.StatusCreated)
	var response APIResponse
	decodePortfolioJSON(t, recorder, &response)
	if response.Status != "success" || !strings.Contains(response.Message, "18") {
		t.Fatalf("unexpected response: %+v", response)
	}
	assertPortfolioExpectations(t, mock)
}

func TestReadPortfolioStructureEmptySuccess(t *testing.T) {
	mock := usePortfolioMockDB(t)
	expectOwnedPortfolio(mock, 5)
	expectEmptyPortfolioItems(mock, 5)

	recorder := performPortfolioRequest(http.MethodGet, "/portfolios/5/structure", "/portfolios/{id}/structure", ReadPortfolioStructure, "")

	assertPortfolioStatus(t, recorder, http.StatusOK)
	var response PortfolioStructureResponse
	decodePortfolioJSON(t, recorder, &response)
	if response.PortfolioStructure == nil || len(response.PortfolioStructure) != 0 {
		t.Fatalf("unexpected response: %+v", response)
	}
	assertPortfolioExpectations(t, mock)
}

func TestReadPortfolioOverviewEmptySuccess(t *testing.T) {
	mock := usePortfolioMockDB(t)
	expectOwnedPortfolio(mock, 5)
	expectEmptyPortfolioItems(mock, 5)

	recorder := performPortfolioRequest(http.MethodGet, "/portfolios/5/overview", "/portfolios/{id}/overview", ReadPortfolioOverview, "")

	assertPortfolioStatus(t, recorder, http.StatusOK)
	var response PortfolioOverviewResponse
	decodePortfolioJSON(t, recorder, &response)
	if len(response.PLHistory["portfolio"]) != 1 || response.PLHistory["portfolio"][0] != 0 {
		t.Fatalf("unexpected response: %+v", response)
	}
	assertPortfolioExpectations(t, mock)
}

func TestReadPortfolioYieldEmptySuccess(t *testing.T) {
	mock := usePortfolioMockDB(t)
	expectOwnedPortfolio(mock, 5)
	expectEmptyPortfolioItems(mock, 5)

	recorder := performPortfolioRequest(http.MethodGet, "/portfolios/5/yield", "/portfolios/{id}/yield", ReadPortfolioYield, "")

	assertPortfolioStatus(t, recorder, http.StatusOK)
	var response PortfolioYieldResponse
	decodePortfolioJSON(t, recorder, &response)
	if response.PortfolioStructure == nil || len(response.PortfolioStructure) != 0 {
		t.Fatalf("unexpected response: %+v", response)
	}
	assertPortfolioExpectations(t, mock)
}

func TestGetPortfolioRiskMetricsEmptySuccess(t *testing.T) {
	mock := usePortfolioMockDB(t)
	expectOwnedPortfolio(mock, 5)
	expectEmptyPortfolioItems(mock, 5)

	recorder := performPortfolioRequest(http.MethodGet, "/portfolios/5/risks", "/portfolios/{id}/risks", GetPortfolioRiskMetrics, "")

	assertPortfolioStatus(t, recorder, http.StatusOK)
	var response PortfolioRiskMetricsResponse
	decodePortfolioJSON(t, recorder, &response)
	if response.PartialRisks == nil || len(response.PartialRisks) != 0 {
		t.Fatalf("unexpected response: %+v", response)
	}
	assertPortfolioExpectations(t, mock)
}

func usePortfolioMockDB(t *testing.T) pgxmock.PgxConnIface {
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

func expectOwnedPortfolio(mock pgxmock.PgxConnIface, id int32) {
	createdAt := time.Now().UTC()
	mock.ExpectQuery("SELECT id, user_id, risk_free_rate_id, name, created_at FROM portfolios WHERE id").
		WithArgs(id).
		WillReturnRows(mock.NewRows([]string{"id", "user_id", "risk_free_rate_id", "name", "created_at"}).
			AddRow(&id, testUserID, int32(2), "Main", &createdAt))
}

func expectEmptyPortfolioItems(mock pgxmock.PgxConnIface, portfolioID int32) {
	mock.ExpectQuery("SELECT id, portfolio_id, ticker, risk_free_rate_id, total_amount").
		WithArgs(portfolioID).
		WillReturnRows(mock.NewRows([]string{"id", "portfolio_id", "ticker", "risk_free_rate_id", "total_amount"}))
}

func performPortfolioRequest(method, path, pattern string, handler http.HandlerFunc, body string) *httptest.ResponseRecorder {
	router := chi.NewRouter()
	router.Method(method, pattern, handler)
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request = request.WithContext(context.WithValue(request.Context(), middlewares.UserIDKey, testUserID))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func assertSuccessAPIResponse(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus int) {
	t.Helper()
	assertPortfolioStatus(t, recorder, wantStatus)
	var response APIResponse
	decodePortfolioJSON(t, recorder, &response)
	if response.Status != "success" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func assertPortfolioStatus(t *testing.T, recorder *httptest.ResponseRecorder, want int) {
	t.Helper()
	if recorder.Code != want {
		t.Fatalf("status: want %d, got %d; body: %s", want, recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type: want application/json, got %q", got)
	}
}

func decodePortfolioJSON(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(recorder.Body).Decode(target); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
}

func assertPortfolioExpectations(t *testing.T, mock pgxmock.PgxConnIface) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet database expectations: %v", err)
	}
}
