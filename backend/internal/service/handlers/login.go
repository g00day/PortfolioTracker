package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	database "portfolio-tracker/internal/adapter/db"
	"portfolio-tracker/internal/adapter/db/repos"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Status string `json:"status"`
	Token  string `json:"token,omitempty"`
	Error  string `json:"error,omitempty"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Status: "error", Error: "Invalid JSON payload"})
		return
	}

	if req.Login == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Status: "error", Error: "Both login and password are required"})
		return
	}

	dbConn, err := database.GetDBConnection()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(LoginResponse{Status: "error", Error: "Database connection unavailable"})
		return
	}

	u := repos.User{DBConn: dbConn}
	userInstance, err := u.GetByUsername(req.Login)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{Status: "error", Error: "Invalid login or password"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(userInstance.PasswordHash), []byte(req.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{Status: "error", Error: "Invalid login or password"})
		return
	}

	// --- JWT token generation ---
	var userID int32
	if userInstance.Id != nil {
		userID = *userInstance.Id
	}

	claims := jwt.MapClaims{
		"sub":   userID,                                // Subject (user ID)
		"login": userInstance.Username,                 // login
		"exp":   time.Now().Add(time.Hour * 72).Unix(), // expiration date
		"iat":   time.Now().Unix(),                     // time when the token was created
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // pass all claims and generating algorithm

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default_fallback_secret_key"
	}

	// sign the payload with secret
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Status: "error", Error: "Failed to generate token"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		Status: "success",
		Token:  tokenString,
	})
}
