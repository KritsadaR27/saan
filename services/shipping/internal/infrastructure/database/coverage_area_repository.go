package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"shipping/internal/domain/entity"
	"shipping/internal/domain/repository"
)

type coverageAreaRepository struct {
	db *sqlx.DB
}

// NewCoverageAreaRepository creates a new coverage area repository implementation
func NewCoverageAreaRepository(db *sqlx.DB) repository.CoverageAreaRepository {
	return &coverageAreaRepository{db: db}
}

// Create creates a new coverage area
func (r *coverageAreaRepository) Create(ctx context.Context, area *entity.CoverageArea) error {
	query := `
		INSERT INTO coverage_areas (
			id, province, district, subdistrict, postal_code,
			is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			base_delivery_fee, per_km_rate, free_delivery_threshold,
			standard_delivery_hours, express_delivery_hours, same_day_available,
			is_active, auto_assign, max_daily_capacity, created_at, updated_at
		) VALUES (
			:id, :province, :district, :subdistrict, :postal_code,
			:is_self_delivery_area, :delivery_route, :delivery_zone, :priority_order,
			:base_delivery_fee, :per_km_rate, :free_delivery_threshold,
			:standard_delivery_hours, :express_delivery_hours, :same_day_available,
			:is_active, :auto_assign, :max_daily_capacity, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":                       area.ID,
		"province":                 area.Province,
		"district":                 area.District,
		"subdistrict":              area.Subdistrict,
		"postal_code":              area.PostalCode,
		"is_self_delivery_area":    area.IsSelfDeliveryArea,
		"delivery_route":           area.DeliveryRoute,
		"delivery_zone":            area.DeliveryZone,
		"priority_order":           area.PriorityOrder,
		"base_delivery_fee":        area.BaseDeliveryFee,
		"per_km_rate":              area.PerKmRate,
		"free_delivery_threshold":  area.FreeDeliveryThreshold,
		"standard_delivery_hours":  area.StandardDeliveryHours,
		"express_delivery_hours":   area.ExpressDeliveryHours,
		"same_day_available":       area.SameDayAvailable,
		"is_active":                area.IsActive,
		"auto_assign":              area.AutoAssign,
		"max_daily_capacity":       area.MaxDailyCapacity,
		"created_at":               area.CreatedAt,
		"updated_at":               area.UpdatedAt,
	})

	return err
}

// GetByID retrieves a coverage area by ID
func (r *coverageAreaRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE id = $1`

	var area entity.CoverageArea

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&area.ID, &area.Province, &area.District, &area.Subdistrict, &area.PostalCode,
		&area.IsSelfDeliveryArea, &area.DeliveryRoute, &area.DeliveryZone, &area.PriorityOrder,
		&area.BaseDeliveryFee, &area.PerKmRate, &area.FreeDeliveryThreshold,
		&area.StandardDeliveryHours, &area.ExpressDeliveryHours, &area.SameDayAvailable,
		&area.IsActive, &area.AutoAssign, &area.MaxDailyCapacity, &area.CreatedAt, &area.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrCoverageAreaNotFound
		}
		return nil, fmt.Errorf("failed to get coverage area: %w", err)
	}

	return &area, nil
}

// Update updates an existing coverage area
func (r *coverageAreaRepository) Update(ctx context.Context, area *entity.CoverageArea) error {
	query := `
		UPDATE coverage_areas SET
			province = :province,
			district = :district,
			subdistrict = :subdistrict,
			postal_code = :postal_code,
			is_self_delivery_area = :is_self_delivery_area,
			delivery_route = :delivery_route,
			delivery_zone = :delivery_zone,
			priority_order = :priority_order,
			base_delivery_fee = :base_delivery_fee,
			per_km_rate = :per_km_rate,
			free_delivery_threshold = :free_delivery_threshold,
			standard_delivery_hours = :standard_delivery_hours,
			express_delivery_hours = :express_delivery_hours,
			same_day_available = :same_day_available,
			is_active = :is_active,
			auto_assign = :auto_assign,
			max_daily_capacity = :max_daily_capacity,
			updated_at = :updated_at
		WHERE id = :id`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":                       area.ID,
		"province":                 area.Province,
		"district":                 area.District,
		"subdistrict":              area.Subdistrict,
		"postal_code":              area.PostalCode,
		"is_self_delivery_area":    area.IsSelfDeliveryArea,
		"delivery_route":           area.DeliveryRoute,
		"delivery_zone":            area.DeliveryZone,
		"priority_order":           area.PriorityOrder,
		"base_delivery_fee":        area.BaseDeliveryFee,
		"per_km_rate":              area.PerKmRate,
		"free_delivery_threshold":  area.FreeDeliveryThreshold,
		"standard_delivery_hours":  area.StandardDeliveryHours,
		"express_delivery_hours":   area.ExpressDeliveryHours,
		"same_day_available":       area.SameDayAvailable,
		"is_active":                area.IsActive,
		"auto_assign":              area.AutoAssign,
		"max_daily_capacity":       area.MaxDailyCapacity,
		"updated_at":               area.UpdatedAt,
	})

	return err
}

// Delete deletes a coverage area
func (r *coverageAreaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM coverage_areas WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete coverage area: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrCoverageAreaNotFound
	}

	return nil
}

// GetAll retrieves all coverage areas with pagination
func (r *coverageAreaRepository) GetAll(ctx context.Context, limit, offset int) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		ORDER BY province, district, subdistrict
		LIMIT $1 OFFSET $2`

	return r.queryAreas(ctx, query, limit, offset)
}

// GetActive retrieves all active coverage areas
func (r *coverageAreaRepository) GetActive(ctx context.Context) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE is_active = true
		ORDER BY priority_order, province, district`

	return r.queryAreas(ctx, query)
}

// GetByProvince retrieves coverage areas by province
func (r *coverageAreaRepository) GetByProvince(ctx context.Context, province string) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE province = $1 AND is_active = true
		ORDER BY priority_order, district`

	return r.queryAreas(ctx, query, province)
}

// GetSelfDeliveryAreas retrieves self-delivery areas
func (r *coverageAreaRepository) GetSelfDeliveryAreas(ctx context.Context) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE is_self_delivery_area = true AND is_active = true
		ORDER BY priority_order, province, district`

	return r.queryAreas(ctx, query)
}

// GetThirdPartyAreas retrieves third-party delivery areas
func (r *coverageAreaRepository) GetThirdPartyAreas(ctx context.Context) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE is_self_delivery_area = false AND is_active = true
		ORDER BY priority_order, province, district`

	return r.queryAreas(ctx, query)
}

// FindByLocation finds coverage areas matching a location
func (r *coverageAreaRepository) FindByLocation(ctx context.Context, province, district, subdistrict, postalCode string) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE is_active = true
		AND province = $1
		AND (district = $2 OR district = '' OR district IS NULL)
		AND (subdistrict = $3 OR subdistrict = '' OR subdistrict IS NULL)
		AND (postal_code = $4 OR postal_code = '' OR postal_code IS NULL)
		ORDER BY priority_order,
			CASE WHEN subdistrict = $3 THEN 1 ELSE 0 END DESC,
			CASE WHEN district = $2 THEN 1 ELSE 0 END DESC,
			CASE WHEN postal_code = $4 THEN 1 ELSE 0 END DESC`

	return r.queryAreas(ctx, query, province, district, subdistrict, postalCode)
}

// GetBestMatchForLocation finds the best matching coverage area for a location
func (r *coverageAreaRepository) GetBestMatchForLocation(ctx context.Context, province, district, subdistrict, postalCode string) (*entity.CoverageArea, error) {
	areas, err := r.FindByLocation(ctx, province, district, subdistrict, postalCode)
	if err != nil {
		return nil, err
	}

	if len(areas) == 0 {
		return nil, repository.ErrLocationNotCovered
	}

	return areas[0], nil
}

// GetByPostalCode retrieves coverage areas by postal code
func (r *coverageAreaRepository) GetByPostalCode(ctx context.Context, postalCode string) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE postal_code = $1 AND is_active = true
		ORDER BY priority_order`

	return r.queryAreas(ctx, query, postalCode)
}

// GetByDeliveryZone retrieves coverage areas by delivery zone
func (r *coverageAreaRepository) GetByDeliveryZone(ctx context.Context, zone string) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE delivery_zone = $1 AND is_active = true
		ORDER BY priority_order`

	return r.queryAreas(ctx, query, zone)
}

// GetByRoute retrieves coverage areas by delivery route
func (r *coverageAreaRepository) GetByRoute(ctx context.Context, route string) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE delivery_route = $1 AND is_active = true
		ORDER BY priority_order`

	return r.queryAreas(ctx, query, route)
}

// GetSameDayAreas retrieves coverage areas with same-day delivery available
func (r *coverageAreaRepository) GetSameDayAreas(ctx context.Context) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE same_day_available = true AND is_active = true
		ORDER BY priority_order`

	return r.queryAreas(ctx, query)
}

// ActivateArea activates a coverage area
func (r *coverageAreaRepository) ActivateArea(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE coverage_areas SET is_active = true, updated_at = NOW() WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to activate coverage area: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrCoverageAreaNotFound
	}

	return nil
}

// DeactivateArea deactivates a coverage area
func (r *coverageAreaRepository) DeactivateArea(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE coverage_areas SET is_active = false, updated_at = NOW() WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate coverage area: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrCoverageAreaNotFound
	}

	return nil
}

// UpdateFee updates the base delivery fee for a coverage area
func (r *coverageAreaRepository) UpdateFee(ctx context.Context, id uuid.UUID, fee decimal.Decimal) error {
	query := `UPDATE coverage_areas SET base_delivery_fee = $1, updated_at = NOW() WHERE id = $2`
	
	result, err := r.db.ExecContext(ctx, query, fee, id)
	if err != nil {
		return fmt.Errorf("failed to update delivery fee: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrCoverageAreaNotFound
	}

	return nil
}

// UpdateCapacity updates the maximum daily capacity for a coverage area
func (r *coverageAreaRepository) UpdateCapacity(ctx context.Context, id uuid.UUID, capacity int) error {
	query := `UPDATE coverage_areas SET max_daily_capacity = $1, updated_at = NOW() WHERE id = $2`
	
	result, err := r.db.ExecContext(ctx, query, capacity, id)
	if err != nil {
		return fmt.Errorf("failed to update capacity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrCoverageAreaNotFound
	}

	return nil
}

// EnableAutoAssign enables auto-assign for a coverage area
func (r *coverageAreaRepository) EnableAutoAssign(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE coverage_areas SET auto_assign = true, updated_at = NOW() WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to enable auto assign: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrCoverageAreaNotFound
	}

	return nil
}

// DisableAutoAssign disables auto-assign for a coverage area
func (r *coverageAreaRepository) DisableAutoAssign(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE coverage_areas SET auto_assign = false, updated_at = NOW() WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to disable auto assign: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrCoverageAreaNotFound
	}

	return nil
}

// GetAreasByPriceRange retrieves areas within a price range (alias for GetAreasByFeeRange)
func (r *coverageAreaRepository) GetAreasByPriceRange(ctx context.Context, minFee, maxFee decimal.Decimal) ([]*entity.CoverageArea, error) {
	return r.GetAreasByFeeRange(ctx, minFee, maxFee)
}

// GetAreasByFeeRange retrieves areas within a fee range
func (r *coverageAreaRepository) GetAreasByFeeRange(ctx context.Context, minFee, maxFee decimal.Decimal) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE base_delivery_fee >= $1 AND base_delivery_fee <= $2
		AND is_active = true
		ORDER BY base_delivery_fee ASC`

	return r.queryAreas(ctx, query, minFee, maxFee)
}

// GetAreasBySameDaySupport retrieves areas by same day delivery support
func (r *coverageAreaRepository) GetAreasBySameDaySupport(ctx context.Context, supported bool) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE same_day_available = $1 AND is_active = true
		ORDER BY priority_order ASC`

	return r.queryAreas(ctx, query, supported)
}

// GetAreasByServiceLevel retrieves areas by service level
func (r *coverageAreaRepository) GetAreasByServiceLevel(ctx context.Context, serviceLevel string) ([]*entity.CoverageArea, error) {
	// Service level based on delivery hours: standard (>24h), express (<=24h), same_day
	var query string
	switch serviceLevel {
	case "same_day":
		query = `
			SELECT id, province, district, subdistrict, postal_code,
				   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
				   base_delivery_fee, per_km_rate, free_delivery_threshold,
				   standard_delivery_hours, express_delivery_hours, same_day_available,
				   is_active, auto_assign, max_daily_capacity, created_at, updated_at
			FROM coverage_areas 
			WHERE same_day_available = true AND is_active = true
			ORDER BY priority_order ASC`
	case "express":
		query = `
			SELECT id, province, district, subdistrict, postal_code,
				   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
				   base_delivery_fee, per_km_rate, free_delivery_threshold,
				   standard_delivery_hours, express_delivery_hours, same_day_available,
				   is_active, auto_assign, max_daily_capacity, created_at, updated_at
			FROM coverage_areas 
			WHERE express_delivery_hours <= 24 AND is_active = true
			ORDER BY express_delivery_hours ASC`
	default: // standard
		query = `
			SELECT id, province, district, subdistrict, postal_code,
				   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
				   base_delivery_fee, per_km_rate, free_delivery_threshold,
				   standard_delivery_hours, express_delivery_hours, same_day_available,
				   is_active, auto_assign, max_daily_capacity, created_at, updated_at
			FROM coverage_areas 
			WHERE is_active = true
			ORDER BY standard_delivery_hours ASC`
	}

	return r.queryAreas(ctx, query)
}

// GetAreasByCapacityRange retrieves areas within a capacity range
func (r *coverageAreaRepository) GetAreasByCapacityRange(ctx context.Context, minCapacity, maxCapacity int) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE max_daily_capacity >= $1 AND max_daily_capacity <= $2
		AND is_active = true
		ORDER BY max_daily_capacity DESC`

	return r.queryAreas(ctx, query, minCapacity, maxCapacity)
}

// GetAreasWithAvailableCapacity retrieves areas with available capacity
func (r *coverageAreaRepository) GetAreasWithAvailableCapacity(ctx context.Context, requiredCapacity int) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE max_daily_capacity >= $1 AND is_active = true
		ORDER BY max_daily_capacity DESC`

	return r.queryAreas(ctx, query, requiredCapacity)
}

// GetAutoAssignAreas retrieves areas with auto assignment enabled
func (r *coverageAreaRepository) GetAutoAssignAreas(ctx context.Context) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE auto_assign = true AND is_active = true
		ORDER BY priority_order ASC`

	return r.queryAreas(ctx, query)
}

// SearchAreas searches areas with filters
func (r *coverageAreaRepository) SearchAreas(ctx context.Context, filters *repository.CoverageAreaQueryFilters) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if filters.Province != nil {
		query += fmt.Sprintf(" AND province = $%d", argIndex)
		args = append(args, *filters.Province)
		argIndex++
	}

	if filters.District != nil {
		query += fmt.Sprintf(" AND district = $%d", argIndex)
		args = append(args, *filters.District)
		argIndex++
	}

	if filters.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, *filters.IsActive)
		argIndex++
	}

	query += " ORDER BY priority_order ASC"
	
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
		
		if filters.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, filters.Offset)
		}
	}

	return r.queryAreasWithArgs(ctx, query, args...)
}

// GetAreasByPattern searches areas by location pattern
func (r *coverageAreaRepository) GetAreasByPattern(ctx context.Context, locationPattern string) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE (province ILIKE $1 OR district ILIKE $1 OR subdistrict ILIKE $1 OR postal_code ILIKE $1)
		AND is_active = true
		ORDER BY priority_order ASC`

	pattern := "%" + locationPattern + "%"
	return r.queryAreas(ctx, query, pattern)
}

// UpdateMultipleAreaStatuses updates status for multiple areas
func (r *coverageAreaRepository) UpdateMultipleAreaStatuses(ctx context.Context, areaIDs []uuid.UUID, isActive bool) error {
	if len(areaIDs) == 0 {
		return nil
	}

	query := `UPDATE coverage_areas SET is_active = $1, updated_at = NOW() WHERE id = ANY($2)`
	
	_, err := r.db.ExecContext(ctx, query, isActive, areaIDs)
	return err
}

// GetAreasByIDs retrieves areas by their IDs
func (r *coverageAreaRepository) GetAreasByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.CoverageArea, error) {
	if len(ids) == 0 {
		return []*entity.CoverageArea{}, nil
	}

	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE id = ANY($1)
		ORDER BY priority_order ASC`

	return r.queryAreas(ctx, query, ids)
}

// CreateBulkAreas creates multiple areas in bulk
func (r *coverageAreaRepository) CreateBulkAreas(ctx context.Context, areas []*entity.CoverageArea) error {
	if len(areas) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO coverage_areas (
			id, province, district, subdistrict, postal_code,
			is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			base_delivery_fee, per_km_rate, free_delivery_threshold,
			standard_delivery_hours, express_delivery_hours, same_day_available,
			is_active, auto_assign, max_daily_capacity, created_at, updated_at
		) VALUES (
			:id, :province, :district, :subdistrict, :postal_code,
			:is_self_delivery_area, :delivery_route, :delivery_zone, :priority_order,
			:base_delivery_fee, :per_km_rate, :free_delivery_threshold,
			:standard_delivery_hours, :express_delivery_hours, :same_day_available,
			:is_active, :auto_assign, :max_daily_capacity, :created_at, :updated_at
		)`

	for _, area := range areas {
		_, err := tx.NamedExecContext(ctx, query, map[string]interface{}{
			"id":                       area.ID,
			"province":                 area.Province,
			"district":                 area.District,
			"subdistrict":              area.Subdistrict,
			"postal_code":              area.PostalCode,
			"is_self_delivery_area":    area.IsSelfDeliveryArea,
			"delivery_route":           area.DeliveryRoute,
			"delivery_zone":            area.DeliveryZone,
			"priority_order":           area.PriorityOrder,
			"base_delivery_fee":        area.BaseDeliveryFee,
			"per_km_rate":              area.PerKmRate,
			"free_delivery_threshold":  area.FreeDeliveryThreshold,
			"standard_delivery_hours":  area.StandardDeliveryHours,
			"express_delivery_hours":   area.ExpressDeliveryHours,
			"same_day_available":       area.SameDayAvailable,
			"is_active":                area.IsActive,
			"auto_assign":              area.AutoAssign,
			"max_daily_capacity":       area.MaxDailyCapacity,
			"created_at":               area.CreatedAt,
			"updated_at":               area.UpdatedAt,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetCoverageStats retrieves coverage statistics
func (r *coverageAreaRepository) GetCoverageStats(ctx context.Context) (*repository.CoverageStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_areas,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_areas,
			COUNT(CASE WHEN is_self_delivery_area = true THEN 1 END) as self_delivery_areas,
			COUNT(CASE WHEN same_day_available = true THEN 1 END) as same_day_areas,
			COUNT(CASE WHEN auto_assign = true THEN 1 END) as auto_assign_areas,
			AVG(base_delivery_fee) as avg_delivery_fee,
			AVG(max_daily_capacity) as avg_capacity
		FROM coverage_areas`

	var stats repository.CoverageStats
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalAreas,
		&stats.ActiveAreas,
		&stats.SelfDeliveryAreas,
		&stats.SameDayAreas,
		&stats.AutoAssignAreas,
		&stats.AverageBaseFee,
		&stats.AverageCapacity,
	)

	return &stats, err
}

// GetProvinceCoverage retrieves coverage by province
func (r *coverageAreaRepository) GetProvinceCoverage(ctx context.Context) (map[string]*repository.ProvinceCoverage, error) {
	query := `
		SELECT 
			province,
			COUNT(*) as total_areas,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_areas,
			COUNT(CASE WHEN is_self_delivery_area = true THEN 1 END) as self_delivery_areas,
			AVG(base_delivery_fee) as avg_fee
		FROM coverage_areas
		GROUP BY province
		ORDER BY province`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]*repository.ProvinceCoverage)
	for rows.Next() {
		var coverage repository.ProvinceCoverage
		var province string
		
		err := rows.Scan(&province, &coverage.TotalAreas, &coverage.ActiveAreas, &coverage.SelfDeliveryAreas, &coverage.AverageBaseFee)
		if err != nil {
			return nil, err
		}
		
		result[province] = &coverage
	}

	return result, rows.Err()
}

// GetUnservicedAreas retrieves areas not covered
func (r *coverageAreaRepository) GetUnservicedAreas(ctx context.Context) ([]string, error) {
	// This would typically query against a master location table
	// For now, return an empty list as this needs business logic
	return []string{}, nil
}

// GetByZone retrieves areas by delivery zone
func (r *coverageAreaRepository) GetByZone(ctx context.Context, zone string) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE delivery_zone = $1 AND is_active = true
		ORDER BY priority_order ASC`

	return r.queryAreas(ctx, query, zone)
}

// GetAvailableRoutes retrieves all available delivery routes
func (r *coverageAreaRepository) GetAvailableRoutes(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT delivery_route
		FROM coverage_areas 
		WHERE delivery_route IS NOT NULL AND delivery_route != ''
		AND is_active = true
		ORDER BY delivery_route`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []string
	for rows.Next() {
		var route string
		if err := rows.Scan(&route); err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}

	return routes, rows.Err()
}

// GetAvailableZones retrieves all available delivery zones
func (r *coverageAreaRepository) GetAvailableZones(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT delivery_zone
		FROM coverage_areas 
		WHERE delivery_zone IS NOT NULL AND delivery_zone != ''
		AND is_active = true
		ORDER BY delivery_zone`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []string
	for rows.Next() {
		var zone string
		if err := rows.Scan(&zone); err != nil {
			return nil, err
		}
		zones = append(zones, zone)
	}

	return zones, rows.Err()
}

// GetByPriorityOrder retrieves areas ordered by priority
func (r *coverageAreaRepository) GetByPriorityOrder(ctx context.Context) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE is_active = true
		ORDER BY priority_order ASC`

	return r.queryAreas(ctx, query)
}

// GetDeliveryOptions retrieves delivery options for a location
func (r *coverageAreaRepository) GetDeliveryOptions(ctx context.Context, province, district, subdistrict, postalCode string) ([]*repository.DeliveryOption, error) {
	areas, err := r.FindByLocation(ctx, province, district, subdistrict, postalCode)
	if err != nil {
		return nil, err
	}

	var options []*repository.DeliveryOption
	for _, area := range areas {
		option := &repository.DeliveryOption{
			CoverageAreaID:        area.ID,
			LocationString:        fmt.Sprintf("%s, %s, %s %s", area.Subdistrict, area.District, area.Province, area.PostalCode),
			IsSelfDelivery:        area.IsSelfDeliveryArea,
			DeliveryRoute:         area.DeliveryRoute,
			DeliveryZone:          area.DeliveryZone,
			BaseDeliveryFee:       area.BaseDeliveryFee,
			PerKmRate:             area.PerKmRate,
			FreeDeliveryThreshold: area.FreeDeliveryThreshold,
			StandardDeliveryHours: area.StandardDeliveryHours,
			ExpressDeliveryHours:  area.ExpressDeliveryHours,
			SameDayAvailable:      area.SameDayAvailable,
			MaxDailyCapacity:      area.MaxDailyCapacity,
			PriorityOrder:         area.PriorityOrder,
			IsActive:              area.IsActive,
			AutoAssign:            area.AutoAssign,
		}
		options = append(options, option)
	}

	return options, nil
}

// GetAreasWithFreeDelivery retrieves areas that offer free delivery for the given order value
func (r *coverageAreaRepository) GetAreasWithFreeDelivery(ctx context.Context, orderValue decimal.Decimal) ([]*entity.CoverageArea, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code,
			   is_self_delivery_area, delivery_route, delivery_zone, priority_order,
			   base_delivery_fee, per_km_rate, free_delivery_threshold,
			   standard_delivery_hours, express_delivery_hours, same_day_available,
			   is_active, auto_assign, max_daily_capacity, created_at, updated_at
		FROM coverage_areas 
		WHERE free_delivery_threshold <= $1 AND is_active = true
		ORDER BY priority_order ASC`

	return r.queryAreas(ctx, query, orderValue)
}

// UpdatePriority updates the priority order of an area
func (r *coverageAreaRepository) UpdatePriority(ctx context.Context, id uuid.UUID, priority int) error {
	query := `UPDATE coverage_areas SET priority_order = $1, updated_at = NOW() WHERE id = $2`
	
	result, err := r.db.ExecContext(ctx, query, priority, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrCoverageAreaNotFound
	}

	return nil
}

// UpdatePricing updates the pricing information for an area
func (r *coverageAreaRepository) UpdatePricing(ctx context.Context, id uuid.UUID, baseDeliveryFee, perKmRate, freeDeliveryThreshold decimal.Decimal) error {
	query := `
		UPDATE coverage_areas SET 
			base_delivery_fee = $1, 
			per_km_rate = $2, 
			free_delivery_threshold = $3,
			updated_at = NOW()
		WHERE id = $4`
	
	result, err := r.db.ExecContext(ctx, query, baseDeliveryFee, perKmRate, freeDeliveryThreshold, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrCoverageAreaNotFound
	}

	return nil
}

// UpdateServiceLevels updates the service level configuration for an area
func (r *coverageAreaRepository) UpdateServiceLevels(ctx context.Context, id uuid.UUID, standardHours, expressHours int, sameDayAvailable bool) error {
	query := `
		UPDATE coverage_areas SET 
			standard_delivery_hours = $1, 
			express_delivery_hours = $2, 
			same_day_available = $3,
			updated_at = NOW()
		WHERE id = $4`
	
	result, err := r.db.ExecContext(ctx, query, standardHours, expressHours, sameDayAvailable, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrCoverageAreaNotFound
	}

	return nil
}

// Helper methods
func (r *coverageAreaRepository) queryAreasWithArgs(ctx context.Context, query string, args ...interface{}) ([]*entity.CoverageArea, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []*entity.CoverageArea
	for rows.Next() {
		var area entity.CoverageArea

		err := rows.Scan(
			&area.ID, &area.Province, &area.District, &area.Subdistrict, &area.PostalCode,
			&area.IsSelfDeliveryArea, &area.DeliveryRoute, &area.DeliveryZone, &area.PriorityOrder,
			&area.BaseDeliveryFee, &area.PerKmRate, &area.FreeDeliveryThreshold,
			&area.StandardDeliveryHours, &area.ExpressDeliveryHours, &area.SameDayAvailable,
			&area.IsActive, &area.AutoAssign, &area.MaxDailyCapacity, &area.CreatedAt, &area.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coverage area: %w", err)
		}

		areas = append(areas, &area)
	}

	return areas, nil
}

// Helper method to query coverage areas and handle common scanning logic
func (r *coverageAreaRepository) queryAreas(ctx context.Context, query string, args ...interface{}) ([]*entity.CoverageArea, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query coverage areas: %w", err)
	}
	defer rows.Close()

	var areas []*entity.CoverageArea
	for rows.Next() {
		var area entity.CoverageArea

		err := rows.Scan(
			&area.ID, &area.Province, &area.District, &area.Subdistrict, &area.PostalCode,
			&area.IsSelfDeliveryArea, &area.DeliveryRoute, &area.DeliveryZone, &area.PriorityOrder,
			&area.BaseDeliveryFee, &area.PerKmRate, &area.FreeDeliveryThreshold,
			&area.StandardDeliveryHours, &area.ExpressDeliveryHours, &area.SameDayAvailable,
			&area.IsActive, &area.AutoAssign, &area.MaxDailyCapacity, &area.CreatedAt, &area.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coverage area: %w", err)
		}

		areas = append(areas, &area)
	}

	return areas, nil
}
