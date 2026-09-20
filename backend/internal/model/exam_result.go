package model

import "time"

// ExamResult 检查结果。
type ExamResult struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	RegistrationID uint       `gorm:"index;not null" json:"registration_id"`
	ExamineeID     uint       `gorm:"index;not null" json:"examinee_id"`
	PackageItemID  uint       `gorm:"index;not null" json:"package_item_id"`
	ResultValue    string     `gorm:"size:100" json:"result_value"`
	ResultText     string     `gorm:"size:500" json:"result_text"`
	IsAbnormal     bool       `gorm:"default:false" json:"is_abnormal"`
	Status         string     `gorm:"size:20;default:pending" json:"status"`
	ImageURL       string     `gorm:"size:255" json:"image_url"`
	DoctorID       uint       `json:"doctor_id"`
	EnteredAt      *time.Time `json:"entered_at"`
	CreatedAt      time.Time  `json:"created_at"`
	PackageItem    PackageItem `gorm:"foreignKey:PackageItemID" json:"package_item,omitempty"`
}
