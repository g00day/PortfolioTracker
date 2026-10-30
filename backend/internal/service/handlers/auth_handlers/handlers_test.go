package authhandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pashagolub/pgxmock/v5"
	"golang.org/x/crypto/bcrypt"

	"portfolio-tracker/internal/adapter/db/repos"
)

func TestRegistrationHandlerSuccess(t *testing.T) {
	mock := useMockDB(t)
	mock.ExpectQuery("INSERT INTO users").
		WithArgs("alice", "alice@example.com", pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(int32(7)))

	recorder := performAuthRequest(RegistrationHandler, `{`+
		`"login":"alice","email":"alice@example.com","password":"secret123"}`,
	)

	assertAuthStatus(t, recorder, http.StatusCreated)
	var response RegisterResponse
	decodeAuthJSON(t, recorder, &response)
	if response.Status != "success" || !strings.Contains(response.Message, "7") {
		t.Fatalf("unexpected response: %+v", response)
	}
	assertAuthExpectations(t, mock)
}

func TestLoginHandlerSuccess(t *testing.T) {
	t.Setenv("JWT_SECRET", "access-test-secret")
	mock := useMockDB(t)

	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	now := time.Now().UTC()
	id := int32(7)
	mock.ExpectQuery("SELECT id, username, email, password_hash, created_at FROM users WHERE username").
		WithArgs("alice").
		WillReturnRows(mock.NewRows([]string{"id", "username", "email", "password_hash", "created_at"}).
			AddRow(&id, "alice", "alice@example.com", string(hash), now))

	recorder := performAuthRequest(LoginHandler, `{"login":"alice","password":"secret123"}`)

	assertAuthStatus(t, recorder, http.StatusOK)
	var response LoginResponse
	decodeAuthJSON(t, recorder, &response)
	if response.Status != "success" || response.Token == "" {
		t.Fatalf("unexpected response: %+v", response)
	}
	token, err := jwt.Parse(response.Token, func(*jwt.Token) (any, error) {
		return []byte("access-test-secret"), nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("invalid access token: %v", err)
	}
	assertAuthExpectations(t, mock)
}

func TestRequestPasswordResetHandlerSuccess(t *testing.T) {
	mock := useMockDB(t)
	now := time.Now().UTC()
	id := int32(7)
	mock.ExpectQuery("SELECT id, username, email, password_hash, created_at FROM users WHERE email").
		WithArgs("alice@example.com").
		WillReturnRows(mock.NewRows([]string{"id", "username", "email", "password_hash", "created_at"}).
			AddRow(&id, "alice", "alice@example.com", "hash", now))
	mock.ExpectExec("DELETE FROM one_time_codes WHERE email").
		WithArgs("alice@example.com").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectQuery("INSERT INTO one_time_codes").
		WithArgs("alice@example.com", pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(int32(11)))

	oldSend := sendPasswordResetEmail
	sendPasswordResetEmail = func(recipient, code string) error {
		if recipient != "alice@example.com" {
			t.Errorf("recipient: want alice@example.com, got %q", recipient)
		}
		if len(code) != 6 {
			t.Errorf("code must contain six digits, got %q", code)
		}
		return nil
	}
	t.Cleanup(func() { sendPasswordResetEmail = oldSend })

	recorder := performAuthRequest(RequestPasswordResetHandler, `{"email":"alice@example.com"}`)

	assertAuthStatus(t, recorder, http.StatusOK)
	var response APIResponse
	decodeAuthJSON(t, recorder, &response)
	if response.Status != "success" {
		t.Fatalf("unexpected response: %+v", response)
	}
	assertAuthExpectations(t, mock)
}

func TestVerifyOneTimeCodeHandlerSuccess(t *testing.T) {
	t.Setenv("PURPOSE_JWT_SECRET", "reset-test-secret")
	mock := useMockDB(t)
	mock.ExpectQuery("DELETE FROM one_time_codes").
		WithArgs("alice@example.com", pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(int32(11)))

	recorder := performAuthRequest(VerifyOneTimeCodeHandler, `{"email":"alice@example.com","code":"123456"}`)

	assertAuthStatus(t, recorder, http.StatusOK)
	var response VerifyOneTimeCodeResponse
	decodeAuthJSON(t, recorder, &response)
	if response.Status != "success" || response.ResetToken == "" {
		t.Fatalf("unexpected response: %+v", response)
	}
	token, err := jwt.Parse(response.ResetToken, func(*jwt.Token) (any, error) {
		return []byte("reset-test-secret"), nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("invalid reset token: %v", err)
	}
	assertAuthExpectations(t, mock)
}

func TestResetPasswordHandlerSuccess(t *testing.T) {
	t.Setenv("PURPOSE_JWT_SECRET", "reset-test-secret")
	mock := useMockDB(t)
	mock.ExpectExec("UPDATE users SET password_hash").
		WithArgs(pgxmock.AnyArg(), "alice@example.com").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	resetToken := signedResetToken(t, "alice@example.com")
	request := httptest.NewRequest(http.MethodPost, "/auth/reset-password", strings.NewReader(`{"new_password":"new-secret"}`))
	request.Header.Set("Authorization", "Bearer "+resetToken)
	recorder := httptest.NewRecorder()
	ResetPasswordHandler(recorder, request)

	assertAuthStatus(t, recorder, http.StatusOK)
	var response APIResponse
	decodeAuthJSON(t, recorder, &response)
	if response.Status != "success" {
		t.Fatalf("unexpected response: %+v", response)
	}
	assertAuthExpectations(t, mock)
}

func useMockDB(t *testing.T) pgxmock.PgxConnIface {
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

func performAuthRequest(handler http.HandlerFunc, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	recorder := httptest.NewRecorder()
	handler(recorder, request)
	return recorder
}

func signedResetToken(t *testing.T, email string) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     email,
		"purpose": "password_reset",
		"exp":     time.Now().Add(time.Minute).Unix(),
	})
	signed, err := token.SignedString([]byte("reset-test-secret"))
	if err != nil {
		t.Fatalf("sign reset token: %v", err)
	}
	return signed
}

func assertAuthStatus(t *testing.T, recorder *httptest.ResponseRecorder, want int) {
	t.Helper()
	if recorder.Code != want {
		t.Fatalf("status: want %d, got %d; body: %s", want, recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type: want application/json, got %q", got)
	}
}

func decodeAuthJSON(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(recorder.Body).Decode(target); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
}

func assertAuthExpectations(t *testing.T, mock pgxmock.PgxConnIface) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet database expectations: %v", err)
	}
}
