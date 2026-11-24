package workers

import (
	"fmt"
	"log"
	"portfolio-tracker/internal/adapter/db/repos"
	"portfolio-tracker/internal/adapter/exchanges_datafetcher"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5"
)

type SnapshotWorker struct {
	DBConn *pgx.Conn
}

func NewSnapshotWorker(dbConn *pgx.Conn) *SnapshotWorker {
	return &SnapshotWorker{DBConn: dbConn}
}

// StartCronJob initializes and starts the daily cron job scheduled for midnight
func (w *SnapshotWorker) StartCronJob() (gocron.Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	// Schedule the job to run every day at 00:00 (midnight)
	_, err = s.NewJob(
		gocron.CronJob("0 0 * * *", false),
		gocron.NewTask(w.MakePortfolioSnapshots),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to schedule snapshot job: %w", err)
	}

	s.Start()
	return s, nil
}

// MakePortfolioSnapshots executes the core operational logic of the worker task
func (w *SnapshotWorker) MakePortfolioSnapshots() {
	log.Println("[CRON WORKER] --- Starting daily portfolio snapshots process ---")

	cs := exchanges_datafetcher.NewConnectionSettings()
	df := exchanges_datafetcher.NewCrossMarketDataFetcher(&cs, 1) // create a datafetcher with only one - actual kline

	piRepo := repos.PortfolioItem{DBConn: w.DBConn}
	snapshotRepo := repos.PortfolioItemSnapshot{DBConn: w.DBConn}

	items, err := piRepo.ReadAll()
	if err != nil {
		log.Printf("[CRON WORKER ERROR] Failed to fetch portfolio items: %v\n", err)
		return
	}

	priceCache := make(map[string]float64) // using cache to avoid duplicating requests to exchanges for the same ticker

	snapshotTime := time.Now().UTC() // unified time for all tickers in that run
	savedCount := 0

	for _, item := range items {
		price, err := w.resolveAssetPrice(df, item.Ticker, priceCache)
		if err != nil {
			log.Printf("[CRON WORKER WARN] Skipping item ID %d: %v\n", item.Id, err)
			continue
		}

		totalUSDValue := float64(item.TotalAmount) * price

		snapshotInput := repos.PortfolioItemSnapshotInstance{
			PortfolioItemId: item.Id,
			Amount:          totalUSDValue,
			Date:            &snapshotTime,
		}

		if _, err := snapshotRepo.Create(snapshotInput); err != nil {
			log.Printf("[CRON WORKER ERROR] Failed to save snapshot for item ID %d: %v\n", item.Id, err)
			continue
		}

		savedCount++
	}

	log.Printf("[CRON WORKER] --- Finished: saved %d/%d snapshots ---\n", savedCount, len(items))
}

func (w *SnapshotWorker) resolveAssetPrice(
	df *exchanges_datafetcher.CrossMarketDataFetcher,
	ticker *string,
	cache map[string]float64,
) (float64, error) {
	if ticker == nil || *ticker == "" {
		return 1.0, nil // risk-free rate
	}

	symbol := *ticker
	if cachedPrice, ok := cache[symbol]; ok {
		return cachedPrice, nil
	}

	klines := df.GetKlines(symbol, "D", "spot")
	if len(klines) == 0 || len(klines[0]) < 5 {
		return 0, fmt.Errorf("no kline data returned for ticker %s", symbol)
	}

	closePrice := klines[0][4]
	if closePrice <= 0 {
		return 0, fmt.Errorf("invalid non-positive price (%f) for ticker %s", closePrice, symbol)
	}

	cache[symbol] = closePrice
	return closePrice, nil
}
