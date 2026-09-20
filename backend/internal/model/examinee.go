package model

import "time"

// Examinee 体检人。
type Examinee struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:50;not null" json:"name"`
	IDCardNo     string    `gorm:"size:18;uniqueIndex;not null" json:"id_card_no"`
	Phone        string    `gorm:"size:20" json:"phone"`
	Gender       string    `gorm:"size:10" json:"gender"`
	Age          int       `json:"age"`
	SourceType   string    `gorm:"size:20;default:personal" json:"source_type"`
	EnterpriseID *uint     `json:"enterprise_id"`
	CreatedAt    time.Time `json:"created_at"`
}
