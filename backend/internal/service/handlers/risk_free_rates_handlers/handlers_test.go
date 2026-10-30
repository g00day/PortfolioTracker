package riskfreerateshandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/pashagolub/pgxmock/v5"

	"portfolio-tracker/internal/adapter/db/repos"
)

func TestCreateRiskFreeRateHandlerSuccess(t *testing.T) {
	mock := useRiskFreeMockDB(t)
	mock.ExpectQuery("INSERT INTO risk_free_rate").
		WithArgs("US Treasury", 0.045).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(int32(3)))

	request := httptest.NewRequest(http.MethodPost, "/risk-free-rates/", strings.NewReader(`{"name":"US Treasury","value":0.045}`))
	recorder := httptest.NewRecorder()
	CreateRiskFreeRateHandler(recorder, request)

	assertRiskFreeStatus(t, recorder, http.StatusCreated)
	var response APIResponse
	decodeRiskFreeJSON(t, recorder, &response)
	if response.Status != "success" || !strings.Contains(response.Message, "3") {
		t.Fatalf("unexpected response: %+v", response)
	}
	assertRiskFreeExpectations(t, mock)
}

func TestReadRiskFreeRatesSuccess(t *testing.T) {
	mock := useRiskFreeMockDB(t)
	id1, id2 := int32(1), int32(2)
	mock.ExpectQuery("SELECT id, name, value FROM risk_free_rate").
		WillReturnRows(mock.NewRows([]string{"id", "name", "value"}).
			AddRow(&id1, "US Treasury", 0.045).
			AddRow(&id2, "EUR Bond", 0.031))

	request := httptest.NewRequest(http.MethodGet, "/risk-free-rates/", nil)
	recorder := httptest.NewRecorder()
	ReadRiskFreeRates(recorder, request)

	assertRiskFreeStatus(t, recorder, http.StatusOK)
	var response []RiskFreeRateResponse
	decodeRiskFreeJSON(t, recorder, &response)
	if len(response) != 2 || response[0].ID != id1 || response[1].Name != "EUR Bond" {
		t.Fatalf("unexpected response: %+v", response)
	}
	assertRiskFreeExpectations(t, mock)
}

func TestReadRiskFreeRateSuccess(t *testing.T) {
	mock := useRiskFreeMockDB(t)
	id := int32(3)
	mock.ExpectQuery("SELECT id, name, value FROM risk_free_rate WHERE id").
		WithArgs(int32(3)).
		WillReturnRows(mock.NewRows([]string{"id", "name", "value"}).AddRow(&id, "US Treasury", 0.045))

	router := chi.NewRouter()
	router.Get("/{id}", ReadRiskFreeRate)
	request := httptest.NewRequest(http.MethodGet, "/3", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	assertRiskFreeStatus(t, recorder, http.StatusOK)
	var response ReadRiskFreeRateResponse
	decodeRiskFreeJSON(t, recorder, &response)
	if response.ID != id || response.Name != "US Treasury" || response.Value != 0.045 {
		t.Fatalf("unexpected response: %+v", response)
	}
	assertRiskFreeExpectations(t, mock)
}

func useRiskFreeMockDB(t *testing.T) pgxmock.PgxConnIface {
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

func assertRiskFreeStatus(t *testing.T, recorder *httptest.ResponseRecorder, want int) {
	t.Helper()
	if recorder.Code != want {
		t.Fatalf("status: want %d, got %d; body: %s", want, recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type: want application/json, got %q", got)
	}
}

func decodeRiskFreeJSON(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(recorder.Body).Decode(target); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
}

func assertRiskFreeExpectations(t *testing.T, mock pgxmock.PgxConnIface) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet database expectations: %v", err)
	}
}
