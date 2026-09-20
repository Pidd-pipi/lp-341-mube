package model

import "time"

// ReportWithdrawRequest 报告撤回重签申请。
// 同一报告同一时刻只允许存在一份 pending 申请（由 service 事务保证）。
type ReportWithdrawRequest struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	ReportID      uint   `gorm:"index;not null" json:"report_id"`
	ApplicantID   uint   `gorm:"not null" json:"applicant_id"`
	Reason        string `gorm:"size:500" json:"reason"`
	Status        string `gorm:"size:20;default:pending" json:"status"` // pending/approved/rejected
	ReviewerID    uint   `gorm:"default:0" json:"reviewer_id"`
	ReviewComment string `gorm:"size:500" json:"review_comment"`
	// 批准撤回后生成的待重签新版本（驳回时为 0）。
	NewReportID uint       `gorm:"default:0" json:"new_report_id"`
	ExpiresAt   time.Time  `gorm:"index;not null" json:"expires_at"`
	ReviewedAt  *time.Time `json:"reviewed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Report      *Report    `gorm:"foreignKey:ReportID" json:"report,omitempty"`
}
