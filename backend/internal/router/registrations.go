package router

import (
	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerRegistrationRoutes 登记路由。
func registerRegistrationRoutes(g *gin.RouterGroup, h Handlers) {
	reg := g.Group("/registrations", middleware.RequireRole(constants.RoleAdmin, constants.RoleFrontDesk, constants.RoleDoctor))
	reg.POST("", h.Registration.Register)
	reg.GET("", h.Registration.List)
	reg.GET("/:id", h.Registration.Detail)
	reg.PUT("/:id/status", h.Registration.UpdateStatus)
}
