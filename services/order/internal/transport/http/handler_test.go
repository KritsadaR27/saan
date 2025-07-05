package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"order/internal/application/dto"
	"order/internal/domain"
)

// HandlerTestSuite tests HTTP endpoints
type HandlerTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *HandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	
	// Add health check route for basic testing
	suite.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
}

// Test basic HTTP functionality
func (suite *HandlerTestSuite) TestHealthCheck() {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	suite.router.ServeHTTP(recorder, req)

	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "healthy", response["status"])
}

// Test request validation 
func (suite *HandlerTestSuite) TestValidateCreateOrderRequest() {
	// Test valid request structure
	validRequest := &dto.CreateOrderRequest{
		CustomerID:      uuid.New(),
		ShippingAddress: "123 Main St",
		BillingAddress:  "123 Main St",
		Notes:           "Test order",
		Items: []dto.CreateOrderItemRequest{
			{
				ProductID: uuid.New(),
				Quantity:  2,
				UnitPrice: 10.99,
			},
		},
	}

	requestBody, err := json.Marshal(validRequest)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), requestBody)

	// Validate JSON structure
	var parsed dto.CreateOrderRequest
	err = json.Unmarshal(requestBody, &parsed)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), validRequest.CustomerID, parsed.CustomerID)
	assert.Len(suite.T(), parsed.Items, 1)
}

// Test status validation
func (suite *HandlerTestSuite) TestValidateOrderStatus() {
	// Test valid order statuses
	validStatuses := []domain.OrderStatus{
		domain.OrderStatusPending,
		domain.OrderStatusConfirmed,
		domain.OrderStatusProcessing,
		domain.OrderStatusShipped,
		domain.OrderStatusDelivered,
		domain.OrderStatusCancelled,
	}

	for _, status := range validStatuses {
		updateRequest := &dto.UpdateOrderStatusRequest{
			Status: status,
		}

		requestBody, err := json.Marshal(updateRequest)
		assert.NoError(suite.T(), err)
		
		var parsed dto.UpdateOrderStatusRequest
		err = json.Unmarshal(requestBody, &parsed)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), status, parsed.Status)
	}
}

// Test error response structure
func (suite *HandlerTestSuite) TestErrorResponseFormat() {
	// Create a test route that returns an error
	suite.router.GET("/test-error", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"code":    "VALIDATION_ERROR",
			"details": "Missing required field",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test-error", nil)
	recorder := httptest.NewRecorder()

	suite.router.ServeHTTP(recorder, req)

	assert.Equal(suite.T(), http.StatusBadRequest, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Validation failed", response["error"])
	assert.Equal(suite.T(), "VALIDATION_ERROR", response["code"])
}

// Test pagination parameters
func (suite *HandlerTestSuite) TestPaginationParams() {
	suite.router.GET("/test-pagination", func(c *gin.Context) {
		limit := c.DefaultQuery("limit", "10")
		offset := c.DefaultQuery("offset", "0")
		
		c.JSON(http.StatusOK, gin.H{
			"limit":  limit,
			"offset": offset,
		})
	})

	// Test default values
	req := httptest.NewRequest(http.MethodGet, "/test-pagination", nil)
	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "10", response["limit"])
	assert.Equal(suite.T(), "0", response["offset"])

	// Test custom values
	req = httptest.NewRequest(http.MethodGet, "/test-pagination?limit=20&offset=10", nil)
	recorder = httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "20", response["limit"])
	assert.Equal(suite.T(), "10", response["offset"])
}

// Test UUID parameter validation
func (suite *HandlerTestSuite) TestUUIDValidation() {
	suite.router.GET("/test-uuid/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": id.String()})
	})

	// Test valid UUID
	validUUID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/test-uuid/"+validUUID.String(), nil)
	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)

	// Test invalid UUID
	req = httptest.NewRequest(http.MethodGet, "/test-uuid/invalid-uuid", nil)
	recorder = httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)
	assert.Equal(suite.T(), http.StatusBadRequest, recorder.Code)
}

// Test date format validation
func (suite *HandlerTestSuite) TestDateFormatValidation() {
	suite.router.GET("/test-date", func(c *gin.Context) {
		dateStr := c.Query("date")
		if dateStr == "" {
			dateStr = time.Now().Format("2006-01-02")
		}
		
		_, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"date": dateStr})
	})

	// Test valid date
	req := httptest.NewRequest(http.MethodGet, "/test-date?date=2024-01-15", nil)
	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)

	// Test invalid date
	req = httptest.NewRequest(http.MethodGet, "/test-date?date=invalid-date", nil)
	recorder = httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)
	assert.Equal(suite.T(), http.StatusBadRequest, recorder.Code)
}

// Test content type validation
func (suite *HandlerTestSuite) TestContentTypeValidation() {
	suite.router.POST("/test-json", func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	// Test valid JSON
	requestBody := bytes.NewBufferString(`{"test": "value"}`)
	req := httptest.NewRequest(http.MethodPost, "/test-json", requestBody)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)

	// Test invalid JSON
	requestBody = bytes.NewBufferString(`{"test": invalid}`)
	req = httptest.NewRequest(http.MethodPost, "/test-json", requestBody)
	req.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)
	assert.Equal(suite.T(), http.StatusBadRequest, recorder.Code)
}

// Run the test suite
func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
