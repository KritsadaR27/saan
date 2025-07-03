package handler

import (
	"net/http"
	"strconv"

	"product-service/internal/application"
	"product-service/internal/domain/entity"
	"product-service/internal/domain/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ProductHandler handles product HTTP requests
type ProductHandler struct {
	productUsecase *application.ProductUsecase
	logger         *logrus.Logger
}

// NewProductHandler creates a new product handler
func NewProductHandler(productUsecase *application.ProductUsecase, logger *logrus.Logger) *ProductHandler {
	return &ProductHandler{
		productUsecase: productUsecase,
		logger:         logger,
	}
}

// CreateProduct creates a new product
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req application.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.productUsecase.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create product")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// GetProduct retrieves a product by ID
func (h *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	product, err := h.productUsecase.GetProduct(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get product")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product"})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// GetProducts retrieves products with filtering
func (h *ProductHandler) GetProducts(c *gin.Context) {
	filter := repository.ProductFilter{
		Limit:  50, // Default limit
		Offset: 0,
	}

	// Parse query parameters
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if categoryID, err := uuid.Parse(categoryIDStr); err == nil {
			filter.CategoryID = &categoryID
		}
	}

	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filter.IsActive = &isActive
		}
	}

	// Handle search
	query := c.Query("search")
	var products []*entity.Product
	var err error

	if query != "" {
		products, err = h.productUsecase.SearchProducts(c.Request.Context(), query, filter)
	} else {
		products, err = h.productUsecase.ListProducts(c.Request.Context(), filter)
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to get products")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

// UpdateProduct updates an existing product
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req application.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.productUsecase.UpdateProduct(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update product")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// DeleteProduct deletes a product
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	if err := h.productUsecase.DeleteProduct(c.Request.Context(), id); err != nil {
		h.logger.WithError(err).Error("Failed to delete product")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
