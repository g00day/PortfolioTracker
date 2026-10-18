package main

import (
	edf "portfolio-tracker/internal/adapter/exchanges_datafetcher"

	service "portfolio-tracker/internal/service"
)

func main() {
	connectionSettings := edf.NewConnectionSettings()
	df := edf.NewCrossMarketDataFetcher(&connectionSettings, 1000)

	vc := service.VolatilityCursor{Datafetcher: *df}

	//fmt.Print(vc.PrepareParkinsonVolatilityDistributionData("BTCUSDT", "D", "spot"))
	//fmt.Print(vc.CalculateAnnualizedParkinsonVolatility("BTCUSDT", "spot"))
	//fmt.Print(vc.CalculateAnnualizedVolatilityDrag("BTCUSDT", "spot"))
}
