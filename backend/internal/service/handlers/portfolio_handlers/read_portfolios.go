package portfoliohandlers

import (
	"encoding/json"
	"net/http"
	"portfolio-tracker/internal/adapter/db/repos"
	"portfolio-tracker/internal/service/middlewares"
	"time"
)

type PortfolioResponse struct {
	ID             int32     `json:"id"`
	UserID         int32     `json:"user_id"`
	RiskFreeRateID int32     `json:"risk_free_rate_id"`
	Name           string    `json:"name"`
	CreatedAt      time.Time `json:"created_at"`
}

func ReadPortfolios(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(middlewares.UserIDKey).(int32)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Unauthorized: user identity not found"})
		return
	}

	dbConn, err := getDBConnection()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to get DB connection"})
		return
	}

	pRepo := repos.Portfolio{DBConn: dbConn}

	portfolios, err := pRepo.ReadAllByUserID(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to read portfolios"})
		return
	}

	response := make([]PortfolioResponse, 0, len(portfolios))
	for _, p := range portfolios {
		var id int32
		if p.Id != nil {
			id = *p.Id
		}

		var createdAt time.Time
		if p.CreatedAt != nil {
			createdAt = *p.CreatedAt
		}

		response = append(response, PortfolioResponse{
			ID:             id,
			UserID:         p.UserId,
			RiskFreeRateID: p.RiskFreeRateId,
			Name:           p.Name,
			CreatedAt:      createdAt,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
