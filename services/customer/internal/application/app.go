package application

import (
	"go.uber.org/zap"

	"github.com/saan-system/services/customer/internal/domain/repository"
)

// Application holds all application usecases as per Clean Architecture
type Application struct {
	CustomerUsecase *CustomerUsecase
	AddressUsecase  *AddressUsecase
	PointsUsecase   *PointsUsecase
}

// Dependencies represents external dependencies for the application
type Dependencies struct {
	// Repository implementations (will be created in main.go)
	CustomerRepo       repository.CustomerRepository
	AddressRepo        repository.CustomerAddressRepository
	VIPBenefitsRepo    repository.VIPTierBenefitsRepository
	PointsRepo         repository.CustomerPointsRepository
	AnalyticsRepo      repository.CustomerAnalyticsRepository
	ThaiAddressRepo    repository.ThaiAddressRepository
	DeliveryRouteRepo  repository.DeliveryRouteRepository
	CacheRepo          repository.CacheRepository
	EventPublisher     repository.EventPublisher
	LoyverseClient     repository.LoyverseClient
	Logger             *zap.Logger
}

// New creates a new application instance with all usecases
func New(deps Dependencies) *Application {
	// Create usecases with dependency injection (orchestrating domain logic)
	customerUsecase := NewCustomerUsecase(
		deps.CustomerRepo,
		deps.AddressRepo,
		deps.VIPBenefitsRepo,
		deps.PointsRepo,
		deps.AnalyticsRepo,
		deps.ThaiAddressRepo,
		deps.DeliveryRouteRepo,
		deps.EventPublisher,
		deps.CacheRepo,
		deps.LoyverseClient,
	)

	addressUsecase := NewAddressUsecase(
		deps.AddressRepo,
		deps.CustomerRepo,
		deps.ThaiAddressRepo,
		deps.EventPublisher,
	)

	pointsUsecase := NewPointsUsecase(
		deps.PointsRepo,
		deps.CustomerRepo,
		deps.VIPBenefitsRepo,
		deps.EventPublisher,
	)

	return &Application{
		CustomerUsecase: customerUsecase,
		AddressUsecase:  addressUsecase,
		PointsUsecase:   pointsUsecase,
	}
}
