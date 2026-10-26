package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	dbu "portfolio-tracker/internal/adapter/db/utils"
	"portfolio-tracker/internal/service/handlers"

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
	r.Post("auth/register", handlers.RegistrationHandler)
	r.Post("auth/login", handlers.LoginHandler)
	r.Post("auth/request-password-reset", handlers.RequestPasswordResetHandler)
	r.Post("auth/verify-otc", handlers.VerifyOneTimeCodeHandler)
	r.Post("auth/reset-password", handlers.ResetPasswordHandler)

	port := os.Getenv("APP_PORT")

	fmt.Printf("--- starting server on :%s ---", port)
	http.ListenAndServe(":8080", r)

}
