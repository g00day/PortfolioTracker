package service

import "math"

func ProjectRiskFreePnL(history []float64, principal, annualRate float64, days int) []float64 {
	result := append([]float64(nil), history...)
	if len(result) == 0 { result = []float64{0} }
	last := result[len(result)-1]
	for day := 1; day <= days; day++ {
		growth := principal * (math.Pow(1+annualRate, float64(day)/365.0) - 1)
		result = append(result, last+growth)
	}
	return result
}
