package portfoliohandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-tracker/internal/adapter/db/repos"
	"portfolio-tracker/internal/adapter/exchanges_datafetcher"
	"portfolio-tracker/internal/service/middlewares"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type AddPortfolioItemRequest struct {
	Ticker         *string `json:"ticker"`
	RiskFreeRateId *int32  `json:"risk_free_rate_id"`
	TotalAmount    float32 `json:"total_amount"`
}

func AddPortfolioItem(w http.ResponseWriter, r *http.Request) {
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
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Missing portfolio id parameter in URL"})
		return
	}

	portfolioID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid portfolio id parameter"})
		return
	}

	var req AddPortfolioItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid JSON body"})
		return
	}

	isTickerValid := req.Ticker != nil && *req.Ticker != ""
	if (isTickerValid && req.RiskFreeRateId != nil) || (!isTickerValid && req.RiskFreeRateId == nil) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Either ticker or risk_free_rate_id must be provided, but not both"})
		return
	}

	if req.TotalAmount <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Total amount must be greater than zero"})
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

	piRepo := repos.PortfolioItem{DBConn: dbConn}

	input := repos.PortfolioItemInstance{
		PortfolioId:    int32(portfolioID),
		Ticker:         req.Ticker,
		RiskFreeRateId: req.RiskFreeRateId,
		TotalAmount:    req.TotalAmount,
	}

	id, err := piRepo.Create(input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to add item to portfolio"})
		return
	}

	// Store the opening cost basis so the structure endpoint can calculate
	// actual P&L from the first moment the position exists.
	openingPrice := 1.0
	if isTickerValid {
		settings := exchanges_datafetcher.NewConnectionSettings()
		dataFetcher := exchanges_datafetcher.NewCrossMarketDataFetcher(&settings, 1000)
		klines := dataFetcher.GetKlines(*req.Ticker, "D", "spot")
		if len(klines) > 0 && klines[0][4] > 0 {
			openingPrice = klines[0][4]
		} else {
			openingPrice = 0
		}
	}

	if openingPrice > 0 {
		tRepo := repos.Transaction{DBConn: dbConn}
		_, _ = tRepo.Create(repos.TransactionInstance{
			PortfolioId:   int32(portfolioID),
			PortfolioItem: id,
			Type:          "deposit",
			Amount:        float64(req.TotalAmount),
			PriceUSD:      openingPrice,
		})
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(APIResponse{
		Status:  "success",
		Message: fmt.Sprintf("Portfolio item with %d id created successfully", id),
	})
}
