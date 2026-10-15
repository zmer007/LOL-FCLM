package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"roadside-league/backend/internal/response"
	"roadside-league/backend/internal/service"
)

type SyncHandler struct {
	sync *service.SyncService
}

func NewSyncHandler(s *service.SyncService) *SyncHandler {
	return &SyncHandler{sync: s}
}

type syncRequest struct {
	Force bool `json:"force"`
}

// Trigger 启动一次 Data Dragon 同步（异步执行，立即返回）。
func (h *SyncHandler) Trigger(c *gin.Context) {
	var req syncRequest
	// body 可空，绑定失败时按 force=false 处理
	_ = c.ShouldBindJSON(&req)

	if err := h.sync.Start(req.Force); err != nil {
		if errors.Is(err, service.ErrSyncRunning) {
			response.Fail(c, 409, 409, "同步正在进行中，请稍候")
			return
		}
		response.Fail(c, 500, 500, "启动同步失败")
		return
	}
	response.OK(c, h.sync.State())
}

func (h *SyncHandler) Status(c *gin.Context) {
	response.OK(c, h.sync.State())
}
