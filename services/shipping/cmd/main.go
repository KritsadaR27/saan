package main

import (
	"log"
	"os"

	"saan/shipping/internal/application"
	"saan/shipping/internal/infrastructure/database"
	"saan/shipping/internal/infrastructure/redis"
	"saan/shipping/internal/infrastructure/kafka"
	"saan/shipping/internal/transport/http"

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
	shippingService := application.NewShippingService(db, redisClient, kafkaProducer)
	routeService := application.NewRouteService(db, redisClient)
	carrierService := application.NewCarrierService(db, redisClient)

	// Initialize HTTP server
	router := http.NewRouter(shippingService, routeService, carrierService)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8086"
	}

	log.Printf("🚚 Shipping Service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
