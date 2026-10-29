package riskfreerateshandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-tracker/internal/adapter/db/repos"

	database "portfolio-tracker/internal/adapter/db"
)

type CreateRiskFreeRateRequest struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type APIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func CreateRiskFreeRateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req CreateRiskFreeRateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid JSON body"})
		return
	}

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Name field is required"})
		return
	}

	dbConn, err := database.GetDBConnection()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to get DB connection"})
		return
	}

	rfr := repos.RiskFreeRate{DBConn: dbConn}

	input := repos.RiskFreeRateInstance{
		Name:  req.Name,
		Value: req.Value,
	}

	id, err := rfr.Create(input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to save record"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(APIResponse{
		Status:  "success",
		Message: fmt.Sprintf("Risk free rate with %d id created successfully", id),
	})
}
