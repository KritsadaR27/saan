package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/saan-system/services/customer/internal/domain/entity"
	"github.com/saan-system/services/customer/internal/domain/repository"
)

// AddressUsecase handles customer address business logic
type AddressUsecase struct {
	addressRepo      repository.CustomerAddressRepository
	customerRepo     repository.CustomerRepository
	thaiAddressRepo  repository.ThaiAddressRepository
	eventPublisher   repository.EventPublisher
}

// NewAddressUsecase creates a new address usecase
func NewAddressUsecase(
	addressRepo repository.CustomerAddressRepository,
	customerRepo repository.CustomerRepository,
	thaiAddressRepo repository.ThaiAddressRepository,
	eventPublisher repository.EventPublisher,
) *AddressUsecase {
	return &AddressUsecase{
		addressRepo:     addressRepo,
		customerRepo:    customerRepo,
		thaiAddressRepo: thaiAddressRepo,
		eventPublisher:  eventPublisher,
	}
}

// CreateAddressRequest represents a request to create a customer address
type CreateAddressRequest struct {
	CustomerID    uuid.UUID `json:"customer_id" validate:"required"`
	Type          string    `json:"type" validate:"required"`
	Label         string    `json:"label"`
	AddressLine1  string    `json:"address_line1" validate:"required"`
	AddressLine2  *string   `json:"address_line2"`
	SubDistrict   string    `json:"sub_district" validate:"required"`
	District      string    `json:"district" validate:"required"`
	Province      string    `json:"province" validate:"required"`
	PostalCode    string    `json:"postal_code" validate:"required"`
	IsDefault     bool      `json:"is_default"`
	DeliveryNotes *string   `json:"delivery_notes"`
}

// UpdateAddressRequest represents a request to update a customer address
type UpdateAddressRequest struct {
	ID            uuid.UUID `json:"id" validate:"required"`
	Type          string    `json:"type" validate:"required"`
	Label         string    `json:"label"`
	AddressLine1  string    `json:"address_line1" validate:"required"`
	AddressLine2  *string   `json:"address_line2"`
	SubDistrict   string    `json:"sub_district" validate:"required"`
	District      string    `json:"district" validate:"required"`
	Province      string    `json:"province" validate:"required"`
	PostalCode    string    `json:"postal_code" validate:"required"`
	IsDefault     bool      `json:"is_default"`
	DeliveryNotes *string   `json:"delivery_notes"`
}

// CreateCustomerAddress creates a new customer address
func (uc *AddressUsecase) CreateCustomerAddress(ctx context.Context, req CreateAddressRequest) (*entity.CustomerAddress, error) {
	// 1. Validation
	if err := uc.validateCreateAddressRequest(req); err != nil {
		return nil, err
	}

	// 2. Verify customer exists
	_, err := uc.customerRepo.GetByID(ctx, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// 3. Create address entity
	address := &entity.CustomerAddress{
		ID:            uuid.New(),
		CustomerID:    req.CustomerID,
		Type:          req.Type,
		Label:         req.Label,
		AddressLine1:  req.AddressLine1,
		AddressLine2:  req.AddressLine2,
		SubDistrict:   req.SubDistrict,
		District:      req.District,
		Province:      req.Province,
		PostalCode:    req.PostalCode,
		IsDefault:     req.IsDefault,
		DeliveryNotes: req.DeliveryNotes,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 4. If this is set as default, unset other defaults first
	if req.IsDefault {
		if err := uc.unsetOtherDefaults(ctx, req.CustomerID); err != nil {
			return nil, fmt.Errorf("failed to unset other defaults: %w", err)
		}
	}

	// 5. Persistence
	if err := uc.addressRepo.Create(ctx, address); err != nil {
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	return address, nil
}

// GetCustomerAddresses retrieves all addresses for a customer
func (uc *AddressUsecase) GetCustomerAddresses(ctx context.Context, customerID uuid.UUID) ([]entity.CustomerAddress, error) {
	addresses, err := uc.addressRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer addresses: %w", err)
	}

	return addresses, nil
}

// GetDefaultAddress retrieves the default address for a customer
func (uc *AddressUsecase) GetDefaultAddress(ctx context.Context, customerID uuid.UUID) (*entity.CustomerAddress, error) {
	address, err := uc.addressRepo.GetDefaultAddress(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get default address: %w", err)
	}

	return address, nil
}

// UpdateCustomerAddress updates a customer address
func (uc *AddressUsecase) UpdateCustomerAddress(ctx context.Context, req UpdateAddressRequest) (*entity.CustomerAddress, error) {
	// 1. Get existing address
	existing, err := uc.addressRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	// 2. Update fields
	existing.Type = req.Type
	existing.Label = req.Label
	existing.AddressLine1 = req.AddressLine1
	existing.AddressLine2 = req.AddressLine2
	existing.SubDistrict = req.SubDistrict
	existing.District = req.District
	existing.Province = req.Province
	existing.PostalCode = req.PostalCode
	existing.DeliveryNotes = req.DeliveryNotes
	existing.UpdatedAt = time.Now()

	// 3. Handle default address logic
	if req.IsDefault && !existing.IsDefault {
		if err := uc.unsetOtherDefaults(ctx, existing.CustomerID); err != nil {
			return nil, fmt.Errorf("failed to unset other defaults: %w", err)
		}
		existing.IsDefault = true
	} else if !req.IsDefault && existing.IsDefault {
		existing.IsDefault = false
	}

	// 4. Update in database
	if err := uc.addressRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update address: %w", err)
	}

	return existing, nil
}

// DeleteCustomerAddress deletes a customer address
func (uc *AddressUsecase) DeleteCustomerAddress(ctx context.Context, addressID uuid.UUID) error {
	// 1. Check if address exists
	existing, err := uc.addressRepo.GetByID(ctx, addressID)
	if err != nil {
		return fmt.Errorf("failed to get address: %w", err)
	}

	// 2. Delete address
	if err := uc.addressRepo.Delete(ctx, addressID); err != nil {
		return fmt.Errorf("failed to delete address: %w", err)
	}

	// 3. If this was the default address, set another as default
	if existing.IsDefault {
		addresses, err := uc.addressRepo.GetByCustomerID(ctx, existing.CustomerID)
		if err == nil && len(addresses) > 0 {
			// Set the first remaining address as default
			if err := uc.addressRepo.SetAsDefault(ctx, addresses[0].ID, existing.CustomerID); err != nil {
				// Log error but don't fail the operation
			}
		}
	}

	return nil
}

// SetDefaultAddress sets an address as default for a customer
func (uc *AddressUsecase) SetDefaultAddress(ctx context.Context, addressID uuid.UUID, customerID uuid.UUID) error {
	// 1. Verify address belongs to customer
	address, err := uc.addressRepo.GetByID(ctx, addressID)
	if err != nil {
		return fmt.Errorf("failed to get address: %w", err)
	}

	if address.CustomerID != customerID {
		return fmt.Errorf("address does not belong to customer")
	}

	// 2. Set as default (this will unset others)
	if err := uc.addressRepo.SetAsDefault(ctx, addressID, customerID); err != nil {
		return fmt.Errorf("failed to set default address: %w", err)
	}

	return nil
}

// GetCustomerAddress gets a specific customer address
func (uc *AddressUsecase) GetCustomerAddress(ctx context.Context, customerID, addressID uuid.UUID) (*entity.CustomerAddress, error) {
	address, err := uc.addressRepo.GetByID(ctx, addressID)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	// Verify the address belongs to the customer
	if address.CustomerID != customerID {
		return nil, entity.ErrAddressNotFound
	}

	return address, nil
}

// GetAddressSuggestions gets address suggestions based on search query
func (uc *AddressUsecase) GetAddressSuggestions(ctx context.Context, query string) ([]entity.AddressSuggestion, error) {
	return uc.thaiAddressRepo.GetAddressSuggestions(ctx, query, 10)
}

// SearchThaiAddresses searches Thai addresses
func (uc *AddressUsecase) SearchThaiAddresses(ctx context.Context, query string) ([]entity.ThaiAddress, error) {
	return uc.thaiAddressRepo.AutoComplete(ctx, query, 20)
}

// GetThaiAddressByPostalCode gets Thai address by postal code
func (uc *AddressUsecase) GetThaiAddressByPostalCode(ctx context.Context, postalCode string) ([]entity.ThaiAddress, error) {
	return uc.thaiAddressRepo.GetByPostalCode(ctx, postalCode)
}

// Private helper methods

func (uc *AddressUsecase) validateCreateAddressRequest(req CreateAddressRequest) error {
	if req.CustomerID == uuid.Nil {
		return entity.ErrInvalidCustomerID
	}
	if req.AddressLine1 == "" {
		return entity.ErrInvalidAddressLine1
	}
	if req.PostalCode == "" {
		return entity.ErrInvalidPostalCode
	}
	if req.SubDistrict == "" {
		return entity.ErrInvalidSubDistrict
	}
	if req.District == "" {
		return entity.ErrInvalidDistrict
	}
	if req.Province == "" {
		return entity.ErrInvalidProvince
	}
	return nil
}

func (uc *AddressUsecase) unsetOtherDefaults(ctx context.Context, customerID uuid.UUID) error {
	// Get all addresses for customer
	addresses, err := uc.addressRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return err
	}

	// Update all default addresses to non-default
	for _, addr := range addresses {
		if addr.IsDefault {
			addr.IsDefault = false
			addr.UpdatedAt = time.Now()
			if err := uc.addressRepo.Update(ctx, &addr); err != nil {
				return err
			}
		}
	}

	return nil
}
