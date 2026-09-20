package model

// DashboardStats 运营统计。
type DashboardStats struct {
	PackageCount      int64         `json:"package_count"`
	RegistrationCount int64         `json:"registration_count"`
	ReportCount       int64         `json:"report_count"`
	AbnormalCount     int64         `json:"abnormal_count"`
	Revenue           float64       `json:"revenue"`
	PackageSold       []NameCount   `json:"package_sold"`
	DeptWorkload      []NameCount   `json:"dept_workload"`
	AbnormalTop       []NameCount   `json:"abnormal_top"`
	MonthlyRevenue    []MonthAmount `json:"monthly_revenue"`
}

// NameCount 名称-数量统计。
type NameCount struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// MonthAmount 月度金额。
type MonthAmount struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
}
