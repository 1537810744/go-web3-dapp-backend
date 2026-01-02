package models

import "time"

type User struct {
	ID            uint   `gorm:"primaryKey"`
	Username      string `gorm:"unique"`
	Password      string
	EthAddress    string `gorm:"unique"`
	EthPrivateKey string
	CreatedAt     time.Time
}
