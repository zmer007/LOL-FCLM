package handler

import (
	"github.com/gin-gonic/gin"

	"roadside-league/backend/internal/ddragon"
	"roadside-league/backend/internal/repository"
	"roadside-league/backend/internal/response"
	"roadside-league/backend/internal/storage"
)

// RuneHandler 提供公开的结构化符文树与属性碎片接口。
type RuneHandler struct {
	runes *repository.RuneRepository
	store *storage.LocalStore
}

func NewRuneHandler(runes *repository.RuneRepository, store *storage.LocalStore) *RuneHandler {
	return &RuneHandler{runes: runes, store: store}
}

type shardVO struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

// Trees GET /api/rune-trees → {trees, shards}
func (h *RuneHandler) Trees(c *gin.Context) {
	trees, err := h.runes.ListTreesWithRunes()
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}

	shards := make([][]shardVO, len(ddragon.ShardGrid))
	for row := range ddragon.ShardGrid {
		shards[row] = make([]shardVO, len(ddragon.ShardGrid[row]))
		for col, d := range ddragon.ShardGrid[row] {
			shards[row][col] = shardVO{
				Name: d.Name,
				Icon: h.store.IconURL("shards", d.File+".png"),
			}
		}
	}
	response.OK(c, gin.H{"trees": trees, "shards": shards})
}
