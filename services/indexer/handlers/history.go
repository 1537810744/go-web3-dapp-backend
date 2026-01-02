package handlers

import (
	"net/http"
	"strconv"

	"go-web3-dapp-backend/services/indexer/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) History(c *gin.Context) {
	address := c.Query("address")
	eventType := c.Query("type")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	var txs []models.Transaction
	query := h.DB
	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if address != "" {
		query = query.Where("user_address = ? OR to_address = ?", address, address)
	}
	query.Limit(limit).Offset(offset).Order("id desc").Find(&txs)
	c.JSON(http.StatusOK, txs)
}
