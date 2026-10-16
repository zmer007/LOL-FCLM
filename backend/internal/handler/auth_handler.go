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

type sendCodeRequest struct {
	Email string `json:"email"`
}

// SendCode 发送邮箱注册验证码。
func (h *AuthHandler) SendCode(c *gin.Context) {
	var req sendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, 400, "请求参数格式错误")
		return
	}
	err := h.auth.SendVerificationCode(req.Email)
	var ve *service.ValidationError
	switch {
	case errors.As(err, &ve):
		response.Fail(c, 400, 400, err.Error())
		return
	case errors.Is(err, service.ErrCodeRateLimited):
		response.Fail(c, 429, 429, "验证码发送过于频繁，请稍后再试")
		return
	case errors.Is(err, service.ErrMailNotConfigured):
		response.Fail(c, 503, 503, "邮件服务未配置，请联系管理员")
		return
	case err != nil:
		response.Fail(c, 500, 500, "验证码发送失败，请稍后重试")
		return
	}
	response.OK(c, gin.H{"msg": "验证码已发送，请注意查收"})
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, 400, "请求参数格式错误")
		return
	}

	u, token, err := h.auth.Register(req.Username, req.Email, req.Password, req.Code)

	var ve *service.ValidationError
	switch {
	case errors.As(err, &ve):
		response.Fail(c, 400, 400, err.Error())
		return
	case errors.Is(err, service.ErrUsernameTaken):
		response.Fail(c, 409, 409, "用户名已被占用")
		return
	case errors.Is(err, service.ErrEmailTaken):
		response.Fail(c, 409, 409, "邮箱已被注册")
		return
	case errors.Is(err, service.ErrInvalidCode):
		response.Fail(c, 400, 400, "验证码错误或已过期")
		return
	case err != nil:
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}

	response.OK(c, gin.H{"token": token, "user": userVO(u)})
}

type loginRequest struct {
	Account  string `json:"account"`
	Username string `json:"username"` // 兼容旧客户端
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, 400, "请求参数格式错误")
		return
	}
	account := req.Account
	if account == "" {
		account = req.Username
	}

	u, token, err := h.auth.Login(account, req.Password)

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
