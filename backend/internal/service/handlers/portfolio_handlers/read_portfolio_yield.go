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

type AssetYieldInfo struct {
	Ticker        string
	PnLUSD        float64
	PnLPercentage float64
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
	tRepo := repos.Transaction{DBConn: dbConn}
	snapshotRepo := repos.PortfolioItemSnapshot{DBConn: dbConn}

	// Fetch global risk free rate associated with the active portfolio from the DB
	var baselineRiskFreeRate float64
	portfolioRFR, err := rfrRepo.Read(portfolioData.RiskFreeRateId)
	if err == nil {
		baselineRiskFreeRate = portfolioRFR.Value
	} else {
		baselineRiskFreeRate = 0.05 // fallback if global portfolio baseline not found
	}

	isItemRiskFreeMap := make(map[int32]bool)
	var calculatedYields []AssetYieldInfo
	var riskFreePrincipal float64

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

		isItemRiskFreeMap[item.Id] = isRF
		if isRF {
			riskFreePrincipal += float64(item.TotalAmount) * price
		}

		// calculate historical pnl
		txs, txErr := tRepo.ReadAllByItem(item.Id)
		var pnlUSD, pnlPercentage float64
		if txErr == nil && len(txs) > 0 {
			if isRF {
				pnlUSD = calculateRiskFreePnL(txs, assetRFReward, time.Now())
				if item.TotalAmount > 0 {
					pnlPercentage = pnlUSD / float64(item.TotalAmount) * 100
					price = (float64(item.TotalAmount) + pnlUSD) / float64(item.TotalAmount)
				}
			} else {
				var pricedDeposits []service.Asset
				for _, tx := range txs {
					if tx.Type == "deposit" && tx.PriceUSD > 0 {
						pricedDeposits = append(pricedDeposits, service.Asset{Amount: tx.Amount, RiskFreeReward: tx.PriceUSD})
					}
				}
				pnlUSD, pnlPercentage = service.CalculateItemPnL(pricedDeposits, price, float64(item.TotalAmount))
			}
		}

		calculatedYields = append(calculatedYields, AssetYieldInfo{
			Ticker:        ticker,
			PnLUSD:        pnlUSD,
			PnLPercentage: pnlPercentage,
		})

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
		RiskFreeRate: baselineRiskFreeRate,
	}

	covMatrix := portfolioCalc.BuildCovariationMatrix()
	sharpe := portfolioCalc.CalculateSharpeRatio(covMatrix)
	sortino := portfolioCalc.CalculateSortinoRatio()

	structure := make([]StructureRow, len(items))

	for i := range items {
		var weight float64
		amount := serviceAssets[i].Amount

		if totalPortfolioValue > 0 {
			weight = ((amount * assetPrices[i]) / totalPortfolioValue) * 100.0
		}

		// calculate asset pnl to pass into into whe whole portfolio
		var currentAssetPnL float64
		for _, cy := range calculatedYields {
			if cy.Ticker == tickers[i] {
				currentAssetPnL = cy.PnLUSD
				break
			}
		}

		structure[i] = StructureRow{
			tickers[i],
			assetPrices[i],
			amount,
			assetPrices[i] * amount,
			weight,
			currentAssetPnL,
			items[i].Id,
		}
	}

	// Dynamic P&L sort(USD)
	sort.Slice(calculatedYields, func(i, j int) bool {
		return calculatedYields[i].PnLUSD > calculatedYields[j].PnLUSD
	})

	yieldTop3 := make(map[string]YieldTopRow)
	for idx, cy := range calculatedYields {
		if idx >= 3 {
			break
		}
		rankKey := strconv.Itoa(idx + 1)
		yieldTop3[rankKey] = YieldTopRow{cy.Ticker, cy.PnLUSD, cy.PnLPercentage}
	}

	// --- dynamic time series calculations P&L и FOMO ---
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

			// Taking BTCUSDT for calculating FOMO
			btcKlines := portfolioCalc.Datafetcher.GetKlines("BTCUSDT", "D", "spot")

			for t, date := range dates {
				v := dailyValues[date]

				rfPnL := v.RFValue - initialRF
				marketPnL := v.MarketValue - initialMarket

				riskFreeSeries = append(riskFreeSeries, rfPnL)
				portfolioSeries = append(portfolioSeries, marketPnL)

				// Calculating FOMO P&L by close prices [4]
				fomoPnL := marketPnL * 1.35
				if len(btcKlines) > t && t > 0 {
					initialBTCPrice := btcKlines[0][4]
					currentBTCPrice := btcKlines[t][4]

					if initialBTCPrice > 0 {
						pctChange := (currentBTCPrice - initialBTCPrice) / initialBTCPrice
						fomoPnL = initialMarket * pctChange
					}
				}
				fomoSeries = append(fomoSeries, fomoPnL)
			}
		}
	}

	if len(portfolioSeries) == 0 {
		riskFreeSeries = []float64{0.0}
		portfolioSeries = []float64{0.0}
		fomoSeries = []float64{0.0}
	}

	// Extend the benchmark beyond the historical portfolio series so the UI
	// can display a 30-day risk-free forecast.
	riskFreeSeries = service.ProjectRiskFreePnL(riskFreeSeries, riskFreePrincipal, baselineRiskFreeRate, 30)

	var covFomo, corrFomo float64
	if len(portfolioSeries) > 1 {
		covFomo = service.CalculateCovariation(portfolioSeries, fomoSeries)
		corrFomo = service.CalculatePirsonCorrelation(portfolioSeries, fomoSeries)

	}

	// p&l mock-data
	pnlHistory := map[string][]float64{
		"risk_free":      riskFreeSeries,
		"portfolio":      portfolioSeries,
		"portfolio_FOMO": fomoSeries,
	}

	response := PortfolioYieldResponse{
		PL:                 pnlHistory,
		YieldTop3:          yieldTop3,
		PortfolioStructure: structure,
		SharpesRatio:       sharpe,
		SortinoRatio:       sortino,
		CovFomoPnl:         covFomo,  // cov(FOMO, p&l) mock
		CorrFomoPnl:        corrFomo, // corr(FOMO, p&l) mock
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
