package router

import (
	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerExamineeRoutes 体检人路由（前台/医生/管理员）。
func registerExamineeRoutes(g *gin.RouterGroup, h Handlers) {
	examinee := g.Group("/examinees", middleware.RequireRole(constants.RoleAdmin, constants.RoleFrontDesk, constants.RoleDoctor))
	examinee.POST("", h.Examinee.Create)
	examinee.GET("", h.Examinee.List)
	examinee.GET("/:id", h.Examinee.Detail)
	examinee.POST("/batch-import", h.Examinee.BatchImport)
}
