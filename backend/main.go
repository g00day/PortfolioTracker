package main

import (
	"fmt"
	edf "portfolio-tracker/internal/adapter/exchanges_datafetcher"

	service "portfolio-tracker/internal/service"
)

func main() {
	connectionSettings := edf.NewConnectionSettings()
	df := edf.NewCrossMarketDataFetcher(&connectionSettings, 1000)

	vc := service.VolatilityCursor{Datafetcher: *df}

	assets := []service.Asset{}
	p := service.Portfolio{Datafetcher: *df, Assets: assets}

	fmt.Print(p.CalculateDailyLogarithmicIncome("BTCUSDT"))

	//fmt.Print(vc.PrepareParkinsonVolatilityDistributionData("BTCUSDT", "D", "spot"))
	//fmt.Print(vc.CalculateAnnualizedParkinsonVolatility("BTCUSDT", "spot"))
	//fmt.Print(vc.CalculateAnnualizedVolatilityDrag("BTCUSDT", "spot"))

	fmt.Print(vc.CalculateMediumParkinsonVolatility("BTCUSDT", "D", "spot"))
}
