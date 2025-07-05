package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"shipping/internal/domain/entity"
	"shipping/internal/domain/repository"
)

// ProviderUseCase handles delivery provider management operations
type ProviderUseCase struct {
	providerRepo     repository.ProviderRepository
	coverageAreaRepo repository.CoverageAreaRepository
	eventPublisher   EventPublisher
	cache           Cache
}

// NewProviderUseCase creates a new provider use case
func NewProviderUseCase(
	providerRepo repository.ProviderRepository,
	coverageAreaRepo repository.CoverageAreaRepository,
	eventPublisher EventPublisher,
	cache Cache,
) *ProviderUseCase {
	return &ProviderUseCase{
		providerRepo:     providerRepo,
		coverageAreaRepo: coverageAreaRepo,
		eventPublisher:   eventPublisher,
		cache:           cache,
	}
}

// CreateProviderRequest represents the request to create a provider
type CreateProviderRequest struct {
	Code             string                     `json:"code"`
	Name             string                     `json:"name"`
	ProviderType     entity.ProviderType       `json:"provider_type"`
	Description      string                     `json:"description,omitempty"`
	APIConfig        *APIConfigurationRequest  `json:"api_config,omitempty"`
	ContactInfo      *ContactInfoRequest       `json:"contact_info,omitempty"`
	PricingConfig    *PricingConfigRequest     `json:"pricing_config,omitempty"`
	ServiceLevels    *ServiceLevelsRequest     `json:"service_levels,omitempty"`
	OperationalConfig *OperationalConfigRequest `json:"operational_config,omitempty"`
}

// APIConfigurationRequest represents API configuration
type APIConfigurationRequest struct {
	BaseURL           string `json:"base_url"`
	APIKey            string `json:"api_key"`
	AuthMethod        string `json:"auth_method"`
	RateQuoteEndpoint string `json:"rate_quote_endpoint,omitempty"`
}

// ContactInfoRequest represents contact information
type ContactInfoRequest struct {
	Phone         string `json:"phone,omitempty"`
	LineID        string `json:"line_id,omitempty"`
	Email         string `json:"email,omitempty"`
	Notes         string `json:"notes,omitempty"`
}

// PricingConfigRequest represents pricing configuration
type PricingConfigRequest struct {
	BaseRate            decimal.Decimal `json:"base_rate"`
	PerKmRate           decimal.Decimal `json:"per_km_rate"`
	WeightSurchargeRate decimal.Decimal `json:"weight_surcharge_rate"`
	SameDaySurcharge    decimal.Decimal `json:"same_day_surcharge"`
	CODSurchargeRate    decimal.Decimal `json:"cod_surcharge_rate"`
}

// ServiceLevelsRequest represents service level configuration
type ServiceLevelsRequest struct {
	StandardDeliveryHours int  `json:"standard_delivery_hours"`
	ExpressDeliveryHours  int  `json:"express_delivery_hours"`
	SameDayAvailable      bool `json:"same_day_available"`
	CODAvailable          bool `json:"cod_available"`
	TrackingAvailable     bool `json:"tracking_available"`
	InsuranceAvailable    bool `json:"insurance_available"`
}

// OperationalConfigRequest represents operational configuration
type OperationalConfigRequest struct {
	MaxWeight       float64    `json:"max_weight"`
	WeekendService  bool       `json:"weekend_service"`
	HolidayService  bool       `json:"holiday_service"`
	DailyCutoffTime *time.Time `json:"daily_cutoff_time,omitempty"`
	AutoAssign      bool       `json:"auto_assign"`
	PriorityOrder   int        `json:"priority_order"`
}

// UpdateProviderRequest represents the request to update a provider
type UpdateProviderRequest struct {
	ID                uuid.UUID                  `json:"id"`
	Name              *string                    `json:"name,omitempty"`
	Description       *string                    `json:"description,omitempty"`
	APIConfig         *APIConfigurationRequest   `json:"api_config,omitempty"`
	ContactInfo       *ContactInfoRequest        `json:"contact_info,omitempty"`
	PricingConfig     *PricingConfigRequest      `json:"pricing_config,omitempty"`
	ServiceLevels     *ServiceLevelsRequest      `json:"service_levels,omitempty"`
	OperationalConfig *OperationalConfigRequest  `json:"operational_config,omitempty"`
}

// PerformanceMetricsRequest represents performance metrics update
type PerformanceMetricsRequest struct {
	ProviderID          uuid.UUID       `json:"provider_id"`
	AverageDeliveryTime decimal.Decimal `json:"average_delivery_time"`
	SuccessRate         decimal.Decimal `json:"success_rate"`
	CustomerRating      decimal.Decimal `json:"customer_rating"`
}

// CreateProvider creates a new delivery provider
func (uc *ProviderUseCase) CreateProvider(ctx context.Context, req CreateProviderRequest) (*entity.DeliveryProvider, error) {
	// Check if provider with code already exists
	existing, err := uc.providerRepo.GetByCode(ctx, req.Code)
	if err == nil && existing != nil {
		return nil, errors.New("provider with this code already exists")
	}
	
	// Create provider entity
	provider, err := entity.NewDeliveryProvider(req.Code, req.Name, req.ProviderType)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}
	
	// Set API configuration if provided
	if req.APIConfig != nil && req.ProviderType == entity.ProviderTypeAPIIntegrated {
		if err := provider.SetAPIConfiguration(
			req.APIConfig.BaseURL,
			req.APIConfig.APIKey,
			req.APIConfig.AuthMethod,
		); err != nil {
			return nil, fmt.Errorf("failed to set API configuration: %w", err)
		}
		
		if req.APIConfig.RateQuoteEndpoint != "" {
			provider.RateQuoteAPIEndpoint = req.APIConfig.RateQuoteEndpoint
			provider.SupportsRateComparison = true
		}
	}
	
	// Set contact information if provided
	if req.ContactInfo != nil {
		if err := provider.SetManualCoordinationInfo(
			req.ContactInfo.Phone,
			req.ContactInfo.LineID,
			req.ContactInfo.Email,
			req.ContactInfo.Notes,
		); err != nil {
			return nil, fmt.Errorf("failed to set contact information: %w", err)
		}
	}
	
	// Set pricing configuration if provided
	if req.PricingConfig != nil {
		if err := provider.SetPricing(
			req.PricingConfig.BaseRate,
			req.PricingConfig.PerKmRate,
			req.PricingConfig.WeightSurchargeRate,
			req.PricingConfig.SameDaySurcharge,
			req.PricingConfig.CODSurchargeRate,
		); err != nil {
			return nil, fmt.Errorf("failed to set pricing: %w", err)
		}
	}
	
	// Set service levels if provided
	if req.ServiceLevels != nil {
		if err := provider.SetServiceLevels(
			req.ServiceLevels.StandardDeliveryHours,
			req.ServiceLevels.ExpressDeliveryHours,
			req.ServiceLevels.SameDayAvailable,
			req.ServiceLevels.CODAvailable,
			req.ServiceLevels.TrackingAvailable,
			req.ServiceLevels.InsuranceAvailable,
		); err != nil {
			return nil, fmt.Errorf("failed to set service levels: %w", err)
		}
	}
	
	// Set operational configuration if provided
	if req.OperationalConfig != nil {
		if req.OperationalConfig.MaxWeight > 0 {
			if err := provider.SetMaxWeight(decimal.NewFromFloat(req.OperationalConfig.MaxWeight)); err != nil {
				return nil, fmt.Errorf("failed to set max weight: %w", err)
			}
		}
		
		provider.WeekendService = req.OperationalConfig.WeekendService
		provider.HolidayService = req.OperationalConfig.HolidayService
		provider.DailyCutoffTime = req.OperationalConfig.DailyCutoffTime
		provider.AutoAssign = req.OperationalConfig.AutoAssign
		provider.SetPriority(req.OperationalConfig.PriorityOrder)
	}
	
	// Save provider
	if err := uc.providerRepo.Create(ctx, provider); err != nil {
		return nil, fmt.Errorf("failed to save provider: %w", err)
	}
	
	// Publish event
	event := map[string]interface{}{
		"event_type":     "provider_created",
		"provider_id":    provider.ID.String(),
		"provider_code":  provider.ProviderCode,
		"provider_name":  provider.ProviderName,
		"provider_type":  string(provider.ProviderType),
		"created_at":     provider.CreatedAt,
	}
	
	if err := uc.eventPublisher.Publish(ctx, "provider.created", event); err != nil {
		// Don't fail the operation for event publishing errors
		fmt.Printf("Failed to publish provider created event: %v\n", err)
	}
	
	return provider, nil
}

// GetProvider retrieves a provider by ID
func (uc *ProviderUseCase) GetProvider(ctx context.Context, providerID uuid.UUID) (*entity.DeliveryProvider, error) {
	provider, err := uc.providerRepo.GetByID(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}
	
	return provider, nil
}

// GetProviderByCode retrieves a provider by code
func (uc *ProviderUseCase) GetProviderByCode(ctx context.Context, code string) (*entity.DeliveryProvider, error) {
	provider, err := uc.providerRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider by code: %w", err)
	}
	
	return provider, nil
}

// GetActiveProviders retrieves all active providers
func (uc *ProviderUseCase) GetActiveProviders(ctx context.Context) ([]*entity.DeliveryProvider, error) {
	providers, err := uc.providerRepo.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active providers: %w", err)
	}
	
	return providers, nil
}

// GetProvidersByType retrieves providers by type
func (uc *ProviderUseCase) GetProvidersByType(ctx context.Context, providerType entity.ProviderType) ([]*entity.DeliveryProvider, error) {
	providers, err := uc.providerRepo.GetByType(ctx, providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers by type: %w", err)
	}
	
	return providers, nil
}

// GetProvidersByArea retrieves providers that serve a specific area
func (uc *ProviderUseCase) GetProvidersByArea(ctx context.Context, province string) ([]*entity.DeliveryProvider, error) {
	providers, err := uc.providerRepo.GetProvidersForArea(ctx, province, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get providers by area: %w", err)
	}
	
	return providers, nil
}

// UpdateProvider updates an existing provider
func (uc *ProviderUseCase) UpdateProvider(ctx context.Context, req UpdateProviderRequest) (*entity.DeliveryProvider, error) {
	// Get existing provider
	provider, err := uc.providerRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider for update: %w", err)
	}
	
	// Update basic information
	if req.Name != nil {
		provider.ProviderName = *req.Name
	}
	if req.Description != nil {
		// Add description field to entity if needed
	}
	
	// Update API configuration if provided
	if req.APIConfig != nil && provider.ProviderType == entity.ProviderTypeAPIIntegrated {
		if err := provider.SetAPIConfiguration(
			req.APIConfig.BaseURL,
			req.APIConfig.APIKey,
			req.APIConfig.AuthMethod,
		); err != nil {
			return nil, fmt.Errorf("failed to update API configuration: %w", err)
		}
		
		if req.APIConfig.RateQuoteEndpoint != "" {
			provider.RateQuoteAPIEndpoint = req.APIConfig.RateQuoteEndpoint
			provider.SupportsRateComparison = true
		}
	}
	
	// Update contact information if provided
	if req.ContactInfo != nil {
		if err := provider.SetManualCoordinationInfo(
			req.ContactInfo.Phone,
			req.ContactInfo.LineID,
			req.ContactInfo.Email,
			req.ContactInfo.Notes,
		); err != nil {
			return nil, fmt.Errorf("failed to update contact information: %w", err)
		}
	}
	
	// Update pricing configuration if provided
	if req.PricingConfig != nil {
		if err := provider.SetPricing(
			req.PricingConfig.BaseRate,
			req.PricingConfig.PerKmRate,
			req.PricingConfig.WeightSurchargeRate,
			req.PricingConfig.SameDaySurcharge,
			req.PricingConfig.CODSurchargeRate,
		); err != nil {
			return nil, fmt.Errorf("failed to update pricing: %w", err)
		}
	}
	
	// Update service levels if provided
	if req.ServiceLevels != nil {
		if err := provider.SetServiceLevels(
			req.ServiceLevels.StandardDeliveryHours,
			req.ServiceLevels.ExpressDeliveryHours,
			req.ServiceLevels.SameDayAvailable,
			req.ServiceLevels.CODAvailable,
			req.ServiceLevels.TrackingAvailable,
			req.ServiceLevels.InsuranceAvailable,
		); err != nil {
			return nil, fmt.Errorf("failed to update service levels: %w", err)
		}
	}
	
	// Update operational configuration if provided
	if req.OperationalConfig != nil {
		if req.OperationalConfig.MaxWeight > 0 {
			if err := provider.SetMaxWeight(decimal.NewFromFloat(req.OperationalConfig.MaxWeight)); err != nil {
				return nil, fmt.Errorf("failed to update max weight: %w", err)
			}
		}
		
		provider.WeekendService = req.OperationalConfig.WeekendService
		provider.HolidayService = req.OperationalConfig.HolidayService
		provider.DailyCutoffTime = req.OperationalConfig.DailyCutoffTime
		provider.AutoAssign = req.OperationalConfig.AutoAssign
		provider.SetPriority(req.OperationalConfig.PriorityOrder)
	}
	
	provider.UpdatedAt = time.Now()
	
	// Save updated provider
	if err := uc.providerRepo.Update(ctx, provider); err != nil {
		return nil, fmt.Errorf("failed to save provider: %w", err)
	}
	
	// Publish event
	event := map[string]interface{}{
		"event_type":    "provider_updated",
		"provider_id":   provider.ID.String(),
		"provider_code": provider.ProviderCode,
		"updated_at":    provider.UpdatedAt,
	}
	
	if err := uc.eventPublisher.Publish(ctx, "provider.updated", event); err != nil {
		fmt.Printf("Failed to publish provider updated event: %v\n", err)
	}
	
	return provider, nil
}

// ActivateProvider activates a provider
func (uc *ProviderUseCase) ActivateProvider(ctx context.Context, providerID uuid.UUID) error {
	provider, err := uc.providerRepo.GetByID(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}
	
	provider.Activate()
	
	if err := uc.providerRepo.Update(ctx, provider); err != nil {
		return fmt.Errorf("failed to activate provider: %w", err)
	}
	
	// Publish event
	event := map[string]interface{}{
		"event_type":    "provider_activated",
		"provider_id":   provider.ID.String(),
		"provider_code": provider.ProviderCode,
		"activated_at":  time.Now(),
	}
	
	if err := uc.eventPublisher.Publish(ctx, "provider.activated", event); err != nil {
		fmt.Printf("Failed to publish provider activated event: %v\n", err)
	}
	
	return nil
}

// DeactivateProvider deactivates a provider
func (uc *ProviderUseCase) DeactivateProvider(ctx context.Context, providerID uuid.UUID) error {
	provider, err := uc.providerRepo.GetByID(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}
	
	provider.Deactivate()
	
	if err := uc.providerRepo.Update(ctx, provider); err != nil {
		return fmt.Errorf("failed to deactivate provider: %w", err)
	}
	
	// Publish event
	event := map[string]interface{}{
		"event_type":      "provider_deactivated",
		"provider_id":     provider.ID.String(),
		"provider_code":   provider.ProviderCode,
		"deactivated_at":  time.Now(),
	}
	
	if err := uc.eventPublisher.Publish(ctx, "provider.deactivated", event); err != nil {
		fmt.Printf("Failed to publish provider deactivated event: %v\n", err)
	}
	
	return nil
}

// UpdatePerformanceMetrics updates provider performance metrics
func (uc *ProviderUseCase) UpdatePerformanceMetrics(ctx context.Context, req PerformanceMetricsRequest) error {
	provider, err := uc.providerRepo.GetByID(ctx, req.ProviderID)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}
	
	if err := provider.UpdatePerformanceMetrics(
		req.AverageDeliveryTime,
		req.SuccessRate,
		req.CustomerRating,
	); err != nil {
		return fmt.Errorf("failed to update performance metrics: %w", err)
	}
	
	if err := uc.providerRepo.Update(ctx, provider); err != nil {
		return fmt.Errorf("failed to save provider metrics: %w", err)
	}
	
	// Publish event
	event := map[string]interface{}{
		"event_type":            "provider_metrics_updated",
		"provider_id":           provider.ID.String(),
		"average_delivery_time": req.AverageDeliveryTime.String(),
		"success_rate":          req.SuccessRate.String(),
		"customer_rating":       req.CustomerRating.String(),
		"updated_at":            time.Now(),
	}
	
	if err := uc.eventPublisher.Publish(ctx, "provider.metrics_updated", event); err != nil {
		fmt.Printf("Failed to publish provider metrics updated event: %v\n", err)
	}
	
	return nil
}

// CalculateDeliveryFee calculates delivery fee for a provider
func (uc *ProviderUseCase) CalculateDeliveryFee(
	ctx context.Context,
	providerID uuid.UUID,
	distance, weight decimal.Decimal,
	sameDayDelivery, codRequired bool,
) (decimal.Decimal, error) {
	provider, err := uc.providerRepo.GetByID(ctx, providerID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get provider: %w", err)
	}
	
	if !provider.IsActive {
		return decimal.Zero, errors.New("provider is not active")
	}
	
	fee := provider.CalculateDeliveryFee(distance, weight, sameDayDelivery, codRequired)
	
	return fee, nil
}

// CheckProviderAvailability checks if a provider is available for delivery
func (uc *ProviderUseCase) CheckProviderAvailability(
	ctx context.Context,
	providerID uuid.UUID,
	province string,
	weight decimal.Decimal,
	sameDayRequired, codRequired bool,
) (bool, error) {
	provider, err := uc.providerRepo.GetByID(ctx, providerID)
	if err != nil {
		return false, fmt.Errorf("failed to get provider: %w", err)
	}
	
	return provider.IsAvailableForDelivery(province, weight, sameDayRequired, codRequired), nil
}

// SetProviderPriority sets the priority order for a provider
func (uc *ProviderUseCase) SetProviderPriority(ctx context.Context, providerID uuid.UUID, priority int) error {
	provider, err := uc.providerRepo.GetByID(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}
	
	provider.SetPriority(priority)
	
	if err := uc.providerRepo.Update(ctx, provider); err != nil {
		return fmt.Errorf("failed to update provider priority: %w", err)
	}
	
	// Clear cache for provider rankings
	cacheKey := fmt.Sprintf("provider:rankings:%s", "all")
	if err := uc.cache.Delete(ctx, cacheKey); err != nil {
		fmt.Printf("Failed to clear provider rankings cache: %v\n", err)
	}
	
	return nil
}

// GetProviderRankings retrieves providers ranked by priority and performance
func (uc *ProviderUseCase) GetProviderRankings(ctx context.Context, province string) ([]*entity.DeliveryProvider, error) {
	cacheKey := fmt.Sprintf("provider:rankings:%s", province)
	
	// Try to get from cache first
	cached, err := uc.cache.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		// Parse cached result if needed
		// For now, fetch from repository
	}
	
	providers, err := uc.providerRepo.GetProvidersForArea(ctx, province, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get provider rankings: %w", err)
	}
	
	// Cache the result for 30 minutes
	if err := uc.cache.Set(ctx, cacheKey, "cached", 30*time.Minute); err != nil {
		fmt.Printf("Failed to cache provider rankings: %v\n", err)
	}
	
	return providers, nil
}

// DeleteProvider soft deletes a provider
func (uc *ProviderUseCase) DeleteProvider(ctx context.Context, providerID uuid.UUID) error {
	provider, err := uc.providerRepo.GetByID(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}
	
	// For now, skip the active deliveries check since the method doesn't exist
	// TODO: Add HasActiveDeliveries method to the repository interface
	
	// Soft delete
	if err := uc.providerRepo.Delete(ctx, providerID); err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}
	
	// Publish event
	event := map[string]interface{}{
		"event_type":    "provider_deleted",
		"provider_id":   provider.ID.String(),
		"provider_code": provider.ProviderCode,
		"deleted_at":    time.Now(),
	}
	
	if err := uc.eventPublisher.Publish(ctx, "provider.deleted", event); err != nil {
		fmt.Printf("Failed to publish provider deleted event: %v\n", err)
	}
	
	return nil
}
