package model

import "time"

// Registration 体检登记/导检。
type Registration struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ExamineeID     uint      `gorm:"index;not null" json:"examinee_id"`
	PackageID      uint      `gorm:"index;not null" json:"package_id"`
	GuideNo        string    `gorm:"size:32;uniqueIndex;not null" json:"guide_no"`
	Status         string    `gorm:"size:20;default:registered" json:"status"`
	RegisteredAt   time.Time `json:"registered_at"`
	RegisterUserID uint      `json:"register_user_id"`
	CreatedAt      time.Time `json:"created_at"`
	Examinee       Examinee  `gorm:"foreignKey:ExamineeID" json:"examinee,omitempty"`
	Package        Package   `gorm:"foreignKey:PackageID" json:"package,omitempty"`
}
