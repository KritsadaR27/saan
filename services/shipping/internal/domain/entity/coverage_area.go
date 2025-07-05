package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CoverageArea represents a delivery coverage area configuration
type CoverageArea struct {
	ID           uuid.UUID `json:"id"`
	Province     string    `json:"province"`
	District     string    `json:"district,omitempty"`
	Subdistrict  string    `json:"subdistrict,omitempty"`
	PostalCode   string    `json:"postal_code,omitempty"`
	
	// Delivery Configuration
	IsSelfDeliveryArea      bool   `json:"is_self_delivery_area"`
	DeliveryRoute           string `json:"delivery_route,omitempty"`
	DeliveryZone            string `json:"delivery_zone,omitempty"` // A, B, C zones
	PriorityOrder           int    `json:"priority_order"`
	
	// Pricing
	BaseDeliveryFee         decimal.Decimal `json:"base_delivery_fee"`
	PerKmRate               decimal.Decimal `json:"per_km_rate"`
	FreeDeliveryThreshold   decimal.Decimal `json:"free_delivery_threshold"`
	
	// Service Levels
	StandardDeliveryHours   int  `json:"standard_delivery_hours"`
	ExpressDeliveryHours    int  `json:"express_delivery_hours"`
	SameDayAvailable        bool `json:"same_day_available"`
	
	// Admin Configuration
	IsActive                bool `json:"is_active"`
	AutoAssign              bool `json:"auto_assign"`
	MaxDailyCapacity        int  `json:"max_daily_capacity"`
	
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// Domain errors
var (
	ErrCoverageInvalidProvince     = errors.New("province cannot be empty")
	ErrCoverageInvalidRoute        = errors.New("delivery route cannot be empty for self-delivery areas")
	ErrCoverageInvalidFee          = errors.New("delivery fee must be non-negative")
	ErrCoverageInvalidRate         = errors.New("per km rate must be non-negative")
	ErrCoverageInvalidThreshold    = errors.New("free delivery threshold must be positive")
	ErrCoverageInvalidHours        = errors.New("delivery hours must be positive")
	ErrCoverageInvalidCapacity     = errors.New("daily capacity must be positive")
	ErrCoverageInvalidPriority     = errors.New("priority order must be positive")
	ErrCoverageNotSelfDelivery     = errors.New("operation only valid for self-delivery areas")
)

// NewCoverageArea creates a new coverage area with validation
func NewCoverageArea(province string, isSelfDeliveryArea bool) (*CoverageArea, error) {
	if err := validateProvince(province); err != nil {
		return nil, err
	}
	
	now := time.Now()
	return &CoverageArea{
		ID:                    uuid.New(),
		Province:              province,
		IsSelfDeliveryArea:    isSelfDeliveryArea,
		PriorityOrder:         100,
		BaseDeliveryFee:       decimal.NewFromFloat(50.0), // Default 50 THB
		PerKmRate:             decimal.NewFromFloat(5.0),   // Default 5 THB/km
		FreeDeliveryThreshold: decimal.NewFromFloat(500.0), // Default 500 THB
		StandardDeliveryHours: 24,
		ExpressDeliveryHours:  4,
		SameDayAvailable:      false,
		IsActive:              true,
		AutoAssign:            true,
		MaxDailyCapacity:      100,
		CreatedAt:             now,
		UpdatedAt:             now,
	}, nil
}

// SetLocation sets the detailed location information
func (c *CoverageArea) SetLocation(district, subdistrict, postalCode string) {
	c.District = district
	c.Subdistrict = subdistrict
	c.PostalCode = postalCode
	c.UpdatedAt = time.Now()
}

// SetDeliveryRoute sets the delivery route for self-delivery areas
func (c *CoverageArea) SetDeliveryRoute(route, zone string) error {
	if !c.IsSelfDeliveryArea {
		return ErrCoverageNotSelfDelivery
	}
	
	if route == "" {
		return ErrCoverageInvalidRoute
	}
	
	c.DeliveryRoute = route
	c.DeliveryZone = zone
	c.UpdatedAt = time.Now()
	
	return nil
}

// SetPricing sets the pricing configuration
func (c *CoverageArea) SetPricing(baseDeliveryFee, perKmRate, freeDeliveryThreshold decimal.Decimal) error {
	if baseDeliveryFee.IsNegative() {
		return ErrCoverageInvalidFee
	}
	
	if perKmRate.IsNegative() {
		return ErrCoverageInvalidRate
	}
	
	if freeDeliveryThreshold.IsNegative() || freeDeliveryThreshold.IsZero() {
		return ErrCoverageInvalidThreshold
	}
	
	c.BaseDeliveryFee = baseDeliveryFee
	c.PerKmRate = perKmRate
	c.FreeDeliveryThreshold = freeDeliveryThreshold
	c.UpdatedAt = time.Now()
	
	return nil
}

// SetServiceLevels sets the service level configuration
func (c *CoverageArea) SetServiceLevels(standardHours, expressHours int, sameDayAvailable bool) error {
	if standardHours <= 0 {
		return ErrCoverageInvalidHours
	}
	
	if expressHours <= 0 {
		return ErrCoverageInvalidHours
	}
	
	c.StandardDeliveryHours = standardHours
	c.ExpressDeliveryHours = expressHours
	c.SameDayAvailable = sameDayAvailable
	c.UpdatedAt = time.Now()
	
	return nil
}

// SetCapacity sets the daily capacity limit
func (c *CoverageArea) SetCapacity(maxDailyCapacity int) error {
	if maxDailyCapacity <= 0 {
		return ErrCoverageInvalidCapacity
	}
	
	c.MaxDailyCapacity = maxDailyCapacity
	c.UpdatedAt = time.Now()
	
	return nil
}

// SetPriority sets the priority order
func (c *CoverageArea) SetPriority(priority int) error {
	if priority <= 0 {
		return ErrCoverageInvalidPriority
	}
	
	c.PriorityOrder = priority
	c.UpdatedAt = time.Now()
	
	return nil
}

// Activate activates the coverage area
func (c *CoverageArea) Activate() {
	c.IsActive = true
	c.UpdatedAt = time.Now()
}

// Deactivate deactivates the coverage area
func (c *CoverageArea) Deactivate() {
	c.IsActive = false
	c.UpdatedAt = time.Now()
}

// EnableAutoAssign enables auto assignment
func (c *CoverageArea) EnableAutoAssign() {
	c.AutoAssign = true
	c.UpdatedAt = time.Now()
}

// DisableAutoAssign disables auto assignment
func (c *CoverageArea) DisableAutoAssign() {
	c.AutoAssign = false
	c.UpdatedAt = time.Now()
}

// CalculateDeliveryFee calculates the delivery fee based on distance and order value
func (c *CoverageArea) CalculateDeliveryFee(distance decimal.Decimal, orderValue decimal.Decimal) decimal.Decimal {
	// Check for free delivery
	if orderValue.GreaterThanOrEqual(c.FreeDeliveryThreshold) {
		return decimal.Zero
	}
	
	// Calculate fee
	fee := c.BaseDeliveryFee
	
	// Add distance charge
	if distance.GreaterThan(decimal.Zero) {
		distanceCharge := distance.Mul(c.PerKmRate)
		fee = fee.Add(distanceCharge)
	}
	
	return fee
}

// GetEstimatedDeliveryTime gets estimated delivery time based on service level
func (c *CoverageArea) GetEstimatedDeliveryTime(serviceLevel string) time.Duration {
	switch serviceLevel {
	case "express":
		return time.Duration(c.ExpressDeliveryHours) * time.Hour
	case "same_day":
		if c.SameDayAvailable {
			return 8 * time.Hour // Same day within 8 hours
		}
		fallthrough
	default:
		return time.Duration(c.StandardDeliveryHours) * time.Hour
	}
}

// SupportsServiceLevel checks if the area supports the requested service level
func (c *CoverageArea) SupportsServiceLevel(serviceLevel string) bool {
	if !c.IsActive {
		return false
	}
	
	switch serviceLevel {
	case "standard":
		return true
	case "express":
		return c.ExpressDeliveryHours > 0
	case "same_day":
		return c.SameDayAvailable
	default:
		return false
	}
}

// IsAvailableForDelivery checks if the area is available for delivery
func (c *CoverageArea) IsAvailableForDelivery() bool {
	return c.IsActive && c.AutoAssign
}

// GetLocationString returns a formatted location string
func (c *CoverageArea) GetLocationString() string {
	parts := []string{c.Province}
	
	if c.District != "" {
		parts = append(parts, c.District)
	}
	
	if c.Subdistrict != "" {
		parts = append(parts, c.Subdistrict)
	}
	
	if c.PostalCode != "" {
		parts = append(parts, c.PostalCode)
	}
	
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += " > "
		}
		result += part
	}
	
	return result
}

// MatchesLocation checks if the coverage area matches the given location
func (c *CoverageArea) MatchesLocation(province, district, subdistrict, postalCode string) bool {
	// Province must match
	if c.Province != province {
		return false
	}
	
	// Check district if specified
	if c.District != "" && c.District != district {
		return false
	}
	
	// Check subdistrict if specified
	if c.Subdistrict != "" && c.Subdistrict != subdistrict {
		return false
	}
	
	// Check postal code if specified
	if c.PostalCode != "" && c.PostalCode != postalCode {
		return false
	}
	
	return true
}

// GetDeliveryInfo returns delivery information for this coverage area
func (c *CoverageArea) GetDeliveryInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":                      c.ID,
		"location":                c.GetLocationString(),
		"is_self_delivery":        c.IsSelfDeliveryArea,
		"delivery_route":          c.DeliveryRoute,
		"delivery_zone":           c.DeliveryZone,
		"base_delivery_fee":       c.BaseDeliveryFee,
		"per_km_rate":             c.PerKmRate,
		"free_delivery_threshold": c.FreeDeliveryThreshold,
		"standard_delivery_hours": c.StandardDeliveryHours,
		"express_delivery_hours":  c.ExpressDeliveryHours,
		"same_day_available":      c.SameDayAvailable,
		"max_daily_capacity":      c.MaxDailyCapacity,
		"priority_order":          c.PriorityOrder,
		"is_active":               c.IsActive,
		"auto_assign":             c.AutoAssign,
	}
}

// Validation functions
func validateProvince(province string) error {
	if province == "" {
		return ErrCoverageInvalidProvince
	}
	return nil
}
