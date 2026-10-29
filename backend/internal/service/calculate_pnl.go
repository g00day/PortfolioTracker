package service

import "math"

func CalculateItemPnL(transactions []Asset, currentPrice, currentAmount float64) (float64, float64) {
	if len(transactions) == 0 || currentAmount <= 0 {
		return 0.0, 0.0
	}

	totalCost := 0.0
	totalAmountTracked := 0.0

	// calculate WAC only by deposit trasactions
	for _, tx := range transactions {
		if tx.Amount > 0 {
			totalCost += tx.Amount * tx.RiskFreeReward
			totalAmountTracked += tx.Amount
		}
	}

	if totalAmountTracked == 0 {
		return 0.0, 0.0
	}

	// average deposit buy
	avgPrice := totalCost / totalAmountTracked

	// market price for the rest of the assets
	currentValue := currentAmount * currentPrice

	// historical deposit volume of buy/deposit transactions
	initialValue := currentAmount * avgPrice

	pnlUSD := currentValue - initialValue

	// Percentual result
	pnlPercentage := 0.0
	if initialValue > 0 {
		pnlPercentage = (pnlUSD / initialValue) * 100.0
	}

	// defence from returning -0.0, -0.0
	if math.Abs(pnlUSD) < 0.000001 {
		pnlUSD = 0.0
		pnlPercentage = 0.0
	}

	return pnlUSD, pnlPercentage
}
