package main

import (
	"log"

	"shipping/internal/infrastructure/config"
	"shipping/internal/infrastructure/database"
	"shipping/internal/infrastructure/cache"
	"shipping/internal/infrastructure/events"
	"shipping/internal/application"
	"shipping/internal/transport/http"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize infrastructure
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	cacheClient, err := cache.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	eventPublisher, err := events.NewKafkaProducer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.ServiceName)
	if err != nil {
		log.Fatal("Failed to connect to Kafka:", err)
	}
	defer eventPublisher.Close()

	// Initialize repositories
	deliveryRepo := database.NewDeliveryRepository(db)
	vehicleRepo := database.NewVehicleRepository(db)
	routeRepo := database.NewRouteRepository(db)
	providerRepo := database.NewProviderRepository(db)
	snapshotRepo := database.NewSnapshotRepository(db)
	coverageRepo := database.NewCoverageAreaRepository(db)

	// Initialize use cases (using correct constructor names)
	deliveryUseCase := application.NewDeliveryUsecase(
		deliveryRepo, vehicleRepo, routeRepo, providerRepo, 
		snapshotRepo, coverageRepo, eventPublisher, cacheClient)
	vehicleUseCase := application.NewVehicleUseCase(vehicleRepo, eventPublisher)

	// Create placeholder use cases for compilation
	providerUseCase := &application.ProviderUseCase{}
	routingUseCase := &application.RoutingUseCase{}
	trackingUseCase := &application.TrackingUseCase{}
	coverageUseCase := &application.CoverageUseCase{}

	// Initialize HTTP server
	server := http.NewServer(
		cfg.ServerPort,
		deliveryUseCase,
		vehicleUseCase,
		providerUseCase,
		routingUseCase,
		trackingUseCase,
		coverageUseCase,
	)

	log.Printf("🚚 Shipping Service starting on port %s", cfg.ServerPort)
	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
