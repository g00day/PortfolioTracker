package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type UserInstance struct {
	Id           *int32
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type User struct {
	DBConn *pgx.Conn
}

func (u *User) CreateTable() error {
	query := `CREATE TABLE IF NOT EXISTS user (
		id SERIAL PRIMARY KEY,
		username VARCHAR(40),
		email VARCHAR(256),
		password_hash VARCHAR(72) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	)`

	_, err := u.DBConn.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create user table: %w", err)
	}

	return nil
}

func (u *User) Create(data UserInstance) (int32, error) {
	query := `INSERT INTO user (
		username,
		email,
		password_hash,
		created_at)
	VALUES (
		$1,
		$2,
		$3,
		$4
	)
		RETURNING id;
	)`

	var id int32

	err := u.DBConn.QueryRow(context.Background(), query,
		data.Username,
		data.Email,
		data.PasswordHash,
		data.CreatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create user instance: %w", err)
	}

	return id, nil
}
