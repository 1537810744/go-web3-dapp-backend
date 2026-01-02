package models

type SyncStatus struct {
	ID        uint `gorm:"primaryKey"`
	LastBlock uint64
}
