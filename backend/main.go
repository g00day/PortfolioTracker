package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	dbu "portfolio-tracker/internal/adapter/db/utils"

	handlers "portfolio-tracker/internal/service/handlers"

	auth_handlers "portfolio-tracker/internal/service/handlers/auth_handlers"
	p_handlers "portfolio-tracker/internal/service/handlers/portfolio_handlers"
	pi_handlers "portfolio-tracker/internal/service/handlers/portfolio_items_handlers"
	rfr_handlers "portfolio-tracker/internal/service/handlers/risk_free_rates_handlers"

	"portfolio-tracker/internal/service/middlewares"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("failed to get .env file: %w", err)
	}

	fmt.Println("--- Trying set up DB tables ---")

	_, err = dbu.SetUpTables()
	if err != nil {
		log.Fatalf("Failed to connect DB: %v", err)
	}

	fmt.Println("--- All DB tables set up successfully ---")

	// determine a chi router - app core
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// endpoints list
	r.Get("/health", handlers.HealthCheckHandler)

	// Auth endpoints
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", auth_handlers.RegistrationHandler)
		r.Post("/login", auth_handlers.LoginHandler)
		r.Post("/request-password-reset", auth_handlers.RequestPasswordResetHandler)
		r.Post("/verify-otc", auth_handlers.VerifyOneTimeCodeHandler)
		r.Post("/reset-password", auth_handlers.ResetPasswordHandler)
	})

	// risk free rate endpoints
	r.Route("/risk-free-rates", func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware)

		r.Post("/", rfr_handlers.CreateRiskFreeRateHandler)
		r.Get("/", rfr_handlers.ReadRiskFreeRates)
		r.Get("/{id}", rfr_handlers.ReadRiskFreeRate)
	})

	// portfolio endpoints
	r.Route("/portfolios", func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware)

		r.Post("/", p_handlers.CreatePortfolio)
		r.Get("/", p_handlers.ReadPortfolios)
		r.Put("/{id}", p_handlers.UpdatePortfolio)
		r.Delete("/{id}", p_handlers.DeletePortfolio)

		r.Post("/{id}/add-item", p_handlers.AddPortfolioItem)

		r.Get("/{id}/transactions", p_handlers.ReadPortfolioTransactions)
	})

	// portfolio items endpoints
	r.Route("/portfolio-items", func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware)
		r.Delete("/{id}", pi_handlers.DeletePortfolioItem)
	})

	port := os.Getenv("APP_PORT")

	fmt.Printf("--- starting server on :%s ---", port)
	http.ListenAndServe(":8080", r)

}
