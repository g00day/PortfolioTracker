package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type PortfolioInstance struct {
	Id             *int32
	CreatedAt      *time.Time
	UserId         int32
	Name           string
	RiskFreeRateId int32
}

type Portfolio struct {
	DBConn *pgx.Conn
}

func (p *Portfolio) CreateTable() error {
	query := `CREATE TABLE IF NOT EXISTS portfolios (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		risk_free_rate_id INT NOT NULL REFERENCES risk_free_rate(id),
		name VARCHAR(100) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);`

	_, err := p.DBConn.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create portfolios table: %w", err)
	}

	return nil
}

func (p *Portfolio) Create(data PortfolioInstance) (int32, error) {
	query := `INSERT INTO portfolios (
		user_id,
		risk_free_rate_id,
		name
	) VALUES ($1, $2, $3)
	RETURNING id;`

	var id int32

	err := p.DBConn.QueryRow(context.Background(), query,
		data.UserId,
		data.RiskFreeRateId,
		data.Name,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create portfolio instance: %w", err)
	}

	return id, nil
}

func (p *Portfolio) Read(id int32) (PortfolioInstance, error) {
	query := `SELECT id, user_id, risk_free_rate_id, name, created_at FROM portfolios WHERE id = $1;`

	var data PortfolioInstance

	err := p.DBConn.QueryRow(context.Background(), query, id).Scan(
		&data.Id,
		&data.UserId,
		&data.RiskFreeRateId,
		&data.Name,
		&data.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return PortfolioInstance{}, fmt.Errorf("portfolio not found: %w", err)
		}
		return PortfolioInstance{}, fmt.Errorf("failed to read portfolio: %w", err)
	}

	return data, nil
}

func (p *Portfolio) ReadAllByUserID(userID int32) ([]PortfolioInstance, error) {
	query := `SELECT id, user_id, risk_free_rate_id, name, created_at 
	          FROM portfolios 
	          WHERE user_id = $1;`

	rows, err := p.DBConn.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user portfolios: %w", err)
	}
	defer rows.Close()

	var portfolios []PortfolioInstance

	for rows.Next() {
		var data PortfolioInstance
		err := rows.Scan(
			&data.Id,
			&data.UserId,
			&data.RiskFreeRateId,
			&data.Name,
			&data.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan portfolio: %w", err)
		}
		portfolios = append(portfolios, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating portfolios: %w", err)
	}

	return portfolios, nil
}

func (p *Portfolio) Update(id int32, data PortfolioInstance) error {
	query := `UPDATE portfolios 
	          SET name = $1, 
	              risk_free_rate_id = $2 
	          WHERE id = $3;`

	res, err := p.DBConn.Exec(context.Background(), query,
		data.Name,
		data.RiskFreeRateId,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update portfolio: %w", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("portfolio not found")
	}

	return nil
}

func (p *Portfolio) Delete(id int32) error {
	query := `DELETE FROM portfolios WHERE id = $1;`

	res, err := p.DBConn.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("failed to delete portfolio: %w", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("portfolio not found")
	}

	return nil
}
