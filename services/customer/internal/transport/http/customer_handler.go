package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/saan-system/services/customer/internal/application"
	"github.com/saan-system/services/customer/internal/domain"
)

// CustomerHandler handles customer HTTP requests
type CustomerHandler struct {
	app *application.Application
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(app *application.Application) *CustomerHandler {
	return &CustomerHandler{app: app}
}

// CreateCustomer creates a new customer
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req domain.Customer
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := h.app.CustomerService.CreateCustomer(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case domain.ErrCustomerExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case domain.ErrInvalidFirstName, domain.ErrInvalidLastName, domain.ErrInvalidEmail, domain.ErrInvalidPhone:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
		}
		return
	}

	c.JSON(http.StatusCreated, customer)
}

// GetCustomer retrieves a customer by ID
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	customer, err := h.app.CustomerService.GetCustomer(c.Request.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get customer"})
		}
		return
	}

	c.JSON(http.StatusOK, customer)
}

// UpdateCustomer updates a customer
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var req domain.Customer
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = id
	customer, err := h.app.CustomerService.UpdateCustomer(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case domain.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case domain.ErrInvalidFirstName, domain.ErrInvalidLastName, domain.ErrInvalidEmail, domain.ErrInvalidPhone:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer"})
		}
		return
	}

	c.JSON(http.StatusOK, customer)
}

// DeleteCustomer soft deletes a customer
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	err = h.app.CustomerService.DeleteCustomer(c.Request.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete customer"})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListCustomers retrieves customers with filtering and pagination
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	filter := domain.CustomerFilter{}

	// Parse query parameters
	if email := c.Query("email"); email != "" {
		filter.Email = &email
	}
	if phone := c.Query("phone"); phone != "" {
		filter.Phone = &phone
	}
	if tier := c.Query("tier"); tier != "" {
		customerTier := domain.CustomerTier(tier)
		filter.Tier = &customerTier
	}
	if deliveryRouteIDStr := c.Query("delivery_route_id"); deliveryRouteIDStr != "" {
		if deliveryRouteID, err := uuid.Parse(deliveryRouteIDStr); err == nil {
			filter.DeliveryRouteID = &deliveryRouteID
		}
	}

	// Parse pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	// Parse sorting
	filter.SortBy = c.Query("sort_by")
	filter.SortOrder = c.Query("sort_order")

	customers, total, err := h.app.CustomerService.ListCustomers(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list customers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customers": customers,
		"total":     total,
		"limit":     filter.Limit,
		"offset":    filter.Offset,
	})
}

// GetCustomerByEmail retrieves a customer by email
func (h *CustomerHandler) GetCustomerByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	customer, err := h.app.CustomerService.GetCustomerByEmail(c.Request.Context(), email)
	if err != nil {
		switch err {
		case domain.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get customer"})
		}
		return
	}

	c.JSON(http.StatusOK, customer)
}

// GetCustomerByPhone retrieves a customer by phone
func (h *CustomerHandler) GetCustomerByPhone(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone is required"})
		return
	}

	customer, err := h.app.CustomerService.GetCustomerByPhone(c.Request.Context(), phone)
	if err != nil {
		switch err {
		case domain.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get customer"})
		}
		return
	}

	c.JSON(http.StatusOK, customer)
}

// AddCustomerAddress adds a new address for a customer
func (h *CustomerHandler) AddCustomerAddress(c *gin.Context) {
	idStr := c.Param("id")
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var req domain.CustomerAddress
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.CustomerID = customerID
	address, err := h.app.CustomerService.AddCustomerAddress(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case domain.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case domain.ErrInvalidAddressLine1, domain.ErrInvalidPostalCode:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add customer address"})
		}
		return
	}

	c.JSON(http.StatusCreated, address)
}

// UpdateCustomerAddress updates a customer address
func (h *CustomerHandler) UpdateCustomerAddress(c *gin.Context) {
	addressIDStr := c.Param("address_id")
	addressID, err := uuid.Parse(addressIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
		return
	}

	var req domain.CustomerAddress
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = addressID
	address, err := h.app.CustomerService.UpdateCustomerAddress(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case domain.ErrAddressNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case domain.ErrInvalidAddressLine1, domain.ErrInvalidPostalCode:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer address"})
		}
		return
	}

	c.JSON(http.StatusOK, address)
}

// DeleteCustomerAddress soft deletes a customer address
func (h *CustomerHandler) DeleteCustomerAddress(c *gin.Context) {
	addressIDStr := c.Param("address_id")
	addressID, err := uuid.Parse(addressIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
		return
	}

	err = h.app.CustomerService.DeleteCustomerAddress(c.Request.Context(), addressID)
	if err != nil {
		switch err {
		case domain.ErrAddressNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete customer address"})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// SetDefaultAddress sets an address as default for a customer
func (h *CustomerHandler) SetDefaultAddress(c *gin.Context) {
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

	err = h.app.CustomerService.SetDefaultAddress(c.Request.Context(), addressID, customerID)
	if err != nil {
		switch err {
		case domain.ErrAddressNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set default address"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default address set successfully"})
}

// SearchThaiAddresses searches Thai addresses with autocomplete
func (h *CustomerHandler) SearchThaiAddresses(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query is required"})
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	addresses, err := h.app.CustomerService.SearchThaiAddresses(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search Thai addresses"})
		return
	}

	c.JSON(http.StatusOK, addresses)
}

// GetThaiAddressByPostalCode retrieves Thai addresses by postal code
func (h *CustomerHandler) GetThaiAddressByPostalCode(c *gin.Context) {
	postalCode := c.Param("postal_code")
	if postalCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Postal code is required"})
		return
	}

	addresses, err := h.app.CustomerService.GetThaiAddressByPostalCode(c.Request.Context(), postalCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Thai addresses"})
		return
	}

	c.JSON(http.StatusOK, addresses)
}

// SyncWithLoyverse syncs customer with Loyverse
func (h *CustomerHandler) SyncWithLoyverse(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	err = h.app.CustomerService.SyncWithLoyverse(c.Request.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case domain.ErrLoyverseSyncFailed, domain.ErrLoyverseAPIError:
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync with Loyverse"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer synced with Loyverse successfully"})
}

// GetAddressSuggestions returns address suggestions based on query
func (h *CustomerHandler) GetAddressSuggestions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 50 {
		limit = 10
	}

	suggestions, err := h.app.ThaiAddressService.GetAddressSuggestions(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get address suggestions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"suggestions": suggestions,
		"count":      len(suggestions),
	})
}
