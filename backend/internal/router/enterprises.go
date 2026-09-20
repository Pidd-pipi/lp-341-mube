package router

import (
	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerEnterpriseRoutes 团检路由。
func registerEnterpriseRoutes(g *gin.RouterGroup, h Handlers) {
	ent := g.Group("/enterprises", middleware.RequireRole(constants.RoleAdmin, constants.RoleFrontDesk))
	ent.POST("", h.Enterprise.CreateEnterprise)
	ent.GET("", h.Enterprise.ListEnterprises)
	ent.POST("/orders", h.Enterprise.CreateOrder)
	ent.GET("/orders", h.Enterprise.ListOrders)
	ent.POST("/orders/:id/deliver", h.Enterprise.DeliverReports)
}
