package service

import (
	"math"
	df "portfolio-tracker/internal/adapter/exchanges_datafetcher"
)

type Asset struct {
	Ticker string
	Amount float64
	W      float64 // the weight of the asset inside the portfolio

	// specially for risk-free rates
	IsRiskFree     bool
	RiskFreeReward float64 // only percents: 0.06, 0.035
}

type Portfolio struct {
	Assets       []Asset
	Datafetcher  df.CrossMarketDataFetcher
	RiskFreeRate float64
}

const SQRT_365 = 19.1049731745428

func (p *Portfolio) CalculateDailyRiskFreeLogarithmicIncome(value float64) float64 {
	dailyIncome := math.Log(1+value) / 365

	return dailyIncome
}

func (p *Portfolio) CalculateDailyLogarithmicIncome(asset Asset) []float64 {
	// If the asset is a risk-free rate, compute its constant daily log return directly from reward rate
	if asset.IsRiskFree {
		dailyIncome := p.CalculateDailyRiskFreeLogarithmicIncome(asset.RiskFreeReward)
		return []float64{dailyIncome}
	}

	df := p.Datafetcher

	klines := df.GetKlines(asset.Ticker, "D", "spot") // defaultly for D timeframe
	var logarithmicIncomes []float64

	for i := 0; i < len(klines); i++ {

		if i == 0 {
			continue
		}

		p_0 := klines[i-1][4] // close price of previous day
		p_1 := klines[i][4]   // close day of currently iterated day

		logarithmicIncomes = append(logarithmicIncomes, math.Log(p_1/p_0))
	}

	return logarithmicIncomes
}

func (p *Portfolio) CalculateDailyAssetCorrelation(asset1, asset2 Asset) float64 {
	// Risk-free rates have zero volatility, resulting in an undefined mathematical correlation (0/0).
	// We handle this by explicitly defaulting their cross-asset correlation to 0.0.
	if asset1.IsRiskFree || asset2.IsRiskFree {
		return 0.0
	}

	ticker1Income, ticker2Income := p.CalculateDailyLogarithmicIncome(asset1), p.CalculateDailyLogarithmicIncome(asset2)

	correlation := CalculatePirsonCorrelation(ticker1Income, ticker2Income)

	return correlation
}

func (p *Portfolio) BuildCovariationMatrix() [][]float64 {
	assets := p.Assets
	n := len(assets)

	// allocate memory for matrix
	covariationMatrix := make([][]float64, n)
	for i := range covariationMatrix {
		covariationMatrix[i] = make([]float64, n)
	}

	// Cache logarithmic incomes so we don't recalculate them
	logIncomes := make([][]float64, n)
	for i, asset := range assets {
		logIncomes[i] = p.CalculateDailyLogarithmicIncome(asset)
	}

	for i := 0; i < n; i++ {
		// Risk-free rates do not fluctuate, meaning their variance is zero
		if assets[i].IsRiskFree {
			covariationMatrix[i][i] = 0.0
			continue
		}

		// Does not spend time to calculate logarithmic incomes, covariation etc
		// if it is the same asset
		covariationMatrix[i][i] = CalculateVariance(logIncomes[i])

		for j := i + 1; j < n; j++ {
			// Cross-covariance involving a risk-free rate is always zero
			if assets[j].IsRiskFree {
				covariationMatrix[i][j] = 0.0
				covariationMatrix[j][i] = 0.0
				continue
			}

			cov := CalculateCovariation(logIncomes[i], logIncomes[j])

			// Matrix is symmetric, so fill it mirrorwise
			covariationMatrix[i][j] = cov
			covariationMatrix[j][i] = cov
		}
	}

	return covariationMatrix
}

func (p *Portfolio) BuildPirsonCorrelationMatrix() [][]float64 {
	assets := p.Assets
	n := len(assets)

	// allocate memory for matrix
	correlationMatrix := make([][]float64, n)
	for i := range correlationMatrix {
		correlationMatrix[i] = make([]float64, n)
	}

	// Cache logarithmic incomes so we don't recalculate them
	logIncomes := make([][]float64, n)
	for i, asset := range assets {
		logIncomes[i] = p.CalculateDailyLogarithmicIncome(asset)
	}

	for i := 0; i < n; i++ {
		// Does not spend time to calculate logarithmic incomes, pirson correlation etc
		// if it is the same asset
		correlationMatrix[i][i] = 1.0

		for j := i + 1; j < n; j++ {
			// Risk-free rates have zero correlation to market assets
			if assets[i].IsRiskFree || assets[j].IsRiskFree {
				correlationMatrix[i][j] = 0.0
				correlationMatrix[j][i] = 0.0
				continue
			}

			corr := CalculatePirsonCorrelation(logIncomes[i], logIncomes[j])

			// Matrix is symmetric, so fill it mirrorwise
			correlationMatrix[i][j] = corr
			correlationMatrix[j][i] = corr
		}
	}

	return correlationMatrix
}

func (p *Portfolio) CalculateGeneralVolatility(covariationMatrix [][]float64) float64 {
	assets := p.Assets
	n := len(assets)

	// if the portfolio is empty then return zero
	if n == 0 {
		return 0.0
	}

	variance := 0.0

	// w^T * Sigma * w
	for i := 0; i < n; i++ {
		w_i := assets[i].W

		for j := 0; j < n; j++ {
			w_j := assets[j].W
			sigma_ij := covariationMatrix[i][j]

			variance += w_i * w_j * sigma_ij
		}
	}

	// volatility is the square root of variance
	return math.Sqrt(variance)
}

func (p *Portfolio) CalculateDiversificationRate(covariationMatrix [][]float64) float64 {
	weightedVolatilitySumm := 0.0

	assets := p.Assets

	if len(assets) == 0 {
		return 0.0
	}

	generalVolatility := p.CalculateGeneralVolatility(covariationMatrix) // portfolio variance

	// Safeguard against division by zero if portfolio contains only risk-free assets
	if generalVolatility == 0 {
		return 0.0
	}

	for i, asset := range assets {

		assetVolatility := math.Sqrt(covariationMatrix[i][i])
		w_i := asset.W

		weightedVolatilitySumm += w_i * assetVolatility
	}

	return weightedVolatilitySumm / generalVolatility
}

func (p *Portfolio) CalculateMarginalAndComponentContributions(covariationMatrix [][]float64) ([]float64, []float64) {
	assets := p.Assets
	n := len(assets)

	// if the portfolio is empty return an empty slice
	if n == 0 {
		return []float64{}, []float64{}
	}

	generalVolatility := p.CalculateGeneralVolatility(covariationMatrix) // general portfolio volatility

	// if the volatility is equals to zero return empty slices
	if generalVolatility == 0 {
		return []float64{}, []float64{}
	}

	mcr := make([]float64, n)
	cctr := make([]float64, n)

	// 3. Calculate MCR and CCR througout one loop
	for i := 0; i < n; i++ {
		sigmaW_i := 0.0
		for j := 0; j < n; j++ {
			sigmaW_i += covariationMatrix[i][j] * assets[j].W
		}

		// MCR
		mcr[i] = (sigmaW_i / generalVolatility) * SQRT_365

		// CCTR = w_i * MCR_i
		cctr[i] = (assets[i].W * mcr[i])
	}

	return mcr, cctr
}

func (p *Portfolio) CalculateEffectiveNA() float64 {
	assets := p.Assets

	if len(assets) == 0 {
		return 0.0
	}

	weightedSum := 0.0

	for _, asset := range assets {
		w_i := asset.W
		weightedSum += w_i * w_i
	}

	return 1 / weightedSum
}

func (p *Portfolio) CalculateAnnualizedPortfolioReturn() float64 {
	assets := p.Assets
	n := len(assets)
	if n == 0 {
		return 0.0
	}

	portfolioReturn := 0.0

	for _, asset := range assets {
		// Calculate the annualized yield of a risk-free rate directly without klines
		if asset.IsRiskFree {
			annualAssetReturn := math.Log(1 + asset.RiskFreeReward)
			portfolioReturn += asset.W * annualAssetReturn
			continue
		}

		dailyIncomes := p.CalculateDailyLogarithmicIncome(asset)
		if len(dailyIncomes) == 0 {
			continue
		}

		// Calculate the mean of daily logarithmic incomes
		totalDailyIncome := 0.0
		for _, income := range dailyIncomes {
			totalDailyIncome += income
		}
		avgDailyReturn := totalDailyIncome / float64(len(dailyIncomes))

		// Annualize daily log return by multiplying by 365 days
		annualAssetReturn := avgDailyReturn * 365.0

		// Add weighted asset return to the total portfolio return
		portfolioReturn += asset.W * annualAssetReturn
	}

	return portfolioReturn
}

func (p *Portfolio) CalculateSharpeRatio(covariationMatrix [][]float64) float64 {
	annualVolatility := p.CalculateGeneralVolatility(covariationMatrix) * SQRT_365
	if annualVolatility == 0 {
		return 0.0
	}

	// Get the annualized portfolio return
	annualReturn := p.CalculateAnnualizedPortfolioReturn()

	// Sharpe Ratio = (Portfolio Return - Risk-Free Rate) / Portfolio Volatility
	sharpeRatio := (annualReturn - p.RiskFreeRate) / annualVolatility

	return sharpeRatio
}

func (p *Portfolio) CalculateAnnualizedDownsideDeviation() float64 {
	assets := p.Assets
	n := len(assets)
	if n == 0 {
		return 0.0
	}

	// Convert annualized risk-free rate to a daily rate
	dailyRf := p.RiskFreeRate / 365.0

	// Cache logarithmic incomes for all assets to build portfolio daily history
	logIncomes := make([][]float64, n)
	var minHistoryLen int
	hasMarketAssets := false

	for i, asset := range assets {
		if asset.IsRiskFree {
			logIncomes[i] = p.CalculateDailyLogarithmicIncome(asset)
			continue
		}

		incomes := p.CalculateDailyLogarithmicIncome(asset)
		if !hasMarketAssets || len(incomes) < minHistoryLen {
			minHistoryLen = len(incomes)
		}
		logIncomes[i] = incomes
		hasMarketAssets = true
	}

	// If the portfolio only contains risk-free instruments, market downside risk is non-existent
	if !hasMarketAssets || minHistoryLen == 0 {
		return 0.0
	}

	// Construct historical daily portfolio returns and compute downside variance
	downsideSum := 0.0
	totalDays := 0

	for t := 0; t < minHistoryLen; t++ {
		portfolioDailyReturn := 0.0
		for i, asset := range assets {
			if asset.IsRiskFree {
				// Risk free returns use a single scalar index cached on position [0]
				portfolioDailyReturn += asset.W * logIncomes[i][0]
			} else {
				portfolioDailyReturn += asset.W * logIncomes[i][t]
			}
		}

		// Calculate downside deviation relative to the daily risk-free rate
		excessReturn := portfolioDailyReturn - dailyRf
		if excessReturn < 0 {
			downsideSum += excessReturn * excessReturn
		}
		totalDays++
	}

	if totalDays == 0 {
		return 0.0
	}

	// Calculate daily downside variance and convert it to annualized downside deviation
	dailyDownsideVariance := downsideSum / float64(totalDays)
	dailyDownsideDeviation := math.Sqrt(dailyDownsideVariance)

	return dailyDownsideDeviation * SQRT_365
}

func (p *Portfolio) CalculateSortinoRatio() float64 {
	// Annualize portfolio return
	annualReturn := p.CalculateAnnualizedPortfolioReturn()

	// Calculate annualized downside deviation
	annualDownsideDeviation := p.CalculateAnnualizedDownsideDeviation()
	if annualDownsideDeviation == 0 {
		return 0.0
	}
	// Sortino Ratio = (Portfolio Return - Risk-Free Rate) / Downside Deviation
	sortinoRatio := (annualReturn - p.RiskFreeRate) / annualDownsideDeviation
	return sortinoRatio
}
