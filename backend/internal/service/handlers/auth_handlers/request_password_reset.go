package authhandlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	database "portfolio-tracker/internal/adapter/db"
	"portfolio-tracker/internal/adapter/db/repos"
	smtp "portfolio-tracker/internal/service/smtp"
)

type RequestPasswordResetRequest struct {
	Email string `json:"email"`
}

type APIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// code generating repo
func generateRandomCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func RequestPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RequestPasswordResetRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid JSON payload"})
		return
	}

	if req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "email field required"})
		return
	}

	dbConn, err := database.GetDBConnection()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to get database connection"})
		return
	}
	defer dbConn.Close(r.Context())

	u := repos.User{DBConn: dbConn}

	_, err = u.GetByEmail(req.Email)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "User with this email not found"})
		return
	}

	otc := repos.OneTimeCode{DBConn: dbConn}

	err = otc.InvalidateOldCodes(req.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to invalidate old reset codes"})
		return
	}

	code, err := generateRandomCode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to generate security code"})
		return
	}

	_, err = otc.Create(req.Email, code)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to save reset code"})
		return
	}

	err = smtp.SendPasswordResetEmail(req.Email, code)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to send password reset email"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Status:  "success",
		Message: "A password reset code has been sent",
	})
}
