package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-web3-dapp-backend/services/indexer/models"
)

func (h *Handler) SyncStatus(c *gin.Context) {
	var s models.SyncStatus
	if err := h.DB.First(&s, 1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"last_block": s.LastBlock})
}
