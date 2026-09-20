package model

import "time"

// Package 体检套餐。
type Package struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	PackageType string    `gorm:"size:20;not null" json:"package_type"`
	Price       float64   `json:"price"`
	Status      string    `gorm:"size:20;default:active" json:"status"`
	Description string    `gorm:"size:500" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
