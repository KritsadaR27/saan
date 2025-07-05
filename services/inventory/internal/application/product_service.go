package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"inventory/internal/domain"
	"inventory/internal/infrastructure/database"

	"github.com/sirupsen/logrus"
)

type ProductService struct {
	conn   *database.Connection
	logger *logrus.Logger
}

func NewProductService(conn *database.Connection, logger *logrus.Logger) *ProductService {
	return &ProductService{
		conn:   conn,
		logger: logger,
	}
}

// UpsertProduct creates or updates a product from Loyverse event data
func (s *ProductService) UpsertProduct(ctx context.Context, eventData []byte) error {
	// Parse the simplified event data format from Loyverse transformer
	var productEvent struct {
		ProductID   string `json:"product_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		CategoryID  string `json:"category_id"`
		Source      string `json:"source"`
	}

	if err := json.Unmarshal(eventData, &productEvent); err != nil {
		return fmt.Errorf("failed to unmarshal product event: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"product_id":   productEvent.ProductID,
		"product_name": productEvent.Name,
		"source":       productEvent.Source,
	}).Info("Processing product upsert from Loyverse event")

	// Start transaction
	tx, err := s.conn.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Create product from event data
	product := domain.Product{
		ID:           productEvent.ProductID,
		Name:         productEvent.Name,
		Description:  productEvent.Description,
		CategoryID:   productEvent.CategoryID,
		CategoryName: "",    // Will be populated later or via category sync
		CostPrice:    0,     // Will be updated from inventory sync
		SellPrice:    0,     // Will be updated from pricing sync
		Unit:         "pcs", // Default unit
		IsActive:     true,
		LastUpdated:  time.Now(),
	}

	if err := s.upsertProductInTx(ctx, tx, product); err != nil {
		return fmt.Errorf("failed to upsert product %s: %w", productEvent.ProductID, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.WithField("product_id", productEvent.ProductID).Info("Successfully upserted product from Loyverse event")
	return nil
}

func (s *ProductService) upsertProductInTx(ctx context.Context, tx *sql.Tx, product domain.Product) error {
	query := `
		INSERT INTO products (
			id, name, sku, barcode, category_id, category_name, cost_price, sell_price, 
			unit, description, is_active, last_updated
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			sku = EXCLUDED.sku,
			barcode = EXCLUDED.barcode,
			category_id = EXCLUDED.category_id,
			category_name = EXCLUDED.category_name,
			cost_price = EXCLUDED.cost_price,
			sell_price = EXCLUDED.sell_price,
			unit = EXCLUDED.unit,
			description = EXCLUDED.description,
			is_active = EXCLUDED.is_active,
			last_updated = EXCLUDED.last_updated
	`

	_, err := tx.ExecContext(ctx, query,
		product.ID,
		product.Name,
		product.SKU,
		product.Barcode,
		product.CategoryID,
		product.CategoryName,
		product.CostPrice,
		product.SellPrice,
		product.Unit,
		product.Description,
		product.IsActive,
		product.LastUpdated,
	)

	if err != nil {
		s.logger.WithError(err).WithField("product_id", product.ID).Error("Failed to upsert product")
		return err
	}

	s.logger.WithField("product_id", product.ID).Debug("Product upserted successfully")
	return nil
}
