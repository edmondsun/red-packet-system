package model

import "time"

type RedPacket struct {
	ID              uint      `gorm:"primaryKey"`
	TotalAmount     float64   `gorm:"not null"`
	RemainingAmount float64   `gorm:"not null"`
	TotalCount      int       `gorm:"not null"`
	RemainingCount  int       `gorm:"not null"`
	Status          int       `gorm:"default:1"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}
