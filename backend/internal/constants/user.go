package constants

// UserRole 用户角色枚举。
const (
	RoleAdmin     = "admin"      // 管理员
	RoleDoctor    = "doctor"     // 医生
	RoleFrontDesk = "front_desk" // 前台
	RoleExaminee  = "examinee"   // 体检人
)

// Roles 全部角色。
var Roles = []string{RoleAdmin, RoleDoctor, RoleFrontDesk, RoleExaminee}
