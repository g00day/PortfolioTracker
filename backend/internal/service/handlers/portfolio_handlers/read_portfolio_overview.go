package portfoliohandlers

import (
	"encoding/json"
	"net/http"
	"portfolio-tracker/internal/adapter/db/repos"
	"portfolio-tracker/internal/adapter/exchanges_datafetcher"
	"portfolio-tracker/internal/service"
	"portfolio-tracker/internal/service/middlewares"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type RisksScorecard struct {
	RisksMetric int      `json:"risks_metrick"` // 0 to 100
	Reasons     []string `json:"reasons"`
}

type PnLMetricsSummary struct {
	TotalPnL      [2]float64 `json:"total_pnl"` // [in USD, in %]
	DailyPnL      [2]float64 `json:"daily_pnl"`
	HistoricalPnL [2]float64 `json:"historical_pnl"`
}

type PortfolioOverviewResponse struct {
	PLHistory  map[string][]float64 `json:"p&l_history"` // timeline metrics
	Risks      RisksScorecard       `json:"risks"`
	PnLSummary PnLMetricsSummary    `json:"p&l"`
}

func ReadPortfolioOverview(w http.ResponseWriter, r *http.Request) {
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
		json.NewEncoder(w).Encode(PortfolioOverviewResponse{
			PLHistory: map[string][]float64{"portfolio": {0}, "portfolio_FOMO": {0}, "risk_free": {0}},
		})
		return
	}

	cs := exchanges_datafetcher.NewConnectionSettings()
	df := exchanges_datafetcher.NewCrossMarketDataFetcher(&cs, 1000)

	var totalPortfolioValue float64
	var initialCostSum float64
	serviceAssets := make([]service.Asset, len(items))
	assetPrices := make([]float64, len(items))
	tickers := make([]string, len(items))
	rfrRepo := repos.RiskFreeRate{DBConn: dbConn}
	tRepo := repos.Transaction{DBConn: dbConn}
	snapshotRepo := repos.PortfolioItemSnapshot{DBConn: dbConn}

	isItemRiskFreeMap := make(map[int32]bool)
	var riskFreePrincipal float64

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

		isItemRiskFreeMap[item.Id] = isRF
		tickers[i] = ticker

		// Derive historical purchase metrics to calculate true initial value basis
		txs, txErr := tRepo.ReadAllByItem(item.Id)
		if isRF && txErr == nil {
			rateValue := 0.0
			if item.RiskFreeRateId != nil {
				if rate, rateErr := rfrRepo.Read(*item.RiskFreeRateId); rateErr == nil {
					rateValue = rate.Value
				}
			}
			accrued := calculateRiskFreePnL(txs, rateValue, time.Now())
			if item.TotalAmount > 0 {
				price = (float64(item.TotalAmount) + accrued) / float64(item.TotalAmount)
			}
		}
		assetPrices[i] = price
		currentVal := float64(item.TotalAmount) * price
		totalPortfolioValue += currentVal
		if isRF {
			riskFreePrincipal += float64(item.TotalAmount)
		}
		if txErr == nil && len(txs) > 0 {
			var mockTxAssets []service.Asset
			for _, tx := range txs {
				mockTxAssets = append(mockTxAssets, service.Asset{Amount: tx.Amount})
			}
			totalCost := 0.0
			totalTracked := 0.0
			for _, tx := range txs {
				if tx.Type == "deposit" {
					totalCost += tx.Amount * tx.PriceUSD
					totalTracked += tx.Amount
				}
			}
			if totalTracked > 0 {
				avgPrice := totalCost / totalTracked
				initialCostSum += float64(item.TotalAmount) * avgPrice
			} else {
				initialCostSum += currentVal
			}
		} else {
			initialCostSum += currentVal
		}

		serviceAssets[i] = service.Asset{
			Ticker:     ticker,
			Amount:     float64(item.TotalAmount),
			IsRiskFree: isRF,
		}
	}

	if totalPortfolioValue > 0 {
		for i := range serviceAssets {
			serviceAssets[i].W = (serviceAssets[i].Amount * assetPrices[i]) / totalPortfolioValue
		}
	}

	portfolioCalc := &service.Portfolio{
		Assets:      serviceAssets,
		Datafetcher: *df,
	}

	covMatrix := portfolioCalc.BuildCovariationMatrix()
	divRatio := portfolioCalc.CalculateDiversificationRate(covMatrix)
	generalVol := portfolioCalc.CalculateGeneralVolatility(covMatrix)

	volCursor := &service.VolatilityCursor{
		Datafetcher: portfolioCalc.Datafetcher,
	}

	// --- Structural Risk Matrix Breakdown ---
	var totalVolatilityDragSum float64
	for _, asset := range serviceAssets {
		if !asset.IsRiskFree {
			drag := volCursor.CalculateAnnualizedVolatilityDrag(asset.Ticker, "spot")
			totalVolatilityDragSum += asset.W * drag
		}
	}

	// Dynamic calculation of risk metric profile scalar bounds [0, 100]
	riskScore := int(generalVol * 100 * SQRT_365) // baseline scalar bound evaluation mapping
	if riskScore > 100 {
		riskScore = 100
	} else if riskScore < 0 {
		riskScore = 0
	}

	var reasons []string
	if totalVolatilityDragSum*100.0 > 5.0 {
		reasons = append(reasons, "volatility drag")
	}
	if generalVol*SQRT_365 > 0.30 {
		reasons = append(reasons, "High VaR")
	}
	if divRatio < 1.15 {
		reasons = append(reasons, "Low Diversification")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "Optimal allocation balances observed")
	}

	// --- Absolute Trailing P&L Value Assessment Matrix ---
	totalPnLUSD := totalPortfolioValue - initialCostSum
	totalPnLPct := 0.0
	if initialCostSum > 0 {
		totalPnLPct = (totalPnLUSD / initialCostSum) * 100.0
	}

	// --- Historical Timeline Trailing Series Calculations ---
	snapshots, err := snapshotRepo.ReadHistoryByPortfolio(int32(portfolioID))
	var riskFreeSeries []float64
	var portfolioSeries []float64
	var fomoSeries []float64

	if err == nil && len(snapshots) > 0 {
		type DayData struct {
			MarketValue float64
			RFValue     float64
		}
		dailyValues := make(map[string]*DayData)
		var dates []string

		for _, snap := range snapshots {
			if snap.Date == nil {
				continue
			}
			dayStr := snap.Date.Format("2006-01-02")
			if _, exists := dailyValues[dayStr]; !exists {
				dailyValues[dayStr] = &DayData{}
				dates = append(dates, dayStr)
			}
			if isItemRiskFreeMap[snap.PortfolioItemId] {
				dailyValues[dayStr].RFValue += snap.Amount
			} else {
				dailyValues[dayStr].MarketValue += snap.Amount
			}
		}

		sort.Strings(dates)

		if len(dates) > 0 {
			initialRF := dailyValues[dates[0]].RFValue
			initialMarket := dailyValues[dates[0]].MarketValue
			btcKlines := portfolioCalc.Datafetcher.GetKlines("BTCUSDT", "D", "spot")

			for t, date := range dates {
				v := dailyValues[date]
				rfPnL := v.RFValue - initialRF
				marketPnL := v.MarketValue - initialMarket

				riskFreeSeries = append(riskFreeSeries, rfPnL)
				portfolioSeries = append(portfolioSeries, marketPnL)

				fomoPnL := marketPnL * 1.35
				if len(btcKlines) > t && t > 0 {
					initialBTCPrice := btcKlines[0][4]
					currentBTCPrice := btcKlines[t][4]
					if initialBTCPrice > 0 {
						fomoPnL = initialMarket * ((currentBTCPrice - initialBTCPrice) / initialBTCPrice)
					}
				}
				fomoSeries = append(fomoSeries, fomoPnL)
			}
		}
	}

	if len(portfolioSeries) == 0 {
		riskFreeSeries, portfolioSeries, fomoSeries = []float64{0.0}, []float64{0.0}, []float64{0.0}
	}

	annualRiskFreeRate := 0.0
	if rate, rateErr := rfrRepo.Read(portfolioData.RiskFreeRateId); rateErr == nil {
		annualRiskFreeRate = rate.Value
	}
	riskFreeSeries = service.ProjectRiskFreePnL(riskFreeSeries, riskFreePrincipal, annualRiskFreeRate, 30)

	// Deriving chronological Daily and Historical metrics from trailing series
	var dailyPnLUSD, dailyPnLPct float64
	if len(portfolioSeries) > 1 {
		dailyPnLUSD = portfolioSeries[len(portfolioSeries)-1] - portfolioSeries[len(portfolioSeries)-2]
		prevDayTotal := totalPortfolioValue - portfolioSeries[len(portfolioSeries)-1]
		if prevDayTotal > 0 {
			dailyPnLPct = (dailyPnLUSD / prevDayTotal) * 100.0
		}
	}

	var histPnLUSD, histPnLPct float64
	if len(portfolioSeries) > 0 {
		histPnLUSD = portfolioSeries[len(portfolioSeries)-1]
		if initialCostSum > 0 {
			histPnLPct = (histPnLUSD / initialCostSum) * 100.0
		}
	}

	pnlHistory := map[string][]float64{
		"portfolio":      portfolioSeries,
		"portfolio_FOMO": fomoSeries,
		"risk_free":      riskFreeSeries,
	}

	response := PortfolioOverviewResponse{
		PLHistory: pnlHistory,
		Risks: RisksScorecard{
			RisksMetric: riskScore,
			Reasons:     reasons,
		},
		PnLSummary: PnLMetricsSummary{
			TotalPnL:      [2]float64{totalPnLUSD, totalPnLPct},
			DailyPnL:      [2]float64{dailyPnLUSD, dailyPnLPct},
			HistoricalPnL: [2]float64{histPnLUSD, histPnLPct},
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
