// Package handler 处理 HTTP 请求。
package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"roadside-league/backend/internal/middleware"
	"roadside-league/backend/internal/model"
	"roadside-league/backend/internal/response"
	"roadside-league/backend/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type loginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, 400, "请求参数格式错误")
		return
	}

	u, token, err := h.auth.Login(req.Account, req.Password)

	var ve *service.ValidationError
	switch {
	case errors.As(err, &ve):
		response.Fail(c, 400, 400, err.Error())
		return
	case errors.Is(err, service.ErrBadCredentials):
		response.Fail(c, 401, 401, "账号或密码错误")
		return
	case errors.Is(err, service.ErrUserDisabled):
		response.Fail(c, 403, 403, "账号已被禁用，请联系管理员")
		return
	case err != nil:
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}

	response.OK(c, gin.H{"token": token, "user": userVO(u)})
}

func (h *AuthHandler) Me(c *gin.Context) {
	u, err := h.auth.GetUser(middleware.UserID(c))
	if errors.Is(err, service.ErrUserNotFound) {
		response.Fail(c, 404, 404, "用户不存在")
		return
	}
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"user": userVO(u)})
}

// userVO 对外用户结构（不含密码等内部字段）。
func userVO(u *model.User) gin.H {
	return gin.H{
		"id":        u.ID,
		"username":  u.Username,
		"email":     u.Email,
		"avatar":    u.Avatar,
		"role":      u.Role,
		"createdAt": u.CreatedAt,
	}
}
