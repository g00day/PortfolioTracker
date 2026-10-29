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
	Type          string  // "deposit" or "withdraw"
	Amount        float64 // amount in asset
	PriceUSD      float64 // historical price of asset
	ProcessedAt   *time.Time
}

type Transaction struct {
	DBConn *pgx.Conn
}

func (t *Transaction) CreateTable() error {
	createEnumQuery := `
	DO $$ 
	BEGIN 
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'transaction_type') THEN 
			CREATE TYPE transaction_type AS ENUM ('deposit', 'withdraw'); 
		END IF; 
	END $$;`

	_, err := t.DBConn.Exec(context.Background(), createEnumQuery)
	if err != nil {
		return fmt.Errorf("failed to create transaction_type enum: %w", err)
	}

	query := `CREATE TABLE IF NOT EXISTS transactions (
		id SERIAL PRIMARY KEY,
		portfolio_id INT NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
		portfolio_item_id INT NOT NULL REFERENCES portfolio_items(id) ON DELETE CASCADE,
		type transaction_type NOT NULL,
		amount DECIMAL NOT NULL,
		price_usd DECIMAL NOT NULL,
		processed_at TIMESTAMP NOT NULL DEFAULT NOW()
	);`

	_, err = t.DBConn.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create transactions table: %w", err)
	}

	return nil
}

func (t *Transaction) CreateWithTx(ctx context.Context, tx pgx.Tx, data TransactionInstance) (int32, error) {
	query := `INSERT INTO transactions (
		portfolio_id,
		portfolio_item_id,
		type,
		amount,
		price_usd,
		processed_at
	) VALUES ($1, $2, $3, $4, $5, COALESCE($6, NOW()))
	RETURNING id;`

	var id int32
	err := tx.QueryRow(ctx, query,
		data.PortfolioId,
		data.PortfolioItem,
		data.Type,
		data.Amount,
		data.PriceUSD,
		data.ProcessedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create transaction with tx: %w", err)
	}
	return id, nil
}

func (t *Transaction) Create(data TransactionInstance) (int32, error) {
	query := `INSERT INTO transactions (
		portfolio_id,
		portfolio_item_id,
		type,
		amount,
		price_usd,
		processed_at
	) VALUES ($1, $2, $3, $4, $5, COALESCE($6, NOW()))
	RETURNING id;`

	var id int32

	err := t.DBConn.QueryRow(context.Background(), query,
		data.PortfolioId,
		data.PortfolioItem,
		data.Type,
		data.Amount,
		data.PriceUSD,
		data.ProcessedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create transaction instance: %w", err)
	}

	return id, nil
}

func (t *Transaction) Read(id int32) (TransactionInstance, error) {
	query := `SELECT id, portfolio_id, portfolio_item_id, type, amount, price_usd, processed_at 
	          FROM transactions WHERE id = $1;`

	var data TransactionInstance

	err := t.DBConn.QueryRow(context.Background(), query, id).Scan(
		&data.Id,
		&data.PortfolioId,
		&data.PortfolioItem,
		&data.Type,
		&data.Amount,
		&data.PriceUSD,
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

func (t *Transaction) ReadAllByPortfolio(portfolioID int32) ([]TransactionInstance, error) {
	query := `SELECT id, portfolio_id, portfolio_item_id, type, amount, price_usd, processed_at 
	          FROM transactions 
	          WHERE portfolio_id = $1
	          ORDER BY processed_at DESC;`

	rows, err := t.DBConn.Query(context.Background(), query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to query portfolio transactions: %w", err)
	}
	defer rows.Close()

	var transactions []TransactionInstance

	for rows.Next() {
		var data TransactionInstance
		err := rows.Scan(
			&data.Id,
			&data.PortfolioId,
			&data.PortfolioItem,
			&data.Type,
			&data.Amount,
			&data.PriceUSD,
			&data.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}
		transactions = append(transactions, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

func (t *Transaction) ReadAllByItem(portfolioItemID int32) ([]TransactionInstance, error) {
	query := `SELECT id, portfolio_id, portfolio_item_id, type, amount, price_usd, processed_at 
	          FROM transactions 
	          WHERE portfolio_item_id = $1
	          ORDER BY processed_at ASC;`

	rows, err := t.DBConn.Query(context.Background(), query, portfolioItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions for item: %w", err)
	}
	defer rows.Close()

	var transactions []TransactionInstance

	for rows.Next() {
		var data TransactionInstance
		err := rows.Scan(
			&data.Id,
			&data.PortfolioId,
			&data.PortfolioItem,
			&data.Type,
			&data.Amount,
			&data.PriceUSD,
			&data.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}
		transactions = append(transactions, data)
	}

	return transactions, nil
}
