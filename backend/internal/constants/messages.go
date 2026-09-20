package constants

// messages.go 集中管理后端返回文案与日志文案（屎山耦合点）。

const (
	MsgOK               = "ok"
	MsgLoginSuccess     = "登录成功"
	MsgRegisterSuccess  = "注册成功"
	MsgCreated          = "创建成功"
	MsgUpdated          = "更新成功"
	MsgDeleted          = "删除成功"
	MsgRegistered       = "体检登记成功"
	MsgResultEntered    = "检查结果录入成功"
	MsgReportGenerated  = "报告生成成功"
	MsgReportPublished  = "报告发布成功"
	MsgImported         = "团体批量导入成功"
	MsgAbnormalTracked  = "异常指标已记录"
	MsgUserNotFound     = "用户（User）不存在"
	MsgDuplicatePhone   = "手机号（User.phone）已注册"
	MsgLoginFailed      = "手机号或密码（User）错误"
	MsgRoleForbidden    = "当前角色（UserRole）无权执行该操作"
	MsgPackageNotFound  = "体检套餐（Package）不存在"
	MsgExamineeNotFound = "体检人（Examinee）不存在"
	MsgRegNotFound      = "体检登记（Registration）不存在"
	MsgResultNotFound   = "检查结果（ExamResult）不存在"
	MsgReportNotFound   = "体检报告（Report）不存在"
	MsgReportStatusInvalid = "报告状态（Report.status）流转不合法"
	MsgAbnormalLevelInvalid = "异常等级（AbnormalLevel）不合法"
	MsgInternalError    = "服务内部错误"
	MsgParamInvalid     = "请求参数校验失败"
	MsgRateLimited      = "请求过于频繁，请稍后重试"
)
