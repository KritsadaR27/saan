package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/saan-system/services/customer/internal/domain/entity"
	"github.com/saan-system/services/customer/internal/domain/repository"
)

// CustomerUsecase handles customer business logic
type CustomerUsecase struct {
	customerRepo       repository.CustomerRepository
	addressRepo        repository.CustomerAddressRepository
	vipBenefitsRepo    repository.VIPTierBenefitsRepository
	pointsRepo         repository.CustomerPointsRepository
	analyticsRepo      repository.CustomerAnalyticsRepository
	thaiAddressRepo    repository.ThaiAddressRepository
	deliveryRouteRepo  repository.DeliveryRouteRepository
	eventPublisher     repository.EventPublisher
	cache              repository.CacheRepository
	loyverseClient     repository.LoyverseClient
}

// NewCustomerUsecase creates a new customer usecase
func NewCustomerUsecase(
	customerRepo repository.CustomerRepository,
	addressRepo repository.CustomerAddressRepository,
	vipBenefitsRepo repository.VIPTierBenefitsRepository,
	pointsRepo repository.CustomerPointsRepository,
	analyticsRepo repository.CustomerAnalyticsRepository,
	thaiAddressRepo repository.ThaiAddressRepository,
	deliveryRouteRepo repository.DeliveryRouteRepository,
	eventPublisher repository.EventPublisher,
	cache repository.CacheRepository,
	loyverseClient repository.LoyverseClient,
) *CustomerUsecase {
	return &CustomerUsecase{
		customerRepo:       customerRepo,
		addressRepo:        addressRepo,
		vipBenefitsRepo:    vipBenefitsRepo,
		pointsRepo:         pointsRepo,
		analyticsRepo:      analyticsRepo,
		thaiAddressRepo:    thaiAddressRepo,
		deliveryRouteRepo:  deliveryRouteRepo,
		eventPublisher:     eventPublisher,
		cache:              cache,
		loyverseClient:     loyverseClient,
	}
}

// CreateCustomerRequest represents a request to create a customer
type CreateCustomerRequest struct {
	FirstName    string     `json:"first_name" validate:"required,min=1,max=100"`
	LastName     string     `json:"last_name" validate:"required,min=1,max=100"`
	Email        string     `json:"email" validate:"required,email"`
	Phone        string     `json:"phone" validate:"required,min=10,max=20"`
	DateOfBirth  *time.Time `json:"date_of_birth"`
	Gender       *string    `json:"gender" validate:"omitempty,oneof=male female other"`
	LoyverseID   *string    `json:"loyverse_id"`
	LineUserID   *string    `json:"line_user_id"`
	LineDisplayName *string `json:"line_display_name"`
}

// UpdateCustomerRequest represents a request to update a customer
type UpdateCustomerRequest struct {
	FirstName    *string    `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName     *string    `json:"last_name" validate:"omitempty,min=1,max=100"`
	Email        *string    `json:"email" validate:"omitempty,email"`
	Phone        *string    `json:"phone" validate:"omitempty,min=10,max=20"`
	DateOfBirth  *time.Time `json:"date_of_birth"`
	Gender       *string    `json:"gender" validate:"omitempty,oneof=male female other"`
	LineDisplayName *string `json:"line_display_name"`
}

// CreateCustomer creates a new customer
func (uc *CustomerUsecase) CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*entity.Customer, error) {
	// Generate customer code
	code := entity.GenerateCustomerCode()
	
	// Create customer entity
	customer := &entity.Customer{
		ID:              uuid.New(),
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Email:           req.Email,
		Phone:           req.Phone,
		DateOfBirth:     req.DateOfBirth,
		Gender:          req.Gender,
		CustomerCode:    code,
		Tier:            entity.CustomerTier(1), // Bronze
		PointsBalance:   0,
		TotalSpent:      0,
		LoyverseID:      req.LoyverseID,
		LineUserID:      req.LineUserID,
		LineDisplayName: req.LineDisplayName,
		IsActive:        true,
		OrderCount:      0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Validate customer
	if err := customer.ValidateCustomer(); err != nil {
		return nil, fmt.Errorf("invalid customer data: %w", err)
	}

	// Create in database
	if err := uc.customerRepo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// Publish event
	if err := uc.eventPublisher.PublishCustomerCreated(ctx, customer); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}

	return customer, nil
}

// GetCustomerByID retrieves a customer by ID
func (uc *CustomerUsecase) GetCustomerByID(ctx context.Context, id uuid.UUID) (*entity.Customer, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("customer:%s", id.String())
	if customer, err := uc.cache.GetCustomer(ctx, cacheKey); err == nil && customer != nil {
		return customer, nil
	}

	// Get from database
	customer, err := uc.customerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	// Cache result
	if err := uc.cache.SetCustomer(ctx, cacheKey, customer, 3600); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}

	return customer, nil
}

// GetCustomerByPhone retrieves a customer by phone
func (uc *CustomerUsecase) GetCustomerByPhone(ctx context.Context, phone string) (*entity.Customer, error) {
	return uc.customerRepo.GetByPhone(ctx, phone)
}

// GetCustomerByEmail retrieves a customer by email
func (uc *CustomerUsecase) GetCustomerByEmail(ctx context.Context, email string) (*entity.Customer, error) {
	return uc.customerRepo.GetByEmail(ctx, email)
}

// GetCustomerByLoyverseID retrieves a customer by Loyverse ID
func (uc *CustomerUsecase) GetCustomerByLoyverseID(ctx context.Context, loyverseID string) (*entity.Customer, error) {
	return uc.customerRepo.GetByLoyverseID(ctx, loyverseID)
}

// GetCustomerByLineUserID retrieves a customer by LINE user ID
func (uc *CustomerUsecase) GetCustomerByLineUserID(ctx context.Context, lineUserID string) (*entity.Customer, error) {
	return uc.customerRepo.GetByLineUserID(ctx, lineUserID)
}

// UpdateCustomer updates a customer
func (uc *CustomerUsecase) UpdateCustomer(ctx context.Context, id uuid.UUID, req *UpdateCustomerRequest) (*entity.Customer, error) {
	// Get existing customer
	customer, err := uc.customerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	// Update fields
	if req.FirstName != nil {
		customer.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		customer.LastName = *req.LastName
	}
	if req.Email != nil {
		customer.Email = *req.Email
	}
	if req.Phone != nil {
		customer.Phone = *req.Phone
	}
	if req.DateOfBirth != nil {
		customer.DateOfBirth = req.DateOfBirth
	}
	if req.Gender != nil {
		customer.Gender = req.Gender
	}
	if req.LineDisplayName != nil {
		customer.LineDisplayName = req.LineDisplayName
	}

	customer.UpdatedAt = time.Now()

	// Validate updated customer
	if err := customer.ValidateCustomer(); err != nil {
		return nil, fmt.Errorf("invalid customer data: %w", err)
	}

	// Update in database
	if err := uc.customerRepo.Update(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("customer:%s", id.String())
	if err := uc.cache.DeleteCustomer(ctx, cacheKey); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}

	// Publish event
	if err := uc.eventPublisher.PublishCustomerUpdated(ctx, customer); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}

	return customer, nil
}

// DeleteCustomer deletes a customer (soft delete)
func (uc *CustomerUsecase) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	// Get customer first to verify it exists
	customer, err := uc.customerRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	// Soft delete
	customer.IsActive = false
	customer.UpdatedAt = time.Now()

	if err := uc.customerRepo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("customer:%s", id.String())
	if err := uc.cache.DeleteCustomer(ctx, cacheKey); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}

	// Publish event
	if err := uc.eventPublisher.PublishCustomerDeleted(ctx, id); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}

	return nil
}

// ListCustomers retrieves customers with pagination
func (uc *CustomerUsecase) ListCustomers(ctx context.Context, limit, offset int) ([]*entity.Customer, int, error) {
	filter := repository.CustomerFilter{
		Limit:  limit,
		Offset: offset,
	}
	customers, total, err := uc.customerRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	
	// Convert []entity.Customer to []*entity.Customer
	result := make([]*entity.Customer, len(customers))
	for i := range customers {
		result[i] = &customers[i]
	}
	
	return result, total, nil
}

// SearchCustomers searches customers by query - implementing with List and basic filtering
func (uc *CustomerUsecase) SearchCustomers(ctx context.Context, query string, limit, offset int) ([]*entity.Customer, int, error) {
	// For now, just return all customers - can be improved with proper search
	filter := repository.CustomerFilter{
		Limit:  limit,
		Offset: offset,
	}
	customers, total, err := uc.customerRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	
	// Convert []entity.Customer to []*entity.Customer  
	result := make([]*entity.Customer, len(customers))
	for i := range customers {
		result[i] = &customers[i]
	}
	
	return result, total, nil
}

// GetCustomersByTier retrieves customers by tier - implementing with List and filtering
func (uc *CustomerUsecase) GetCustomersByTier(ctx context.Context, tier int, limit, offset int) ([]*entity.Customer, int, error) {
	filter := repository.CustomerFilter{
		Tier:   func() *entity.CustomerTier { t := entity.CustomerTier(tier); return &t }(),
		Limit:  limit,
		Offset: offset,
	}
	customers, total, err := uc.customerRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	
	// Convert []entity.Customer to []*entity.Customer
	result := make([]*entity.Customer, len(customers))
	for i := range customers {
		result[i] = &customers[i]
	}
	
	return result, total, nil
}

// GetCustomerWithAddresses retrieves a customer with their addresses
func (uc *CustomerUsecase) GetCustomerWithAddresses(ctx context.Context, id uuid.UUID) (*entity.CustomerWithAddresses, error) {
	// Get customer
	customer, err := uc.customerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	// Get addresses
	addresses, err := uc.addressRepo.GetByCustomerID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer addresses: %w", err)
	}

	return &entity.CustomerWithAddresses{
		Customer:  customer,
		Addresses: addresses,
	}, nil
}

// UpdateCustomerSpending updates customer's total spent and recalculates tier
func (uc *CustomerUsecase) UpdateCustomerSpending(ctx context.Context, customerID uuid.UUID, amount float64) error {
	// Get customer
	customer, err := uc.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	// Update spending
	customer.TotalSpent += amount
	customer.OrderCount++
	customer.LastOrderDate = &time.Time{}
	*customer.LastOrderDate = time.Now()
	customer.UpdatedAt = time.Now()

	// Calculate new tier
	newTierInt := entity.CalculateTierFromSpent(customer.TotalSpent)
	newTier := entity.CustomerTier(newTierInt)
	oldTier := customer.Tier
	
	if newTier > customer.Tier {
		customer.Tier = newTier
		customer.TierAchievedDate = &time.Time{}
		*customer.TierAchievedDate = time.Now()
	}

	// Update in database
	if err := uc.customerRepo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("customer:%s", customerID.String())
	if err := uc.cache.DeleteCustomer(ctx, cacheKey); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}

	// Publish tier change event if tier changed
	if newTier > oldTier {
		if err := uc.eventPublisher.PublishCustomerTierUpdated(ctx, customerID, oldTier, newTier); err != nil {
			// Log error but don't fail the operation
			// TODO: Add proper logging
		}
	}

	return nil
}

// SyncWithLoyverse syncs customer data with Loyverse
func (uc *CustomerUsecase) SyncWithLoyverse(ctx context.Context, customerID uuid.UUID) error {
	// Get customer
	customer, err := uc.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	if customer.LoyverseID == nil {
		return fmt.Errorf("customer has no Loyverse ID")
	}

	// Get data from Loyverse
	loyverseData, err := uc.loyverseClient.GetCustomer(ctx, *customer.LoyverseID)
	if err != nil {
		return fmt.Errorf("failed to get Loyverse data: %w", err)
	}

	// Update customer with Loyverse data
	customer.LoyverseTotalVisits = loyverseData.LoyverseTotalVisits
	customer.LoyverseTotalSpent = loyverseData.LoyverseTotalSpent
	customer.LoyversePoints = loyverseData.LoyversePoints
	customer.FirstVisit = loyverseData.FirstVisit
	customer.LastVisit = loyverseData.LastVisit
	customer.LastSyncAt = &time.Time{}
	*customer.LastSyncAt = time.Now()
	customer.UpdatedAt = time.Now()

	// Update in database
	if err := uc.customerRepo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("customer:%s", customerID.String())
	if err := uc.cache.DeleteCustomer(ctx, cacheKey); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}

	return nil
}

// GetVIPBenefits retrieves VIP benefits for a customer's tier
func (uc *CustomerUsecase) GetVIPBenefits(ctx context.Context, customerID uuid.UUID) (*entity.VIPTierBenefits, error) {
	// Get customer
	customer, err := uc.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	// Get VIP benefits
	return uc.vipBenefitsRepo.GetByTier(ctx, customer.Tier)
}

// GetCustomerAnalytics retrieves analytics for a customer
func (uc *CustomerUsecase) GetCustomerAnalytics(ctx context.Context, customerID uuid.UUID) (*entity.CustomerAnalytics, error) {
	return uc.analyticsRepo.GetCustomerInsights(ctx, customerID)
}
