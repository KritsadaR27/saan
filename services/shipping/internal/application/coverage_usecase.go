package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"shipping/internal/domain/entity"
	"shipping/internal/domain/repository"
)

// CoverageUseCase handles coverage area operations
type CoverageUseCase struct {
	coverageRepo repository.CoverageAreaRepository
	eventPub     EventPublisher
	cache        Cache
}

// NewCoverageUseCase creates a new coverage use case
func NewCoverageUseCase(
	coverageRepo repository.CoverageAreaRepository,
	eventPub EventPublisher,
	cache Cache,
) *CoverageUseCase {
	return &CoverageUseCase{
		coverageRepo: coverageRepo,
		eventPub:     eventPub,
		cache:        cache,
	}
}

// CreateCoverageAreaRequest represents a request to create a coverage area
type CreateCoverageAreaRequest struct {
	Province             string          `json:"province" validate:"required"`
	District             string          `json:"district,omitempty"`
	Subdistrict          string          `json:"subdistrict,omitempty"`
	PostalCode           string          `json:"postal_code,omitempty"`
	IsSelfDeliveryArea   bool            `json:"is_self_delivery_area"`
	DeliveryRoute        string          `json:"delivery_route,omitempty"`
	DeliveryZone         string          `json:"delivery_zone,omitempty"`
	BaseDeliveryFee      decimal.Decimal `json:"base_delivery_fee"`
	PerKmRate            decimal.Decimal `json:"per_km_rate"`
	FreeDeliveryThreshold decimal.Decimal `json:"free_delivery_threshold"`
	StandardDeliveryHours int             `json:"standard_delivery_hours"`
	ExpressDeliveryHours  int             `json:"express_delivery_hours"`
	SameDayAvailable     bool            `json:"same_day_available"`
	MaxDailyCapacity     int             `json:"max_daily_capacity"`
	CreatedBy            string          `json:"created_by" validate:"required"`
}

// UpdateCoverageAreaRequest represents a request to update a coverage area
type UpdateCoverageAreaRequest struct {
	ID                   uuid.UUID        `json:"id" validate:"required"`
	District             *string          `json:"district,omitempty"`
	Subdistrict          *string          `json:"subdistrict,omitempty"`
	PostalCode           *string          `json:"postal_code,omitempty"`
	DeliveryRoute        *string          `json:"delivery_route,omitempty"`
	DeliveryZone         *string          `json:"delivery_zone,omitempty"`
	BaseDeliveryFee      *decimal.Decimal `json:"base_delivery_fee,omitempty"`
	PerKmRate            *decimal.Decimal `json:"per_km_rate,omitempty"`
	FreeDeliveryThreshold *decimal.Decimal `json:"free_delivery_threshold,omitempty"`
	StandardDeliveryHours *int             `json:"standard_delivery_hours,omitempty"`
	ExpressDeliveryHours  *int             `json:"express_delivery_hours,omitempty"`
	SameDayAvailable     *bool            `json:"same_day_available,omitempty"`
	MaxDailyCapacity     *int             `json:"max_daily_capacity,omitempty"`
	IsActive             *bool            `json:"is_active,omitempty"`
	UpdatedBy            string           `json:"updated_by" validate:"required"`
}

// CreateCoverageArea creates a new coverage area
func (uc *CoverageUseCase) CreateCoverageArea(ctx context.Context, req CreateCoverageAreaRequest) (*entity.CoverageArea, error) {
	// Create new coverage area entity
	coverage, err := entity.NewCoverageArea(req.Province, req.IsSelfDeliveryArea)
	if err != nil {
		return nil, fmt.Errorf("failed to create coverage area entity: %w", err)
	}

	// Set location details
	if req.District != "" || req.Subdistrict != "" || req.PostalCode != "" {
		coverage.SetLocation(req.District, req.Subdistrict, req.PostalCode)
	}

	// Set delivery route for self-delivery areas
	if req.IsSelfDeliveryArea && req.DeliveryRoute != "" {
		if err := coverage.SetDeliveryRoute(req.DeliveryRoute, req.DeliveryZone); err != nil {
			return nil, fmt.Errorf("failed to set delivery route: %w", err)
		}
	}

	// Set pricing if provided
	if !req.BaseDeliveryFee.IsZero() || !req.PerKmRate.IsZero() || !req.FreeDeliveryThreshold.IsZero() {
		if err := coverage.SetPricing(req.BaseDeliveryFee, req.PerKmRate, req.FreeDeliveryThreshold); err != nil {
			return nil, fmt.Errorf("failed to set pricing: %w", err)
		}
	}

	// Set service levels if provided
	if req.StandardDeliveryHours > 0 || req.ExpressDeliveryHours > 0 {
		if err := coverage.SetServiceLevels(req.StandardDeliveryHours, req.ExpressDeliveryHours, req.SameDayAvailable); err != nil {
			return nil, fmt.Errorf("failed to set service levels: %w", err)
		}
	}

	// Set capacity if provided
	if req.MaxDailyCapacity > 0 {
		if err := coverage.SetCapacity(req.MaxDailyCapacity); err != nil {
			return nil, fmt.Errorf("failed to set capacity: %w", err)
		}
	}

	// Save to repository
	if err := uc.coverageRepo.Create(ctx, coverage); err != nil {
		return nil, fmt.Errorf("failed to save coverage area: %w", err)
	}

	// Publish event
	uc.eventPub.Publish(ctx, "coverage_area.created", map[string]interface{}{
		"coverage_id":          coverage.ID.String(),
		"province":             coverage.Province,
		"district":             coverage.District,
		"is_self_delivery_area": coverage.IsSelfDeliveryArea,
		"created_by":           req.CreatedBy,
		"created_at":           coverage.CreatedAt,
	})

	return coverage, nil
}

// GetCoverageArea retrieves a coverage area by ID
func (uc *CoverageUseCase) GetCoverageArea(ctx context.Context, id uuid.UUID) (*entity.CoverageArea, error) {
	coverage, err := uc.coverageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get coverage area: %w", err)
	}

	return coverage, nil
}

// GetCoverageAreasByProvince retrieves all coverage areas for a province
func (uc *CoverageUseCase) GetCoverageAreasByProvince(ctx context.Context, province string) ([]*entity.CoverageArea, error) {
	coverageAreas, err := uc.coverageRepo.GetByProvince(ctx, province)
	if err != nil {
		return nil, fmt.Errorf("failed to get coverage areas for province: %w", err)
	}

	return coverageAreas, nil
}

// GetSelfDeliveryAreas retrieves all self-delivery coverage areas
func (uc *CoverageUseCase) GetSelfDeliveryAreas(ctx context.Context) ([]*entity.CoverageArea, error) {
	coverageAreas, err := uc.coverageRepo.GetSelfDeliveryAreas(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get self-delivery areas: %w", err)
	}

	return coverageAreas, nil
}

// FindCoverageForLocation finds the best coverage area for a location
func (uc *CoverageUseCase) FindCoverageForLocation(ctx context.Context, province, district, subdistrict, postalCode string) (*entity.CoverageArea, error) {
	coverage, err := uc.coverageRepo.GetBestMatchForLocation(ctx, province, district, subdistrict, postalCode)
	if err != nil {
		return nil, fmt.Errorf("failed to find coverage for location: %w", err)
	}

	return coverage, nil
}

// UpdateCoverageArea updates an existing coverage area
func (uc *CoverageUseCase) UpdateCoverageArea(ctx context.Context, req UpdateCoverageAreaRequest) (*entity.CoverageArea, error) {
	// Get existing coverage area
	coverage, err := uc.coverageRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coverage area: %w", err)
	}

	// Update location if provided
	if req.District != nil || req.Subdistrict != nil || req.PostalCode != nil {
		district := coverage.District
		subdistrict := coverage.Subdistrict
		postalCode := coverage.PostalCode

		if req.District != nil {
			district = *req.District
		}
		if req.Subdistrict != nil {
			subdistrict = *req.Subdistrict
		}
		if req.PostalCode != nil {
			postalCode = *req.PostalCode
		}

		coverage.SetLocation(district, subdistrict, postalCode)
	}

	// Update delivery route if provided
	if req.DeliveryRoute != nil || req.DeliveryZone != nil {
		route := coverage.DeliveryRoute
		zone := coverage.DeliveryZone

		if req.DeliveryRoute != nil {
			route = *req.DeliveryRoute
		}
		if req.DeliveryZone != nil {
			zone = *req.DeliveryZone
		}

		if coverage.IsSelfDeliveryArea {
			if err := coverage.SetDeliveryRoute(route, zone); err != nil {
				return nil, fmt.Errorf("failed to update delivery route: %w", err)
			}
		}
	}

	// Update pricing if provided
	if req.BaseDeliveryFee != nil || req.PerKmRate != nil || req.FreeDeliveryThreshold != nil {
		baseDeliveryFee := coverage.BaseDeliveryFee
		perKmRate := coverage.PerKmRate
		freeDeliveryThreshold := coverage.FreeDeliveryThreshold

		if req.BaseDeliveryFee != nil {
			baseDeliveryFee = *req.BaseDeliveryFee
		}
		if req.PerKmRate != nil {
			perKmRate = *req.PerKmRate
		}
		if req.FreeDeliveryThreshold != nil {
			freeDeliveryThreshold = *req.FreeDeliveryThreshold
		}

		if err := coverage.SetPricing(baseDeliveryFee, perKmRate, freeDeliveryThreshold); err != nil {
			return nil, fmt.Errorf("failed to update pricing: %w", err)
		}
	}

	// Update service levels if provided
	if req.StandardDeliveryHours != nil || req.ExpressDeliveryHours != nil || req.SameDayAvailable != nil {
		standardHours := coverage.StandardDeliveryHours
		expressHours := coverage.ExpressDeliveryHours
		sameDayAvailable := coverage.SameDayAvailable

		if req.StandardDeliveryHours != nil {
			standardHours = *req.StandardDeliveryHours
		}
		if req.ExpressDeliveryHours != nil {
			expressHours = *req.ExpressDeliveryHours
		}
		if req.SameDayAvailable != nil {
			sameDayAvailable = *req.SameDayAvailable
		}

		if err := coverage.SetServiceLevels(standardHours, expressHours, sameDayAvailable); err != nil {
			return nil, fmt.Errorf("failed to update service levels: %w", err)
		}
	}

	// Update capacity if provided
	if req.MaxDailyCapacity != nil {
		if err := coverage.SetCapacity(*req.MaxDailyCapacity); err != nil {
			return nil, fmt.Errorf("failed to update capacity: %w", err)
		}
	}

	// Update active status if provided
	if req.IsActive != nil {
		if *req.IsActive {
			coverage.Activate()
		} else {
			coverage.Deactivate()
		}
	}

	// Save updated coverage area
	if err := uc.coverageRepo.Update(ctx, coverage); err != nil {
		return nil, fmt.Errorf("failed to update coverage area: %w", err)
	}

	// Publish event
	uc.eventPub.Publish(ctx, "coverage_area.updated", map[string]interface{}{
		"coverage_id": coverage.ID.String(),
		"updated_by":  req.UpdatedBy,
		"updated_at":  coverage.UpdatedAt,
	})

	return coverage, nil
}

// DeleteCoverageArea deletes a coverage area
func (uc *CoverageUseCase) DeleteCoverageArea(ctx context.Context, id uuid.UUID, deletedBy string) error {
	// Get coverage area to ensure it exists
	coverage, err := uc.coverageRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get coverage area: %w", err)
	}

	// Delete coverage area
	if err := uc.coverageRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete coverage area: %w", err)
	}

	// Publish event
	uc.eventPub.Publish(ctx, "coverage_area.deleted", map[string]interface{}{
		"coverage_id": id.String(),
		"province":    coverage.Province,
		"deleted_by":  deletedBy,
		"deleted_at":  time.Now(),
	})

	return nil
}

// CalculateDeliveryFee calculates delivery fee for a location
func (uc *CoverageUseCase) CalculateDeliveryFee(ctx context.Context, province, district, subdistrict, postalCode string, orderValue decimal.Decimal, distance float64) (*FeeCalculation, error) {
	// Find best coverage area for location
	coverage, err := uc.coverageRepo.GetBestMatchForLocation(ctx, province, district, subdistrict, postalCode)
	if err != nil {
		return nil, fmt.Errorf("no coverage available for location: %w", err)
	}

	// Calculate fee using coverage area logic
	distanceDecimal := decimal.NewFromFloat(distance)
	deliveryFee := coverage.CalculateDeliveryFee(distanceDecimal, orderValue)

	calculation := &FeeCalculation{
		CoverageID:      coverage.ID,
		Province:        coverage.Province,
		District:        coverage.District,
		BaseDeliveryFee: coverage.BaseDeliveryFee,
		PerKmRate:       coverage.PerKmRate,
		Distance:        distanceDecimal,
		OrderValue:      orderValue,
		DeliveryFee:     deliveryFee,
		IsFreeDelivery:  deliveryFee.IsZero(),
		CalculatedAt:    time.Now(),
	}

	return calculation, nil
}

// GetCoverageStats retrieves coverage statistics
func (uc *CoverageUseCase) GetCoverageStats(ctx context.Context) (*CoverageStats, error) {
	allAreas, err := uc.coverageRepo.GetAll(ctx, 1000, 0) // Get all areas
	if err != nil {
		return nil, fmt.Errorf("failed to get coverage areas: %w", err)
	}

	stats := &CoverageStats{
		TotalAreas:          len(allAreas),
		ActiveAreas:         0,
		InactiveAreas:       0,
		SelfDeliveryAreas:   0,
		ThirdPartyAreas:     0,
		ProvincesCovered:    make(map[string]int),
		RoutesCovered:       make(map[string]int),
	}

	for _, area := range allAreas {
		if area.IsActive {
			stats.ActiveAreas++
		} else {
			stats.InactiveAreas++
		}

		if area.IsSelfDeliveryArea {
			stats.SelfDeliveryAreas++
			if area.DeliveryRoute != "" {
				stats.RoutesCovered[area.DeliveryRoute]++
			}
		} else {
			stats.ThirdPartyAreas++
		}

		stats.ProvincesCovered[area.Province]++
	}

	return stats, nil
}

// FeeCalculation represents a delivery fee calculation result
type FeeCalculation struct {
	CoverageID      uuid.UUID       `json:"coverage_id"`
	Province        string          `json:"province"`
	District        string          `json:"district"`
	BaseDeliveryFee decimal.Decimal `json:"base_delivery_fee"`
	PerKmRate       decimal.Decimal `json:"per_km_rate"`
	Distance        decimal.Decimal `json:"distance"`
	OrderValue      decimal.Decimal `json:"order_value"`
	DeliveryFee     decimal.Decimal `json:"delivery_fee"`
	IsFreeDelivery  bool            `json:"is_free_delivery"`
	CalculatedAt    time.Time       `json:"calculated_at"`
}

// CoverageStats represents coverage statistics
type CoverageStats struct {
	TotalAreas        int            `json:"total_areas"`
	ActiveAreas       int            `json:"active_areas"`
	InactiveAreas     int            `json:"inactive_areas"`
	SelfDeliveryAreas int            `json:"self_delivery_areas"`
	ThirdPartyAreas   int            `json:"third_party_areas"`
	ProvincesCovered  map[string]int `json:"provinces_covered"`
	RoutesCovered     map[string]int `json:"routes_covered"`
}
