package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"customer/internal/application"
	"customer/internal/domain/entity"
)

// CustomerHandler handles customer HTTP requests
type CustomerHandler struct {
	customerUsecase *application.CustomerUsecase
	addressUsecase  *application.AddressUsecase
	pointsUsecase   *application.PointsUsecase
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(
	customerUsecase *application.CustomerUsecase,
	addressUsecase *application.AddressUsecase,
	pointsUsecase *application.PointsUsecase,
) *CustomerHandler {
	return &CustomerHandler{
		customerUsecase: customerUsecase,
		addressUsecase:  addressUsecase,
		pointsUsecase:   pointsUsecase,
	}
}

// CreateCustomerHTTPRequest represents the HTTP request body for creating a customer
type CreateCustomerHTTPRequest struct {
	FirstName       string     `json:"first_name" binding:"required,min=1,max=100"`
	LastName        string     `json:"last_name" binding:"required,min=1,max=100"`
	Email           string     `json:"email" binding:"required,email"`
	Phone           string     `json:"phone" binding:"required,min=10,max=20"`
	DateOfBirth     *time.Time `json:"date_of_birth" binding:"omitempty"`
	Gender          *string    `json:"gender" binding:"omitempty"`
	LoyverseID      *string    `json:"loyverse_id" binding:"omitempty"`
	LineUserID      *string    `json:"line_user_id" binding:"omitempty"`
	LineDisplayName *string    `json:"line_display_name" binding:"omitempty"`
}

// UpdateCustomerHTTPRequest represents the HTTP request body for updating a customer
type UpdateCustomerHTTPRequest struct {
	FirstName       *string    `json:"first_name" binding:"omitempty,min=1,max=100"`
	LastName        *string    `json:"last_name" binding:"omitempty,min=1,max=100"`
	Email           *string    `json:"email" binding:"omitempty,email"`
	Phone           *string    `json:"phone" binding:"omitempty,min=10,max=20"`
	DateOfBirth     *time.Time `json:"date_of_birth" binding:"omitempty"`
	Gender          *string    `json:"gender" binding:"omitempty"`
	LineDisplayName *string    `json:"line_display_name" binding:"omitempty"`
}

// CreateCustomer creates a new customer
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req CreateCustomerHTTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert HTTP request to usecase request
	createReq := application.CreateCustomerRequest{
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Email:           req.Email,
		Phone:           req.Phone,
		DateOfBirth:     req.DateOfBirth,
		Gender:          req.Gender,
		LoyverseID:      req.LoyverseID,
		LineUserID:      req.LineUserID,
		LineDisplayName: req.LineDisplayName,
	}

	createdCustomer, err := h.customerUsecase.CreateCustomer(c.Request.Context(), &createReq)
	if err != nil {
		switch err {
		case entity.ErrCustomerExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case entity.ErrInvalidFirstName, entity.ErrInvalidLastName, entity.ErrInvalidEmail, entity.ErrInvalidPhone:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
		}
		return
	}

	c.JSON(http.StatusCreated, createdCustomer)
}

// GetCustomer retrieves a customer by ID
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	customer, err := h.customerUsecase.GetCustomerByID(c.Request.Context(), id)
	if err != nil {
		switch err {
		case entity.ErrCustomerNotFound:
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

	var req UpdateCustomerHTTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing customer first (we're not using the result but the API expects it)
	_, err = h.customerUsecase.GetCustomerByID(c.Request.Context(), id)
	if err != nil {
		switch err {
		case entity.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get customer"})
		}
		return
	}

	// Build update request with only provided fields  
	updateReq := application.UpdateCustomerRequest{}

	// Update only provided fields
	if req.FirstName != nil {
		updateReq.FirstName = req.FirstName
	}
	if req.LastName != nil {
		updateReq.LastName = req.LastName
	}
	if req.Phone != nil {
		updateReq.Phone = req.Phone
	}
	if req.Email != nil {
		updateReq.Email = req.Email
	}
	if req.DateOfBirth != nil {
		updateReq.DateOfBirth = req.DateOfBirth
	}
	if req.Gender != nil {
		updateReq.Gender = req.Gender
	}
	if req.LineDisplayName != nil {
		updateReq.LineDisplayName = req.LineDisplayName
	}

	updatedCustomer, err := h.customerUsecase.UpdateCustomer(c.Request.Context(), id, &updateReq)
	if err != nil {
		switch err {
		case entity.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case entity.ErrInvalidFirstName, entity.ErrInvalidLastName, entity.ErrInvalidEmail, entity.ErrInvalidPhone:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer"})
		}
		return
	}

	c.JSON(http.StatusOK, updatedCustomer)
}

// DeleteCustomer deletes a customer
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	err = h.customerUsecase.DeleteCustomer(c.Request.Context(), id)
	if err != nil {
		switch err {
		case entity.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete customer"})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetCustomerByEmail retrieves a customer by email
func (h *CustomerHandler) GetCustomerByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email parameter is required"})
		return
	}

	customer, err := h.customerUsecase.GetCustomerByEmail(c.Request.Context(), email)
	if err != nil {
		switch err {
		case entity.ErrCustomerNotFound:
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone parameter is required"})
		return
	}

	customer, err := h.customerUsecase.GetCustomerByPhone(c.Request.Context(), phone)
	if err != nil {
		switch err {
		case entity.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get customer"})
		}
		return
	}

	c.JSON(http.StatusOK, customer)
}

// ListCustomers retrieves a list of customers with pagination
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Calculate offset
	offset := (page - 1) * limit

	customers, total, err := h.customerUsecase.ListCustomers(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list customers"})
		return
	}

	response := gin.H{
		"customers": customers,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	}

	c.JSON(http.StatusOK, response)
}

// SyncWithLoyverse syncs customer data with Loyverse
func (h *CustomerHandler) SyncWithLoyverse(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	err = h.customerUsecase.SyncWithLoyverse(c.Request.Context(), id)
	if err != nil {
		switch err {
		case entity.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync with Loyverse"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer synced with Loyverse successfully"})
}
