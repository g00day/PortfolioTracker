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
	query := `CREATE TABLE IF NOT EXISTS users (
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
	query := `INSERT INTO users (
		username,
		email,
		password_hash,
		created_at
	) VALUES ($1, $2, $3, $4)
	RETURNING id;`

	var id int32

	if data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}

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

func (u *User) Read(id int32) (UserInstance, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE id = $1`

	var user UserInstance

	err := u.DBConn.QueryRow(context.Background(), query, id).Scan(
		&user.Id,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return UserInstance{}, fmt.Errorf("user not found: %w", err)
		}
		return UserInstance{}, fmt.Errorf("failed to read user: %w", err)
	}

	return user, nil
}

func (u *User) GetByUsername(username string) (UserInstance, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE username = $1`

	var user UserInstance

	err := u.DBConn.QueryRow(context.Background(), query, username).Scan(
		&user.Id,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return UserInstance{}, fmt.Errorf("user with username '%s' not found: %w", username, err)
		}
		return UserInstance{}, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

func (u *User) Update(id int32, data UserInstance) error {
	query := `UPDATE users 
	SET username = $1, email = $2, password_hash = $3 
	WHERE id = $4`

	result, err := u.DBConn.Exec(context.Background(), query,
		data.Username,
		data.Email,
		data.PasswordHash,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// check if the user exists
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user with id %d not found for update", id)
	}

	return nil
}

func (u *User) Delete(id int32) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := u.DBConn.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user with id %d not found for deletion", id)
	}

	return nil
}
