package model

import "time"

// Report 体检报告（含撤回重签版本链字段）。
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
	// 撤回重签版本链：version_no 从 1 起递增；root_report_id 指向版本链首份（原版）；
	// parent_report_id 指向本版本直接来源的已撤回原版；首版三者为 0/1/0。
	VersionNo      uint       `gorm:"default:1" json:"version_no"`
	RootReportID   uint       `gorm:"index" json:"root_report_id"`
	ParentReportID uint       `gorm:"index" json:"parent_report_id"`
	PublishedAt    *time.Time `json:"published_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Examinee       Examinee   `gorm:"foreignKey:ExamineeID" json:"examinee,omitempty"`
}
