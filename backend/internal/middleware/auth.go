package middleware

import (
	"net/http"
	"strings"

	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/util"
	"github.com/gin-gonic/gin"
)

const (
	UserKey   = "user"
	RoleKey   = "userRole"
	UserIDKey = "userID"
)

// AuthRequired 校验 Bearer JWT。
func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			abortJSON(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgLoginFailed)
			return
		}
		claims, err := util.ParseToken(jwtSecret, strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			abortJSON(c, http.StatusUnauthorized, constants.CodeUnauthorized, "登录状态已失效（User token）")
			return
		}
		c.Set(UserIDKey, claims.UserID)
		c.Set(RoleKey, claims.Role)
		c.Set(UserKey, claims)
		c.Next()
	}
}

func abortJSON(c *gin.Context, status, code int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"code": code, "message": message, "data": nil})
}
