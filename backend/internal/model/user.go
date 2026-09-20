package model

import "time"

// User 用户：管理员/医生/前台/体检人。
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Phone        string    `gorm:"size:20;uniqueIndex;not null" json:"phone"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Name         string    `gorm:"size:50" json:"name"`
	Avatar       string    `gorm:"size:255" json:"avatar"`
	Role         string    `gorm:"size:20;default:examinee" json:"role"`
	Department   string    `gorm:"size:50" json:"department"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
