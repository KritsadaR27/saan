package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/saan-system/services/customer/internal/domain"
)

type customerRepository struct {
	db *sql.DB
}

// NewCustomerRepository creates a new customer repository
func NewCustomerRepository(db *sql.DB) domain.CustomerRepository {
	return &customerRepository{db: db}
}

// Create creates a new customer
func (r *customerRepository) Create(ctx context.Context, customer *domain.Customer) error {
	query := `
		INSERT INTO customers (id, first_name, last_name, email, phone, date_of_birth, 
			gender, tier, loyverse_id, total_spent, order_count, last_order_date, 
			delivery_route_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err := r.db.ExecContext(ctx, query,
		customer.ID, customer.FirstName, customer.LastName, customer.Email,
		customer.Phone, customer.DateOfBirth, customer.Gender, customer.Tier,
		customer.LoyverseID, customer.TotalSpent, customer.OrderCount,
		customer.LastOrderDate, customer.DeliveryRouteID, customer.IsActive,
		customer.CreatedAt, customer.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

// GetByID retrieves a customer by ID
func (r *customerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, date_of_birth, gender, 
			tier, loyverse_id, total_spent, order_count, last_order_date, 
			delivery_route_id, is_active, created_at, updated_at
		FROM customers 
		WHERE id = $1 AND is_active = true`

	customer := &domain.Customer{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&customer.ID, &customer.FirstName, &customer.LastName, &customer.Email,
		&customer.Phone, &customer.DateOfBirth, &customer.Gender, &customer.Tier,
		&customer.LoyverseID, &customer.TotalSpent, &customer.OrderCount,
		&customer.LastOrderDate, &customer.DeliveryRouteID, &customer.IsActive,
		&customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

// GetByEmail retrieves a customer by email
func (r *customerRepository) GetByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, date_of_birth, gender, 
			tier, loyverse_id, total_spent, order_count, last_order_date, 
			delivery_route_id, is_active, created_at, updated_at
		FROM customers 
		WHERE email = $1 AND is_active = true`

	customer := &domain.Customer{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&customer.ID, &customer.FirstName, &customer.LastName, &customer.Email,
		&customer.Phone, &customer.DateOfBirth, &customer.Gender, &customer.Tier,
		&customer.LoyverseID, &customer.TotalSpent, &customer.OrderCount,
		&customer.LastOrderDate, &customer.DeliveryRouteID, &customer.IsActive,
		&customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}

	return customer, nil
}

// GetByPhone retrieves a customer by phone
func (r *customerRepository) GetByPhone(ctx context.Context, phone string) (*domain.Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, date_of_birth, gender, 
			tier, loyverse_id, total_spent, order_count, last_order_date, 
			delivery_route_id, is_active, created_at, updated_at
		FROM customers 
		WHERE phone = $1 AND is_active = true`

	customer := &domain.Customer{}
	err := r.db.QueryRowContext(ctx, query, phone).Scan(
		&customer.ID, &customer.FirstName, &customer.LastName, &customer.Email,
		&customer.Phone, &customer.DateOfBirth, &customer.Gender, &customer.Tier,
		&customer.LoyverseID, &customer.TotalSpent, &customer.OrderCount,
		&customer.LastOrderDate, &customer.DeliveryRouteID, &customer.IsActive,
		&customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer by phone: %w", err)
	}

	return customer, nil
}

// GetByLoyverseID retrieves a customer by Loyverse ID
func (r *customerRepository) GetByLoyverseID(ctx context.Context, loyverseID string) (*domain.Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, date_of_birth, gender, 
			tier, loyverse_id, total_spent, order_count, last_order_date, 
			delivery_route_id, is_active, created_at, updated_at
		FROM customers 
		WHERE loyverse_id = $1 AND is_active = true`

	customer := &domain.Customer{}
	err := r.db.QueryRowContext(ctx, query, loyverseID).Scan(
		&customer.ID, &customer.FirstName, &customer.LastName, &customer.Email,
		&customer.Phone, &customer.DateOfBirth, &customer.Gender, &customer.Tier,
		&customer.LoyverseID, &customer.TotalSpent, &customer.OrderCount,
		&customer.LastOrderDate, &customer.DeliveryRouteID, &customer.IsActive,
		&customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer by loyverse ID: %w", err)
	}

	return customer, nil
}

// Update updates a customer
func (r *customerRepository) Update(ctx context.Context, customer *domain.Customer) error {
	query := `
		UPDATE customers 
		SET first_name = $2, last_name = $3, email = $4, phone = $5, 
			date_of_birth = $6, gender = $7, tier = $8, loyverse_id = $9, 
			total_spent = $10, order_count = $11, last_order_date = $12, 
			delivery_route_id = $13, is_active = $14, updated_at = $15
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		customer.ID, customer.FirstName, customer.LastName, customer.Email,
		customer.Phone, customer.DateOfBirth, customer.Gender, customer.Tier,
		customer.LoyverseID, customer.TotalSpent, customer.OrderCount,
		customer.LastOrderDate, customer.DeliveryRouteID, customer.IsActive,
		customer.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrCustomerNotFound
	}

	return nil
}

// Delete soft deletes a customer
func (r *customerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE customers SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrCustomerNotFound
	}

	return nil
}

// List retrieves customers with filtering and pagination
func (r *customerRepository) List(ctx context.Context, filter domain.CustomerFilter) ([]domain.Customer, int, error) {
	// Build WHERE clause based on filters
	whereClause := "WHERE is_active = true"
	args := []interface{}{}
	argCount := 0

	if filter.Email != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND email ILIKE $%d", argCount)
		args = append(args, "%"+*filter.Email+"%")
	}

	if filter.Phone != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND phone ILIKE $%d", argCount)
		args = append(args, "%"+*filter.Phone+"%")
	}

	if filter.Tier != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND tier = $%d", argCount)
		args = append(args, *filter.Tier)
	}

	if filter.DeliveryRouteID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND delivery_route_id = $%d", argCount)
		args = append(args, *filter.DeliveryRouteID)
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM customers %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY created_at DESC"
	if filter.SortBy != "" {
		direction := "ASC"
		if filter.SortOrder == "desc" {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf("ORDER BY %s %s", filter.SortBy, direction)
	}

	// Build LIMIT and OFFSET
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	argCount++
	limitClause := fmt.Sprintf("LIMIT $%d", argCount)
	args = append(args, limit)

	argCount++
	offsetClause := fmt.Sprintf("OFFSET $%d", argCount)
	args = append(args, offset)

	// Execute query
	query := fmt.Sprintf(`
		SELECT id, first_name, last_name, email, phone, date_of_birth, gender, 
			tier, loyverse_id, total_spent, order_count, last_order_date, 
			delivery_route_id, is_active, created_at, updated_at
		FROM customers 
		%s %s %s %s`, whereClause, orderBy, limitClause, offsetClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list customers: %w", err)
	}
	defer rows.Close()

	var customers []domain.Customer
	for rows.Next() {
		customer := domain.Customer{}
		err := rows.Scan(
			&customer.ID, &customer.FirstName, &customer.LastName, &customer.Email,
			&customer.Phone, &customer.DateOfBirth, &customer.Gender, &customer.Tier,
			&customer.LoyverseID, &customer.TotalSpent, &customer.OrderCount,
			&customer.LastOrderDate, &customer.DeliveryRouteID, &customer.IsActive,
			&customer.CreatedAt, &customer.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating rows: %w", err)
	}

	return customers, total, nil
}

// UpdateTotalSpent updates customer's total spent and tier
func (r *customerRepository) UpdateTotalSpent(ctx context.Context, customerID uuid.UUID, amount float64) error {
	// First get current customer to calculate new tier
	customer, err := r.GetByID(ctx, customerID)
	if err != nil {
		return err
	}

	// Update total spent and recalculate tier
	customer.UpdateTotalSpent(amount)

	// Update in database
	query := `
		UPDATE customers 
		SET total_spent = $2, tier = $3, order_count = $4, last_order_date = $5, updated_at = $6
		WHERE id = $1`

	_, err = r.db.ExecContext(ctx, query,
		customerID, customer.TotalSpent, customer.Tier, customer.OrderCount,
		customer.LastOrderDate, customer.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update customer total spent: %w", err)
	}

	return nil
}

// GetTopCustomers retrieves top customers by total spent
func (r *customerRepository) GetTopCustomers(ctx context.Context, limit int) ([]domain.Customer, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, date_of_birth, gender, 
			tier, loyverse_id, total_spent, order_count, last_order_date, 
			delivery_route_id, is_active, created_at, updated_at
		FROM customers 
		WHERE is_active = true
		ORDER BY total_spent DESC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top customers: %w", err)
	}
	defer rows.Close()

	var customers []domain.Customer
	for rows.Next() {
		customer := domain.Customer{}
		err := rows.Scan(
			&customer.ID, &customer.FirstName, &customer.LastName, &customer.Email,
			&customer.Phone, &customer.DateOfBirth, &customer.Gender, &customer.Tier,
			&customer.LoyverseID, &customer.TotalSpent, &customer.OrderCount,
			&customer.LastOrderDate, &customer.DeliveryRouteID, &customer.IsActive,
			&customer.CreatedAt, &customer.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return customers, nil
}
