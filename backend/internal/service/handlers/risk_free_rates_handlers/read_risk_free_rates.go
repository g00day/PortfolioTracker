package riskfreerateshandlers

import (
	"encoding/json"
	"net/http"
	"portfolio-tracker/internal/adapter/db/repos"
)

type RiskFreeRateResponse struct {
	ID    int32   `json:"id"`
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

func ReadRiskFreeRates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbConn, err := getDBConnection()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to get DB connection"})
		return
	}

	rfr := repos.RiskFreeRate{DBConn: dbConn}

	rates, err := rfr.ReadAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to read records"})
		return
	}

	response := make([]RiskFreeRateResponse, 0, len(rates))
	for _, rate := range rates {
		var id int32
		if rate.Id != nil {
			id = *rate.Id
		}
		response = append(response, RiskFreeRateResponse{
			ID:    id,
			Name:  rate.Name,
			Value: rate.Value,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
