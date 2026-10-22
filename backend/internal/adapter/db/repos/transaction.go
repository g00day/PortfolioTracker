package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type TransactionInstance struct {
	Id            *int32
	PortfolioId   int32
	PortfolioItem int32
	ProcessedAt   *time.Time
	PriceUSD      float64
	FeeUSDT       float64
}

type Transaction struct {
	DBConn *pgx.Conn
}

func (t *Transaction) CreateTable() error {
	query := `CREATE TABLE IF NOT EXISTS transactions (
		id SERIAL PRIMARY KEY,
		portfolio_id INT NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
		portfolio_item_id INT NOT NULL REFERENCES portfolio_items(id) ON DELETE CASCADE,
		price_usd DECIMAL NOT NULL,
		fee_usdt DECIMAL NOT NULL,
		processed_at TIMESTAMP NOT NULL DEFAULT NOW()
	);`

	_, err := t.DBConn.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create transactions table: %w", err)
	}

	return nil
}

func (t *Transaction) Create(data TransactionInstance) (int32, error) {
	query := `INSERT INTO transactions (
		portfolio_id,
		portfolio_item_id,
		price_usd,
		fee_usdt,
		processed_at
	) VALUES ($1, $2, $3, $4, COALESCE($5, NOW()))
	RETURNING id;`

	var id int32

	err := t.DBConn.QueryRow(context.Background(), query,
		data.PortfolioId,
		data.PortfolioItem,
		data.PriceUSD,
		data.FeeUSDT,
		data.ProcessedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create transaction instance: %w", err)
	}

	return id, nil
}

func (t *Transaction) Read(id int32) (TransactionInstance, error) {
	query := `SELECT id, portfolio_id, portfolio_item_id, price_usd, fee_usdt, processed_at 
	          FROM transactions WHERE id = $1;`

	var data TransactionInstance

	err := t.DBConn.QueryRow(context.Background(), query, id).Scan(
		&data.Id,
		&data.PortfolioId,
		&data.PortfolioItem,
		&data.PriceUSD,
		&data.FeeUSDT,
		&data.ProcessedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return TransactionInstance{}, fmt.Errorf("transaction not found: %w", err)
		}
		return TransactionInstance{}, fmt.Errorf("failed to read transaction: %w", err)
	}

	return data, nil
}
