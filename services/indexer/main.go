package main

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"go-web3-dapp-backend/services/indexer/handlers"
	"go-web3-dapp-backend/services/indexer/listener"
	"go-web3-dapp-backend/services/indexer/models"
)

func main() {
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
	db.AutoMigrate(&models.Transaction{}, &models.SyncStatus{})

	ws := viper.GetString("ethereum.ws_url")
	client, err := ethclient.Dial(ws)
	if err != nil {
		log.Fatal(err)
	}

	contractAddr := viper.GetString("ethereum.contract_address")
	listener.InitContract(contractAddr)

	go listener.StartListener(client, db)

	r := gin.Default()
	api := r.Group("/api/v1")
	h := handlers.NewHandler(db)
	api.GET("/history", h.History)
	api.GET("/stats", h.Stats)
	api.GET("/sync-status", h.SyncStatus)
	r.Run(":8081")
}
