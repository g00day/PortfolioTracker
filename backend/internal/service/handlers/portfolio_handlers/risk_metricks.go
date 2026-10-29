package portfoliohandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-tracker/internal/adapter/db/repos"
	"portfolio-tracker/internal/adapter/exchanges_datafetcher"
	"portfolio-tracker/internal/service"
	"portfolio-tracker/internal/service/middlewares"
	"strconv"

	database "portfolio-tracker/internal/adapter/db"

	"github.com/go-chi/chi/v5"
)

type QuantileTuple [2]float64 // [Percentage, USD_Value]

type VaRQuantiles struct {
	Mu  float64 `json:"Mu"`
	Q90 float64 `json:"Q90"`
	Q95 float64 `json:"Q95"`
	Q99 float64 `json:"Q99"`
}

const SQRT_365 = 19.1049731745428

// Order: [Ticker, [Mu, MuUSD], [Q90, Q90USD], [Q95, Q95USD], [Q99, Q99USD], mcr_i, ccrt_i]
type PartialRiskRow []interface{}

type PortfolioRiskMetricsResponse struct {
	PortfolioVaRQuantiles   VaRQuantiles     `json:"portfolioVaR_quantiles"`
	PortfolioValueUSD       float64          `json:"portfolio_value_usd"`
	PartialRisks            []PartialRiskRow `json:"partial_risks"`
	VolatilityDrag          float64          `json:"volatility_drag"` // in percents
	DiversificationRatio    float64          `json:"diversification_ratio"`
	EffectiveNumberOfAssets float64          `json:"effective_number_of_assets"`
}

func GetPortfolioRiskMetrics(w http.ResponseWriter, r *http.Request) {
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
		json.NewEncoder(w).Encode(PortfolioRiskMetricsResponse{PartialRisks: []PartialRiskRow{}})
		return
	}

	cs := exchanges_datafetcher.NewConnectionSettings()
	df := exchanges_datafetcher.NewCrossMarketDataFetcher(&cs, 1000)

	var totalPortfolioValue float64
	serviceAssets := make([]service.Asset, len(items))
	assetPrices := make([]float64, len(items))
	rfrRepo := repos.RiskFreeRate{DBConn: dbConn}

	for i, item := range items {
		var ticker string
		var isRF bool
		var price float64

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
			} else {
				ticker = "RISK_FREE_RATE"
			}
		}

		assetPrices[i] = price
		totalPortfolioValue += float64(item.TotalAmount) * price

		serviceAssets[i] = service.Asset{
			Ticker:     ticker,
			Amount:     float64(item.TotalAmount),
			IsRiskFree: isRF,
		}
	}

	if totalPortfolioValue > 0 {
		for i := range serviceAssets {
			assetUSDValue := serviceAssets[i].Amount * assetPrices[i]
			serviceAssets[i].W = assetUSDValue / totalPortfolioValue
		}
	}

	fmt.Println(serviceAssets)

	// Initializing cross-market calculation modules
	portfolioCalc := &service.Portfolio{
		Assets:      serviceAssets,
		Datafetcher: *df,
	}

	covMatrix := portfolioCalc.BuildCovariationMatrix()
	mcr, cctr := portfolioCalc.CalculateMarginalAndComponentContributions(covMatrix)
	divRatio := portfolioCalc.CalculateDiversificationRate(covMatrix)
	effNA := portfolioCalc.CalculateEffectiveNA()

	volCursor := &service.VolatilityCursor{
		Datafetcher: portfolioCalc.Datafetcher,
	}

	partialRisks := make([]PartialRiskRow, len(serviceAssets))
	var totalVolatilityDragSum float64

	// Accumulators for aggregating portfolio-wide unified Parkinson VaR Quantiles from partial risks
	var sumPortMuUSD, sumPortQ90USD, sumPortQ95USD, sumPortQ99USD float64

	// Preparing arrays to construct integrated portfolio historical logs for asset allocation
	var combinedKlines [][]float64
	hasMarketAssets := false

	for i, asset := range serviceAssets {

		var mu, q90, q95, q99 float64
		var muUSD, q90USD, q95USD, q99USD float64
		var mcrVal, cctrVal float64

		if i < len(mcr) {
			mcrVal = mcr[i]
			cctrVal = cctr[i]
		}

		if !asset.IsRiskFree {
			klines := portfolioCalc.Datafetcher.GetKlines(asset.Ticker, "D", "spot")
			lastPrice := assetPrices[i]

			// Compute lognormal Parkinson risk distribution benchmarks
			q90, q95, q99 = volCursor.CalculateLognormalParkinsonVaR(klines)
			_, mu, err = volCursor.PrepareParkinsonVolatilityDistributionData(klines)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to prepare volatility distribution data"})
				return
			}

			// Transfer lognormal Parkinson risk distribution benchmarks to USD
			amount := asset.Amount
			muUSD, q90USD, q95USD, q99USD = mu*lastPrice*amount, q90*lastPrice*amount, q95*lastPrice*amount, q99*lastPrice*amount

			// Sum partial USD VaR components to calculate unified portfolio quantiles
			sumPortMuUSD += muUSD
			sumPortQ90USD += q90USD
			sumPortQ95USD += q95USD
			sumPortQ99USD += q99USD

			// Aggregate overall structural volatility drag dynamically across asset weights
			drag := volCursor.CalculateAnnualizedVolatilityDrag(asset.Ticker, "spot")
			totalVolatilityDragSum += asset.W * drag

			if !hasMarketAssets || len(klines) < len(combinedKlines) {
				combinedKlines = klines
			}
			hasMarketAssets = true
		} else {
			// For risk-free assets, risk and volatility benchmarks default to zero, pricing nominal is 1.0 base
			muUSD, q90USD, q95USD, q99USD = 0.0, 0.0, 0.0, 0.0
		}

		partialRisks[i] = PartialRiskRow{
			asset.Ticker,
			[]float64{mu * 100.0, muUSD},
			[]float64{q90 * 100.0, q90USD},
			[]float64{q95 * 100.0, q95USD},
			[]float64{q99 * 100.0, q99USD},
			mcrVal * 100,  // in percents - 0.1 -> 10%
			cctrVal * 100, // in percents - 0.1 -> 10%
		}
	}

	response := PortfolioRiskMetricsResponse{
		PortfolioVaRQuantiles: VaRQuantiles{
			Mu:  sumPortMuUSD,
			Q90: sumPortQ90USD,
			Q95: sumPortQ95USD,
			Q99: sumPortQ99USD,
		},
		PortfolioValueUSD:       totalPortfolioValue,
		PartialRisks:            partialRisks,
		VolatilityDrag:          totalVolatilityDragSum * 100.0,
		DiversificationRatio:    divRatio,
		EffectiveNumberOfAssets: effNA,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
