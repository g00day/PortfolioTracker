package service

import (
	"math"
	df "portfolio-tracker/internal/adapter/exchanges_datafetcher"
)

type Asset struct {
	Ticker     string
	W          float64 // the weight of the asset inside the portfolio
	Volatility float64 // volatility(whatever is needed: minutely, monthly, annualy)
}

type Portfolio struct {
	Assets      []Asset
	Datafetcher df.CrossMarketDataFetcher
}

type screenedAsset struct {
	Ticker     string
	MeanIncome float64
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
