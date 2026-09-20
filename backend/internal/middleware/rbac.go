package middleware

import (
	"net/http"

	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/gin-gonic/gin"
)

// RequireRole 角色权限校验。
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role, _ := c.Get(RoleKey)
		roleStr, _ := role.(string)
		if !allowed[roleStr] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": constants.CodeForbidden, "message": constants.MsgRoleForbidden, "data": nil,
			})
			return
		}
		c.Next()
	}
}
