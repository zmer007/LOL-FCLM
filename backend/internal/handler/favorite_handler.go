package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"roadside-league/backend/internal/middleware"
	"roadside-league/backend/internal/response"
	"roadside-league/backend/internal/service"
)

type FavoriteHandler struct {
	svc *service.FavoriteService
}

func NewFavoriteHandler(svc *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{svc: svc}
}

// Toggle POST /api/favorites/:contentId 切换收藏状态。
func (h *FavoriteHandler) Toggle(c *gin.Context) {
	contentID, err := strconv.ParseInt(c.Param("contentId"), 10, 64)
	if err != nil || contentID <= 0 {
		response.Fail(c, 400, 400, "内容 ID 无效")
		return
	}
	favorited, err := h.svc.Toggle(middleware.UserID(c), contentID)
	if err != nil {
		response.Fail(c, 500, 500, "操作失败，请稍后重试")
		return
	}
	response.OK(c, gin.H{"favorited": favorited})
}

// ListMine GET /api/favorites/mine 查看当前用户的收藏列表。
func (h *FavoriteHandler) ListMine(c *gin.Context) {
	list, err := h.svc.ListMine(middleware.UserID(c))
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"list": list})
}
