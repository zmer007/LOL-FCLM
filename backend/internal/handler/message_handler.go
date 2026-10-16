package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"roadside-league/backend/internal/middleware"
	"roadside-league/backend/internal/response"
	"roadside-league/backend/internal/service"
)

type MessageHandler struct {
	svc *service.MessageService
}

func NewMessageHandler(svc *service.MessageService) *MessageHandler {
	return &MessageHandler{svc: svc}
}

type sendMessageRequest struct {
	Content string `json:"content"`
}

// Send POST /api/messages 普通用户提交留言。
func (h *MessageHandler) Send(c *gin.Context) {
	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, 400, "请求参数格式错误")
		return
	}
	msg, err := h.svc.Send(middleware.UserID(c), req.Content)
	switch {
	case errors.Is(err, service.ErrMessageEmpty):
		response.Fail(c, 400, 400, "留言内容不能为空")
		return
	case errors.Is(err, service.ErrMessageTooLong):
		response.Fail(c, 400, 400, "留言内容不能超过 500 字")
		return
	case errors.Is(err, service.ErrMessageRateLimited):
		response.Fail(c, 429, 429, "留言提交过于频繁，请 10 秒后再试")
		return
	case err != nil:
		response.Fail(c, 500, 500, "留言提交失败，请稍后重试")
		return
	}
	response.OK(c, msg)
}

// Mine GET /api/messages/mine 查看自己的留言。
func (h *MessageHandler) Mine(c *gin.Context) {
	list, err := h.svc.ListMine(middleware.UserID(c))
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"list": list})
}

// AdminList GET /api/admin/messages 管理员查看所有留言。
func (h *MessageHandler) AdminList(c *gin.Context) {
	list, err := h.svc.ListAll()
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"list": list})
}
