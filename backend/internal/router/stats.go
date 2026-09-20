package router

import (
	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerStatsRoutes 统计路由（管理员）。
func registerStatsRoutes(g *gin.RouterGroup, h Handlers) {
	g.GET("/stats/dashboard", middleware.RequireRole(constants.RoleAdmin), h.Stats.Dashboard)
}
