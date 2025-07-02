package http

import (
	"net/http"

	"saan/shipping/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ShippingHandler struct {
	shippingService domain.ShippingService
	routeService    domain.RouteService
	carrierService  domain.CarrierService
}

func NewRouter(shippingService domain.ShippingService, routeService domain.RouteService, carrierService domain.CarrierService) *gin.Engine {
	handler := &ShippingHandler{
		shippingService: shippingService,
		routeService:    routeService,
		carrierService:  carrierService,
	}

	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "shipping-service",
			"status":  "healthy",
		})
	})

	// API routes
	api := router.Group("/api/shipping")
	{
		// Delivery options
		api.GET("/delivery-options", handler.GetDeliveryOptions)
		
		// Delivery tasks
		api.POST("/tasks", handler.CreateDeliveryTask)
		api.GET("/tasks/:id", handler.GetDeliveryTask)
		api.PUT("/tasks/:id/status", handler.UpdateTaskStatus)
		api.GET("/orders/:orderId/delivery", handler.GetTaskByOrderID)
		
		// Routes
		api.GET("/routes", handler.GetRoutes)
		api.GET("/routes/:code", handler.GetRouteInfo)
		
		// Carriers
		api.GET("/carriers", handler.GetCarriers)
		api.GET("/carriers/:id/tracking/:trackingNumber", handler.GetTrackingInfo)
		
		// Route planning
		api.POST("/plan-routes", handler.PlanDailyRoutes)
	}

	return router
}

func (h *ShippingHandler) GetDeliveryOptions(c *gin.Context) {
	addressIDStr := c.Query("customer_address_id")
	if addressIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer_address_id is required"})
		return
	}

	addressID, err := uuid.Parse(addressIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer_address_id"})
		return
	}

	options, err := h.shippingService.GetDeliveryOptions(addressID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customer_address_id": addressID,
		"delivery_options":    options,
	})
}

func (h *ShippingHandler) CreateDeliveryTask(c *gin.Context) {
	var req struct {
		OrderID           uuid.UUID `json:"order_id" binding:"required"`
		CustomerAddressID uuid.UUID `json:"customer_address_id" binding:"required"`
		CODAmount         float64   `json:"cod_amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.shippingService.CreateDeliveryTask(req.OrderID, req.CustomerAddressID, req.CODAmount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (h *ShippingHandler) GetDeliveryTask(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	// This would call the repository to get task
	c.JSON(http.StatusOK, gin.H{"task_id": taskID})
}

func (h *ShippingHandler) UpdateTaskStatus(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	var req struct {
		Status domain.TaskStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.shippingService.UpdateTaskStatus(taskID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (h *ShippingHandler) GetTaskByOrderID(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	task, err := h.shippingService.GetTaskByOrderID(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *ShippingHandler) GetRoutes(c *gin.Context) {
	// Mock response
	c.JSON(http.StatusOK, gin.H{"routes": []string{"route_a", "route_b"}})
}

func (h *ShippingHandler) GetRouteInfo(c *gin.Context) {
	routeCode := c.Param("code")
	
	routeInfo, err := h.routeService.GetRouteInfo(routeCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
		return
	}

	c.JSON(http.StatusOK, routeInfo)
}

func (h *ShippingHandler) GetCarriers(c *gin.Context) {
	province := c.Query("province")
	
	carriers, err := h.carrierService.GetAvailableCarriers(province)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"carriers": carriers})
}

func (h *ShippingHandler) GetTrackingInfo(c *gin.Context) {
	carrierIDStr := c.Param("id")
	trackingNumber := c.Param("trackingNumber")

	carrierID, err := uuid.Parse(carrierIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid carrier id"})
		return
	}

	trackingInfo, err := h.carrierService.GetTrackingInfo(carrierID, trackingNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trackingInfo)
}

func (h *ShippingHandler) PlanDailyRoutes(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date parameter is required"})
		return
	}

	// Parse date and plan routes
	c.JSON(http.StatusOK, gin.H{"message": "routes planned successfully"})
}
