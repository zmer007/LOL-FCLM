// Package middleware 提供 Gin 中间件。
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"roadside-league/backend/internal/response"
	"roadside-league/backend/internal/service"
)

// 上下文键
const (
	CtxUserID   = "uid"
	CtxUsername = "username"
	CtxRole     = "role"
)

// RequiredAuth 校验 Authorization: Bearer <token>，通过后写入用户信息。
func RequiredAuth(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		if token == "" || token == header {
			response.Fail(c, 401, 401, "未登录")
			c.Abort()
			return
		}
		claims, err := auth.ParseToken(token)
		if err != nil {
			response.Fail(c, 401, 401, "登录已过期，请重新登录")
			c.Abort()
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxUsername, claims.Username)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}

// UserID 从上下文取当前登录用户 ID。
func UserID(c *gin.Context) int64 {
	if v, ok := c.Get(CtxUserID); ok {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}
