package utils

import (
	db "portfolio-tracker/internal/adapter/db"
	r "portfolio-tracker/internal/adapter/db/repos"
)

func SetUpTables() ([]error, error) {

	dbConn, err := db.GetDBConnection()

	if err != nil {
		return nil, err
	}

	mms := &r.MacroMarketSnapshot{DBConn: dbConn}

	var errSlice []error

	errSlice = append(errSlice, mms.CreateTable())

	var finalErr error
	for _, err := range errSlice {
		if err != nil {
			finalErr = err
		}
	}

	return errSlice, finalErr

}
