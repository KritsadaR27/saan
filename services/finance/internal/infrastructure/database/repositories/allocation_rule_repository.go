package repositories

import (
	"database/sql"

	"finance/internal/domain"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type allocationRuleRepository struct {
	db *sql.DB
}

func NewAllocationRuleRepository(db *sql.DB) domain.AllocationRuleRepository {
	return &allocationRuleRepository{
		db: db,
	}
}

func (r *allocationRuleRepository) GetActiveRule(branchID, vehicleID *uuid.UUID) (*domain.ProfitAllocationRule, error) {
	// Try to find specific entity rule first, then fallback to global rule
	query := `
		SELECT 
			id, branch_id, vehicle_id, profit_percentage, owner_pay_percentage, 
			tax_percentage, effective_from, effective_to, is_active, 
			updated_by_user_id, created_at, updated_at
		FROM profit_allocation_rules 
		WHERE is_active = true 
		AND (effective_to IS NULL OR effective_to >= CURRENT_DATE)
		AND effective_from <= CURRENT_DATE
		AND (
			-- Exact entity match
			(branch_id = $1 AND vehicle_id = $2) OR
			-- Branch match with null vehicle
			(branch_id = $1 AND $2::uuid IS NULL AND vehicle_id IS NULL) OR
			-- Vehicle match with null branch  
			(vehicle_id = $2 AND $1::uuid IS NULL AND branch_id IS NULL) OR
			-- Global rule (both null)
			(branch_id IS NULL AND vehicle_id IS NULL)
		)
		ORDER BY 
			-- Prioritize specific entity rules over global
			CASE 
				WHEN branch_id IS NOT NULL OR vehicle_id IS NOT NULL THEN 1 
				ELSE 2 
			END,
			effective_from DESC
		LIMIT 1`

	rule := &domain.ProfitAllocationRule{}
	err := r.db.QueryRow(query, branchID, vehicleID).Scan(
		&rule.ID,
		&rule.BranchID,
		&rule.VehicleID,
		&rule.ProfitPercentage,
		&rule.OwnerPayPercentage,
		&rule.TaxPercentage,
		&rule.EffectiveFrom,
		&rule.EffectiveTo,
		&rule.IsActive,
		&rule.UpdatedByUserID,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAllocationRuleNotFound
		}
		return nil, err
	}

	return rule, nil
}

func (r *allocationRuleRepository) Create(rule *domain.ProfitAllocationRule) error {
	// Validate percentage sum
	totalPercentage := rule.ProfitPercentage + rule.OwnerPayPercentage + rule.TaxPercentage
	if totalPercentage > 100.0 {
		return domain.ErrPercentageSum
	}

	query := `
		INSERT INTO profit_allocation_rules (
			id, branch_id, vehicle_id, profit_percentage, owner_pay_percentage,
			tax_percentage, effective_from, effective_to, is_active,
			updated_by_user_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)`

	_, err := r.db.Exec(query,
		rule.ID,
		rule.BranchID,
		rule.VehicleID,
		rule.ProfitPercentage,
		rule.OwnerPayPercentage,
		rule.TaxPercentage,
		rule.EffectiveFrom,
		rule.EffectiveTo,
		rule.IsActive,
		rule.UpdatedByUserID,
		rule.CreatedAt,
		rule.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23514": // check_violation
				if pqErr.Constraint == "check_percentage_sum" {
					return domain.ErrPercentageSum
				}
			}
		}
		return err
	}

	return nil
}

func (r *allocationRuleRepository) Update(rule *domain.ProfitAllocationRule) error {
	// Validate percentage sum
	totalPercentage := rule.ProfitPercentage + rule.OwnerPayPercentage + rule.TaxPercentage
	if totalPercentage > 100.0 {
		return domain.ErrPercentageSum
	}

	query := `
		UPDATE profit_allocation_rules 
		SET profit_percentage = $2,
			owner_pay_percentage = $3,
			tax_percentage = $4,
			effective_from = $5,
			effective_to = $6,
			is_active = $7,
			updated_by_user_id = $8,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.Exec(query,
		rule.ID,
		rule.ProfitPercentage,
		rule.OwnerPayPercentage,
		rule.TaxPercentage,
		rule.EffectiveFrom,
		rule.EffectiveTo,
		rule.IsActive,
		rule.UpdatedByUserID,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23514": // check_violation
				if pqErr.Constraint == "check_percentage_sum" {
					return domain.ErrPercentageSum
				}
			}
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrAllocationRuleNotFound
	}

	return nil
}

func (r *allocationRuleRepository) DeactivateRule(id uuid.UUID, updatedBy uuid.UUID) error {
	query := `
		UPDATE profit_allocation_rules 
		SET is_active = false,
			updated_by_user_id = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.Exec(query, id, updatedBy)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrAllocationRuleNotFound
	}

	return nil
}

func (r *allocationRuleRepository) GetByEntity(branchID, vehicleID *uuid.UUID) ([]*domain.ProfitAllocationRule, error) {
	query := `
		SELECT 
			id, branch_id, vehicle_id, profit_percentage, owner_pay_percentage, 
			tax_percentage, effective_from, effective_to, is_active, 
			updated_by_user_id, created_at, updated_at
		FROM profit_allocation_rules 
		WHERE ($1::uuid IS NULL OR branch_id = $1) 
		AND ($2::uuid IS NULL OR vehicle_id = $2)
		ORDER BY effective_from DESC, created_at DESC`

	rows, err := r.db.Query(query, branchID, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*domain.ProfitAllocationRule
	for rows.Next() {
		rule := &domain.ProfitAllocationRule{}
		err := rows.Scan(
			&rule.ID,
			&rule.BranchID,
			&rule.VehicleID,
			&rule.ProfitPercentage,
			&rule.OwnerPayPercentage,
			&rule.TaxPercentage,
			&rule.EffectiveFrom,
			&rule.EffectiveTo,
			&rule.IsActive,
			&rule.UpdatedByUserID,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

func (r *allocationRuleRepository) CreateWithTransaction(tx *sql.Tx, rule *domain.ProfitAllocationRule) error {
	// First, deactivate any existing active rules for the same entity
	deactivateQuery := `
		UPDATE profit_allocation_rules 
		SET is_active = false,
			updated_by_user_id = $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE is_active = true
		AND ($2::uuid IS NULL OR branch_id = $2)
		AND ($3::uuid IS NULL OR vehicle_id = $3)`

	_, err := tx.Exec(deactivateQuery, rule.UpdatedByUserID, rule.BranchID, rule.VehicleID)
	if err != nil {
		return err
	}

	// Then create the new rule
	insertQuery := `
		INSERT INTO profit_allocation_rules (
			id, branch_id, vehicle_id, profit_percentage, owner_pay_percentage,
			tax_percentage, effective_from, effective_to, is_active,
			updated_by_user_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)`

	_, err = tx.Exec(insertQuery,
		rule.ID,
		rule.BranchID,
		rule.VehicleID,
		rule.ProfitPercentage,
		rule.OwnerPayPercentage,
		rule.TaxPercentage,
		rule.EffectiveFrom,
		rule.EffectiveTo,
		rule.IsActive,
		rule.UpdatedByUserID,
		rule.CreatedAt,
		rule.UpdatedAt,
	)

	return err
}
