package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func GetDBConnection() (*pgx.Conn, error) {
	/* Returns the actual connection to the main Database
	-----------------------------------------------------
	Mostly important that the connection must be closed
	after using it
	*/
	// .env удобен при локальном запуске, но в Docker значения передаются
	// непосредственно через environment.
	if err := godotenv.Load(); err != nil {
		log.Print(".env file not found; using environment variables")
	}

	// Connecting to postgres Database
	// Unpacking all env vars that rely to db
	dbUsername := os.Getenv("POSTGRES_USER_NAME")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB_NAME")
	dbHost := os.Getenv("POSTGRES_HOST")

	connectionCreds := fmt.Sprintf("postgres://%s:%s@%s:5432/%s", dbUsername, dbPassword, dbHost, dbName)
	dbconn, err := pgx.Connect(context.Background(), connectionCreds)

	return dbconn, err
}
