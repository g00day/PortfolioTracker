package utils

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const DemoLogin = "demo"
const DemoPassword = "Demo12345!"

// SeedDemoData creates an idempotent account with enough history to make all
// analytics screens useful immediately after docker compose up.
func SeedDemoData(db *pgx.Conn) error {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	hash, err := bcrypt.GenerateFromPassword([]byte(DemoPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	var userID int32
	err = tx.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, 'demo@portfolio.local', $2)
		ON CONFLICT (username) DO UPDATE SET password_hash = EXCLUDED.password_hash
		RETURNING id`, DemoLogin, string(hash)).Scan(&userID)
	if err != nil {
		return fmt.Errorf("seed demo user: %w", err)
	}

	var rateID int32
	err = tx.QueryRow(ctx, `SELECT id FROM risk_free_rate WHERE name = 'US Treasury Demo 5.10%' LIMIT 1`).Scan(&rateID)
	if err == pgx.ErrNoRows {
		err = tx.QueryRow(ctx, `INSERT INTO risk_free_rate (name, value) VALUES ('US Treasury Demo 5.10%', 0.051) RETURNING id`).Scan(&rateID)
	}
	if err != nil {
		return fmt.Errorf("seed demo rate: %w", err)
	}

	var portfolioID int32
	err = tx.QueryRow(ctx, `SELECT id FROM portfolios WHERE user_id = $1 AND name = 'Демо-портфель' LIMIT 1`, userID).Scan(&portfolioID)
	if err == pgx.ErrNoRows {
		err = tx.QueryRow(ctx, `INSERT INTO portfolios (user_id, risk_free_rate_id, name) VALUES ($1, $2, 'Демо-портфель') RETURNING id`, userID, rateID).Scan(&portfolioID)
	}
	if err != nil {
		return fmt.Errorf("seed demo portfolio: %w", err)
	}

	btcID, err := seedDemoItem(ctx, tx, portfolioID, "BTCUSDT", nil, 0.35, 62000)
	if err != nil { return err }
	ethID, err := seedDemoItem(ctx, tx, portfolioID, "ETHUSDT", nil, 4, 2500)
	if err != nil { return err }
	rfID, err := seedDemoItem(ctx, tx, portfolioID, "", &rateID, 8000, 1)
	if err != nil { return err }

	var snapshotCount int
	if err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM portfolio_item_snapshots pis JOIN portfolio_items pi ON pi.id = pis.portfolio_item_id WHERE pi.portfolio_id = $1`, portfolioID).Scan(&snapshotCount); err != nil {
		return err
	}
	if snapshotCount == 0 {
		start := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -44)
		for day := 0; day < 45; day++ {
			date := start.AddDate(0, 0, day)
			btcValue := 21700 + float64(day)*155 + math.Sin(float64(day)*0.48)*1250
			ethValue := 10000 + float64(day)*62 + math.Sin(float64(day)*0.39+1.2)*720
			rfValue := 8000 * math.Pow(1.051, float64(day)/365)
			for _, snapshot := range []struct{ id int32; value float64 }{{btcID, btcValue}, {ethID, ethValue}, {rfID, rfValue}} {
				if _, err = tx.Exec(ctx, `INSERT INTO portfolio_item_snapshots (portfolio_item_id, amount, date) VALUES ($1, $2, $3)`, snapshot.id, snapshot.value, date); err != nil {
					return fmt.Errorf("seed demo snapshot: %w", err)
				}
			}
		}
	}

	return tx.Commit(ctx)
}

func seedDemoItem(ctx context.Context, tx pgx.Tx, portfolioID int32, ticker string, rateID *int32, amount, price float64) (int32, error) {
	var id int32
	if ticker != "" {
		err := tx.QueryRow(ctx, `SELECT id FROM portfolio_items WHERE portfolio_id = $1 AND ticker = $2 LIMIT 1`, portfolioID, ticker).Scan(&id)
		if err == pgx.ErrNoRows { err = tx.QueryRow(ctx, `INSERT INTO portfolio_items (portfolio_id, ticker, total_amount) VALUES ($1, $2, $3) RETURNING id`, portfolioID, ticker, amount).Scan(&id) }
		if err != nil { return 0, fmt.Errorf("seed demo item %s: %w", ticker, err) }
	} else {
		err := tx.QueryRow(ctx, `SELECT id FROM portfolio_items WHERE portfolio_id = $1 AND risk_free_rate_id = $2 LIMIT 1`, portfolioID, *rateID).Scan(&id)
		if err == pgx.ErrNoRows { err = tx.QueryRow(ctx, `INSERT INTO portfolio_items (portfolio_id, risk_free_rate_id, total_amount) VALUES ($1, $2, $3) RETURNING id`, portfolioID, *rateID, amount).Scan(&id) }
		if err != nil { return 0, fmt.Errorf("seed demo risk-free item: %w", err) }
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE portfolio_item_id = $1`, id).Scan(&count); err != nil { return 0, err }
	if count == 0 {
		if _, err := tx.Exec(ctx, `INSERT INTO transactions (portfolio_id, portfolio_item_id, type, amount, price_usd, processed_at) VALUES ($1, $2, 'deposit', $3, $4, NOW() - INTERVAL '45 days')`, portfolioID, id, amount, price); err != nil { return 0, err }
	}
	return id, nil
}
