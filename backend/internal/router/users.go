package router

import (
	"github.com/gin-gonic/gin"
)

// registerUserRoutes 用户路由。
func registerUserRoutes(g *gin.RouterGroup, h Handlers) {
	g.GET("/users/me", h.User.Me)
	g.PUT("/users/me", h.User.UpdateProfile)
}
