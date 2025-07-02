package main

import (
	"log"
	"os"

	"saan/payment/internal/application"
	"saan/payment/internal/infrastructure/database"
	"saan/payment/internal/infrastructure/redis"
	"saan/payment/internal/infrastructure/kafka"
	"saan/payment/internal/transport/http"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize infrastructure
	db, err := database.NewConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	redisClient := redis.NewClient()
	defer redisClient.Close()

	kafkaProducer := kafka.NewProducer()
	defer kafkaProducer.Close()

	// Initialize application services
	paymentService := application.NewPaymentService(db, redisClient, kafkaProducer)
	refundService := application.NewRefundService(db, redisClient, kafkaProducer)

	// Initialize HTTP server
	router := http.NewRouter(paymentService, refundService)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8087"
	}

	log.Printf("💳 Payment Service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
