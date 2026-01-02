package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-web3-dapp-backend/services/indexer/models"
)

func (h *Handler) Stats(c *gin.Context) {
	address := c.Query("address")
	var totalDeposit, totalWithdraw string
	db := h.DB
	if address != "" {
		db.Model(&models.Transaction{}).
			Where("user_address = ? AND event_type = ?", address, "deposit").
			Select("SUM(amount)").Row().Scan(&totalDeposit)
		db.Model(&models.Transaction{}).
			Where("user_address = ? AND event_type = ?", address, "withdraw").
			Select("SUM(amount)").Row().Scan(&totalWithdraw)
	} else {
		db.Model(&models.Transaction{}).
			Where("event_type = ?", "deposit").
			Select("SUM(amount)").Row().Scan(&totalDeposit)
		db.Model(&models.Transaction{}).
			Where("event_type = ?", "withdraw").
			Select("SUM(amount)").Row().Scan(&totalWithdraw)
	}
	c.JSON(http.StatusOK, gin.H{"total_deposit": totalDeposit, "total_withdraw": totalWithdraw})
}
