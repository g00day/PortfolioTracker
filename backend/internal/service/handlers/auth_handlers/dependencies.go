package authhandlers

import (
	database "portfolio-tracker/internal/adapter/db"
	"portfolio-tracker/internal/adapter/db/repos"
	"portfolio-tracker/internal/service/smtp"
)

var getDBConnection = func() (repos.DBConnection, error) {
	return database.GetDBConnection()
}

var sendPasswordResetEmail = smtp.SendPasswordResetEmail
