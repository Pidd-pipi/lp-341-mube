package dto

// RegisterRequest 注册请求。
type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required,len=11"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required,max=50"`
}

// LoginRequest 登录请求。
type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest 更新资料请求。
type UpdateProfileRequest struct {
	Name       string `json:"name" binding:"max=50"`
	Avatar     string `json:"avatar" binding:"max=255"`
	Department string `json:"department" binding:"max=50"`
}

// PackageRequest 套餐请求。
type PackageRequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	PackageType string  `json:"package_type" binding:"required,oneof=entry annual premium other"`
	Price       float64 `json:"price" binding:"gte=0"`
	Status      string  `json:"status" binding:"oneof=active inactive"`
	Description string  `json:"description" binding:"max=500"`
}

// PackageItemRequest 检查项目请求。
type PackageItemRequest struct {
	ItemName      string `json:"item_name" binding:"required,max=100"`
	ItemGroup     string `json:"item_group" binding:"max=50"`
	RefValueRange string `json:"ref_value_range" binding:"max=100"`
	Department    string `json:"department" binding:"max=50"`
	SortOrder     int    `json:"sort_order"`
}

// ExamineeRequest 体检人请求。
type ExamineeRequest struct {
	Name         string `json:"name" binding:"required,max=50"`
	IDCardNo     string `json:"id_card_no" binding:"required,max=18"`
	Phone        string `json:"phone" binding:"max=20"`
	Gender       string `json:"gender" binding:"oneof=male female"`
	Age          int    `json:"age" binding:"gte=0,lte=150"`
	SourceType   string `json:"source_type" binding:"oneof=personal group"`
	EnterpriseID *uint  `json:"enterprise_id"`
}

// BatchImportRequest 团体批量导入请求。
type BatchImportRequest struct {
	EnterpriseID *uint  `json:"enterprise_id"`
	CSVText      string `json:"csv_text" binding:"required"`
}

// RegisterRequestDTO 登记请求。
type RegisterRequestDTO struct {
	ExamineeID uint `json:"examinee_id" binding:"required"`
	PackageID  uint `json:"package_id" binding:"required"`
}

// UpdateStatusRequest 状态更新请求。
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// EnterResultRequest 结果录入请求。
type EnterResultRequest struct {
	ResultValue string `json:"result_value"`
	ResultText  string `json:"result_text"`
	ImageURL    string `json:"image_url" binding:"max=255"`
}

// ReportContentRequest 报告内容请求。
type ReportContentRequest struct {
	Conclusion       string `json:"conclusion"`
	HealthAdvice     string `json:"health_advice"`
	FollowUpReminder string `json:"follow_up_reminder"`
}

// EnterpriseRequest 企业请求。
type EnterpriseRequest struct {
	Name    string `json:"name" binding:"required,max=100"`
	Contact string `json:"contact" binding:"max=50"`
	Phone   string `json:"phone" binding:"max=20"`
	Address string `json:"address" binding:"max=200"`
}

// GroupOrderRequest 团检订单请求。
type GroupOrderRequest struct {
	EnterpriseID  uint `json:"enterprise_id" binding:"required"`
	PackageID     uint `json:"package_id" binding:"required"`
	ExamineeCount int  `json:"examinee_count" binding:"gte=1"`
}

// FollowUpRequest 复查跟踪请求。
type FollowUpRequest struct {
	Status  string `json:"status" binding:"required,oneof=pending done"`
	Advice  string `json:"specialist_advice" binding:"max=500"`
}

// TokenResponse 令牌响应。
type TokenResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}
