package application

import (
	"database/sql"

	"go.uber.org/zap"

	"github.com/saan-system/services/customer/internal/domain"
	"github.com/saan-system/services/customer/internal/infrastructure/repository"
	"github.com/saan-system/services/customer/internal/infrastructure/cache"
	"github.com/saan-system/services/customer/internal/infrastructure/messaging"
)

// Application holds all application services
type Application struct {
	CustomerService    domain.CustomerService
	ThaiAddressService domain.ThaiAddressService
}

// New creates a new application instance
func New(
	db *sql.DB,
	redisClient cache.RedisClient,
	kafkaProducer messaging.KafkaProducer,
	loyverseClient domain.LoyverseClient,
	logger *zap.Logger,
) *Application {
	// Initialize repositories
	customerRepo := repository.NewCustomerRepository(db)
	addressRepo := repository.NewCustomerAddressRepository(db)
	thaiAddressRepo := repository.NewThaiAddressRepository(db)
	deliveryRouteRepo := repository.NewDeliveryRouteRepository(db)
	
	// Initialize cache repository
	cacheRepo := cache.NewCacheRepository(redisClient)
	
	// Initialize event publisher
	eventPublisher := messaging.NewEventPublisher(kafkaProducer)
	
	// Initialize services
	customerService := NewCustomerService(
		customerRepo,
		addressRepo,
		thaiAddressRepo,
		deliveryRouteRepo,
		cacheRepo,
		eventPublisher,
		loyverseClient,
		logger,
	)

	thaiAddressService := NewThaiAddressService(thaiAddressRepo, logger)

	return &Application{
		CustomerService:    customerService,
		ThaiAddressService: thaiAddressService,
	}
}
