package main

import (
	"fmt"

	"portfolio-tracker/datafetcher"
)

func main() {
	fetcher := datafetcher.NewCrossMarketDataFetcher(nil, 2000)

	klines := fetcher.GetKlines("BTCUSDT", "D", "spot")

	fmt.Printf("%d\n", len(klines))
	for _, kline := range klines {
		fmt.Println(kline)
	}
}
