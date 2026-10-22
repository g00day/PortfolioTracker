package main

import (
	"fmt"
	edf "portfolio-tracker/internal/adapter/exchanges_datafetcher"

	service "portfolio-tracker/internal/service"
)

func main() {
	connectionSettings := edf.NewConnectionSettings()
	df := edf.NewCrossMarketDataFetcher(&connectionSettings, 1000)

	//vc := service.VolatilityCursor{Datafetcher: *df}

	assets := []service.Asset{
		service.Asset{Ticker: "BTCUSDT", W: 0.5},
		service.Asset{Ticker: "ETHUSDT", W: 0.5},
	}
	p := service.Portfolio{Datafetcher: *df, Assets: assets}

	//fmt.Print(p.BuildCovariationMatrix())

	//fmt.Print(p.CalculateDailyLogarithmicIncome("BTCUSDT"))

	//fmt.Print(vc.PrepareParkinsonVolatilityDistributionData("BTCUSDT", "D", "spot"))
	//fmt.Print(vc.CalculateAnnualizedParkinsonVolatility("BTCUSDT", "spot"))
	//fmt.Print(vc.CalculateAnnualizedVolatilityDrag("BTCUSDT", "spot"))

	//fmt.Println(vc.CalculateMediumParkinsonVolatility("BTCUSDT", "D", "spot"))

	//fmt.Print(p.CalculateGeneralVolatility())
	fmt.Println(p.CalculateMarginalAndComponentContributions())
	fmt.Println(p.CalculateDiversificationRate())
	fmt.Println(p.CalculateEffectiveNA())
}
