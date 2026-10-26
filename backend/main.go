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

	_, err = dbu.SetUpTables()
	if err != nil {
		log.Fatalf("Failed to connect DB: %v", err)
	}

	// determine a chi router - app core
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", handlers.HealthCheckHandler)
	r.Post("/register", handlers.RegistrationHandler)

	port := os.Getenv("APP_PORT")

	fmt.Printf("starting server on :%s", port)
	http.ListenAndServe(":8080", r)

}
