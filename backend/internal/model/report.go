package model

import "time"

// Report 体检报告。
type Report struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	RegistrationID   uint       `gorm:"index;not null" json:"registration_id"`
	ExamineeID       uint       `gorm:"index;not null" json:"examinee_id"`
	ReportNo         string     `gorm:"size:32;uniqueIndex;not null" json:"report_no"`
	Status           string     `gorm:"size:20;default:draft" json:"status"`
	Conclusion       string     `gorm:"type:text" json:"conclusion"`
	HealthAdvice     string     `gorm:"type:text" json:"health_advice"`
	FollowUpReminder string     `gorm:"size:500" json:"follow_up_reminder"`
	PDFURL           string     `gorm:"size:255" json:"pdf_url"`
	DoctorID         uint       `json:"doctor_id"`
	GeneratedAt      *time.Time `json:"generated_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	Examinee         Examinee   `gorm:"foreignKey:ExamineeID" json:"examinee,omitempty"`
}
