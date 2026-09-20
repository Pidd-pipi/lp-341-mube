package model

import "time"

// GroupOrder 团检订单。
type GroupOrder struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	EnterpriseID        uint      `gorm:"index;not null" json:"enterprise_id"`
	PackageID           uint      `gorm:"index;not null" json:"package_id"`
	ExamineeCount       int       `json:"examinee_count"`
	Status              string    `gorm:"size:20;default:pending" json:"status"`
	ReportDeliveryStatus string   `gorm:"size:20;default:pending" json:"report_delivery_status"`
	CreatedAt           time.Time `json:"created_at"`
	Enterprise          Enterprise `gorm:"foreignKey:EnterpriseID" json:"enterprise,omitempty"`
	Package             Package    `gorm:"foreignKey:PackageID" json:"package,omitempty"`
}
