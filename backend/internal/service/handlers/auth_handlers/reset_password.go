package authhandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"portfolio-tracker/internal/adapter/db/repos"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get a temporary one-time-code
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Authorization header is required"})
		return
	}

	// expected format "Bearer <token>"
	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid authorization header format"})
		return
	}

	tokenString := headerParts[1]

	// jwt validation
	secret := os.Getenv("PURPOSE_JWT_SECRET")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// check signature
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid or expired reset token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid token claims"})
		return
	}

	// check if the token was needed for changing password
	purpose, _ := claims["purpose"].(string)
	if purpose != "password_reset" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid token purpose"})
		return
	}

	email, _ := claims["sub"].(string)
	if email == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Email not found in token"})
		return
	}

	var req ResetPasswordRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid JSON payload"})
		return
	}

	if req.NewPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "new_password field is required"})
		return
	}

	dbConn, err := getDBConnection()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to get database connection"})
		return
	}
	defer dbConn.Close(r.Context())

	// hashing new password
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 10)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to process new password"})
		return
	}
	hashedPassword := string(hashedPasswordBytes)

	userRepo := repos.User{DBConn: dbConn}

	err = userRepo.UpdatePassword(r.Context(), email, hashedPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to update password"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Status:  "success",
		Message: "Password has been successfully updated",
	})
}
