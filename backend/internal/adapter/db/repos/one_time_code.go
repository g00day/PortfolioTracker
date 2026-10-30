package repos

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type OneTimeCodeInstance struct {
	Id        int32
	Email     string
	CodeHash  string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type OneTimeCode struct {
	DBConn DBConnection
}

// code hashing helper
func hashCode(code string) string {
	hash := sha256.Sum256([]byte(code))
	return hex.EncodeToString(hash[:])
}

func (otc *OneTimeCode) CreateTable() error {
	query := `CREATE TABLE IF NOT EXISTS one_time_codes (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) NOT NULL,
		code_hash VARCHAR(64) NOT NULL,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_otc_email ON one_time_codes(email);`

	_, err := otc.DBConn.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create one_time_codes table: %w", err)
	}

	return nil
}

func (otc *OneTimeCode) Create(email string, code string) (int32, error) {
	query := `INSERT INTO one_time_codes (
		email,
		code_hash,
		expires_at
	) VALUES ($1, $2, $3)
	RETURNING id;`

	var id int32
	expiresAt := time.Now().Add(3 * time.Minute)
	codeHash := hashCode(code)

	err := otc.DBConn.QueryRow(context.Background(), query,
		email,
		codeHash,
		expiresAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create one time code instance: %w", err)
	}

	return id, nil
}

func (otc *OneTimeCode) VerifyAndUse(email string, code string) (bool, error) {
	query := `DELETE FROM one_time_codes
	          WHERE id = (
	              SELECT id FROM one_time_codes
	              WHERE email = $1 
	                AND code_hash = $2 
	                AND expires_at > $3
	              ORDER BY created_at DESC
	              LIMIT 1
	          )
	          RETURNING id;`

	codeHash := hashCode(code)
	now := time.Now()
	var id int32

	err := otc.DBConn.QueryRow(context.Background(), query, email, codeHash, now).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to verify and delete one time code: %w", err)
	}

	return true, nil
}

func (otc *OneTimeCode) InvalidateOldCodes(email string) error {
	query := `DELETE FROM one_time_codes WHERE email = $1;`

	_, err := otc.DBConn.Exec(context.Background(), query, email)
	if err != nil {
		return fmt.Errorf("failed to invalidate (delete) old codes: %w", err)
	}

	return nil
}

func (otc *OneTimeCode) DeleteExpired() (int64, error) {
	query := `DELETE FROM one_time_codes WHERE expires_at <= $1;`

	res, err := otc.DBConn.Exec(context.Background(), query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired codes: %w", err)
	}

	return res.RowsAffected(), nil
}
