package portfoliohandlers

import (
	"encoding/json"
	"net/http"
	"portfolio-tracker/internal/adapter/db/repos"
	"portfolio-tracker/internal/adapter/exchanges_datafetcher"
	"portfolio-tracker/internal/service"
	"portfolio-tracker/internal/service/middlewares"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// Order: [Ticker, CurrentPrice, Amount, Cost, WeightPercentage, PnL]
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

	dbConn, err := getDBConnection()
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
	pnls := make([]float64, len(items))
	rfrRepo := repos.RiskFreeRate{DBConn: dbConn}
	tRepo := repos.Transaction{DBConn: dbConn}

	for i, item := range items {
		var ticker string
		var price float64
		var riskFreeRate float64

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
					riskFreeRate = rfrData.Value
				} else {
					ticker = "RISK_FREE_RATE"
				}
			} else {
				ticker = "RISK_FREE_RATE"
			}
		}

		// P&L is based on the weighted average acquisition price from deposit
		// transactions. Withdrawals do not change the cost basis of the
		// remaining position. Positions without priced deposits return zero.
		transactions, txErr := tRepo.ReadAllByItem(item.Id)
		if txErr == nil {
			if item.RiskFreeRateId != nil {
				pnls[i] = calculateRiskFreePnL(transactions, riskFreeRate, time.Now())
				if item.TotalAmount > 0 {
					price = (float64(item.TotalAmount) + pnls[i]) / float64(item.TotalAmount)
				}
			} else {
				pricedDeposits := make([]service.Asset, 0, len(transactions))
				for _, tx := range transactions {
					if tx.Type == "deposit" && tx.Amount > 0 && tx.PriceUSD > 0 {
						pricedDeposits = append(pricedDeposits, service.Asset{Amount: tx.Amount, RiskFreeReward: tx.PriceUSD})
					}
				}
				pnls[i], _ = service.CalculateItemPnL(pricedDeposits, price, float64(item.TotalAmount))
			}
		}
		tickers[i] = ticker
		prices[i] = price
		totalPortfolioValue += float64(item.TotalAmount) * price
	}

	structure := make([]StructureRow, len(items))

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
			prices[i] * amount,
			weight,
			pnls[i],
			item.Id,
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(PortfolioStructureResponse{
		PortfolioStructure: structure,
	})
}
