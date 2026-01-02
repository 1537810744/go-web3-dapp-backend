package handlers

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

type Handler struct {
	DB     *gorm.DB
	Client *ethclient.Client
}

func NewHandler(db *gorm.DB, client *ethclient.Client) *Handler {
	return &Handler{DB: db, Client: client}
}
