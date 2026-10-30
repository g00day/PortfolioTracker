package portfoliohandlers

import (
	database "portfolio-tracker/internal/adapter/db"
	"portfolio-tracker/internal/adapter/db/repos"
)

var getDBConnection = func() (repos.DBConnection, error) {
	return database.GetDBConnection()
}
