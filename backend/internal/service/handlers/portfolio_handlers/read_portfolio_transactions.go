package portfoliohandlers

import (
	"encoding/json"
	"net/http"
	"portfolio-tracker/internal/adapter/db/repos"
	"portfolio-tracker/internal/service/middlewares"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type TransactionResponse struct {
	ID              int32     `json:"id"`
	PortfolioID     int32     `json:"portfolio_id"`
	PortfolioItemID int32     `json:"portfolio_item_id"`
	Type            string    `json:"type"` // "deposit" or "withdraw"
	Amount          float64   `json:"amount"`
	ProcessedAt     time.Time `json:"processed_at"`
}

func ReadPortfolioTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(middlewares.UserIDKey).(int32)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Unauthorized: user identity not found"})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Missing portfolio id parameter"})
		return
	}

	portfolioID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid portfolio id parameter"})
		return
	}

	dbConn, err := getDBConnection()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to get DB connection"})
		return
	}

	pRepo := repos.Portfolio{DBConn: dbConn}
	portfolio, err := pRepo.Read(int32(portfolioID))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Portfolio not found"})
		return
	}

	if portfolio.UserId != userID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Forbidden: you do not own this portfolio"})
		return
	}

	tRepo := repos.Transaction{DBConn: dbConn}
	transactions, err := tRepo.ReadAllByPortfolio(int32(portfolioID))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to read portfolio transactions"})
		return
	}

	response := make([]TransactionResponse, 0, len(transactions))
	for _, tx := range transactions {
		var id int32
		if tx.Id != nil {
			id = *tx.Id
		}

		var processedAt time.Time
		if tx.ProcessedAt != nil {
			processedAt = *tx.ProcessedAt
		}

		response = append(response, TransactionResponse{
			ID:              id,
			PortfolioID:     tx.PortfolioId,
			PortfolioItemID: tx.PortfolioItem,
			Type:            tx.Type,
			Amount:          tx.Amount,
			ProcessedAt:     processedAt,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
