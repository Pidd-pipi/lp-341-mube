package router

import (
	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerAbnormalMetricRoutes 异常指标路由。
func registerAbnormalMetricRoutes(g *gin.RouterGroup, h Handlers) {
	abnormal := g.Group("/abnormal-metrics", middleware.RequireRole(constants.RoleAdmin, constants.RoleDoctor, constants.RoleExaminee))
	abnormal.GET("", h.Abnormal.List)
	abnormal.PUT("/:id/follow-up", h.Abnormal.UpdateFollowUp)
}
