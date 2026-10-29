package utils

import (
	r "portfolio-tracker/internal/adapter/db/repos"

	"github.com/jackc/pgx/v5"
)

func SetUpTables(dbConn *pgx.Conn) ([]error, error) {

	otc := &r.OneTimeCode{DBConn: dbConn}
	mms := &r.MacroMarketSnapshot{DBConn: dbConn}
	u := &r.User{DBConn: dbConn}
	rfr := &r.RiskFreeRate{DBConn: dbConn}
	p := &r.Portfolio{DBConn: dbConn}
	pi := &r.PortfolioItem{DBConn: dbConn}
	pis := &r.PortfolioItemSnapshot{DBConn: dbConn}
	t := &r.Transaction{DBConn: dbConn}

	var errSlice []error

	// layer - 1. independent tables
	errSlice = append(errSlice, u.CreateTable())
	errSlice = append(errSlice, rfr.CreateTable())
	errSlice = append(errSlice, mms.CreateTable())

	// layer-2. rely on layer-1
	errSlice = append(errSlice, otc.CreateTable())
	errSlice = append(errSlice, p.CreateTable())

	// layer-3. rely on layer-2
	errSlice = append(errSlice, pi.CreateTable())

	// layer-4. rely on layer-3
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
