package handlers

import (
	"context"
	"math/big"
	"net/http"
	"strings"
	"time"

	"go-web3-dapp-backend/services/transactor/models"
	"go-web3-dapp-backend/services/transactor/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	// create ethereum key
	privKey, err := crypto.GenerateKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate key"})
		return
	}
	privBytes := crypto.FromECDSA(privKey)
	privHex := common.Bytes2Hex(privBytes)
	encKey := viper.GetString("security.encryption_key")
	encryptedKey, err := utils.Encrypt(privHex, encKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "encryption failed"})
		return
	}
	address := crypto.PubkeyToAddress(privKey.PublicKey).Hex()

	user := models.User{Username: req.Username, Password: string(hashed), EthAddress: address, EthPrivateKey: encryptedKey}
	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed"})
		return
	}

	// fund from faucet (best-effort, non-blocking)
	faucetKeyHex := viper.GetString("dev.faucet_private_key")
	if faucetKeyHex != "" && faucetKeyHex != "0x..." {
		go func() {
			client := h.Client
			// convert amount
			amtStr := viper.GetString("dev.auto_fund_amount")
			f := new(big.Float)
			f.SetString(amtStr)
			f.Mul(f, big.NewFloat(1e18))
			wei := new(big.Int)
			f.Int(wei)

			faucetKey, _ := crypto.HexToECDSA(strings.TrimPrefix(faucetKeyHex, "0x"))
			fromAddr := crypto.PubkeyToAddress(faucetKey.PublicKey)
			nonce, _ := client.PendingNonceAt(context.Background(), fromAddr)
			gasPrice, _ := client.SuggestGasPrice(context.Background())
			tx := types.NewTransaction(nonce, common.HexToAddress(address), wei, 21000, gasPrice, nil)
			signedTx, _ := types.SignTx(tx, types.LatestSignerForChainID(big.NewInt(int64(viper.GetInt("ethereum.chain_id")))), faucetKey)
			_ = client.SendTransaction(context.Background(), signedTx)
		}()
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully", "address": address})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	var user models.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(viper.GetString("security.jwt_secret")))
	c.JSON(http.StatusOK, gin.H{"token": tokenString, "address": user.EthAddress})
}
