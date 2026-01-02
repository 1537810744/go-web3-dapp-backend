package main

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"go-web3-dapp-backend/services/transactor/blockchain"
	"go-web3-dapp-backend/services/transactor/handlers"
	"go-web3-dapp-backend/services/transactor/middleware"
	"go-web3-dapp-backend/services/transactor/models"
)

func main() {
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Error reading config file:", err)
	}

	dbUser := viper.GetString("database.username")
	dbPass := viper.GetString("database.password")
	dbHost := viper.GetString("database.host")
	dbPort := viper.GetInt("database.port")
	dbName := viper.GetString("database.dbname")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
		dbUser, dbPass, dbHost, dbPort, dbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}
	db.AutoMigrate(&models.User{})

	rpc := viper.GetString("ethereum.rpc_url")
	client, err := ethclient.Dial(rpc)
	if err != nil {
		log.Fatal("failed to connect to ethereum rpc:", err)
	}

	contractAddr := viper.GetString("ethereum.contract_address")
	if err := blockchain.InitContract(contractAddr); err != nil {
		log.Fatal("failed to init contract:", err)
	}

	h := handlers.NewHandler(db, client)

	r := gin.Default()
	api := r.Group("/api/v1")
	api.POST("/register", h.Register)
	api.POST("/login", h.Login)

	auth := api.Group("/")
	auth.Use(middleware.AuthMiddleware(viper.GetString("security.jwt_secret")))
	auth.GET("/balance", h.GetBalance)
	auth.POST("/invest", h.Invest)
	auth.POST("/withdraw", h.Withdraw)

	r.Run(":8080")
}
