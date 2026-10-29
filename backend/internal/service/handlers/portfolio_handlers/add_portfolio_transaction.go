package portfoliohandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-tracker/internal/adapter/db/repos"
	"portfolio-tracker/internal/service/middlewares"
	"strconv"
	"time"

	database "portfolio-tracker/internal/adapter/db"

	"github.com/go-chi/chi/v5"
)

type AddPortfolioTransactionRequest struct {
	PortfolioItemID int32      `json:"portfolio_item_id"`
	Type            string     `json:"type"` // "deposit" or "withdraw"
	Amount          float64    `json:"amount"`
	PriceUSD        float64    `json:"price_usd"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty"`
}

func AddPortfolioTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	userID, ok := ctx.Value(middlewares.UserIDKey).(int32)
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

	var req AddPortfolioTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid JSON body"})
		return
	}

	if req.PortfolioItemID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Valid portfolio_item_id is required"})
		return
	}

	if req.Type != "deposit" && req.Type != "withdraw" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Transaction type must be 'deposit' or 'withdraw'"})
		return
	}

	if req.Amount <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Transaction amount must be greater than zero"})
		return
	}

	if req.PriceUSD <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "price_usd must be greater than zero"})
		return
	}

	dbConn, err := database.GetDBConnection()
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

	piRepo := repos.PortfolioItem{DBConn: dbConn}
	item, err := piRepo.Read(req.PortfolioItemID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Portfolio item not found"})
		return
	}

	if item.PortfolioId != int32(portfolioID) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Portfolio item does not belong to the specified portfolio"})
		return
	}

	var newAmount float32
	if req.Type == "deposit" {
		newAmount = item.TotalAmount + float32(req.Amount)
	} else {
		newAmount = item.TotalAmount - float32(req.Amount)
		if newAmount < 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Insufficient balance for withdrawal"})
			return
		}
	}

	tx, err := dbConn.Begin(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to start database transaction"})
		return
	}
	defer tx.Rollback(ctx)

	err = piRepo.UpdateAmount(ctx, tx, item.Id, newAmount)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to update item balance"})
		return
	}

	tRepo := repos.Transaction{DBConn: dbConn}
	input := repos.TransactionInstance{
		PortfolioId:   int32(portfolioID),
		PortfolioItem: req.PortfolioItemID,
		Type:          req.Type,
		Amount:        req.Amount,
		PriceUSD:      req.PriceUSD,
		ProcessedAt:   req.ProcessedAt,
	}

	txID, err := tRepo.CreateWithTx(ctx, tx, input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to record transaction log"})
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to commit database transaction"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(APIResponse{
		Status:  "success",
		Message: fmt.Sprintf("Transaction recorded and balance updated successfully. Transaction ID: %d", txID),
	})
}
