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
	report.GET("/:id", h.Report.Detail)
	report.POST("/:id/generate", h.Report.Generate)
	report.POST("/:id/review", h.Report.Review)
	report.POST("/:id/publish", h.Report.Publish)
	report.GET("/:id/pdf", h.Report.DownloadPDF)

	// 撤回重签闭环：申请与版本链医生/管理员可用，审批仅管理员。
	report.POST("/:id/withdraw", h.ReportWithdraw.Request)
	report.GET("/:id/versions", h.ReportWithdraw.VersionChain)
	review := report.Group("/withdrawals/:requestId", middleware.RequireRole(constants.RoleAdmin))
	review.POST("/approve", h.ReportWithdraw.Approve)
	review.POST("/reject", h.ReportWithdraw.Reject)
}
