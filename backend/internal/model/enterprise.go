package model

import "time"

// Enterprise 团检企业。
type Enterprise struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Contact   string    `gorm:"size:50" json:"contact"`
	Phone     string    `gorm:"size:20" json:"phone"`
	Address   string    `gorm:"size:200" json:"address"`
	CreatedAt time.Time `json:"created_at"`
}
