package application

import (
	"context"
	"fmt"
	"time"

	"product-service/internal/domain/entity"
	"product-service/internal/domain/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// PricingUsecase handles pricing business logic
type PricingUsecase struct {
	priceRepo   repository.PriceRepository
	productRepo repository.ProductRepository
	logger      *logrus.Logger
}

// NewPricingUsecase creates a new pricing usecase
func NewPricingUsecase(priceRepo repository.PriceRepository, productRepo repository.ProductRepository, logger *logrus.Logger) *PricingUsecase {
	return &PricingUsecase{
		priceRepo:   priceRepo,
		productRepo: productRepo,
		logger:      logger,
	}
}

// CreatePriceRequest represents the request to create a price
type CreatePriceRequest struct {
	ProductID        uuid.UUID   `json:"product_id" validate:"required"`
	PriceType        string      `json:"price_type" validate:"required"` // "base", "vip", "bulk", "promotional"
	Price            float64     `json:"price" validate:"required,min=0"`
	Currency         string      `json:"currency"`
	MinQuantity      *int        `json:"min_quantity"`
	MaxQuantity      *int        `json:"max_quantity"`
	VIPTierID        *uuid.UUID  `json:"vip_tier_id"`
	ValidFrom        *time.Time  `json:"valid_from"`
	ValidTo          *time.Time  `json:"valid_to"`
	PromotionName    *string     `json:"promotion_name"`
	DiscountPercent  *float64    `json:"discount_percent"`
	LocationIDs      []uuid.UUID `json:"location_ids"`
	CustomerGroupIDs []uuid.UUID `json:"customer_group_ids"`
	Priority         int         `json:"priority"`
}

// UpdatePriceRequest represents the request to update a price
type UpdatePriceRequest struct {
	Price            *float64    `json:"price"`
	MinQuantity      *int        `json:"min_quantity"`
	MaxQuantity      *int        `json:"max_quantity"`
	ValidFrom        *time.Time  `json:"valid_from"`
	ValidTo          *time.Time  `json:"valid_to"`
	PromotionName    *string     `json:"promotion_name"`
	DiscountPercent  *float64    `json:"discount_percent"`
	LocationIDs      []uuid.UUID `json:"location_ids"`
	CustomerGroupIDs []uuid.UUID `json:"customer_group_ids"`
	Priority         *int        `json:"priority"`
	IsActive         *bool       `json:"is_active"`
}

// PriceCalculationRequest represents request for price calculation
type PriceCalculationRequest struct {
	ProductID     uuid.UUID   `json:"product_id" validate:"required"`
	CustomerID    *uuid.UUID  `json:"customer_id"`
	VIPTierID     *uuid.UUID  `json:"vip_tier_id"`
	GroupIDs      []uuid.UUID `json:"group_ids"`
	LocationID    uuid.UUID   `json:"location_id" validate:"required"`
	Quantity      int         `json:"quantity" validate:"required,min=1"`
	CalculateDate *time.Time  `json:"calculate_date"`
}

// CreatePrice creates a new price
func (uc *PricingUsecase) CreatePrice(ctx context.Context, req *CreatePriceRequest) (*entity.Price, error) {
	// Validate request
	if req.ProductID == uuid.Nil {
		return nil, fmt.Errorf("product ID is required")
	}
	if req.PriceType == "" {
		return nil, fmt.Errorf("price type is required")
	}
	if req.Price < 0 {
		return nil, fmt.Errorf("price must be non-negative")
	}

	// Check if product exists
	product, err := uc.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to check product: %w", err)
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Create new price
	price, err := entity.NewPrice(req.ProductID, req.PriceType, req.Price)
	if err != nil {
		return nil, fmt.Errorf("failed to create price: %w", err)
	}

	// Set optional fields
	if req.Currency != "" {
		price.SetCurrency(req.Currency)
	}
	if req.MinQuantity != nil {
		price.SetQuantityRange(req.MinQuantity, req.MaxQuantity)
	}
	if req.VIPTierID != nil {
		price.SetVIPTier(*req.VIPTierID)
	}
	if req.ValidFrom != nil || req.ValidTo != nil {
		price.SetValidityPeriod(req.ValidFrom, req.ValidTo)
	}
	if req.PromotionName != nil {
		price.SetPromotionDetails(*req.PromotionName, req.DiscountPercent)
	}
	price.SetLocationIDs(req.LocationIDs)
	price.SetCustomerGroupIDs(req.CustomerGroupIDs)
	price.SetPriority(req.Priority)

	// Save to database
	if err := uc.priceRepo.Create(ctx, price); err != nil {
		return nil, fmt.Errorf("failed to save price: %w", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"price_id":   price.ID,
		"product_id": req.ProductID,
		"price_type": req.PriceType,
	}).Info("Price created successfully")

	return price, nil
}

// GetPrice retrieves a price by ID
func (uc *PricingUsecase) GetPrice(ctx context.Context, id uuid.UUID) (*entity.Price, error) {
	price, err := uc.priceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	if price == nil {
		return nil, fmt.Errorf("price not found")
	}

	return price, nil
}

// GetProductPrices retrieves all prices for a product
func (uc *PricingUsecase) GetProductPrices(ctx context.Context, productID uuid.UUID) ([]*entity.Price, error) {
	prices, err := uc.priceRepo.GetByProductID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product prices: %w", err)
	}

	return prices, nil
}

// UpdatePrice updates an existing price
func (uc *PricingUsecase) UpdatePrice(ctx context.Context, id uuid.UUID, req *UpdatePriceRequest) (*entity.Price, error) {
	// Get existing price
	price, err := uc.priceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	if price == nil {
		return nil, fmt.Errorf("price not found")
	}

	// Update fields if provided
	if req.Price != nil {
		if err := price.UpdatePrice(*req.Price); err != nil {
			return nil, fmt.Errorf("failed to update price: %w", err)
		}
	}

	if req.MinQuantity != nil {
		price.SetQuantityRange(req.MinQuantity, req.MaxQuantity)
	}

	if req.ValidFrom != nil || req.ValidTo != nil {
		price.SetValidityPeriod(req.ValidFrom, req.ValidTo)
	}

	if req.PromotionName != nil {
		price.SetPromotionDetails(*req.PromotionName, req.DiscountPercent)
	}

	if req.LocationIDs != nil {
		price.SetLocationIDs(req.LocationIDs)
	}

	if req.CustomerGroupIDs != nil {
		price.SetCustomerGroupIDs(req.CustomerGroupIDs)
	}

	if req.Priority != nil {
		price.SetPriority(*req.Priority)
	}

	if req.IsActive != nil {
		if *req.IsActive {
			price.Activate()
		} else {
			price.Deactivate()
		}
	}

	// Save changes
	if err := uc.priceRepo.Update(ctx, price); err != nil {
		return nil, fmt.Errorf("failed to update price: %w", err)
	}

	uc.logger.WithField("price_id", price.ID).Info("Price updated successfully")
	return price, nil
}

// DeletePrice deletes a price
func (uc *PricingUsecase) DeletePrice(ctx context.Context, id uuid.UUID) error {
	// Get existing price
	price, err := uc.priceRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get price: %w", err)
	}

	if price == nil {
		return fmt.Errorf("price not found")
	}

	// Delete price
	if err := uc.priceRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete price: %w", err)
	}

	uc.logger.WithField("price_id", id).Info("Price deleted successfully")
	return nil
}

// CalculatePrice calculates the best price for a product based on conditions
func (uc *PricingUsecase) CalculatePrice(ctx context.Context, req *PriceCalculationRequest) (*entity.PriceCalculation, error) {
	// Get product to ensure it exists and get base price
	product, err := uc.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Create customer info for price calculation
	customerInfo := &repository.CustomerInfo{
		LocationID: req.LocationID,
		Quantity:   req.Quantity,
	}

	if req.CustomerID != nil {
		customerInfo.CustomerID = *req.CustomerID
	}
	if req.VIPTierID != nil {
		customerInfo.VIPTierID = req.VIPTierID
	}
	customerInfo.GroupIDs = req.GroupIDs

	// Try to get effective price based on conditions
	effectivePrice, err := uc.priceRepo.GetEffectivePrice(ctx, req.ProductID, "", customerInfo)
	if err != nil {
		uc.logger.WithError(err).Warn("Failed to get effective price, using base price")
	}

	// Create price calculation result
	calculation := &entity.PriceCalculation{
		ProductID:    req.ProductID,
		BasePrice:    product.BasePrice,
		AppliedPrice: product.BasePrice,
		PriceType:    "base",
		Currency:     "THB",
	}

	if effectivePrice != nil {
		calculation.AppliedPrice = effectivePrice.Price
		calculation.PriceType = effectivePrice.PriceType
		calculation.Currency = effectivePrice.Currency
		calculation.ValidFrom = effectivePrice.ValidFrom
		calculation.ValidTo = effectivePrice.ValidTo
		calculation.DiscountPercent = effectivePrice.DiscountPercent

		// Calculate savings
		if effectivePrice.Price < product.BasePrice {
			savings := product.BasePrice - effectivePrice.Price
			calculation.Savings = &savings
		}
	}

	return calculation, nil
}

// GetVIPPrices gets VIP prices for a product
func (uc *PricingUsecase) GetVIPPrices(ctx context.Context, productID uuid.UUID) ([]*entity.Price, error) {
	prices, err := uc.priceRepo.GetVIPPrices(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VIP prices: %w", err)
	}

	return prices, nil
}

// GetBulkPrices gets bulk prices for a product
func (uc *PricingUsecase) GetBulkPrices(ctx context.Context, productID uuid.UUID) ([]*entity.Price, error) {
	prices, err := uc.priceRepo.GetBulkPrices(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bulk prices: %w", err)
	}

	return prices, nil
}

// GetPromotionalPrices gets promotional prices for a product
func (uc *PricingUsecase) GetPromotionalPrices(ctx context.Context, productID uuid.UUID) ([]*entity.Price, error) {
	prices, err := uc.priceRepo.GetPromotionalPrices(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get promotional prices: %w", err)
	}

	return prices, nil
}
