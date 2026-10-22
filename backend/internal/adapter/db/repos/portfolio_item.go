package repos

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type PortfolioItemInstance struct {
	Id             int32
	PortfolioId    int32
	AssetId        *int32 // Может быть null
	RiskFreeRateId *int32 // Может быть null
	TotalAmount    float32
}

type PortfolioItem struct {
	DBConn *pgx.Conn
}

func (pi *PortfolioItem) CreateTable() error {
	query := `CREATE TABLE IF NOT EXISTS portfolio_items (
		id SERIAL PRIMARY KEY,
		portfolio_id INT NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
		asset_id INT REFERENCES assets(id) ON DELETE CASCADE,
		risk_free_rate_id INT REFERENCES risk_free_rate(id) ON DELETE CASCADE,
		total_amount DECIMAL NOT NULL,
		
		-- Ограничение: либо asset_id, либо risk_free_rate_id должен быть заполнен, но не оба сразу
		CONSTRAINT chk_item_source CHECK (
			(asset_id IS NOT NULL AND risk_free_rate_id IS NULL) OR
			(asset_id IS NULL AND risk_free_rate_id IS NOT NULL)
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
		asset_id,
		risk_free_rate_id,
		total_amount
	) VALUES ($1, $2, $3, $4)
	RETURNING id;`

	var id int32

	err := pi.DBConn.QueryRow(context.Background(), query,
		data.PortfolioId,
		data.AssetId,
		data.RiskFreeRateId,
		data.TotalAmount,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create portfolio item instance: %w", err)
	}

	return id, nil
}

func (pi *PortfolioItem) Read(id int32) (PortfolioItemInstance, error) {
	query := `SELECT id, portfolio_id, asset_id, risk_free_rate_id, total_amount 
	          FROM portfolio_items WHERE id = $1;`

	var data PortfolioItemInstance

	err := pi.DBConn.QueryRow(context.Background(), query, id).Scan(
		&data.Id,
		&data.PortfolioId,
		&data.AssetId,
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
