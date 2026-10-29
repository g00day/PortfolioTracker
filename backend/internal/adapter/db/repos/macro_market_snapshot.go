package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type MacroMarketSnapshotInstance struct {
	Id             *int32
	Date           *time.Time
	OpenInterest   float64
	CapitalFlowUSD float64
	BTCHashrate    float64
	ETHHashrate    float64
}

type MacroMarketSnapshot struct {
	DBConn *pgx.Conn
}

func (mms *MacroMarketSnapshot) CreateTable() error {
	query := `CREATE TABLE IF NOT EXISTS macro_market_snapshot (
		id SERIAL PRIMARY KEY,
		date TIMESTAMP NOT NULL DEFAULT NOW(),
		open_interest DECIMAL NOT NULL,
		capital_flow_usd DECIMAL NOT NULL,
		btc_hashrate DECIMAL NOT NULL,
		eth_hashrate DECIMAL NOT NULL
	);`

	_, err := mms.DBConn.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create macro_market_snapshot table: %w", err)
	}

	return nil
}

func (mms *MacroMarketSnapshot) Create(data MacroMarketSnapshotInstance) (int32, error) {
	query := `INSERT INTO macro_market_snapshot (
		open_interest,
		capital_flow_usd,
		btc_hashrate,
		eth_hashrate
	) VALUES ($1, $2, $3, $4) 
	RETURNING id;`

	var id int32

	err := mms.DBConn.QueryRow(context.Background(), query,
		data.OpenInterest,
		data.CapitalFlowUSD,
		data.BTCHashrate,
		data.ETHHashrate,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create macro market snapshot instance: %w", err)
	}

	return id, nil
}
