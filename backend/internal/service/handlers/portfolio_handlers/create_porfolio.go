package portfoliohandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-tracker/internal/adapter/db/repos"
	"portfolio-tracker/internal/service/middlewares"

	database "portfolio-tracker/internal/adapter/db"
)

type CreatePortfolioRequest struct {
	Name           string `json:"name"`
	RiskFreeRateId int32  `json:"risk_free_rate_id"`
}

type APIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func CreatePortfolio(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(middlewares.UserIDKey).(int32)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Unauthorized: user identity not found"})
		return
	}

	var req CreatePortfolioRequest
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

	if req.RiskFreeRateId <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Valid risk_free_rate_id is required"})
		return
	}

	dbConn, err := database.GetDBConnection()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to get DB connection"})
		return
	}

	pRepo := repos.Portfolio{DBConn: dbConn}

	input := repos.PortfolioInstance{
		UserId:         userID,
		Name:           req.Name,
		RiskFreeRateId: req.RiskFreeRateId,
	}

	id, err := pRepo.Create(input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to create portfolio"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(APIResponse{
		Status:  "success",
		Message: fmt.Sprintf("Portfolio with %d id created successfully", id),
	})
}
