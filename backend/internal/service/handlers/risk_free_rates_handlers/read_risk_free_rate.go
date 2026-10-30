package riskfreerateshandlers

import (
	"encoding/json"
	"net/http"
	"portfolio-tracker/internal/adapter/db/repos"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ReadRiskFreeRateResponse struct {
	ID    int32   `json:"id"`
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

func ReadRiskFreeRate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Missing id parameter"})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid id parameter"})
		return
	}

	dbConn, err := getDBConnection()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to get DB connection"})
		return
	}

	rfr := repos.RiskFreeRate{DBConn: dbConn}

	data, err := rfr.Read(int32(id))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Risk free rate not found"})
		return
	}

	var resID int32
	if data.Id != nil {
		resID = *data.Id
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ReadRiskFreeRateResponse{
		ID:    resID,
		Name:  data.Name,
		Value: data.Value,
	})
}
