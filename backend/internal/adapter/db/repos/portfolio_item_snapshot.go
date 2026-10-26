package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type PortfolioItemSnapshotInstance struct {
	Id              int32
	Date            *time.Time
	PortfolioItemId int32
	Amount          float64
}

type PortfolioItemSnapshot struct {
	DBConn *pgx.Conn
}

func (r *PortfolioItemSnapshot) CreateTable() error {
	query := `CREATE TABLE IF NOT EXISTS portfolio_item_snapshots (
		id SERIAL PRIMARY KEY,
		portfolio_item_id INT NOT NULL REFERENCES portfolio_items(id) ON DELETE CASCADE,
		amount DECIMAL NOT NULL,
		date TIMESTAMP NOT NULL DEFAULT NOW()
	);`

	_, err := r.DBConn.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create portfolio_item_snapshots table: %w", err)
	}

	return nil
}

func (r *PortfolioItemSnapshot) Create(data PortfolioItemSnapshotInstance) (int32, error) {
	query := `INSERT INTO portfolio_item_snapshots (
		portfolio_item_id,
		amount,
		date
	) VALUES ($1, $2, COALESCE($3, NOW()))
	RETURNING id;`

	var id int32

	err := r.DBConn.QueryRow(context.Background(), query,
		data.PortfolioItemId,
		data.Amount,
		data.Date,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create portfolio item snapshot instance: %w", err)
	}

	return id, nil
}

func (r *PortfolioItemSnapshot) Read(id int32) (PortfolioItemSnapshotInstance, error) {
	query := `SELECT id, portfolio_item_id, amount, date FROM portfolio_item_snapshots WHERE id = $1;`

	var data PortfolioItemSnapshotInstance

	err := r.DBConn.QueryRow(context.Background(), query, id).Scan(
		&data.Id,
		&data.PortfolioItemId,
		&data.Amount,
		&data.Date,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return PortfolioItemSnapshotInstance{}, fmt.Errorf("portfolio item snapshot not found: %w", err)
		}
		return PortfolioItemSnapshotInstance{}, fmt.Errorf("failed to read portfolio item snapshot: %w", err)
	}

	return data, nil
}
