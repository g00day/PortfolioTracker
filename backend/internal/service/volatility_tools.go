package service

import (
	"math"
	df "portfolio-tracker/internal/adapter/exchanges_datafetcher"
)

type VolatilityCursor struct {
	Datafetcher df.CrossMarketDataFetcher
}

func (vc *VolatilityCursor) CalculateAnnualizedParkinsonVolatility(symbol, marketType string) float64 {
	df := vc.Datafetcher

	klines := df.GetKlines(symbol, "D", marketType) // using D timeframe to calculate volatility

	partialVolatilitiesSumm := 0.0
	validCandlesCount := 0 // count only valid candles

	for _, candle := range klines {
		high := candle[2]
		low := candle[3]

		// defence from anomalies(zero division, ln(0), high <= 0)
		if low <= 0 || high <= 0 {
			continue
		}

		partialVolatility := math.Pow(math.Log(high/low), 2)
		partialVolatilitiesSumm += partialVolatility
		validCandlesCount++
	}

	parkinsonConstant := 4 * float64(validCandlesCount) * math.Log(2) // calculate parkinson contant for only a valid days

	return math.Sqrt((365 / parkinsonConstant) * partialVolatilitiesSumm)
}

func (vc *VolatilityCursor) PrepareParkinsonVolatilityDistributionData(
	symbol string, period string, marketType string,
) ([][2]float64, float64, error) {

	df := vc.Datafetcher
	klines := df.GetKlines(symbol, period, marketType)

	totalKlines := len(klines)
	if totalKlines == 0 {
		return [][2]float64{}, 0.0, nil
	}

	volatilities := make([]float64, 0, totalKlines)
	// Parkinson constant denominator: 2 * sqrt(ln(2))
	parkinsonConstant := 2.0 * math.Sqrt(math.Log(2.0))

	for _, candle := range klines {
		high := candle[2]
		low := candle[3]

		// Calculate volatility for a single candle
		vol := math.Log(high/low) / parkinsonConstant
		volatilities = append(volatilities, vol)
	}

	totalValid := len(volatilities)
	if totalValid == 0 {
		return [][2]float64{}, 0.0, nil
	}

	// Base price projection from the Close price of the latest candle
	lastPrice := klines[0][4]
	if lastPrice <= 0 {
		lastPrice = 1.0
	}

	// Partition volatility range [0.0, 1.0] into 1000 bins with a step of 0.001
	const binCount = 1000
	const binStep = 0.001
	counts := make([]int, binCount)

	for _, vol := range volatilities {
		if vol >= 0.0 && vol < 1.0 {
			binIdx := int(vol / binStep)
			if binIdx >= binCount {
				binIdx = binCount - 1 // Handle edge case where vol is exactly 1.0
			}
			counts[binIdx]++
		}
	}

	distributionData := make([][2]float64, 0, binCount)
	var expectedValue float64

	for i := 0; i < binCount; i++ {
		// Upper interval boundary
		n := float64(i+1) * binStep

		probability := float64(counts[i]) / float64(totalValid)
		volPrice := n * lastPrice // Absolute price delta at current volatility

		distributionData = append(distributionData, [2]float64{volPrice, probability})
		expectedValue += volPrice * probability
	}

	return distributionData, expectedValue, nil
}
