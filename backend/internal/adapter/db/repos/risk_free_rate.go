package repos

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type RiskFreeRateInstance struct {
	Id    *int32
	Name  string
	Value float64
}

type RiskFreeRate struct {
	DBConn DBConnection
}

func (rfr *RiskFreeRate) CreateTable() error {
	query := `CREATE TABLE IF NOT EXISTS risk_free_rate (
		id SERIAL PRIMARY KEY,
		name VARCHAR(50) NOT NULL,
		value DECIMAL NOT NULL
	);`

	_, err := rfr.DBConn.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create risk_free_rate table: %w", err)
	}

	return nil
}

func (rfr *RiskFreeRate) Create(data RiskFreeRateInstance) (int32, error) {
	query := `INSERT INTO risk_free_rate (
		name,
		value
	) VALUES ($1, $2) 
	RETURNING id;`

	var id int32

	err := rfr.DBConn.QueryRow(context.Background(), query,
		data.Name,
		data.Value,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create risk_free_rate instance: %w", err)
	}

	return id, nil
}

func (rfr *RiskFreeRate) Read(id int32) (RiskFreeRateInstance, error) {
	query := `SELECT id, name, value FROM risk_free_rate WHERE id = $1;`

	var data RiskFreeRateInstance

	err := rfr.DBConn.QueryRow(context.Background(), query, id).Scan(
		&data.Id,
		&data.Name,
		&data.Value,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return RiskFreeRateInstance{}, fmt.Errorf("risk free rate not found: %w", err)
		}
		return RiskFreeRateInstance{}, fmt.Errorf("failed to read risk free rate: %w", err)
	}

	return data, nil
}

func (rfr *RiskFreeRate) ReadAll() ([]RiskFreeRateInstance, error) {
	query := `SELECT id, name, value FROM risk_free_rate;`

	rows, err := rfr.DBConn.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to query risk free rates: %w", err)
	}
	defer rows.Close()

	var rates []RiskFreeRateInstance

	for rows.Next() {
		var data RiskFreeRateInstance
		err := rows.Scan(
			&data.Id,
			&data.Name,
			&data.Value,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan risk free rate: %w", err)
		}
		rates = append(rates, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating risk free rates: %w", err)
	}

	return rates, nil
}
