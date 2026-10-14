package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"roadside-league/backend/internal/middleware"
	"roadside-league/backend/internal/model"
	"roadside-league/backend/internal/response"
	"roadside-league/backend/internal/service"
)

type AdminUserHandler struct {
	users *service.AdminUserService
}

func NewAdminUserHandler(users *service.AdminUserService) *AdminUserHandler {
	return &AdminUserHandler{users: users}
}

func (h *AdminUserHandler) List(c *gin.Context) {
	list, err := h.users.List()
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"list": list})
}

type addUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (h *AdminUserHandler) Add(c *gin.Context) {
	var req addUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, 400, "请求参数格式错误")
		return
	}
	u, err := h.users.Add(req.Username, req.Password, req.Role)
	if ve := asValidation(err); ve != nil {
		response.Fail(c, 400, 400, ve.Msg)
		return
	}
	if errors.Is(err, service.ErrUsernameTaken) {
		response.Fail(c, 409, 409, "用户名已被占用")
		return
	}
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"user": adminUserVO(u)})
}

type statusRequest struct {
	Status string `json:"status"`
}

func (h *AdminUserHandler) SetStatus(c *gin.Context) {
	id, ok := h.pathID(c)
	if !ok {
		return
	}
	var req statusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, 400, "请求参数格式错误")
		return
	}
	h.guardSelf(c, id, "不能对自己的账号执行该操作")
	if c.IsAborted() {
		return
	}
	h.exec(c, h.users.SetStatus(id, req.Status))
}

type roleRequest struct {
	Role string `json:"role"`
}

func (h *AdminUserHandler) SetRole(c *gin.Context) {
	id, ok := h.pathID(c)
	if !ok {
		return
	}
	var req roleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, 400, "请求参数格式错误")
		return
	}
	h.guardSelf(c, id, "不能修改自己的角色")
	if c.IsAborted() {
		return
	}
	h.exec(c, h.users.SetRole(id, req.Role))
}

type passwordRequest struct {
	Password string `json:"password"`
}

func (h *AdminUserHandler) ResetPassword(c *gin.Context) {
	id, ok := h.pathID(c)
	if !ok {
		return
	}
	var req passwordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, 400, "请求参数格式错误")
		return
	}
	h.exec(c, h.users.ResetPassword(id, req.Password))
}

func (h *AdminUserHandler) Delete(c *gin.Context) {
	id, ok := h.pathID(c)
	if !ok {
		return
	}
	h.guardSelf(c, id, "不能删除自己的账号")
	if c.IsAborted() {
		return
	}
	h.exec(c, h.users.Delete(id))
}

// ---------- 辅助 ----------

func (h *AdminUserHandler) pathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, 400, "用户 ID 不合法")
		return 0, false
	}
	return id, true
}

// guardSelf 防止管理员对自己执行禁用/降级/删除，避免锁死后台。
func (h *AdminUserHandler) guardSelf(c *gin.Context, id int64, msg string) {
	if middleware.UserID(c) == id {
		response.Fail(c, 400, 400, msg)
		c.Abort()
	}
}

func (h *AdminUserHandler) exec(c *gin.Context, err error) {
	if c.IsAborted() {
		return
	}
	if ve := asValidation(err); ve != nil {
		response.Fail(c, 400, 400, ve.Msg)
		return
	}
	if errors.Is(err, service.ErrUserNotFound) {
		response.Fail(c, 404, 404, "用户不存在")
		return
	}
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func adminUserVO(u *model.User) gin.H {
	return gin.H{
		"id":        u.ID,
		"username":  u.Username,
		"role":      u.Role,
		"status":    u.Status,
		"createdAt": u.CreatedAt,
	}
}
