package service

import (
	"math"
	df "portfolio-tracker/internal/adapter/exchanges_datafetcher"
)

type Asset struct {
	Ticker string
	W      float64 // the weight of the asset inside the portfolio
}

type Portfolio struct {
	Assets      []Asset
	Datafetcher df.CrossMarketDataFetcher
}

func (p *Portfolio) CalculateDailyLogarithmicIncome(ticker string) []float64 {
	df := p.Datafetcher

	klines := df.GetKlines(ticker, "D", "spot") // defaultly for D timeframe
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

func (p *Portfolio) CalculateDailyAssetCorrelation(ticker1, ticker2 string) float64 {
	ticker1Income, ticker2Income := p.CalculateDailyLogarithmicIncome(ticker1), p.CalculateDailyLogarithmicIncome(ticker2)

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
		logIncomes[i] = p.CalculateDailyLogarithmicIncome(asset.Ticker)
	}

	for i := 0; i < n; i++ {
		// Does not spend time to calculate logarithmic incomes, covariation etc
		// if it is the same asset
		covariationMatrix[i][i] = 1.0

		for j := i + 1; j < n; j++ {
			corr := CalculateCovariation(logIncomes[i], logIncomes[j])

			// Matrix is symmetric, so fill it mirrorwise
			covariationMatrix[i][j] = corr
			covariationMatrix[j][i] = corr
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
		logIncomes[i] = p.CalculateDailyLogarithmicIncome(asset.Ticker)
	}

	for i := 0; i < n; i++ {
		// Does not spend time to calculate logarithmic incomes, pirson correlation etc
		// if it is the same asset
		correlationMatrix[i][i] = 1.0

		for j := i + 1; j < n; j++ {
			corr := CalculatePirsonCorrelation(logIncomes[i], logIncomes[j])

			// Matrix is symmetric, so fill it mirrorwise
			correlationMatrix[i][j] = corr
			correlationMatrix[j][i] = corr
		}
	}

	return correlationMatrix
}
