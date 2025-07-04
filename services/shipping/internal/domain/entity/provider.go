package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// DeliveryProvider represents a delivery service provider configuration
type DeliveryProvider struct {
	ID           uuid.UUID `json:"id"`
	ProviderCode string    `json:"provider_code"`
	ProviderName string    `json:"provider_name"`
	ProviderType ProviderType `json:"provider_type"`
	
	// API Configuration
	APIBaseURL    string `json:"api_base_url,omitempty"`
	APIVersion    string `json:"api_version"`
	HasAPI        bool   `json:"has_api"`
	AuthMethod    string `json:"auth_method,omitempty"`
	
	// Service Configuration
	CoverageAreas        map[string]interface{} `json:"coverage_areas"`
	SupportedPackageTypes map[string]interface{} `json:"supported_package_types"`
	MaxWeight            decimal.Decimal        `json:"max_weight_kg"`
	MaxDimensions        map[string]interface{} `json:"max_dimensions"`
	
	// Pricing Configuration
	BaseRate              decimal.Decimal `json:"base_rate"`
	PerKmRate             decimal.Decimal `json:"per_km_rate"`
	WeightSurchargeRate   decimal.Decimal `json:"weight_surcharge_rate"`
	SameDaySurcharge      decimal.Decimal `json:"same_day_surcharge"`
	CODSurchargeRate      decimal.Decimal `json:"cod_surcharge_rate"`
	
	// Service Levels
	StandardDeliveryHours int  `json:"standard_delivery_hours"`
	ExpressDeliveryHours  int  `json:"express_delivery_hours"`
	SameDayAvailable      bool `json:"same_day_available"`
	CODAvailable          bool `json:"cod_available"`
	TrackingAvailable     bool `json:"tracking_available"`
	InsuranceAvailable    bool `json:"insurance_available"`
	
	// Operational
	DailyCutoffTime   *time.Time            `json:"daily_cutoff_time"`
	WeekendService    bool                  `json:"weekend_service"`
	HolidayService    bool                  `json:"holiday_service"`
	BusinessHours     map[string]interface{} `json:"business_hours"`
	
	// Manual Coordination
	ContactPhone         string `json:"contact_phone,omitempty"`
	ContactLineID        string `json:"contact_line_id,omitempty"`
	ContactEmail         string `json:"contact_email,omitempty"`
	ManualCoordination   bool   `json:"manual_coordination"`
	CoordinationNotes    string `json:"coordination_notes,omitempty"`
	
	// Inter Express Specific
	DailyAutoPickup               bool       `json:"daily_auto_pickup"`
	PickupCancellationDeadline    *time.Time `json:"pickup_cancellation_deadline"`
	CancellationFee               decimal.Decimal `json:"cancellation_fee"`
	AutoCancelCheckTime           *time.Time `json:"auto_cancel_check_time"`
	
	// Performance Metrics
	AverageDeliveryTime   decimal.Decimal `json:"average_delivery_time_hours"`
	SuccessRate           decimal.Decimal `json:"success_rate_percentage"`
	CustomerRating        decimal.Decimal `json:"customer_rating"`
	LastPerformanceUpdate *time.Time      `json:"last_performance_update"`
	
	// Rate Comparison
	SupportsRateComparison  bool   `json:"supports_rate_comparison"`
	RateQuoteAPIEndpoint    string `json:"rate_quote_api_endpoint,omitempty"`
	RateCacheDuration       int    `json:"rate_cache_duration_minutes"`
	
	// Admin Configuration
	IsActive            bool `json:"is_active"`
	PriorityOrder       int  `json:"priority_order"`
	AutoAssign          bool `json:"auto_assign"`
	RequiresApproval    bool `json:"requires_approval"`
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProviderType represents the type of delivery provider
type ProviderType string

const (
	ProviderTypeAPIIntegrated    ProviderType = "api_integrated"
	ProviderTypeManualCoordination ProviderType = "manual_coordination"
	ProviderTypeAutoPickup       ProviderType = "auto_pickup"
)

// Domain errors
var (
	ErrProviderInvalidCode         = errors.New("provider code cannot be empty")
	ErrProviderInvalidName         = errors.New("provider name cannot be empty")
	ErrProviderInvalidType         = errors.New("invalid provider type")
	ErrProviderInvalidWeight       = errors.New("max weight must be positive")
	ErrProviderInvalidRate         = errors.New("rates must be non-negative")
	ErrProviderInvalidRating       = errors.New("customer rating must be between 1.0 and 5.0")
	ErrProviderInvalidDeliveryTime = errors.New("delivery time must be positive")
	ErrProviderInvalidSuccessRate  = errors.New("success rate must be between 0 and 100")
	ErrProviderAPIRequired         = errors.New("API configuration required for API integrated provider")
	ErrProviderContactRequired     = errors.New("contact information required for manual coordination provider")
)

// NewDeliveryProvider creates a new delivery provider with validation
func NewDeliveryProvider(code, name string, providerType ProviderType) (*DeliveryProvider, error) {
	if err := validateProviderCode(code); err != nil {
		return nil, err
	}
	
	if err := validateProviderName(name); err != nil {
		return nil, err
	}
	
	if err := validateProviderType(providerType); err != nil {
		return nil, err
	}
	
	now := time.Now()
	return &DeliveryProvider{
		ID:                      uuid.New(),
		ProviderCode:            code,
		ProviderName:            name,
		ProviderType:            providerType,
		APIVersion:              "v1",
		HasAPI:                  providerType == ProviderTypeAPIIntegrated,
		CoverageAreas:           make(map[string]interface{}),
		SupportedPackageTypes:   make(map[string]interface{}),
		MaxWeight:               decimal.NewFromFloat(100.0), // Default 100kg
		MaxDimensions:           make(map[string]interface{}),
		BaseRate:                decimal.Zero,
		PerKmRate:               decimal.Zero,
		WeightSurchargeRate:     decimal.Zero,
		SameDaySurcharge:        decimal.Zero,
		CODSurchargeRate:        decimal.Zero,
		StandardDeliveryHours:   24,
		ExpressDeliveryHours:    4,
		SameDayAvailable:        false,
		CODAvailable:            false,
		TrackingAvailable:       true,
		InsuranceAvailable:      false,
		WeekendService:          false,
		HolidayService:          false,
		BusinessHours:           make(map[string]interface{}),
		ManualCoordination:      providerType == ProviderTypeManualCoordination,
		DailyAutoPickup:         providerType == ProviderTypeAutoPickup,
		CancellationFee:         decimal.NewFromFloat(50.0), // Default 50 THB
		AverageDeliveryTime:     decimal.NewFromFloat(24.0),
		SuccessRate:             decimal.NewFromFloat(95.0),
		CustomerRating:          decimal.NewFromFloat(4.0),
		SupportsRateComparison:  providerType == ProviderTypeAPIIntegrated,
		RateCacheDuration:       30,
		IsActive:                true,
		PriorityOrder:           100,
		AutoAssign:              true,
		RequiresApproval:        false,
		CreatedAt:               now,
		UpdatedAt:               now,
	}, nil
}

// SetAPIConfiguration sets API configuration for API integrated providers
func (p *DeliveryProvider) SetAPIConfiguration(baseURL, apiKey, authMethod string) error {
	if p.ProviderType != ProviderTypeAPIIntegrated {
		return errors.New("API configuration only allowed for API integrated providers")
	}
	
	if baseURL == "" {
		return errors.New("API base URL cannot be empty")
	}
	
	p.APIBaseURL = baseURL
	p.AuthMethod = authMethod
	p.HasAPI = true
	p.UpdatedAt = time.Now()
	
	return nil
}

// SetManualCoordinationInfo sets contact information for manual coordination providers
func (p *DeliveryProvider) SetManualCoordinationInfo(phone, lineID, email, notes string) error {
	if p.ProviderType != ProviderTypeManualCoordination && p.ProviderType != ProviderTypeAutoPickup {
		return errors.New("manual coordination info only allowed for manual coordination providers")
	}
	
	if phone == "" && lineID == "" && email == "" {
		return ErrProviderContactRequired
	}
	
	p.ContactPhone = phone
	p.ContactLineID = lineID
	p.ContactEmail = email
	p.CoordinationNotes = notes
	p.ManualCoordination = true
	p.UpdatedAt = time.Now()
	
	return nil
}

// SetPricing sets the pricing configuration
func (p *DeliveryProvider) SetPricing(baseRate, perKmRate, weightSurcharge, sameDaySurcharge, codSurcharge decimal.Decimal) error {
	if baseRate.IsNegative() || perKmRate.IsNegative() || weightSurcharge.IsNegative() || 
	   sameDaySurcharge.IsNegative() || codSurcharge.IsNegative() {
		return ErrProviderInvalidRate
	}
	
	p.BaseRate = baseRate
	p.PerKmRate = perKmRate
	p.WeightSurchargeRate = weightSurcharge
	p.SameDaySurcharge = sameDaySurcharge
	p.CODSurchargeRate = codSurcharge
	p.UpdatedAt = time.Now()
	
	return nil
}

// SetServiceLevels sets the service level configuration
func (p *DeliveryProvider) SetServiceLevels(standardHours, expressHours int, sameDayAvailable, codAvailable, trackingAvailable, insuranceAvailable bool) error {
	if standardHours <= 0 || expressHours <= 0 {
		return ErrProviderInvalidDeliveryTime
	}
	
	p.StandardDeliveryHours = standardHours
	p.ExpressDeliveryHours = expressHours
	p.SameDayAvailable = sameDayAvailable
	p.CODAvailable = codAvailable
	p.TrackingAvailable = trackingAvailable
	p.InsuranceAvailable = insuranceAvailable
	p.UpdatedAt = time.Now()
	
	return nil
}

// UpdatePerformanceMetrics updates the provider's performance metrics
func (p *DeliveryProvider) UpdatePerformanceMetrics(avgDeliveryTime, successRate, customerRating decimal.Decimal) error {
	if avgDeliveryTime.IsNegative() {
		return ErrProviderInvalidDeliveryTime
	}
	
	if successRate.IsNegative() || successRate.GreaterThan(decimal.NewFromFloat(100.0)) {
		return ErrProviderInvalidSuccessRate
	}
	
	if customerRating.LessThan(decimal.NewFromFloat(1.0)) || customerRating.GreaterThan(decimal.NewFromFloat(5.0)) {
		return ErrProviderInvalidRating
	}
	
	now := time.Now()
	p.AverageDeliveryTime = avgDeliveryTime
	p.SuccessRate = successRate
	p.CustomerRating = customerRating
	p.LastPerformanceUpdate = &now
	p.UpdatedAt = now
	
	return nil
}

// SetInterExpressConfig sets Inter Express specific configuration
func (p *DeliveryProvider) SetInterExpressConfig(autoPickup bool, cancellationDeadline, autoCheckTime time.Time, fee decimal.Decimal) error {
	if p.ProviderCode != "inter" {
		return errors.New("Inter Express configuration only allowed for inter provider")
	}
	
	if fee.IsNegative() {
		return ErrProviderInvalidRate
	}
	
	p.DailyAutoPickup = autoPickup
	p.PickupCancellationDeadline = &cancellationDeadline
	p.AutoCancelCheckTime = &autoCheckTime
	p.CancellationFee = fee
	p.UpdatedAt = time.Now()
	
	return nil
}

// SetCoverageAreas sets the coverage areas
func (p *DeliveryProvider) SetCoverageAreas(areas map[string]interface{}) {
	p.CoverageAreas = areas
	p.UpdatedAt = time.Now()
}

// SetMaxWeight sets the maximum weight limit
func (p *DeliveryProvider) SetMaxWeight(weight decimal.Decimal) error {
	if weight.IsNegative() || weight.IsZero() {
		return ErrProviderInvalidWeight
	}
	
	p.MaxWeight = weight
	p.UpdatedAt = time.Now()
	return nil
}

// Activate activates the provider
func (p *DeliveryProvider) Activate() {
	p.IsActive = true
	p.UpdatedAt = time.Now()
}

// Deactivate deactivates the provider
func (p *DeliveryProvider) Deactivate() {
	p.IsActive = false
	p.UpdatedAt = time.Now()
}

// SetPriority sets the provider priority
func (p *DeliveryProvider) SetPriority(priority int) {
	p.PriorityOrder = priority
	p.UpdatedAt = time.Now()
}

// CalculateDeliveryFee calculates the delivery fee based on distance and weight
func (p *DeliveryProvider) CalculateDeliveryFee(distance decimal.Decimal, weight decimal.Decimal, sameDayDelivery bool, codRequired bool) decimal.Decimal {
	fee := p.BaseRate
	
	// Add distance charge
	distanceCharge := distance.Mul(p.PerKmRate)
	fee = fee.Add(distanceCharge)
	
	// Add weight surcharge if applicable
	if weight.GreaterThan(decimal.NewFromFloat(5.0)) { // Over 5kg
		weightCharge := weight.Sub(decimal.NewFromFloat(5.0)).Mul(p.WeightSurchargeRate)
		fee = fee.Add(weightCharge)
	}
	
	// Add same day surcharge
	if sameDayDelivery && p.SameDayAvailable {
		fee = fee.Add(p.SameDaySurcharge)
	}
	
	// Add COD surcharge
	if codRequired && p.CODAvailable {
		fee = fee.Add(p.CODSurchargeRate)
	}
	
	return fee
}

// IsAvailableForDelivery checks if the provider is available for delivery
func (p *DeliveryProvider) IsAvailableForDelivery(province string, weight decimal.Decimal, sameDayRequired bool, codRequired bool) bool {
	if !p.IsActive {
		return false
	}
	
	// Check weight limit
	if weight.GreaterThan(p.MaxWeight) {
		return false
	}
	
	// Check same day availability
	if sameDayRequired && !p.SameDayAvailable {
		return false
	}
	
	// Check COD availability
	if codRequired && !p.CODAvailable {
		return false
	}
	
	// Check coverage area (simplified - would need proper implementation)
	if len(p.CoverageAreas) > 0 {
		if _, exists := p.CoverageAreas[province]; !exists {
			return false
		}
	}
	
	return true
}

// IsWithinCutoffTime checks if current time is within cutoff time
func (p *DeliveryProvider) IsWithinCutoffTime() bool {
	if p.DailyCutoffTime == nil {
		return true
	}
	
	now := time.Now()
	cutoff := time.Date(now.Year(), now.Month(), now.Day(), 
		p.DailyCutoffTime.Hour(), p.DailyCutoffTime.Minute(), 0, 0, now.Location())
	
	return now.Before(cutoff)
}

// Validation functions
func validateProviderCode(code string) error {
	if code == "" {
		return ErrProviderInvalidCode
	}
	return nil
}

func validateProviderName(name string) error {
	if name == "" {
		return ErrProviderInvalidName
	}
	return nil
}

func validateProviderType(providerType ProviderType) error {
	switch providerType {
	case ProviderTypeAPIIntegrated, ProviderTypeManualCoordination, ProviderTypeAutoPickup:
		return nil
	default:
		return ErrProviderInvalidType
	}
}
