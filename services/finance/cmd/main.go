package main

import (
	"log"
	"os"

	"finance/internal/application"
	"finance/internal/infrastructure/database"
	"finance/internal/infrastructure/database/repositories"
	"finance/internal/infrastructure/cache"
	"finance/internal/transport/http"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize infrastructure
	db, err := database.New()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	redisClient, err := cache.New()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	// Initialize repositories
	repos := repositories.NewRepositories(db)

	// Initialize application services
	financeService := application.NewFinanceService(repos, redisClient)
	allocationService := application.NewAllocationService(repos, redisClient)
	cashFlowService := application.NewCashFlowService(repos, redisClient)

	// Initialize HTTP server
	router := http.NewRouter(financeService, allocationService, cashFlowService)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8085"
	}

	log.Printf("💰 Finance Service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
