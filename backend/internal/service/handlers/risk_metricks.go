package handlers

import (
	"fmt"
	edf "portfolio-tracker/internal/adapter/exchanges_datafetcher"
	"portfolio-tracker/internal/service"
)

type LognormalVaRUSD struct {
	Mu  []float64 // like [0.07 - percentage(%), value in USD]
	Q90 []float64
	Q95 []float64
	Q99 []float64
}

type AssetLognormalVaR struct {
	Ticker    string
	quantiles LognormalVaRUSD
}

type ContributionMetricks struct {
	Mcr  float64
	Ccrt float64
}

type PartialRiskAsset struct {
	Ticker        string
	VaR           LognormalVaRUSD
	Contributions ContributionMetricks
}

func PresentRiskMetricks(p *service.Portfolio) {
	cs := edf.NewConnectionSettings()
	df := edf.NewCrossMarketDataFetcher(&cs, 1000)
	vc := service.VolatilityCursor{Datafetcher: *df}
	assets := p.Assets

	// lognormal volatility quantiles(Mu, Q90, Q95, Q99)
	lognormalVaRs := []LognormalVaRUSD{}

	for _, asset := range assets {

		ticker := asset.Ticker
		amount := asset.Amount

		assetKlines := df.GetKlines(ticker, "W", "spot")
		lastAssetPrice := assetKlines[0][4]

		q90, q95, q99 := vc.CalculateLognormalParkinsonVaR(assetKlines)
		_, mu, err := vc.PrepareParkinsonVolatilityDistributionData(assetKlines)
		if err != nil {
			fmt.Print("failed to get volatility distribution")
		}

		q90USD := lastAssetPrice * amount * q90
		q95USD := lastAssetPrice * amount * q95
		q99USD := lastAssetPrice * amount * q99
		muUSD := lastAssetPrice * amount * mu

		lognormalVaR := LognormalVaRUSD{
			Mu:  []float64{mu, muUSD},
			Q90: []float64{q90, q90USD},
			Q95: []float64{q95, q95USD},
			Q99: []float64{q99, q99USD},
		}

		lognormalVaRs = append(lognormalVaRs, lognormalVaR)

	}

	// Value at Risk(USD): mu, Q90, Q95, Q99
	totalUSDVar := []float64{0.0, 0.0, 0.0, 0.0}

	for _, value := range lognormalVaRs {
		muUSD := value.Mu[1]
		q90USD := value.Q90[1]
		q95USD := value.Q95[1]
		q99USD := value.Q99[1]

		totalUSDVar[0] += muUSD
		totalUSDVar[1] += q90USD
		totalUSDVar[2] += q95USD
		totalUSDVar[3] += q99USD
	}

	covariationMatrix := p.BuildCovariationMatrix() // Σ-matrix
	mcr, ccrt := p.CalculateMarginalAndComponentContributions(covariationMatrix)

	totalRiskTable := []PartialRiskAsset{}

	// combining VaRs and it's contribution metricks
	for i := 0; i < len(assets); i++ {
		asset := assets[i]

		ticker := asset.Ticker
		lognomalVaR := lognormalVaRs[i]

		mcri, ccrti := mcr[i], ccrt[i]
		contributionMetricks := ContributionMetricks{Mcr: mcri, Ccrt: ccrti}
		partialRisks := PartialRiskAsset{Ticker: ticker,
			VaR:           lognomalVaR,
			Contributions: contributionMetricks,
		}

		totalRiskTable = append(totalRiskTable, partialRisks)
	}

}
