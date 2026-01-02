package models

import "time"

type Transaction struct {
	ID          uint `gorm:"primaryKey"`
	TxHash      string
	EventType   string
	UserAddress string
	ToAddress   string
	Amount      string
	BlockNumber uint64
	CreatedAt   time.Time
}
