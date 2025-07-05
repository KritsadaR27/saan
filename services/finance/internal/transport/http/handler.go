package http

import (
	"net/http"
	"time"

	"finance/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FinanceHandler struct {
	financeService    domain.FinanceService
	allocationService domain.AllocationService
	cashFlowService   domain.CashFlowService
}

func NewRouter(financeService domain.FinanceService, allocationService domain.AllocationService, cashFlowService domain.CashFlowService) *gin.Engine {
	handler := &FinanceHandler{
		financeService:    financeService,
		allocationService: allocationService,
		cashFlowService:   cashFlowService,
	}

	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "finance-service",
			"status":  "healthy",
		})
	})

	// API routes
	api := router.Group("/api/finance")
	{
		// End of day processing
		api.POST("/end-of-day", handler.ProcessEndOfDay)
		api.POST("/summaries/:id/expenses", handler.AddExpense)
		api.POST("/summaries/:id/reconcile", handler.ReconcileCash)
		
		// Cash transfers
		api.POST("/transfer-batches", handler.CreateTransferBatch)
		api.POST("/transfer-batches/:id/execute", handler.ExecuteTransferBatch)
		
		// Cash status
		api.GET("/cash-status", handler.GetCashStatus)
		api.GET("/entities/:type/:id/cash-flow", handler.GetEntityCashFlow)
		api.GET("/entities/:type/:id/balance", handler.GetCurrentBalance)
		
		// Profit allocations
		api.GET("/allocation-rules", handler.GetAllocationRule)
		api.PUT("/allocation-rules", handler.UpdateAllocationRule)
	}

	return router
}

func (h *FinanceHandler) ProcessEndOfDay(c *gin.Context) {
	var req struct {
		Date           string     `json:"date" binding:"required"`
		BranchID       *uuid.UUID `json:"branch_id,omitempty"`
		VehicleID      *uuid.UUID `json:"vehicle_id,omitempty"`
		Sales          float64    `json:"sales" binding:"required"`
		CODCollections float64    `json:"cod_collections"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
		return
	}

	summary, err := h.financeService.ProcessEndOfDay(date, req.BranchID, req.VehicleID, req.Sales, req.CODCollections)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, summary)
}

func (h *FinanceHandler) AddExpense(c *gin.Context) {
	summaryIDStr := c.Param("id")
	summaryID, err := uuid.Parse(summaryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid summary id"})
		return
	}

	var req struct {
		Category    string    `json:"category" binding:"required"`
		Description string    `json:"description" binding:"required"`
		Amount      float64   `json:"amount" binding:"required"`
		EnteredBy   uuid.UUID `json:"entered_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.financeService.AddExpenseEntry(summaryID, req.Category, req.Description, req.Amount, req.EnteredBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "expense added"})
}

func (h *FinanceHandler) ReconcileCash(c *gin.Context) {
	summaryIDStr := c.Param("id")
	summaryID, err := uuid.Parse(summaryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid summary id"})
		return
	}

	var req struct {
		ActualCash   float64   `json:"actual_cash" binding:"required"`
		ReconciledBy uuid.UUID `json:"reconciled_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.financeService.ReconcileCash(summaryID, req.ActualCash, req.ReconciledBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cash reconciled"})
}

func (h *FinanceHandler) CreateTransferBatch(c *gin.Context) {
	var req struct {
		BranchID     *uuid.UUID               `json:"branch_id,omitempty"`
		VehicleID    *uuid.UUID               `json:"vehicle_id,omitempty"`
		Transfers    []*domain.CashTransfer   `json:"transfers" binding:"required"`
		AuthorizedBy uuid.UUID               `json:"authorized_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	batch, err := h.financeService.CreateTransferBatch(req.BranchID, req.VehicleID, req.Transfers, req.AuthorizedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, batch)
}

func (h *FinanceHandler) ExecuteTransferBatch(c *gin.Context) {
	batchIDStr := c.Param("id")
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batch id"})
		return
	}

	if err := h.financeService.ExecuteTransferBatch(batchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transfer batch executed"})
}

func (h *FinanceHandler) GetCashStatus(c *gin.Context) {
	status, err := h.financeService.GetCashStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *FinanceHandler) GetEntityCashFlow(c *gin.Context) {
	entityType := c.Param("type")
	entityIDStr := c.Param("id")
	
	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity id"})
		return
	}

	cashFlow, err := h.cashFlowService.GetEntityCashFlow(entityType, entityID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cash_flow": cashFlow})
}

func (h *FinanceHandler) GetCurrentBalance(c *gin.Context) {
	entityType := c.Param("type")
	entityIDStr := c.Param("id")
	
	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity id"})
		return
	}

	balance, err := h.cashFlowService.GetCurrentBalance(entityType, entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balance": balance})
}

func (h *FinanceHandler) GetAllocationRule(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	vehicleIDStr := c.Query("vehicle_id")

	var branchID, vehicleID *uuid.UUID
	
	if branchIDStr != "" {
		if id, err := uuid.Parse(branchIDStr); err == nil {
			branchID = &id
		}
	}
	
	if vehicleIDStr != "" {
		if id, err := uuid.Parse(vehicleIDStr); err == nil {
			vehicleID = &id
		}
	}

	rule, err := h.allocationService.GetCurrentRule(branchID, vehicleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

func (h *FinanceHandler) UpdateAllocationRule(c *gin.Context) {
	var rule domain.ProfitAllocationRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.allocationService.UpdateAllocationRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "allocation rule updated"})
}
