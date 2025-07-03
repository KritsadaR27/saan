package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/saan-system/services/customer/internal/domain"
)

type thaiAddressRepository struct {
	db *sql.DB
}

// NewThaiAddressRepository creates a new Thai address repository
func NewThaiAddressRepository(db *sql.DB) domain.ThaiAddressRepository {
	return &thaiAddressRepository{db: db}
}

// GetByID retrieves a Thai address by ID
func (r *thaiAddressRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ThaiAddress, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code, province_code, 
			district_code, created_at, updated_at
		FROM thai_addresses 
		WHERE id = $1`

	address := &domain.ThaiAddress{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&address.ID, &address.Province, &address.District, &address.Subdistrict,
		&address.PostalCode, &address.ProvinceCode, &address.DistrictCode,
		&address.CreatedAt, &address.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrThaiAddressNotFound
		}
		return nil, fmt.Errorf("failed to get Thai address: %w", err)
	}

	return address, nil
}

// GetByPostalCode retrieves Thai addresses by postal code
func (r *thaiAddressRepository) GetByPostalCode(ctx context.Context, postalCode string) ([]domain.ThaiAddress, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code, province_code, 
			district_code, created_at, updated_at
		FROM thai_addresses 
		WHERE postal_code = $1
		ORDER BY province, district, subdistrict`

	rows, err := r.db.QueryContext(ctx, query, postalCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get Thai addresses by postal code: %w", err)
	}
	defer rows.Close()

	var addresses []domain.ThaiAddress
	for rows.Next() {
		address := domain.ThaiAddress{}
		err := rows.Scan(
			&address.ID, &address.Province, &address.District, &address.Subdistrict,
			&address.PostalCode, &address.ProvinceCode, &address.DistrictCode,
			&address.CreatedAt, &address.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan Thai address: %w", err)
		}
		addresses = append(addresses, address)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return addresses, nil
}

// SearchByProvince searches Thai addresses by province
func (r *thaiAddressRepository) SearchByProvince(ctx context.Context, province string) ([]domain.ThaiAddress, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code, province_code, 
			district_code, created_at, updated_at
		FROM thai_addresses 
		WHERE province ILIKE $1
		ORDER BY province, district, subdistrict
		LIMIT 100`

	rows, err := r.db.QueryContext(ctx, query, "%"+province+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search Thai addresses by province: %w", err)
	}
	defer rows.Close()

	var addresses []domain.ThaiAddress
	for rows.Next() {
		address := domain.ThaiAddress{}
		err := rows.Scan(
			&address.ID, &address.Province, &address.District, &address.Subdistrict,
			&address.PostalCode, &address.ProvinceCode, &address.DistrictCode,
			&address.CreatedAt, &address.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan Thai address: %w", err)
		}
		addresses = append(addresses, address)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return addresses, nil
}

// SearchByDistrict searches Thai addresses by district
func (r *thaiAddressRepository) SearchByDistrict(ctx context.Context, district string) ([]domain.ThaiAddress, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code, province_code, 
			district_code, created_at, updated_at
		FROM thai_addresses 
		WHERE district ILIKE $1
		ORDER BY province, district, subdistrict
		LIMIT 100`

	rows, err := r.db.QueryContext(ctx, query, "%"+district+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search Thai addresses by district: %w", err)
	}
	defer rows.Close()

	var addresses []domain.ThaiAddress
	for rows.Next() {
		address := domain.ThaiAddress{}
		err := rows.Scan(
			&address.ID, &address.Province, &address.District, &address.Subdistrict,
			&address.PostalCode, &address.ProvinceCode, &address.DistrictCode,
			&address.CreatedAt, &address.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan Thai address: %w", err)
		}
		addresses = append(addresses, address)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return addresses, nil
}

// SearchBySubdistrict searches Thai addresses by subdistrict
func (r *thaiAddressRepository) SearchBySubdistrict(ctx context.Context, subdistrict string) ([]domain.ThaiAddress, error) {
	query := `
		SELECT id, province, district, subdistrict, postal_code, province_code, 
			district_code, created_at, updated_at
		FROM thai_addresses 
		WHERE subdistrict ILIKE $1
		ORDER BY province, district, subdistrict
		LIMIT 100`

	rows, err := r.db.QueryContext(ctx, query, "%"+subdistrict+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search Thai addresses by subdistrict: %w", err)
	}
	defer rows.Close()

	var addresses []domain.ThaiAddress
	for rows.Next() {
		address := domain.ThaiAddress{}
		err := rows.Scan(
			&address.ID, &address.Province, &address.District, &address.Subdistrict,
			&address.PostalCode, &address.ProvinceCode, &address.DistrictCode,
			&address.CreatedAt, &address.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan Thai address: %w", err)
		}
		addresses = append(addresses, address)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return addresses, nil
}

// AutoComplete provides autocomplete suggestions for Thai addresses
func (r *thaiAddressRepository) AutoComplete(ctx context.Context, query string, limit int) ([]domain.ThaiAddress, error) {
	// Search across all fields (subdistrict, district, province)
	sqlQuery := `
		SELECT id, province, district, subdistrict, postal_code, province_code, 
			district_code, created_at, updated_at
		FROM thai_addresses 
		WHERE subdistrict ILIKE $1 
		   OR district ILIKE $1 
		   OR province ILIKE $1
		ORDER BY 
			CASE 
				WHEN subdistrict ILIKE $2 THEN 1
				WHEN district ILIKE $2 THEN 2
				WHEN province ILIKE $2 THEN 3
				ELSE 4
			END,
			province, district, subdistrict
		LIMIT $3`

	searchPattern := "%" + query + "%"
	prefixPattern := query + "%"

	rows, err := r.db.QueryContext(ctx, sqlQuery, searchPattern, prefixPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to autocomplete Thai addresses: %w", err)
	}
	defer rows.Close()

	var addresses []domain.ThaiAddress
	for rows.Next() {
		address := domain.ThaiAddress{}
		err := rows.Scan(
			&address.ID, &address.Province, &address.District, &address.Subdistrict,
			&address.PostalCode, &address.ProvinceCode, &address.DistrictCode,
			&address.CreatedAt, &address.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan Thai address: %w", err)
		}
		addresses = append(addresses, address)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return addresses, nil
}

// GetAddressSuggestions returns address suggestions based on query
func (r *thaiAddressRepository) GetAddressSuggestions(ctx context.Context, query string, limit int) ([]domain.AddressSuggestion, error) {
	// Use the existing AutoComplete functionality but format as suggestions
	addresses, err := r.AutoComplete(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get address suggestions: %w", err)
	}

	var suggestions []domain.AddressSuggestion
	for _, addr := range addresses {
		suggestion := domain.AddressSuggestion{
			ID:          addr.ID.String(),
			DisplayText: fmt.Sprintf("%s > %s > %s (%s)", 
				addr.Subdistrict, 
				addr.District, 
				addr.Province, 
				addr.PostalCode),
			Subdistrict: addr.Subdistrict,
			District:    addr.District,
			Province:    addr.Province,
			PostalCode:  addr.PostalCode,
		}

		// Determine delivery route based on province (11 provinces for self delivery)
		if r.isSelfDeliveryProvince(addr.Province) {
			suggestion.DeliveryRoute = "self_delivery"
		} else {
			suggestion.DeliveryRoute = "third_party"
		}

		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

// GetBySubdistrict returns Thai addresses by subdistrict
func (r *thaiAddressRepository) GetBySubdistrict(ctx context.Context, subdistrict string) ([]domain.ThaiAddress, error) {
	return r.SearchBySubdistrict(ctx, subdistrict)
}

// GetProvinceDeliveryInfo returns delivery route info for a province
func (r *thaiAddressRepository) GetProvinceDeliveryInfo(ctx context.Context, province string) (*domain.DeliveryRoute, error) {
	var route domain.DeliveryRoute
	
	// Determine route based on province
	if r.isSelfDeliveryProvince(province) {
		route = domain.DeliveryRoute{
			ID:          uuid.New(),
			Name:        "Self Delivery - Route A",
			Description: &[]string{"Self delivery route for central provinces"}[0],
			IsActive:    true,
		}
	} else {
		route = domain.DeliveryRoute{
			ID:          uuid.New(),
			Name:        "Third Party Delivery",
			Description: &[]string{"Third party carrier delivery"}[0],
			IsActive:    true,
		}
	}
	
	return &route, nil
}

// isSelfDeliveryProvince checks if province is in self-delivery area (11 provinces)
func (r *thaiAddressRepository) isSelfDeliveryProvince(province string) bool {
	selfDeliveryProvinces := map[string]bool{
		"กรุงเทพมหานคร":  true,
		"นนทบุรี":       true,
		"ปทุมธานี":      true,
		"สมุทรปราการ":   true,
		"สมุทรสาคร":     true,
		"นครปาธม":      true,
		"สมุทรสงคราม":   true,
		"ฉะเชิงเทรา":    true,
		"ระยอง":        true,
		"ชลบุรี":       true,
		"สระบุรี":      true,
	}
	return selfDeliveryProvinces[province]
}
