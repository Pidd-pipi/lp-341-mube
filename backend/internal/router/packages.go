package router

import (
	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerPackageRoutes 套餐路由（管理员/前台）。
func registerPackageRoutes(g *gin.RouterGroup, h Handlers) {
	pkg := g.Group("/packages", middleware.RequireRole(constants.RoleAdmin, constants.RoleFrontDesk))
	pkg.POST("", h.Package.Create)
	pkg.GET("", h.Package.List)
	pkg.GET("/:id", h.Package.Detail)
	pkg.PUT("/:id", h.Package.Update)
	pkg.POST("/:id/items", h.Package.AddItem)
	pkg.GET("/:id/items", h.Package.ListItems)
	pkg.PUT("/:id/items/:itemId", h.Package.UpdateItem)
	pkg.DELETE("/:id/items/:itemId", h.Package.DeleteItem)
}
