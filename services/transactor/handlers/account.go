package handlers

import (
	"context"
	"go-web3-dapp-backend/services/transactor/blockchain"
	"go-web3-dapp-backend/services/transactor/models"
	"go-web3-dapp-backend/services/transactor/utils"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type InvestRequest struct {
	Amount string `json:"amount"`
}
type WithdrawRequest struct {
	ToAddress string `json:"to_address"`
	Amount    string `json:"amount"`
}

// GetBalance calls contract view function getDeposit to return user's deposit (wei string)
func (h *Handler) GetBalance(c *gin.Context) {
	uid := c.GetInt64("user_id")
	var user models.User
	if err := h.DB.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}
	addr := common.HexToAddress(user.EthAddress)
	data, _ := blockchain.ContractABI.Pack("getDeposit", addr)
	msg := ethereum.CallMsg{To: &blockchain.ContractAddress, Data: data}
	res, err := h.Client.CallContract(context.Background(), msg, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "contract call failed"})
		return
	}
	outs, _ := blockchain.ContractABI.Unpack("getDeposit", res)
	if len(outs) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response"})
		return
	}
	balance := outs[0].(*big.Int)
	c.JSON(http.StatusOK, gin.H{"balance": balance.String()})
}

func (h *Handler) Invest(c *gin.Context) {
	uid := c.GetInt64("user_id")
	var req InvestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	var user models.User
	if err := h.DB.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}
	privHex, err := utils.Decrypt(user.EthPrivateKey, viper.GetString("security.encryption_key"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "decrypt failed"})
		return
	}
	privBytes := common.FromHex(privHex)
	privKey, err := crypto.ToECDSA(privBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid key"})
		return
	}
	// convert amount to wei
	f := new(big.Float)
	f.SetString(req.Amount)
	f.Mul(f, big.NewFloat(1e18))
	wei := new(big.Int)
	f.Int(wei)

	from := common.HexToAddress(user.EthAddress)
	nonce, _ := h.Client.PendingNonceAt(context.Background(), from)
	gasPrice, _ := h.Client.SuggestGasPrice(context.Background())
	data, _ := blockchain.ContractABI.Pack("deposit")
	tx := types.NewTransaction(nonce, blockchain.ContractAddress, wei, 300000, gasPrice, data)
	signed, _ := types.SignTx(tx, types.LatestSignerForChainID(big.NewInt(int64(viper.GetInt("ethereum.chain_id")))), privKey)
	_ = h.Client.SendTransaction(context.Background(), signed)
	c.JSON(http.StatusOK, gin.H{"message": "investment tx sent"})
}

func (h *Handler) Withdraw(c *gin.Context) {
	uid := c.GetInt64("user_id")
	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	var user models.User
	if err := h.DB.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}
	privHex, err := utils.Decrypt(user.EthPrivateKey, viper.GetString("security.encryption_key"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "decrypt failed"})
		return
	}
	privBytes := common.FromHex(privHex)
	privKey, err := crypto.ToECDSA(privBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid key"})
		return
	}
	f := new(big.Float)
	f.SetString(req.Amount)
	f.Mul(f, big.NewFloat(1e18))
	wei := new(big.Int)
	f.Int(wei)
	to := common.HexToAddress(req.ToAddress)
	data, _ := blockchain.ContractABI.Pack("withdraw", to, wei)
	from := common.HexToAddress(user.EthAddress)
	nonce, _ := h.Client.PendingNonceAt(context.Background(), from)
	gasPrice, _ := h.Client.SuggestGasPrice(context.Background())
	tx := types.NewTransaction(nonce, blockchain.ContractAddress, big.NewInt(0), 300000, gasPrice, data)
	signed, _ := types.SignTx(tx, types.LatestSignerForChainID(big.NewInt(int64(viper.GetInt("ethereum.chain_id")))), privKey)
	_ = h.Client.SendTransaction(context.Background(), signed)
	c.JSON(http.StatusOK, gin.H{"message": "withdraw tx sent"})
}
