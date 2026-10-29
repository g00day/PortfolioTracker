package repos

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type PortfolioItemInstance struct {
	Id             int32
	PortfolioId    int32
	Ticker         *string
	RiskFreeRateId *int32
	TotalAmount    float32
}

type PortfolioItem struct {
	DBConn *pgx.Conn
}

func (pi *PortfolioItem) CreateTable() error {
	query := `CREATE TABLE IF NOT EXISTS portfolio_items (
		id SERIAL PRIMARY KEY,
		portfolio_id INT NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
		ticker VARCHAR(20),
		risk_free_rate_id INT REFERENCES risk_free_rate(id) ON DELETE CASCADE,
		total_amount DECIMAL NOT NULL,
		
		CONSTRAINT chk_item_source CHECK (
			(ticker IS NOT NULL AND ticker <> '' AND risk_free_rate_id IS NULL) OR
			(ticker IS NULL AND risk_free_rate_id IS NOT NULL)
		)
	);`

	_, err := pi.DBConn.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create portfolio_items table: %w", err)
	}

	return nil
}

func (pi *PortfolioItem) Create(data PortfolioItemInstance) (int32, error) {
	query := `INSERT INTO portfolio_items (
		portfolio_id,
		ticker,
		risk_free_rate_id,
		total_amount
	) VALUES ($1, $2, $3, $4)
	RETURNING id;`

	var id int32

	err := pi.DBConn.QueryRow(context.Background(), query,
		data.PortfolioId,
		data.Ticker,
		data.RiskFreeRateId,
		data.TotalAmount,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create portfolio item instance: %w", err)
	}

	return id, nil
}

func (pi *PortfolioItem) Read(id int32) (PortfolioItemInstance, error) {
	query := `SELECT id, portfolio_id, ticker, risk_free_rate_id, total_amount 
	          FROM portfolio_items WHERE id = $1;`

	var data PortfolioItemInstance

	err := pi.DBConn.QueryRow(context.Background(), query, id).Scan(
		&data.Id,
		&data.PortfolioId,
		&data.Ticker,
		&data.RiskFreeRateId,
		&data.TotalAmount,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return PortfolioItemInstance{}, fmt.Errorf("portfolio item not found: %w", err)
		}
		return PortfolioItemInstance{}, fmt.Errorf("failed to read portfolio item: %w", err)
	}

	return data, nil
}

func (pi *PortfolioItem) ReadAll() ([]PortfolioItemInstance, error) {
	query := `SELECT id, portfolio_id, ticker, risk_free_rate_id, total_amount FROM portfolio_items;`

	rows, err := pi.DBConn.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all portfolio items: %w", err)
	}
	defer rows.Close()

	var items []PortfolioItemInstance

	for rows.Next() {
		var data PortfolioItemInstance
		err := rows.Scan(
			&data.Id,
			&data.PortfolioId,
			&data.Ticker,
			&data.RiskFreeRateId,
			&data.TotalAmount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan portfolio item row: %w", err)
		}
		items = append(items, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating portfolio items: %w", err)
	}

	return items, nil
}

func (pi *PortfolioItem) UpdateAmount(ctx context.Context, tx pgx.Tx, id int32, newAmount float32) error {
	query := `UPDATE portfolio_items SET total_amount = $1 WHERE id = $2;`

	_, err := tx.Exec(ctx, query, newAmount, id)
	if err != nil {
		return fmt.Errorf("failed to update portfolio item amount: %w", err)
	}
	return nil
}

func (pi *PortfolioItem) Delete(id int32) error {
	query := `DELETE FROM portfolio_items WHERE id = $1;`

	res, err := pi.DBConn.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("failed to delete portfolio item: %w", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("portfolio item not found")
	}

	return nil
}
