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
		// 实时校验账号状态：被禁用后即使 token 未过期也立即失效
		if !auth.IsActive(claims.UserID) {
			response.Fail(c, 403, 403, "账号已被禁用")
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

// RequireAdmin 在已登录基础上要求 role=admin。
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAdmin(c) {
			response.Fail(c, 403, 403, "需要管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

// IsAdmin 从上下文判断当前请求是否来自管理员（公开接口里用于扩展权限，不拦截请求）。
func IsAdmin(c *gin.Context) bool {
	role, _ := c.Get(CtxRole)
	return role == "admin"
}

// OptionalAuth 有合法 token 就写入用户信息；没有或失效也放行（公开接口识别本人用）。
func OptionalAuth(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		if token != "" && token != header {
			if claims, err := auth.ParseToken(token); err == nil && auth.IsActive(claims.UserID) {
				c.Set(CtxUserID, claims.UserID)
				c.Set(CtxUsername, claims.Username)
				c.Set(CtxRole, claims.Role)
			}
		}
		c.Next()
	}
}
