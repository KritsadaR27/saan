package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"customer/internal/application"
	"customer/internal/domain/entity"
)

// PointsHandler handles points HTTP requests
type PointsHandler struct {
	pointsUsecase *application.PointsUsecase
}

// NewPointsHandler creates a new points handler
func NewPointsHandler(pointsUsecase *application.PointsUsecase) *PointsHandler {
	return &PointsHandler{
		pointsUsecase: pointsUsecase,
	}
}

// EarnPointsHTTPRequest represents the HTTP request body for earning points
type EarnPointsHTTPRequest struct {
	Points      int     `json:"points" binding:"required,min=1"`
	Source      string  `json:"source" binding:"required"`
	Description *string `json:"description"`
	OrderID     string  `json:"order_id"` // Optional reference to order
}

// RedeemPointsHTTPRequest represents the HTTP request body for redeeming points
type RedeemPointsHTTPRequest struct {
	Points      int     `json:"points" binding:"required,min=1"`
	Source      string  `json:"source" binding:"required"`
	Description *string `json:"description"`
	OrderID     string  `json:"order_id"` // Optional reference to order
}

// EarnPoints adds points to a customer's balance
func (h *PointsHandler) EarnPoints(c *gin.Context) {
	// Get customer ID from URL parameter
	idStr := c.Param("id")
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var req EarnPointsHTTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert HTTP request to usecase request
	earnReq := application.EarnPointsRequest{
		CustomerID:  customerID,
		Points:      req.Points,
		Source:      req.Source,
		Description: "",
	}

	if req.Description != nil {
		earnReq.Description = *req.Description
	}

	// Add reference if OrderID is provided
	if req.OrderID != "" {
		orderID, err := uuid.Parse(req.OrderID)
		if err == nil {
			earnReq.ReferenceID = &orderID
			refType := "order"
			earnReq.ReferenceType = &refType
		}
	}

	transaction, err := h.pointsUsecase.EarnPoints(c.Request.Context(), &earnReq)
	if err != nil {
		switch err {
		case entity.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case entity.ErrInsufficientPoints:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add points"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Points added successfully",
		"transaction": transaction,
	})
}

// RedeemPoints deducts points from a customer's balance
func (h *PointsHandler) RedeemPoints(c *gin.Context) {
	// Get customer ID from URL parameter
	idStr := c.Param("id")
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var req RedeemPointsHTTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert HTTP request to usecase request
	redeemReq := application.RedeemPointsRequest{
		CustomerID:  customerID,
		Points:      req.Points,
		Source:      req.Source,
		Description: "",
	}

	if req.Description != nil {
		redeemReq.Description = *req.Description
	}

	// Add reference if OrderID is provided
	if req.OrderID != "" {
		orderID, err := uuid.Parse(req.OrderID)
		if err == nil {
			redeemReq.ReferenceID = &orderID
			refType := "order"
			redeemReq.ReferenceType = &refType
		}
	}

	transaction, err := h.pointsUsecase.RedeemPoints(c.Request.Context(), &redeemReq)
	if err != nil {
		switch err {
		case entity.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case entity.ErrInsufficientPoints:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deduct points"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Points redeemed successfully",
		"transaction": transaction,
	})
}

// GetPointsBalance retrieves a customer's current points balance
func (h *PointsHandler) GetPointsBalance(c *gin.Context) {
	idStr := c.Param("id")
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	balance, err := h.pointsUsecase.GetPointsBalance(c.Request.Context(), customerID)
	if err != nil {
		switch err {
		case entity.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get points balance"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customer_id": customerID,
		"balance":     balance,
	})
}

// GetPointsHistory retrieves a customer's points transaction history
func (h *PointsHandler) GetPointsHistory(c *gin.Context) {
	idStr := c.Param("id")
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

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

	transactions, err := h.pointsUsecase.GetPointsHistory(c.Request.Context(), customerID, limit, offset)
	if err != nil {
		switch err {
		case entity.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get points history"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customer_id":    customerID,
		"transactions":   transactions,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
		},
	})
}

// GetPointsStats retrieves customer points statistics
func (h *PointsHandler) GetPointsStats(c *gin.Context) {
	idStr := c.Param("id")
	customerID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	stats, err := h.pointsUsecase.GetPointsStats(c.Request.Context(), customerID)
	if err != nil {
		switch err {
		case entity.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get points stats"})
		}
		return
	}

	c.JSON(http.StatusOK, stats)
}
