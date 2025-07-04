package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/saan-system/services/customer/internal/application"
	"github.com/saan-system/services/customer/internal/domain/entity"
)

// AddressHandler handles address HTTP requests
type AddressHandler struct {
	addressUsecase *application.AddressUsecase
}

// NewAddressHandler creates a new address handler
func NewAddressHandler(addressUsecase *application.AddressUsecase) *AddressHandler {
	return &AddressHandler{
		addressUsecase: addressUsecase,
	}
}

// AddCustomerAddressRequest represents the request body for adding a customer address
type AddCustomerAddressRequest struct {
	Type          string   `json:"type" binding:"required,oneof=home work billing shipping"`
	Label         string   `json:"label" binding:"required,min=1,max=100"`
	AddressLine1  string   `json:"address_line1" binding:"required,min=1,max=255"`
	AddressLine2  *string  `json:"address_line2" binding:"omitempty,max=255"`
	SubDistrict   string   `json:"sub_district" binding:"required,min=1,max=100"`
	District      string   `json:"district" binding:"required,min=1,max=100"`
	Province      string   `json:"province" binding:"required,min=1,max=100"`
	PostalCode    string   `json:"postal_code" binding:"required,len=5"`
	Latitude      *float64 `json:"latitude" binding:"omitempty,min=-90,max=90"`
	Longitude     *float64 `json:"longitude" binding:"omitempty,min=-180,max=180"`
	DeliveryNotes *string  `json:"delivery_notes" binding:"omitempty,max=500"`
}

// UpdateCustomerAddressRequest represents the request body for updating a customer address
type UpdateCustomerAddressRequest struct {
	Type          *string  `json:"type" binding:"omitempty,oneof=home work billing shipping"`
	Label         *string  `json:"label" binding:"omitempty,min=1,max=100"`
	AddressLine1  *string  `json:"address_line1" binding:"omitempty,min=1,max=255"`
	AddressLine2  *string  `json:"address_line2" binding:"omitempty,max=255"`
	SubDistrict   *string  `json:"sub_district" binding:"omitempty,min=1,max=100"`
	District      *string  `json:"district" binding:"omitempty,min=1,max=100"`
	Province      *string  `json:"province" binding:"omitempty,min=1,max=100"`
	PostalCode    *string  `json:"postal_code" binding:"omitempty,len=5"`
	Latitude      *float64 `json:"latitude" binding:"omitempty,min=-90,max=90"`
	Longitude     *float64 `json:"longitude" binding:"omitempty,min=-180,max=180"`
	DeliveryNotes *string  `json:"delivery_notes" binding:"omitempty,max=500"`
}

// AddCustomerAddress adds a new address to a customer
func (h *AddressHandler) AddCustomerAddress(c *gin.Context) {
	customerIDStr := c.Param("id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var req AddCustomerAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert HTTP request to usecase request
	createReq := application.CreateAddressRequest{
		CustomerID:    customerID,
		Type:          req.Type,
		Label:         req.Label,
		AddressLine1:  req.AddressLine1,
		AddressLine2:  req.AddressLine2,
		SubDistrict:   req.SubDistrict,
		District:      req.District,
		Province:      req.Province,
		PostalCode:    req.PostalCode,
		DeliveryNotes: req.DeliveryNotes,
	}

	createdAddress, err := h.addressUsecase.CreateCustomerAddress(c.Request.Context(), createReq)
	if err != nil {
		switch err {
		case entity.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case entity.ErrInvalidAddressData:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add customer address"})
		}
		return
	}

	c.JSON(http.StatusCreated, createdAddress)
}

// UpdateCustomerAddress updates a customer address
func (h *AddressHandler) UpdateCustomerAddress(c *gin.Context) {
	customerIDStr := c.Param("id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	addressIDStr := c.Param("address_id")
	addressID, err := uuid.Parse(addressIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
		return
	}

	var req UpdateCustomerAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing address first
	existingAddress, err := h.addressUsecase.GetCustomerAddress(c.Request.Context(), customerID, addressID)
	if err != nil {
		switch err {
		case entity.ErrAddressNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get address"})
		}
		return
	}

	// Build update request with existing data as defaults
	updateReq := application.UpdateAddressRequest{
		ID:            addressID,
		Type:          existingAddress.Type,
		Label:         existingAddress.Label,
		AddressLine1:  existingAddress.AddressLine1,
		AddressLine2:  existingAddress.AddressLine2,
		SubDistrict:   existingAddress.SubDistrict,
		District:      existingAddress.District,
		Province:      existingAddress.Province,
		PostalCode:    existingAddress.PostalCode,
		IsDefault:     existingAddress.IsDefault,
		DeliveryNotes: existingAddress.DeliveryNotes,
	}

	// Update only provided fields
	if req.Type != nil {
		updateReq.Type = *req.Type
	}
	if req.Label != nil {
		updateReq.Label = *req.Label
	}
	if req.AddressLine1 != nil {
		updateReq.AddressLine1 = *req.AddressLine1
	}
	if req.AddressLine2 != nil {
		updateReq.AddressLine2 = req.AddressLine2
	}
	if req.SubDistrict != nil {
		updateReq.SubDistrict = *req.SubDistrict
	}
	if req.District != nil {
		updateReq.District = *req.District
	}
	if req.Province != nil {
		updateReq.Province = *req.Province
	}
	if req.PostalCode != nil {
		updateReq.PostalCode = *req.PostalCode
	}
	if req.DeliveryNotes != nil {
		updateReq.DeliveryNotes = req.DeliveryNotes
	}

	updatedAddress, err := h.addressUsecase.UpdateCustomerAddress(c.Request.Context(), updateReq)
	if err != nil {
		switch err {
		case entity.ErrAddressNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case entity.ErrInvalidAddressData:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update address"})
		}
		return
	}

	c.JSON(http.StatusOK, updatedAddress)
}

// DeleteCustomerAddress deletes a customer address
func (h *AddressHandler) DeleteCustomerAddress(c *gin.Context) {
	customerIDStr := c.Param("id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	addressIDStr := c.Param("address_id")
	addressID, err := uuid.Parse(addressIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
		return
	}

	// Verify the address belongs to the customer first
	_, err = h.addressUsecase.GetCustomerAddress(c.Request.Context(), customerID, addressID)
	if err != nil {
		switch err {
		case entity.ErrAddressNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify address"})
		}
		return
	}

	err = h.addressUsecase.DeleteCustomerAddress(c.Request.Context(), addressID)
	if err != nil {
		switch err {
		case entity.ErrAddressNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete address"})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// SetDefaultAddress sets an address as the default for a customer
func (h *AddressHandler) SetDefaultAddress(c *gin.Context) {
	customerIDStr := c.Param("id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	addressIDStr := c.Param("address_id")
	addressID, err := uuid.Parse(addressIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
		return
	}

	err = h.addressUsecase.SetDefaultAddress(c.Request.Context(), customerID, addressID)
	if err != nil {
		switch err {
		case entity.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case entity.ErrAddressNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set default address"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default address set successfully"})
}

// GetAddressSuggestions gets address suggestions based on search query
func (h *AddressHandler) GetAddressSuggestions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	suggestions, err := h.addressUsecase.GetAddressSuggestions(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get address suggestions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"suggestions": suggestions})
}

// SearchThaiAddresses searches Thai addresses
func (h *AddressHandler) SearchThaiAddresses(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	addresses, err := h.addressUsecase.SearchThaiAddresses(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search Thai addresses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"addresses": addresses})
}

// GetThaiAddressByPostalCode gets Thai address by postal code
func (h *AddressHandler) GetThaiAddressByPostalCode(c *gin.Context) {
	postalCode := c.Param("postal_code")
	if postalCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Postal code is required"})
		return
	}

	addresses, err := h.addressUsecase.GetThaiAddressByPostalCode(c.Request.Context(), postalCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Thai address"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"addresses": addresses})
}
