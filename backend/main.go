package main

import (
	"fmt"

	edf "portfolio-tracker/internal/adapter/exchanges_datafetcher"
)

func main() {
	fetcher := edf.NewCrossMarketDataFetcher(nil, 1000)

	klines := fetcher.GetKlines("BTCUSDT", "D", "spot")

	fmt.Printf("%d\n", len(klines))
	for _, kline := range klines {
		fmt.Println(kline)
	}
}
