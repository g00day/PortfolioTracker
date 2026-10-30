package authhandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-tracker/internal/adapter/db/repos"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{Status: "error", Message: "Invalid JSON payload"})
		return
	}

	if req.Login == "" || req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{Status: "error", Message: "All fields (login, email, password) are required"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Status: "error", Message: "Failed to process password"})
		return
	}

	ui := repos.UserInstance{
		Username:     req.Login,
		PasswordHash: string(hashedPassword),
		Email:        req.Email,
	}

	dbConn, err := getDBConnection()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(RegisterResponse{Status: "error", Message: "Failed to get database connection"})
		return
	}

	u := repos.User{DBConn: dbConn}
	uid, err := u.Create(ui)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Status: "error", Message: "Failed to create user"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{
		Status:  "success",
		Message: fmt.Sprintf("Successfully registered user with id: %v", uid),
	})
}
