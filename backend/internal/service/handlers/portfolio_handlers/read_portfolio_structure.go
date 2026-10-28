package portfoliohandlers

import (
	"encoding/json"
	"net/http"
	"portfolio-tracker/internal/adapter/db/repos"
	"portfolio-tracker/internal/adapter/exchanges_datafetcher"
	"portfolio-tracker/internal/service/middlewares"
	"strconv"

	database "portfolio-tracker/internal/adapter/db"

	"github.com/go-chi/chi/v5"
)

// Order: [Ticker, CurrentPrice, Amount, WeightPercentage, PnL(mock)]
type StructureRow []interface{}

type PortfolioStructureResponse struct {
	PortfolioStructure []StructureRow `json:"portfolio_structure"`
}

func ReadPortfolioStructure(w http.ResponseWriter, r *http.Request) {
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

	dbConn, err := database.GetDBConnection()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to get DB connection"})
		return
	}

	pRepo := repos.Portfolio{DBConn: dbConn}
	portfolioData, err := pRepo.Read(int32(portfolioID))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Portfolio not found"})
		return
	}

	if portfolioData.UserId != userID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Forbidden: you do not own this portfolio"})
		return
	}

	piRepo := repos.PortfolioItem{DBConn: dbConn}
	items, err := piRepo.ReadAllByPortfolio(int32(portfolioID))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to read portfolio items"})
		return
	}

	if len(items) == 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(PortfolioStructureResponse{PortfolioStructure: []StructureRow{}})
		return
	}

	cs := exchanges_datafetcher.NewConnectionSettings()
	df := exchanges_datafetcher.NewCrossMarketDataFetcher(&cs, 1000)

	var totalPortfolioValue float64
	prices := make([]float64, len(items))
	tickers := make([]string, len(items))
	rfrRepo := repos.RiskFreeRate{DBConn: dbConn}

	for i, item := range items {
		var ticker string
		var price float64

		if item.Ticker != nil {
			ticker = *item.Ticker
			klines := df.GetKlines(ticker, "D", "spot")
			if len(klines) > 0 {
				price = klines[0][4]
			}
		} else {
			price = 1.0
			if item.RiskFreeRateId != nil {
				rfrData, err := rfrRepo.Read(*item.RiskFreeRateId)
				if err == nil {
					ticker = rfrData.Name
				} else {
					ticker = "RISK_FREE_RATE"
				}
			} else {
				ticker = "RISK_FREE_RATE"
			}
		}

		tickers[i] = ticker
		prices[i] = price
		totalPortfolioValue += float64(item.TotalAmount) * price
	}

	structure := make([]StructureRow, len(items))
	const pnlMock = 14000.00

	for i, item := range items {
		var weight float64
		amount := float64(item.TotalAmount)

		if totalPortfolioValue > 0 {
			weight = ((amount * prices[i]) / totalPortfolioValue) * 100.0
		}

		structure[i] = StructureRow{
			tickers[i],
			prices[i],
			amount,
			weight,
			pnlMock,
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(PortfolioStructureResponse{
		PortfolioStructure: structure,
	})
}
