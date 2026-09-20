package router

import (
	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerExamResultRoutes 检查结果路由（医生/管理员）。
func registerExamResultRoutes(g *gin.RouterGroup, h Handlers) {
	result := g.Group("/exam-results", middleware.RequireRole(constants.RoleAdmin, constants.RoleDoctor))
	result.GET("/pending", h.ExamResult.ListPending)
	result.GET("", h.ExamResult.ListByRegistration)
	result.POST("/:id/enter", h.ExamResult.Enter)
	result.POST("/:id/review", h.ExamResult.Review)
}
