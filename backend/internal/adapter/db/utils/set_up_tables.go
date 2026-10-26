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

	otc := &r.OneTimeCode{DBConn: dbConn}
	mms := &r.MacroMarketSnapshot{DBConn: dbConn}
	u := &r.User{DBConn: dbConn}
	a := &r.Asset{DBConn: dbConn}
	rfr := &r.RiskFreeRate{DBConn: dbConn}
	p := &r.Portfolio{DBConn: dbConn}
	pi := &r.PortfolioItem{DBConn: dbConn}
	pis := &r.PortfolioItemSnapshot{DBConn: dbConn}
	t := &r.Transaction{DBConn: dbConn}

	var errSlice []error

	errSlice = append(errSlice, otc.CreateTable())
	errSlice = append(errSlice, mms.CreateTable())
	errSlice = append(errSlice, u.CreateTable())
	errSlice = append(errSlice, a.CreateTable())
	errSlice = append(errSlice, rfr.CreateTable())
	errSlice = append(errSlice, p.CreateTable())
	errSlice = append(errSlice, pi.CreateTable())
	errSlice = append(errSlice, pis.CreateTable())
	errSlice = append(errSlice, t.CreateTable())

	var finalErr error
	for _, err := range errSlice {
		if err != nil {
			finalErr = err
		}
	}

	return errSlice, finalErr

}
