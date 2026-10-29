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
	df := exchanges_datafetcher.NewCrossMarketDataFetcher(&cs, 1000)

	// Instantiating repository abstraction handles
	piRepo := repos.PortfolioItem{DBConn: w.DBConn}
	snapshotRepo := repos.PortfolioItemSnapshot{DBConn: w.DBConn}

	items, err := piRepo.ReadAll()
	if err != nil {
		log.Printf("[CRON WORKER ERROR] Failed to fetch portfolio items from repository: %v\n", err)
		return
	}

	for _, item := range items {
		var currentAssetPrice float64

		// Asset price calculation logic (similar to your analytics handlers)
		if item.Ticker != nil {
			klines := df.GetKlines(*item.Ticker, "D", "spot")
			if len(klines) > 0 {
				currentAssetPrice = klines[0][4]
			}
		} else {
			currentAssetPrice = 1.0 // Fixed risk-free rate price pegged to $1
		}

		// Calculate the final USD valuation of the position
		totalUSDValue := float64(item.TotalAmount) * currentAssetPrice
		now := time.Now()

		snapshotInput := repos.PortfolioItemSnapshotInstance{
			PortfolioItemId: item.Id,
			Amount:          totalUSDValue,
			Date:            &now,
		}

		// Persist the snapshot into the database via the repository layer
		_, err = snapshotRepo.Create(snapshotInput)
		if err != nil {
			log.Printf("[CRON WORKER ERROR] Failed to save snapshot via repository for item ID %d: %v\n", item.Id, err)
			continue
		}
	}

	log.Println("[CRON WORKER] --- Daily portfolio snapshots process finished successfully ---")
}
