package portfoliohandlers

import (
	"math"
	"portfolio-tracker/internal/adapter/db/repos"
	"time"
)

// calculateRiskFreePnL accrues every cash flow independently from its
// transaction date. Deposits add principal and withdrawals remove it.
func calculateRiskFreePnL(transactions []repos.TransactionInstance, annualRate float64, now time.Time) float64 {
	var pnl float64
	for _, transaction := range transactions {
		if transaction.Amount <= 0 || transaction.ProcessedAt == nil {
			continue
		}
		days := now.Sub(*transaction.ProcessedAt).Hours() / 24
		if days <= 0 {
			continue
		}
		sign := 1.0
		if transaction.Type == "withdraw" {
			sign = -1
		}
		pnl += sign * transaction.Amount * (math.Pow(1+annualRate, days/365.0) - 1)
	}
	return pnl
}
