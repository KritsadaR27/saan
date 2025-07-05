package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"customer/internal/domain/entity"
	"customer/internal/domain/repository"
)

// Repository implementations as per SERVICE_ARCHITECTURE_GUIDE.md
// All repository implementations should be in this single file

// customerRepository implements repository.CustomerRepository
type customerRepository struct {
	db *sql.DB
}

// NewCustomerRepository creates a new customer repository
func NewCustomerRepository(db *sql.DB) repository.CustomerRepository {
	return &customerRepository{db: db}
}

// Create creates a new customer
func (r *customerRepository) Create(ctx context.Context, customer *entity.Customer) error {
	query := `
		INSERT INTO customers (
			id, phone, first_name, last_name, email, date_of_birth, gender,
			customer_code, tier, points_balance, total_spent, tier_achieved_date,
			loyverse_id, loyverse_total_visits, loyverse_total_spent, loyverse_points,
			first_visit, last_visit, last_sync_at,
			line_user_id, line_display_name, digital_card_issued_at, last_card_scan,
			order_count, last_order_date, average_order_value, purchase_frequency,
			delivery_route_id, is_active, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31
		)`

	_, err := r.db.ExecContext(ctx, query,
		customer.ID, customer.Phone, customer.FirstName, customer.LastName, customer.Email,
		customer.DateOfBirth, customer.Gender, customer.CustomerCode, customer.Tier,
		customer.PointsBalance, customer.TotalSpent, customer.TierAchievedDate,
		customer.LoyverseID, customer.LoyverseTotalVisits, customer.LoyverseTotalSpent, customer.LoyversePoints,
		customer.FirstVisit, customer.LastVisit, customer.LastSyncAt,
		customer.LineUserID, customer.LineDisplayName, customer.DigitalCardIssuedAt, customer.LastCardScan,
		customer.OrderCount, customer.LastOrderDate, customer.AverageOrderValue, customer.PurchaseFrequency,
		customer.DeliveryRouteID, customer.IsActive, customer.CreatedAt, customer.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

// GetByID retrieves a customer by ID
func (r *customerRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Customer, error) {
	query := `
		SELECT id, phone, first_name, last_name, email, date_of_birth, gender,
			   customer_code, tier, points_balance, total_spent, tier_achieved_date,
			   loyverse_id, loyverse_total_visits, loyverse_total_spent, loyverse_points,
			   first_visit, last_visit, last_sync_at,
			   line_user_id, line_display_name, digital_card_issued_at, last_card_scan,
			   order_count, last_order_date, average_order_value, purchase_frequency,
			   delivery_route_id, is_active, created_at, updated_at
		FROM customers 
		WHERE id = $1 AND is_active = true`

	customer := &entity.Customer{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&customer.ID, &customer.Phone, &customer.FirstName, &customer.LastName, &customer.Email,
		&customer.DateOfBirth, &customer.Gender, &customer.CustomerCode, &customer.Tier,
		&customer.PointsBalance, &customer.TotalSpent, &customer.TierAchievedDate,
		&customer.LoyverseID, &customer.LoyverseTotalVisits, &customer.LoyverseTotalSpent, &customer.LoyversePoints,
		&customer.FirstVisit, &customer.LastVisit, &customer.LastSyncAt,
		&customer.LineUserID, &customer.LineDisplayName, &customer.DigitalCardIssuedAt, &customer.LastCardScan,
		&customer.OrderCount, &customer.LastOrderDate, &customer.AverageOrderValue, &customer.PurchaseFrequency,
		&customer.DeliveryRouteID, &customer.IsActive, &customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

// GetByEmail retrieves a customer by email
func (r *customerRepository) GetByEmail(ctx context.Context, email string) (*entity.Customer, error) {
	query := `
		SELECT id, phone, first_name, last_name, email, date_of_birth, gender,
			   customer_code, tier, points_balance, total_spent, tier_achieved_date,
			   loyverse_id, loyverse_total_visits, loyverse_total_spent, loyverse_points,
			   first_visit, last_visit, last_sync_at,
			   line_user_id, line_display_name, digital_card_issued_at, last_card_scan,
			   order_count, last_order_date, average_order_value, purchase_frequency,
			   delivery_route_id, is_active, created_at, updated_at
		FROM customers 
		WHERE email = $1 AND is_active = true`

	customer := &entity.Customer{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&customer.ID, &customer.Phone, &customer.FirstName, &customer.LastName, &customer.Email,
		&customer.DateOfBirth, &customer.Gender, &customer.CustomerCode, &customer.Tier,
		&customer.PointsBalance, &customer.TotalSpent, &customer.TierAchievedDate,
		&customer.LoyverseID, &customer.LoyverseTotalVisits, &customer.LoyverseTotalSpent, &customer.LoyversePoints,
		&customer.FirstVisit, &customer.LastVisit, &customer.LastSyncAt,
		&customer.LineUserID, &customer.LineDisplayName, &customer.DigitalCardIssuedAt, &customer.LastCardScan,
		&customer.OrderCount, &customer.LastOrderDate, &customer.AverageOrderValue, &customer.PurchaseFrequency,
		&customer.DeliveryRouteID, &customer.IsActive, &customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}

	return customer, nil
}

// GetByPhone retrieves a customer by phone
func (r *customerRepository) GetByPhone(ctx context.Context, phone string) (*entity.Customer, error) {
	query := `
		SELECT id, phone, first_name, last_name, email, date_of_birth, gender,
			   customer_code, tier, points_balance, total_spent, tier_achieved_date,
			   loyverse_id, loyverse_total_visits, loyverse_total_spent, loyverse_points,
			   first_visit, last_visit, last_sync_at,
			   line_user_id, line_display_name, digital_card_issued_at, last_card_scan,
			   order_count, last_order_date, average_order_value, purchase_frequency,
			   delivery_route_id, is_active, created_at, updated_at
		FROM customers 
		WHERE phone = $1 AND is_active = true`

	customer := &entity.Customer{}
	err := r.db.QueryRowContext(ctx, query, phone).Scan(
		&customer.ID, &customer.Phone, &customer.FirstName, &customer.LastName, &customer.Email,
		&customer.DateOfBirth, &customer.Gender, &customer.CustomerCode, &customer.Tier,
		&customer.PointsBalance, &customer.TotalSpent, &customer.TierAchievedDate,
		&customer.LoyverseID, &customer.LoyverseTotalVisits, &customer.LoyverseTotalSpent, &customer.LoyversePoints,
		&customer.FirstVisit, &customer.LastVisit, &customer.LastSyncAt,
		&customer.LineUserID, &customer.LineDisplayName, &customer.DigitalCardIssuedAt, &customer.LastCardScan,
		&customer.OrderCount, &customer.LastOrderDate, &customer.AverageOrderValue, &customer.PurchaseFrequency,
		&customer.DeliveryRouteID, &customer.IsActive, &customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer by phone: %w", err)
	}

	return customer, nil
}

// GetByLoyverseID retrieves a customer by Loyverse ID
func (r *customerRepository) GetByLoyverseID(ctx context.Context, loyverseID string) (*entity.Customer, error) {
	query := `
		SELECT id, phone, first_name, last_name, email, date_of_birth, gender,
			   customer_code, tier, points_balance, total_spent, tier_achieved_date,
			   loyverse_id, loyverse_total_visits, loyverse_total_spent, loyverse_points,
			   first_visit, last_visit, last_sync_at,
			   line_user_id, line_display_name, digital_card_issued_at, last_card_scan,
			   order_count, last_order_date, average_order_value, purchase_frequency,
			   delivery_route_id, is_active, created_at, updated_at
		FROM customers 
		WHERE loyverse_id = $1 AND is_active = true`

	customer := &entity.Customer{}
	err := r.db.QueryRowContext(ctx, query, loyverseID).Scan(
		&customer.ID, &customer.Phone, &customer.FirstName, &customer.LastName, &customer.Email,
		&customer.DateOfBirth, &customer.Gender, &customer.CustomerCode, &customer.Tier,
		&customer.PointsBalance, &customer.TotalSpent, &customer.TierAchievedDate,
		&customer.LoyverseID, &customer.LoyverseTotalVisits, &customer.LoyverseTotalSpent, &customer.LoyversePoints,
		&customer.FirstVisit, &customer.LastVisit, &customer.LastSyncAt,
		&customer.LineUserID, &customer.LineDisplayName, &customer.DigitalCardIssuedAt, &customer.LastCardScan,
		&customer.OrderCount, &customer.LastOrderDate, &customer.AverageOrderValue, &customer.PurchaseFrequency,
		&customer.DeliveryRouteID, &customer.IsActive, &customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer by loyverse ID: %w", err)
	}

	return customer, nil
}

// GetByLineUserID retrieves a customer by LINE user ID
func (r *customerRepository) GetByLineUserID(ctx context.Context, lineUserID string) (*entity.Customer, error) {
	query := `
		SELECT id, phone, first_name, last_name, email, date_of_birth, gender,
			customer_code, tier, points_balance, total_spent, tier_achieved_date,
			loyverse_id, loyverse_total_visits, loyverse_total_spent, loyverse_points,
			first_visit, last_visit, last_sync_at,
			line_user_id, line_display_name, digital_card_issued_at, last_card_scan,
			order_count, last_order_date, average_order_value, purchase_frequency,
			delivery_route_id, is_active, created_at, updated_at
		FROM customers 
		WHERE line_user_id = $1 AND is_active = true`

	customer := &entity.Customer{}
	err := r.db.QueryRowContext(ctx, query, lineUserID).Scan(
		&customer.ID, &customer.Phone, &customer.FirstName, &customer.LastName, &customer.Email,
		&customer.DateOfBirth, &customer.Gender, &customer.CustomerCode, &customer.Tier,
		&customer.PointsBalance, &customer.TotalSpent, &customer.TierAchievedDate,
		&customer.LoyverseID, &customer.LoyverseTotalVisits, &customer.LoyverseTotalSpent, &customer.LoyversePoints,
		&customer.FirstVisit, &customer.LastVisit, &customer.LastSyncAt,
		&customer.LineUserID, &customer.LineDisplayName, &customer.DigitalCardIssuedAt, &customer.LastCardScan,
		&customer.OrderCount, &customer.LastOrderDate, &customer.AverageOrderValue, &customer.PurchaseFrequency,
		&customer.DeliveryRouteID, &customer.IsActive, &customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer by LINE user ID: %w", err)
	}

	return customer, nil
}

// Update updates a customer
func (r *customerRepository) Update(ctx context.Context, customer *entity.Customer) error {
	query := `
		UPDATE customers SET 
			phone = $2, first_name = $3, last_name = $4, email = $5, date_of_birth = $6, gender = $7,
			customer_code = $8, tier = $9, points_balance = $10, total_spent = $11, tier_achieved_date = $12,
			loyverse_id = $13, loyverse_total_visits = $14, loyverse_total_spent = $15, loyverse_points = $16,
			first_visit = $17, last_visit = $18, last_sync_at = $19,
			line_user_id = $20, line_display_name = $21, digital_card_issued_at = $22, last_card_scan = $23,
			order_count = $24, last_order_date = $25, average_order_value = $26, purchase_frequency = $27,
			delivery_route_id = $28, is_active = $29, updated_at = $30
		WHERE id = $1`

	customer.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		customer.ID, customer.Phone, customer.FirstName, customer.LastName, customer.Email,
		customer.DateOfBirth, customer.Gender, customer.CustomerCode, customer.Tier,
		customer.PointsBalance, customer.TotalSpent, customer.TierAchievedDate,
		customer.LoyverseID, customer.LoyverseTotalVisits, customer.LoyverseTotalSpent, customer.LoyversePoints,
		customer.FirstVisit, customer.LastVisit, customer.LastSyncAt,
		customer.LineUserID, customer.LineDisplayName, customer.DigitalCardIssuedAt, customer.LastCardScan,
		customer.OrderCount, customer.LastOrderDate, customer.AverageOrderValue, customer.PurchaseFrequency,
		customer.DeliveryRouteID, customer.IsActive, customer.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

// Delete soft deletes a customer
func (r *customerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE customers SET is_active = false, updated_at = $2 WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	return nil
}

// List retrieves customers with filtering
func (r *customerRepository) List(ctx context.Context, filter repository.CustomerFilter) ([]entity.Customer, int, error) {
	// Build dynamic query based on filters
	baseQuery := `
		SELECT id, phone, first_name, last_name, email, date_of_birth, gender,
			   customer_code, tier, points_balance, total_spent, tier_achieved_date,
			   loyverse_id, loyverse_total_visits, loyverse_total_spent, loyverse_points,
			   first_visit, last_visit, last_sync_at,
			   line_user_id, line_display_name, digital_card_issued_at, last_card_scan,
			   order_count, last_order_date, average_order_value, purchase_frequency,
			   delivery_route_id, is_active, created_at, updated_at
		FROM customers 
		WHERE is_active = true`

	countQuery := `SELECT COUNT(*) FROM customers WHERE is_active = true`

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Apply filters
	if filter.Email != nil {
		conditions = append(conditions, fmt.Sprintf(" AND email = $%d", argIndex))
		args = append(args, *filter.Email)
		argIndex++
	}

	if filter.Phone != nil {
		conditions = append(conditions, fmt.Sprintf(" AND phone = $%d", argIndex))
		args = append(args, *filter.Phone)
		argIndex++
	}

	if filter.Tier != nil {
		conditions = append(conditions, fmt.Sprintf(" AND tier = $%d", argIndex))
		args = append(args, *filter.Tier)
		argIndex++
	}

	if filter.DeliveryRouteID != nil {
		conditions = append(conditions, fmt.Sprintf(" AND delivery_route_id = $%d", argIndex))
		args = append(args, *filter.DeliveryRouteID)
		argIndex++
	}

	// Add conditions to queries
	for _, condition := range conditions {
		baseQuery += condition
		countQuery += condition
	}

	// Add sorting
	if filter.SortBy != "" {
		order := "ASC"
		if filter.SortOrder == "desc" {
			order = "DESC"
		}
		baseQuery += fmt.Sprintf(" ORDER BY %s %s", filter.SortBy, order)
	} else {
		baseQuery += " ORDER BY created_at DESC"
	}

	// Add pagination
	if filter.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		baseQuery += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
		argIndex++
	}

	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get customer count: %w", err)
	}

	// Get customers
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list customers: %w", err)
	}
	defer rows.Close()

	var customers []entity.Customer
	for rows.Next() {
		customer := entity.Customer{}
		err := rows.Scan(
			&customer.ID, &customer.Phone, &customer.FirstName, &customer.LastName, &customer.Email,
			&customer.DateOfBirth, &customer.Gender, &customer.CustomerCode, &customer.Tier,
			&customer.PointsBalance, &customer.TotalSpent, &customer.TierAchievedDate,
			&customer.LoyverseID, &customer.LoyverseTotalVisits, &customer.LoyverseTotalSpent, &customer.LoyversePoints,
			&customer.FirstVisit, &customer.LastVisit, &customer.LastSyncAt,
			&customer.LineUserID, &customer.LineDisplayName, &customer.DigitalCardIssuedAt, &customer.LastCardScan,
			&customer.OrderCount, &customer.LastOrderDate, &customer.AverageOrderValue, &customer.PurchaseFrequency,
			&customer.DeliveryRouteID, &customer.IsActive, &customer.CreatedAt, &customer.UpdatedAt)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan customer: %w", err)
		}

		customers = append(customers, customer)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate customers: %w", err)
	}

	return customers, total, nil
}

// UpdateTotalSpent updates customer's total spent amount and recalculates tier
func (r *customerRepository) UpdateTotalSpent(ctx context.Context, customerID uuid.UUID, amount float64) error {
	query := `
		UPDATE customers 
		SET total_spent = total_spent + $2,
			updated_at = $3
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, customerID, amount, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update total spent: %w", err)
	}

	return nil
}

// GetTopCustomers retrieves top customers by total spent
func (r *customerRepository) GetTopCustomers(ctx context.Context, limit int) ([]entity.Customer, error) {
	query := `
		SELECT id, phone, first_name, last_name, email, date_of_birth, gender,
			   customer_code, tier, points_balance, total_spent, tier_achieved_date,
			   loyverse_id, loyverse_total_visits, loyverse_total_spent, loyverse_points,
			   first_visit, last_visit, last_sync_at,
			   line_user_id, line_display_name, digital_card_issued_at, last_card_scan,
			   order_count, last_order_date, average_order_value, purchase_frequency,
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

	var customers []entity.Customer
	for rows.Next() {
		customer := entity.Customer{}
		err := rows.Scan(
			&customer.ID, &customer.Phone, &customer.FirstName, &customer.LastName, &customer.Email,
			&customer.DateOfBirth, &customer.Gender, &customer.CustomerCode, &customer.Tier,
			&customer.PointsBalance, &customer.TotalSpent, &customer.TierAchievedDate,
			&customer.LoyverseID, &customer.LoyverseTotalVisits, &customer.LoyverseTotalSpent, &customer.LoyversePoints,
			&customer.FirstVisit, &customer.LastVisit, &customer.LastSyncAt,
			&customer.LineUserID, &customer.LineDisplayName, &customer.DigitalCardIssuedAt, &customer.LastCardScan,
			&customer.OrderCount, &customer.LastOrderDate, &customer.AverageOrderValue, &customer.PurchaseFrequency,
			&customer.DeliveryRouteID, &customer.IsActive, &customer.CreatedAt, &customer.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan top customer: %w", err)
		}

		customers = append(customers, customer)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate top customers: %w", err)
	}

	return customers, nil
}

// customerAddressRepository implements repository.CustomerAddressRepository
type customerAddressRepository struct {
	db *sql.DB
}

// NewCustomerAddressRepository creates a new customer address repository
func NewCustomerAddressRepository(db *sql.DB) repository.CustomerAddressRepository {
	return &customerAddressRepository{db: db}
}

// Create creates a new customer address
func (r *customerAddressRepository) Create(ctx context.Context, address *entity.CustomerAddress) error {
	query := `
		INSERT INTO customer_addresses (
			id, customer_id, address_line1, address_line2, sub_district, district, 
			province, postal_code, type, is_default, 
			delivery_notes, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.ExecContext(ctx, query,
		address.ID, address.CustomerID, address.AddressLine1, address.AddressLine2,
		address.SubDistrict, address.District, address.Province, address.PostalCode,
		address.Type, address.IsDefault, 
		address.DeliveryNotes, address.CreatedAt, address.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create customer address: %w", err)
	}

	return nil
}

// GetByID retrieves a customer address by ID
func (r *customerAddressRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.CustomerAddress, error) {
	query := `
		SELECT id, customer_id, address_line1, address_line2, sub_district, district,
			   province, postal_code, type, is_default,
			   delivery_notes, created_at, updated_at
		FROM customer_addresses 
		WHERE id = $1`

	address := &entity.CustomerAddress{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&address.ID, &address.CustomerID, &address.AddressLine1, &address.AddressLine2,
		&address.SubDistrict, &address.District, &address.Province, &address.PostalCode,
		&address.Type, &address.IsDefault,
		&address.DeliveryNotes, &address.CreatedAt, &address.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrAddressNotFound
		}
		return nil, fmt.Errorf("failed to get customer address: %w", err)
	}

	return address, nil
}

// GetByCustomerID retrieves all addresses for a customer
func (r *customerAddressRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]entity.CustomerAddress, error) {
	query := `
		SELECT id, customer_id, address_line1, address_line2, sub_district, district,
			   province, postal_code, type, is_default,
			   delivery_notes, created_at, updated_at
		FROM customer_addresses 
		WHERE customer_id = $1
		ORDER BY is_default DESC, created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer addresses: %w", err)
	}
	defer rows.Close()

	var addresses []entity.CustomerAddress
	for rows.Next() {
		address := entity.CustomerAddress{}
		err := rows.Scan(
			&address.ID, &address.CustomerID, &address.AddressLine1, &address.AddressLine2,
			&address.SubDistrict, &address.District, &address.Province, &address.PostalCode,
			&address.Type, &address.IsDefault,
			&address.DeliveryNotes, &address.CreatedAt, &address.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan customer address: %w", err)
		}

		addresses = append(addresses, address)
	}

	return addresses, nil
}

// GetDefaultAddress retrieves the default address for a customer
func (r *customerAddressRepository) GetDefaultAddress(ctx context.Context, customerID uuid.UUID) (*entity.CustomerAddress, error) {
	query := `
		SELECT id, customer_id, address_line1, address_line2, sub_district, district,
			   province, postal_code, type, is_default,
			   delivery_notes, created_at, updated_at
		FROM customer_addresses 
		WHERE customer_id = $1 AND is_default = true`

	address := &entity.CustomerAddress{}
	err := r.db.QueryRowContext(ctx, query, customerID).Scan(
		&address.ID, &address.CustomerID, &address.AddressLine1, &address.AddressLine2,
		&address.SubDistrict, &address.District, &address.Province, &address.PostalCode,
		&address.Type, &address.IsDefault,
		&address.DeliveryNotes, &address.CreatedAt, &address.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrAddressNotFound
		}
		return nil, fmt.Errorf("failed to get default address: %w", err)
	}

	return address, nil
}

// Update updates a customer address
func (r *customerAddressRepository) Update(ctx context.Context, address *entity.CustomerAddress) error {
	query := `
		UPDATE customer_addresses SET
			address_line1 = $2, address_line2 = $3, sub_district = $4, district = $5,
			province = $6, postal_code = $7, type = $8,
			is_default = $9, delivery_notes = $10, updated_at = $11
		WHERE id = $1`

	address.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		address.ID, address.AddressLine1, address.AddressLine2, address.SubDistrict,
		address.District, address.Province, address.PostalCode, address.Type,
		address.IsDefault, address.DeliveryNotes, address.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update customer address: %w", err)
	}

	return nil
}

// Delete deletes a customer address
func (r *customerAddressRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM customer_addresses WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer address: %w", err)
	}

	return nil
}

// SetAsDefault sets an address as default for a customer
func (r *customerAddressRepository) SetAsDefault(ctx context.Context, addressID uuid.UUID, customerID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Unset all default addresses for this customer
	_, err = tx.ExecContext(ctx, 
		`UPDATE customer_addresses SET is_default = false, updated_at = $1 WHERE customer_id = $2`,
		time.Now(), customerID)
	if err != nil {
		return fmt.Errorf("failed to unset default addresses: %w", err)
	}

	// Set new default address
	_, err = tx.ExecContext(ctx,
		`UPDATE customer_addresses SET is_default = true, updated_at = $1 WHERE id = $2 AND customer_id = $3`,
		time.Now(), addressID, customerID)
	if err != nil {
		return fmt.Errorf("failed to set default address: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// vipTierBenefitsRepository implements repository.VIPTierBenefitsRepository
type vipTierBenefitsRepository struct {
	db *sql.DB
}

// NewVIPTierBenefitsRepository creates a new VIP tier benefits repository
func NewVIPTierBenefitsRepository(db *sql.DB) repository.VIPTierBenefitsRepository {
	return &vipTierBenefitsRepository{db: db}
}

// GetByTier retrieves VIP benefits for a specific tier
func (r *vipTierBenefitsRepository) GetByTier(ctx context.Context, tier entity.CustomerTier) (*entity.VIPTierBenefits, error) {
	query := `
		SELECT tier, tier_name, tier_icon, min_spent, discount_percentage, 
			   points_multiplier, free_shipping_threshold, special_offers,
			   priority_support, early_access, birthday_bonus, referral_bonus,
			   created_at, updated_at
		FROM vip_tier_benefits 
		WHERE tier = $1`

	benefits := &entity.VIPTierBenefits{}
	err := r.db.QueryRowContext(ctx, query, tier).Scan(
		&benefits.Tier, &benefits.TierName, &benefits.TierIcon, &benefits.MinSpent,
		&benefits.DiscountPercentage, &benefits.PointsMultiplier, &benefits.FreeShippingThreshold,
		&benefits.SpecialOffers, &benefits.PrioritySupport, &benefits.EarlyAccess,
		&benefits.BirthdayBonus, &benefits.ReferralBonus, &benefits.CreatedAt, &benefits.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrVIPTierNotFound
		}
		return nil, fmt.Errorf("failed to get VIP tier benefits: %w", err)
	}

	return benefits, nil
}

// GetAll retrieves all VIP tier benefits
func (r *vipTierBenefitsRepository) GetAll(ctx context.Context) ([]entity.VIPTierBenefits, error) {
	query := `
		SELECT tier, tier_name, tier_icon, min_spent, discount_percentage,
			   points_multiplier, free_shipping_threshold, special_offers,
			   priority_support, early_access, birthday_bonus, referral_bonus,
			   created_at, updated_at
		FROM vip_tier_benefits 
		ORDER BY tier ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all VIP tier benefits: %w", err)
	}
	defer rows.Close()

	var benefitsList []entity.VIPTierBenefits
	for rows.Next() {
		benefits := entity.VIPTierBenefits{}
		err := rows.Scan(
			&benefits.Tier, &benefits.TierName, &benefits.TierIcon, &benefits.MinSpent,
			&benefits.DiscountPercentage, &benefits.PointsMultiplier, &benefits.FreeShippingThreshold,
			&benefits.SpecialOffers, &benefits.PrioritySupport, &benefits.EarlyAccess,
			&benefits.BirthdayBonus, &benefits.ReferralBonus, &benefits.CreatedAt, &benefits.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan VIP tier benefits: %w", err)
		}

		benefitsList = append(benefitsList, benefits)
	}

	return benefitsList, nil
}

// Update updates VIP tier benefits
func (r *vipTierBenefitsRepository) Update(ctx context.Context, benefits *entity.VIPTierBenefits) error {
	query := `
		UPDATE vip_tier_benefits SET
			tier_name = $2, tier_icon = $3, min_spent = $4, discount_percentage = $5,
			points_multiplier = $6, free_shipping_threshold = $7, special_offers = $8,
			priority_support = $9, early_access = $10, birthday_bonus = $11,
			referral_bonus = $12, updated_at = $13
		WHERE tier = $1`

	benefits.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		benefits.Tier, benefits.TierName, benefits.TierIcon, benefits.MinSpent,
		benefits.DiscountPercentage, benefits.PointsMultiplier, benefits.FreeShippingThreshold,
		benefits.SpecialOffers, benefits.PrioritySupport, benefits.EarlyAccess,
		benefits.BirthdayBonus, benefits.ReferralBonus, benefits.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update VIP tier benefits: %w", err)
	}

	return nil
}

// customerPointsRepository implements repository.CustomerPointsRepository
type customerPointsRepository struct {
	db *sql.DB
}

// NewCustomerPointsRepository creates a new customer points repository
func NewCustomerPointsRepository(db *sql.DB) repository.CustomerPointsRepository {
	return &customerPointsRepository{db: db}
}

// CreateTransaction creates a new points transaction
func (r *customerPointsRepository) CreateTransaction(ctx context.Context, transaction *entity.CustomerPointsTransaction) error {
	query := `
		INSERT INTO customer_points_transactions (
			id, customer_id, transaction_id, type, points, balance, source, 
			description, reference_id, reference_type, expiry_date, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.db.ExecContext(ctx, query,
		transaction.ID, transaction.CustomerID, transaction.TransactionID, transaction.Type,
		transaction.Points, transaction.Balance, transaction.Source, transaction.Description,
		transaction.ReferenceID, transaction.ReferenceType, transaction.ExpiryDate, transaction.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create points transaction: %w", err)
	}

	return nil
}

// GetTransactionsByCustomer retrieves points transactions for a customer
func (r *customerPointsRepository) GetTransactionsByCustomer(ctx context.Context, customerID uuid.UUID, limit int, offset int) ([]entity.CustomerPointsTransaction, error) {
	query := `
		SELECT id, customer_id, transaction_id, type, points, balance, source,
			   description, reference_id, reference_type, expiry_date, created_at
		FROM customer_points_transactions 
		WHERE customer_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, customerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer points transactions: %w", err)
	}
	defer rows.Close()

	var transactions []entity.CustomerPointsTransaction
	for rows.Next() {
		transaction := entity.CustomerPointsTransaction{}
		err := rows.Scan(
			&transaction.ID, &transaction.CustomerID, &transaction.TransactionID, &transaction.Type,
			&transaction.Points, &transaction.Balance, &transaction.Source, &transaction.Description,
			&transaction.ReferenceID, &transaction.ReferenceType, &transaction.ExpiryDate, &transaction.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan points transaction: %w", err)
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// GetPointsBalance retrieves current points balance for a customer
func (r *customerPointsRepository) GetPointsBalance(ctx context.Context, customerID uuid.UUID) (int, error) {
	query := `SELECT points_balance FROM customers WHERE id = $1`

	var balance int
	err := r.db.QueryRowContext(ctx, query, customerID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, entity.ErrCustomerNotFound
		}
		return 0, fmt.Errorf("failed to get points balance: %w", err)
	}

	return balance, nil
}

// EarnPoints creates an earn points transaction
func (r *customerPointsRepository) EarnPoints(ctx context.Context, customerID uuid.UUID, points int, source, description string, referenceID *uuid.UUID, referenceType *string) error {
	transaction := &entity.CustomerPointsTransaction{
		ID:            uuid.New(),
		CustomerID:    customerID,
		TransactionID: uuid.New(),
		Type:          entity.PointsEarned,
		Points:        points,
		Source:        source,
		Description:   description,
		ReferenceID:   referenceID,
		ReferenceType: referenceType,
		CreatedAt:     time.Now(),
	}

	return r.CreateTransaction(ctx, transaction)
}

// RedeemPoints creates a redeem points transaction
func (r *customerPointsRepository) RedeemPoints(ctx context.Context, customerID uuid.UUID, points int, source, description string, referenceID *uuid.UUID, referenceType *string) error {
	transaction := &entity.CustomerPointsTransaction{
		ID:            uuid.New(),
		CustomerID:    customerID,
		TransactionID: uuid.New(),
		Type:          entity.PointsRedeemed,
		Points:        -points, // Negative for redemption
		Source:        source,
		Description:   description,
		ReferenceID:   referenceID,
		ReferenceType: referenceType,
		CreatedAt:     time.Now(),
	}

	return r.CreateTransaction(ctx, transaction)
}

// ExpirePoints creates an expire points transaction
func (r *customerPointsRepository) ExpirePoints(ctx context.Context, customerID uuid.UUID, points int, description string) error {
	transaction := &entity.CustomerPointsTransaction{
		ID:            uuid.New(),
		CustomerID:    customerID,
		TransactionID: uuid.New(),
		Type:          entity.PointsExpired,
		Points:        -points, // Negative for expiration
		Source:        "system",
		Description:   description,
		CreatedAt:     time.Now(),
	}

	return r.CreateTransaction(ctx, transaction)
}

// thaiAddressRepository implements repository.ThaiAddressRepository
type thaiAddressRepository struct {
	db *sql.DB
}

// NewThaiAddressRepository creates a new Thai address repository
func NewThaiAddressRepository(db *sql.DB) repository.ThaiAddressRepository {
	return &thaiAddressRepository{db: db}
}

// GetAddressSuggestions retrieves address suggestions based on query
func (r *thaiAddressRepository) GetAddressSuggestions(ctx context.Context, query string, limit int) ([]entity.AddressSuggestion, error) {
	sqlQuery := `
		SELECT id, province, district, sub_district, postal_code
		FROM thai_addresses 
		WHERE LOWER(province) LIKE LOWER($1) 
		   OR LOWER(district) LIKE LOWER($1)
		   OR LOWER(sub_district) LIKE LOWER($1)
		   OR postal_code LIKE $1
		ORDER BY 
			CASE 
				WHEN LOWER(province) = LOWER($2) THEN 1
				WHEN LOWER(district) = LOWER($2) THEN 2
				WHEN LOWER(sub_district) = LOWER($2) THEN 3
				ELSE 4
			END
		LIMIT $3`

	searchPattern := "%" + query + "%"
	
	rows, err := r.db.QueryContext(ctx, sqlQuery, searchPattern, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get address suggestions: %w", err)
	}
	defer rows.Close()

	var suggestions []entity.AddressSuggestion
	for rows.Next() {
		var suggestion entity.AddressSuggestion
		err := rows.Scan(
			&suggestion.ID, &suggestion.Province, &suggestion.District,
			&suggestion.Subdistrict, &suggestion.PostalCode)

		if err != nil {
			return nil, fmt.Errorf("failed to scan address suggestion: %w", err)
		}

		// Build full address string
		suggestion.FullAddress = fmt.Sprintf("%s, %s, %s %s", 
			suggestion.Subdistrict, suggestion.District, suggestion.Province, suggestion.PostalCode)

		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

// GetBySubdistrict retrieves Thai addresses by subdistrict
func (r *thaiAddressRepository) GetBySubdistrict(ctx context.Context, subdistrict string) ([]entity.ThaiAddress, error) {
	query := `
		SELECT id, province, district, sub_district, postal_code
		FROM thai_addresses 
		WHERE LOWER(sub_district) = LOWER($1)`

	rows, err := r.db.QueryContext(ctx, query, subdistrict)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses by subdistrict: %w", err)
	}
	defer rows.Close()

	var addresses []entity.ThaiAddress
	for rows.Next() {
		address := entity.ThaiAddress{}
		err := rows.Scan(
			&address.ID, &address.Province, &address.District,
			&address.Subdistrict, &address.PostalCode)

		if err != nil {
			return nil, fmt.Errorf("failed to scan Thai address: %w", err)
		}

		addresses = append(addresses, address)
	}

	return addresses, nil
}

// GetProvinceDeliveryInfo retrieves delivery route info for a province  
func (r *thaiAddressRepository) GetProvinceDeliveryInfo(ctx context.Context, province string) (*entity.DeliveryRoute, error) {
	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM delivery_routes 
		WHERE is_active = true
		LIMIT 1`

	route := &entity.DeliveryRoute{}
	err := r.db.QueryRowContext(ctx, query, province).Scan(
		&route.ID, &route.Name, &route.Description, 
		&route.IsActive, &route.CreatedAt, &route.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrDeliveryRouteNotFound
		}
		return nil, fmt.Errorf("failed to get delivery route: %w", err)
	}

	return route, nil
}

// AutoComplete provides address autocomplete functionality
func (r *thaiAddressRepository) AutoComplete(ctx context.Context, query string, limit int) ([]entity.ThaiAddress, error) {
	sqlQuery := `
		SELECT id, province, district, sub_district, postal_code
		FROM thai_addresses 
		WHERE LOWER(province) LIKE LOWER($1) 
		   OR LOWER(district) LIKE LOWER($1)
		   OR LOWER(sub_district) LIKE LOWER($1)
		LIMIT $2`

	searchPattern := "%" + query + "%"
	
	rows, err := r.db.QueryContext(ctx, sqlQuery, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to autocomplete addresses: %w", err)
	}
	defer rows.Close()

	var addresses []entity.ThaiAddress
	for rows.Next() {
		address := entity.ThaiAddress{}
		err := rows.Scan(
			&address.ID, &address.Province, &address.District,
			&address.Subdistrict, &address.PostalCode)

		if err != nil {
			return nil, fmt.Errorf("failed to scan Thai address: %w", err)
		}

		addresses = append(addresses, address)
	}

	return addresses, nil
}

// GetByPostalCode retrieves addresses by postal code
func (r *thaiAddressRepository) GetByPostalCode(ctx context.Context, postalCode string) ([]entity.ThaiAddress, error) {
	query := `
		SELECT id, province, district, sub_district, postal_code
		FROM thai_addresses 
		WHERE postal_code = $1`

	rows, err := r.db.QueryContext(ctx, query, postalCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses by postal code: %w", err)
	}
	defer rows.Close()

	var addresses []entity.ThaiAddress
	for rows.Next() {
		address := entity.ThaiAddress{}
		err := rows.Scan(
			&address.ID, &address.Province, &address.District,
			&address.Subdistrict, &address.PostalCode)

		if err != nil {
			return nil, fmt.Errorf("failed to scan Thai address: %w", err)
		}

		addresses = append(addresses, address)
	}

	return addresses, nil
}

// SearchByProvince retrieves addresses by province
func (r *thaiAddressRepository) SearchByProvince(ctx context.Context, province string) ([]entity.ThaiAddress, error) {
	query := `
		SELECT id, province, district, sub_district, postal_code
		FROM thai_addresses 
		WHERE LOWER(province) = LOWER($1)`

	rows, err := r.db.QueryContext(ctx, query, province)
	if err != nil {
		return nil, fmt.Errorf("failed to search addresses by province: %w", err)
	}
	defer rows.Close()

	var addresses []entity.ThaiAddress
	for rows.Next() {
		address := entity.ThaiAddress{}
		err := rows.Scan(
			&address.ID, &address.Province, &address.District,
			&address.Subdistrict, &address.PostalCode)

		if err != nil {
			return nil, fmt.Errorf("failed to scan Thai address: %w", err)
		}

		addresses = append(addresses, address)
	}

	return addresses, nil
}

// SearchByDistrict retrieves addresses by district
func (r *thaiAddressRepository) SearchByDistrict(ctx context.Context, district string) ([]entity.ThaiAddress, error) {
	query := `
		SELECT id, province, district, sub_district, postal_code
		FROM thai_addresses 
		WHERE LOWER(district) = LOWER($1)`

	rows, err := r.db.QueryContext(ctx, query, district)
	if err != nil {
		return nil, fmt.Errorf("failed to search addresses by district: %w", err)
	}
	defer rows.Close()

	var addresses []entity.ThaiAddress
	for rows.Next() {
		address := entity.ThaiAddress{}
		err := rows.Scan(
			&address.ID, &address.Province, &address.District,
			&address.Subdistrict, &address.PostalCode)

		if err != nil {
			return nil, fmt.Errorf("failed to scan Thai address: %w", err)
		}

		addresses = append(addresses, address)
	}

	return addresses, nil
}

// SearchBySubdistrict retrieves addresses by subdistrict
func (r *thaiAddressRepository) SearchBySubdistrict(ctx context.Context, subdistrict string) ([]entity.ThaiAddress, error) {
	return r.GetBySubdistrict(ctx, subdistrict)
}

// deliveryRouteRepository implements repository.DeliveryRouteRepository
type deliveryRouteRepository struct {
	db *sql.DB
}

// NewDeliveryRouteRepository creates a new delivery route repository
func NewDeliveryRouteRepository(db *sql.DB) repository.DeliveryRouteRepository {
	return &deliveryRouteRepository{db: db}
}

// Create creates a new delivery route
func (r *deliveryRouteRepository) Create(ctx context.Context, route *entity.DeliveryRoute) error {
	query := `
		INSERT INTO delivery_routes (
			id, name, description, is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		route.ID, route.Name, route.Description, 
		route.IsActive, route.CreatedAt, route.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create delivery route: %w", err)
	}

	return nil
}

// GetByID retrieves a delivery route by ID
func (r *deliveryRouteRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.DeliveryRoute, error) {
	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM delivery_routes 
		WHERE id = $1`

	route := &entity.DeliveryRoute{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&route.ID, &route.Name, &route.Description,
		&route.IsActive, &route.CreatedAt, &route.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrDeliveryRouteNotFound
		}
		return nil, fmt.Errorf("failed to get delivery route: %w", err)
	}

	return route, nil
}

// GetAll retrieves all delivery routes
func (r *deliveryRouteRepository) GetAll(ctx context.Context) ([]entity.DeliveryRoute, error) {
	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM delivery_routes 
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all delivery routes: %w", err)
	}
	defer rows.Close()

	var routes []entity.DeliveryRoute
	for rows.Next() {
		route := entity.DeliveryRoute{}
		err := rows.Scan(
			&route.ID, &route.Name, &route.Description,
			&route.IsActive, &route.CreatedAt, &route.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan delivery route: %w", err)
		}

		routes = append(routes, route)
	}

	return routes, nil
}

// Update updates a delivery route
func (r *deliveryRouteRepository) Update(ctx context.Context, route *entity.DeliveryRoute) error {
	query := `
		UPDATE delivery_routes SET
			name = $2, description = $3, is_active = $4, updated_at = $5
		WHERE id = $1`

	route.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		route.ID, route.Name, route.Description, route.IsActive, route.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update delivery route: %w", err)
	}

	return nil
}

// Delete deletes a delivery route
func (r *deliveryRouteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM delivery_routes WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete delivery route: %w", err)
	}

	return nil
}

// customerAnalyticsRepository implements repository.CustomerAnalyticsRepository  
type customerAnalyticsRepository struct {
	db *sql.DB
}

// NewCustomerAnalyticsRepository creates a new customer analytics repository
func NewCustomerAnalyticsRepository(db *sql.DB) repository.CustomerAnalyticsRepository {
	return &customerAnalyticsRepository{db: db}
}

// GetCustomerInsights retrieves analytics insights for a customer
func (r *customerAnalyticsRepository) GetCustomerInsights(ctx context.Context, customerID uuid.UUID) (*entity.CustomerAnalytics, error) {
	// This would be a complex query joining multiple tables
	// For now, return basic analytics
	analytics := &entity.CustomerAnalytics{
		CustomerID: customerID,
		// TODO: Implement complex analytics queries
	}
	
	return analytics, nil
}

// UpdatePurchaseAnalytics updates purchase analytics for a customer
func (r *customerAnalyticsRepository) UpdatePurchaseAnalytics(ctx context.Context, customerID uuid.UUID, orderValue float64, orderDate time.Time) error {
	// Update customer purchase analytics
	query := `
		UPDATE customers SET
			order_count = order_count + 1,
			last_order_date = $2,
			average_order_value = (average_order_value * order_count + $3) / (order_count + 1),
			updated_at = $4
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, customerID, orderDate, orderValue, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update purchase analytics: %w", err)
	}

	return nil
}

// GetSegmentationData retrieves segmentation data for a customer
func (r *customerAnalyticsRepository) GetSegmentationData(ctx context.Context, customerID uuid.UUID) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	// TODO: Implement segmentation logic
	return data, nil
}

// GetRecommendations retrieves upsell recommendations for a customer
func (r *customerAnalyticsRepository) GetRecommendations(ctx context.Context, customerID uuid.UUID) ([]entity.UpsellSuggestion, error) {
	var recommendations []entity.UpsellSuggestion
	// TODO: Implement recommendation engine
	return recommendations, nil
}


