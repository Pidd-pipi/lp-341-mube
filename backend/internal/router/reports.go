package router

import (
	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerReportRoutes 报告路由（医生/管理员）。
func registerReportRoutes(g *gin.RouterGroup, h Handlers) {
	report := g.Group("/reports", middleware.RequireRole(constants.RoleAdmin, constants.RoleDoctor))
	report.POST("/draft", h.Report.Draft)
	report.GET("", h.Report.List)
	report.GET("/withdraws", h.Withdraw.List) // 撤回申请列表（审批台，静态路径优先）
	report.GET("/withdraws/:id", h.Withdraw.Detail)
	report.GET("/:id", h.Report.Detail)
	report.POST("/:id/generate", h.Report.Generate)
	report.POST("/:id/review", h.Report.Review)
	report.POST("/:id/publish", h.Report.Publish)
	report.GET("/:id/pdf", h.Report.DownloadPDF)

	// 撤回重签闭环：申请撤回（24h 内、唯一待审批）与版本链。
	report.POST("/:id/withdraw", h.Withdraw.Apply)
	report.GET("/:id/versions", h.Withdraw.Chain)

	// 审批仅管理员；超时/重复/并发审批返回 409 且状态不变。
	adminOnly := g.Group("/reports", middleware.RequireRole(constants.RoleAdmin))
	adminOnly.POST("/withdraws/:id/approve", h.Withdraw.Approve)
	adminOnly.POST("/withdraws/:id/reject", h.Withdraw.Reject)
}
