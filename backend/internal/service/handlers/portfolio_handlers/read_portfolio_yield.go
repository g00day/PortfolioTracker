package portfoliohandlers

import (
	"encoding/json"
	"net/http"
	"portfolio-tracker/internal/adapter/db/repos"
	"portfolio-tracker/internal/adapter/exchanges_datafetcher"
	"portfolio-tracker/internal/service"
	"portfolio-tracker/internal/service/middlewares"
	"strconv"

	database "portfolio-tracker/internal/adapter/db"

	"github.com/go-chi/chi/v5"
)

type YieldTopRow []interface{}

type PortfolioYieldResponse struct {
	PL                 map[string][]float64   `json:"p&l"`
	YieldTop3          map[string]YieldTopRow `json:"yield_top_3"`
	PortfolioStructure []StructureRow         `json:"portfolio_structure"`
	SharpesRatio       float64                `json:"sharpes_ratio"`
	SortinoRatio       float64                `json:"sortino_ratio"`
	CovFomoPnl         float64                `json:"cov_fomo_pnl"`
	CorrFomoPnl        float64                `json:"corr_fomo_pnl"`
}

func ReadPortfolioYield(w http.ResponseWriter, r *http.Request) {
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
		json.NewEncoder(w).Encode(PortfolioYieldResponse{PortfolioStructure: []StructureRow{}})
		return
	}

	cs := exchanges_datafetcher.NewConnectionSettings()
	df := exchanges_datafetcher.NewCrossMarketDataFetcher(&cs, 1000)

	var totalPortfolioValue float64
	serviceAssets := make([]service.Asset, len(items))
	assetPrices := make([]float64, len(items))
	tickers := make([]string, len(items))
	rfrRepo := repos.RiskFreeRate{DBConn: dbConn}

	// Fetch global risk free rate associated with the active portfolio from the DB
	var baselineRiskFreeRate float64
	portfolioRFR, err := rfrRepo.Read(portfolioData.RiskFreeRateId)
	if err == nil {
		baselineRiskFreeRate = portfolioRFR.Value
	} else {
		baselineRiskFreeRate = 0.05 // fallback if global portfolio baseline not found
	}

	for i, item := range items {
		var ticker string
		var isRF bool
		var price float64
		var assetRFReward float64

		if item.Ticker != nil {
			ticker = *item.Ticker
			isRF = false

			klines := df.GetKlines(ticker, "D", "spot")
			if len(klines) > 0 {
				price = klines[0][4]
			}
		} else {
			isRF = true
			price = 1.0

			rfrData, err := rfrRepo.Read(*item.RiskFreeRateId)
			if err == nil {
				ticker = rfrData.Name
				assetRFReward = rfrData.Value
			} else {
				ticker = "RISK_FREE_RATE"
			}
		}

		tickers[i] = ticker
		assetPrices[i] = price
		totalPortfolioValue += float64(item.TotalAmount) * price

		serviceAssets[i] = service.Asset{
			Ticker:         ticker,
			Amount:         float64(item.TotalAmount),
			IsRiskFree:     isRF,
			RiskFreeReward: assetRFReward,
		}
	}

	if totalPortfolioValue > 0 {
		for i := range serviceAssets {
			assetUSDValue := serviceAssets[i].Amount * assetPrices[i]
			serviceAssets[i].W = assetUSDValue / totalPortfolioValue
		}
	}

	portfolioCalc := &service.Portfolio{
		Assets:       serviceAssets,
		Datafetcher:  *df,
		RiskFreeRate: baselineRiskFreeRate, // Using dynamic value from DB instead of hardcoded 0.05
	}

	covMatrix := portfolioCalc.BuildCovariationMatrix()
	sharpe := portfolioCalc.CalculateSharpeRatio(covMatrix)
	sortino := portfolioCalc.CalculateSortinoRatio()

	structure := make([]StructureRow, len(items))
	const pnlMock = 14000.00

	for i := range items {
		var weight float64
		amount := serviceAssets[i].Amount

		if totalPortfolioValue > 0 {
			weight = ((amount * assetPrices[i]) / totalPortfolioValue) * 100.0
		}

		structure[i] = StructureRow{
			tickers[i],
			assetPrices[i],
			amount,
			weight,
			pnlMock,
		}
	}

	// Заглушки p&l массивов
	pnlHistory := map[string][]float64{
		"risk_free":      {1, 2, 3, 4, 5, 6},
		"portfolio":      {12, -20, -100, -201, 203},
		"portfolio_FOMO": {12, -50, -200, 300, 500},
	}

	// Заглушки топ-3 доходных активов
	yieldTop3 := map[string]YieldTopRow{
		"1": {"BTCUSDT", 102.0, 48.0},
		"2": {"USD-Tres Oct-2026", 80.0, 12.0},
		"3": {"XAUTUSDT", 40.0, 25.0},
	}

	response := PortfolioYieldResponse{
		PL:                 pnlHistory,
		YieldTop3:          yieldTop3,
		PortfolioStructure: structure,
		SharpesRatio:       sharpe,
		SortinoRatio:       sortino,
		CovFomoPnl:         -0.5, // Заглушка ковариации FOMO
		CorrFomoPnl:        -2.0, // Заглушка корреляции FOMO
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
