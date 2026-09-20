package model

// 显式指定表名，确保与 database/init.sql 中的建表名一致。

func (User) TableName() string            { return "users" }
func (Package) TableName() string         { return "packages" }
func (PackageItem) TableName() string     { return "package_items" }
func (Examinee) TableName() string        { return "examinees" }
func (Registration) TableName() string    { return "registrations" }
func (ExamResult) TableName() string      { return "exam_results" }
func (Report) TableName() string          { return "reports" }
func (AbnormalMetric) TableName() string  { return "abnormal_metrics" }
func (Enterprise) TableName() string      { return "enterprises" }
func (GroupOrder) TableName() string      { return "group_orders" }
