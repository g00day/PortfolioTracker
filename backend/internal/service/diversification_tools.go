package service

import df "portfolio-tracker/internal/adapter/exchanges_datafetcher"

type DiversificationCursor struct {
	Datafetcher df.CrossMarketDataFetcher
}
