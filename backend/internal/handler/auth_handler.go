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

type credentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	h.authenticate(c, false)
}

func (h *AuthHandler) Login(c *gin.Context) {
	h.authenticate(c, true)
}

func (h *AuthHandler) authenticate(c *gin.Context, isLogin bool) {
	var req credentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, 400, "请求参数格式错误")
		return
	}

	var (
		u     *model.User
		token string
		err   error
	)
	if isLogin {
		u, token, err = h.auth.Login(req.Username, req.Password)
	} else {
		u, token, err = h.auth.Register(req.Username, req.Password)
	}

	var ve *service.ValidationError
	switch {
	case errors.As(err, &ve):
		response.Fail(c, 400, 400, err.Error())
		return
	case errors.Is(err, service.ErrUsernameTaken):
		response.Fail(c, 409, 409, "用户名已被占用")
		return
	case errors.Is(err, service.ErrBadCredentials):
		response.Fail(c, 401, 401, "用户名或密码错误")
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
		"avatar":    u.Avatar,
		"role":      u.Role,
		"createdAt": u.CreatedAt,
	}
}
