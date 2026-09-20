package model

import "time"

// ReportWithdraw 报告撤回重签申请。
// 同一报告只能存在一份 status=pending 的申请（部分唯一索引保证）；
// 申请期间对应 Report 状态为 withdrawing，冻结下载与新版本生成。
type ReportWithdraw struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	ReportID      uint       `gorm:"index;not null" json:"report_id"`
	ApplicantID   uint       `gorm:"not null" json:"applicant_id"`
	Reason        string     `gorm:"type:text" json:"reason"`
	Status        string     `gorm:"size:20;default:pending" json:"status"`
	ReviewerID    uint       `json:"reviewer_id"`
	ReviewComment string     `gorm:"size:500" json:"review_comment"`
	ExpiresAt     time.Time  `gorm:"index" json:"expires_at"`
	ReviewedAt    *time.Time `json:"reviewed_at"`
	// NewReportID 批准时生成的待重签版本报告 ID；驳回时为 0。
	NewReportID uint      `gorm:"index" json:"new_report_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Report      *Report   `gorm:"foreignKey:ReportID" json:"report,omitempty"`
}
