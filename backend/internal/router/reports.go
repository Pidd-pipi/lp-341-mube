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
}
