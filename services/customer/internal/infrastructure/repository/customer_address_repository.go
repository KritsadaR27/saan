package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/saan-system/services/customer/internal/domain"
)

type customerAddressRepository struct {
	db *sql.DB
}

// NewCustomerAddressRepository creates a new customer address repository
func NewCustomerAddressRepository(db *sql.DB) domain.CustomerAddressRepository {
	return &customerAddressRepository{db: db}
}

// Create creates a new customer address
func (r *customerAddressRepository) Create(ctx context.Context, address *domain.CustomerAddress) error {
	query := `
		INSERT INTO customer_addresses (id, customer_id, type, address_line1, address_line2, 
			thai_address_id, postal_code, latitude, longitude, is_default, delivery_notes, 
			is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.db.ExecContext(ctx, query,
		address.ID, address.CustomerID, address.Type, address.AddressLine1,
		address.AddressLine2, address.ThaiAddressID, address.PostalCode,
		address.Latitude, address.Longitude, address.IsDefault,
		address.DeliveryNotes, address.IsActive, address.CreatedAt, address.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create customer address: %w", err)
	}

	return nil
}

// GetByID retrieves a customer address by ID
func (r *customerAddressRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CustomerAddress, error) {
	query := `
		SELECT id, customer_id, type, address_line1, address_line2, thai_address_id, 
			postal_code, latitude, longitude, is_default, delivery_notes, is_active, 
			created_at, updated_at
		FROM customer_addresses 
		WHERE id = $1 AND is_active = true`

	address := &domain.CustomerAddress{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&address.ID, &address.CustomerID, &address.Type, &address.AddressLine1,
		&address.AddressLine2, &address.ThaiAddressID, &address.PostalCode,
		&address.Latitude, &address.Longitude, &address.IsDefault,
		&address.DeliveryNotes, &address.IsActive, &address.CreatedAt, &address.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAddressNotFound
		}
		return nil, fmt.Errorf("failed to get customer address: %w", err)
	}

	return address, nil
}

// GetByCustomerID retrieves all addresses for a customer
func (r *customerAddressRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]domain.CustomerAddress, error) {
	query := `
		SELECT id, customer_id, type, address_line1, address_line2, thai_address_id, 
			postal_code, latitude, longitude, is_default, delivery_notes, is_active, 
			created_at, updated_at
		FROM customer_addresses 
		WHERE customer_id = $1 AND is_active = true
		ORDER BY is_default DESC, created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer addresses: %w", err)
	}
	defer rows.Close()

	var addresses []domain.CustomerAddress
	for rows.Next() {
		address := domain.CustomerAddress{}
		err := rows.Scan(
			&address.ID, &address.CustomerID, &address.Type, &address.AddressLine1,
			&address.AddressLine2, &address.ThaiAddressID, &address.PostalCode,
			&address.Latitude, &address.Longitude, &address.IsDefault,
			&address.DeliveryNotes, &address.IsActive, &address.CreatedAt, &address.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer address: %w", err)
		}
		addresses = append(addresses, address)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return addresses, nil
}

// GetDefaultAddress retrieves the default address for a customer
func (r *customerAddressRepository) GetDefaultAddress(ctx context.Context, customerID uuid.UUID) (*domain.CustomerAddress, error) {
	query := `
		SELECT id, customer_id, type, address_line1, address_line2, thai_address_id, 
			postal_code, latitude, longitude, is_default, delivery_notes, is_active, 
			created_at, updated_at
		FROM customer_addresses 
		WHERE customer_id = $1 AND is_default = true AND is_active = true`

	address := &domain.CustomerAddress{}
	err := r.db.QueryRowContext(ctx, query, customerID).Scan(
		&address.ID, &address.CustomerID, &address.Type, &address.AddressLine1,
		&address.AddressLine2, &address.ThaiAddressID, &address.PostalCode,
		&address.Latitude, &address.Longitude, &address.IsDefault,
		&address.DeliveryNotes, &address.IsActive, &address.CreatedAt, &address.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAddressNotFound
		}
		return nil, fmt.Errorf("failed to get default address: %w", err)
	}

	return address, nil
}

// Update updates a customer address
func (r *customerAddressRepository) Update(ctx context.Context, address *domain.CustomerAddress) error {
	query := `
		UPDATE customer_addresses 
		SET type = $2, address_line1 = $3, address_line2 = $4, thai_address_id = $5, 
			postal_code = $6, latitude = $7, longitude = $8, is_default = $9, 
			delivery_notes = $10, is_active = $11, updated_at = $12
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		address.ID, address.Type, address.AddressLine1, address.AddressLine2,
		address.ThaiAddressID, address.PostalCode, address.Latitude, address.Longitude,
		address.IsDefault, address.DeliveryNotes, address.IsActive, address.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update customer address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrAddressNotFound
	}

	return nil
}

// Delete soft deletes a customer address
func (r *customerAddressRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE customer_addresses SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrAddressNotFound
	}

	return nil
}

// SetAsDefault sets an address as default for a customer
func (r *customerAddressRepository) SetAsDefault(ctx context.Context, addressID uuid.UUID, customerID uuid.UUID) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// First, unset all default addresses for the customer
	unsetQuery := `UPDATE customer_addresses SET is_default = false WHERE customer_id = $1`
	_, err = tx.ExecContext(ctx, unsetQuery, customerID)
	if err != nil {
		return fmt.Errorf("failed to unset default addresses: %w", err)
	}

	// Then set the specified address as default
	setQuery := `UPDATE customer_addresses SET is_default = true WHERE id = $1 AND customer_id = $2`
	result, err := tx.ExecContext(ctx, setQuery, addressID, customerID)
	if err != nil {
		return fmt.Errorf("failed to set default address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrAddressNotFound
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
