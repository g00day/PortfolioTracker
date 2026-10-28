package authhandlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	database "portfolio-tracker/internal/adapter/db"
	"portfolio-tracker/internal/adapter/db/repos"

	"github.com/golang-jwt/jwt/v5"
)

type VerifyOneTimeCodeRequest struct {
	Code  string `json:"code"`
	Email string `json:"email"`
}

type VerifyOneTimeCodeResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	ResetToken string `json:"reset_token,omitempty"`
}

func generateResetToken(email string) (string, error) {
	secret := os.Getenv("PURPOSE_JWT_SECRET")

	claims := jwt.MapClaims{
		"sub":     email,
		"purpose": "password_reset",
		"exp":     time.Now().Add(5 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func VerifyOneTimeCodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req VerifyOneTimeCodeRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid JSON payload"})
		return
	}

	if req.Email == "" || req.Code == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "email and code fields are required"})
		return
	}

	dbConn, err := database.GetDBConnection()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to get database connection"})
		return
	}
	defer dbConn.Close(r.Context())

	otcRepo := repos.OneTimeCode{DBConn: dbConn}

	isValid, err := otcRepo.VerifyAndUse(req.Email, req.Code)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to verify code"})
		return
	}

	if !isValid {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid or expired one-time code"})
		return
	}

	// if the one-time-code is right - generate a temporary jwt token
	resetToken, err := generateResetToken(req.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to generate reset token"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(VerifyOneTimeCodeResponse{
		Status:     "success",
		Message:    "Code verified successfully. Use the provided token to reset your password.",
		ResetToken: resetToken,
	})
}
