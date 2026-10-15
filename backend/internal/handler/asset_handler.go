package handler

import (
	"github.com/gin-gonic/gin"

	"roadside-league/backend/internal/repository"
	"roadside-league/backend/internal/response"
)

// AssetHandler 提供公开的游戏素材只读接口（装备/召唤师技能/符文图标）。
type AssetHandler struct {
	assets *repository.AssetRepository
}

func NewAssetHandler(assets *repository.AssetRepository) *AssetHandler {
	return &AssetHandler{assets: assets}
}

// List GET /api/assets?category=&subCategory=
func (h *AssetHandler) List(c *gin.Context) {
	list, err := h.assets.List(c.Query("category"), c.Query("subCategory"))
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"list": list})
}
