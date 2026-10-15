package handler

import (
	"github.com/gin-gonic/gin"

	"roadside-league/backend/internal/repository"
	"roadside-league/backend/internal/response"
)

// HeroHandler 提供公开的英雄只读接口。
type HeroHandler struct {
	heroes *repository.HeroRepository
}

func NewHeroHandler(heroes *repository.HeroRepository) *HeroHandler {
	return &HeroHandler{heroes: heroes}
}

// List 全部英雄，可按分路筛选（?lane=中单）。供首页头像墙、英雄列表页使用。
func (h *HeroHandler) List(c *gin.Context) {
	heroes, err := h.heroes.List(c.Query("lane"))
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"list": heroes})
}
