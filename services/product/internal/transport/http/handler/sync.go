package handler

import (
	"net/http"

	"product-service/internal/application"
	"product-service/internal/infrastructure/loyverse"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SyncHandler handles sync operations
type SyncHandler struct {
	syncUsecase *application.SyncUsecase
	loyverseSync *loyverse.SyncService
	logger      *logrus.Logger
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(syncUsecase *application.SyncUsecase, loyverseSync *loyverse.SyncService, logger *logrus.Logger) *SyncHandler {
	return &SyncHandler{
		syncUsecase:  syncUsecase,
		loyverseSync: loyverseSync,
		logger:       logger,
	}
}

// SyncFromLoyverse handles full sync from Loyverse
func (h *SyncHandler) SyncFromLoyverse(c *gin.Context) {
	if h.loyverseSync == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Loyverse integration not configured",
		})
		return
	}

	h.logger.Info("Starting manual sync from Loyverse")

	result, err := h.loyverseSync.SyncAllProducts(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to sync from Loyverse")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to sync from Loyverse",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sync completed",
		"result":  result,
	})
}

// SyncProductFromLoyverse handles single product sync
func (h *SyncHandler) SyncProductFromLoyverse(c *gin.Context) {
	loyverseProductID := c.Param("loyverse_id")
	if loyverseProductID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "loyverse_id parameter is required",
		})
		return
	}

	if h.loyverseSync == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Loyverse integration not configured",
		})
		return
	}

	h.logger.WithField("loyverse_product_id", loyverseProductID).Info("Starting single product sync from Loyverse")

	err := h.loyverseSync.SyncProduct(c.Request.Context(), loyverseProductID)
	if err != nil {
		h.logger.WithError(err).WithField("loyverse_product_id", loyverseProductID).Error("Failed to sync product from Loyverse")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to sync product from Loyverse",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product sync completed",
		"loyverse_product_id": loyverseProductID,
	})
}

// GetSyncStatus returns sync status
func (h *SyncHandler) GetSyncStatus(c *gin.Context) {
	syncID := c.Param("sync_id")
	if syncID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "sync_id parameter is required",
		})
		return
	}

	result, err := h.syncUsecase.GetSyncStatus(c.Request.Context(), syncID)
	if err != nil {
		h.logger.WithError(err).WithField("sync_id", syncID).Error("Failed to get sync status")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get sync status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetLastSyncTime returns last sync time for a type
func (h *SyncHandler) GetLastSyncTime(c *gin.Context) {
	syncType := c.Query("type")
	if syncType == "" {
		syncType = "loyverse_products"
	}

	lastSyncTime, err := h.syncUsecase.GetLastSyncTime(c.Request.Context(), syncType)
	if err != nil {
		h.logger.WithError(err).WithField("sync_type", syncType).Error("Failed to get last sync time")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get last sync time",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sync_type": syncType,
		"last_sync_time": lastSyncTime,
	})
}