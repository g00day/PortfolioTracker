package repos

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type AssetInstance struct {
	Id     int32
	Ticker string
	Name   string
}

type Asset struct {
	DBConn *pgx.Conn
}

func (a *Asset) CreateTable() error {
	query := `CREATE TABLE IF NOT EXISTS assets (
		id SERIAL PRIMARY KEY,
		ticker VARCHAR(10) NOT NULL UNIQUE,
		name VARCHAR(100) NOT NULL
	);`

	_, err := a.DBConn.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create assets table: %w", err)
	}

	return nil
}

func (a *Asset) Create(data AssetInstance) (int32, error) {
	query := `INSERT INTO assets (
		ticker,
		name
	) VALUES ($1, $2)
	RETURNING id;`

	var id int32

	err := a.DBConn.QueryRow(context.Background(), query,
		data.Ticker,
		data.Name,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create asset instance: %w", err)
	}

	return id, nil
}

func (a *Asset) Read(id int32) (AssetInstance, error) {
	query := `SELECT id, ticker, name FROM assets WHERE id = $1;`

	var data AssetInstance

	err := a.DBConn.QueryRow(context.Background(), query, id).Scan(
		&data.Id,
		&data.Ticker,
		&data.Name,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return AssetInstance{}, fmt.Errorf("asset not found: %w", err)
		}
		return AssetInstance{}, fmt.Errorf("failed to read asset: %w", err)
	}

	return data, nil
}
