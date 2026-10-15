package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"roadside-league/backend/internal/middleware"
	"roadside-league/backend/internal/response"
	"roadside-league/backend/internal/service"
)

// BuildHandler 装备符文方案：公开只读 + 管理员整英雄替换。
type BuildHandler struct {
	svc *service.BuildPlanService
}

func NewBuildHandler(svc *service.BuildPlanService) *BuildHandler {
	return &BuildHandler{svc: svc}
}

// ListByHero GET /api/heroes/:heroId/plans
func (h *BuildHandler) ListByHero(c *gin.Context) {
	heroID, err := strconv.ParseInt(c.Param("heroId"), 10, 64)
	if err != nil {
		response.Fail(c, 400, 400, "英雄 ID 不合法")
		return
	}
	plans, err := h.svc.ListByHero(heroID)
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"list": plans})
}

type replacePlansRequest struct {
	Plans []service.PlanInput `json:"plans"`
}

// Replace PUT /api/admin/heroes/:id/plans（整英雄批量替换，空数组=清空）
func (h *BuildHandler) Replace(c *gin.Context) {
	heroID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, 400, "英雄 ID 不合法")
		return
	}
	var req replacePlansRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, 400, "请求参数格式错误")
		return
	}
	saved, err := h.svc.ReplaceByHero(heroID, middleware.UserID(c), req.Plans)
	if ve := asValidation(err); ve != nil {
		response.Fail(c, 400, 400, ve.Msg)
		return
	}
	if errors.Is(err, service.ErrPlanHeroNotFound) {
		response.Fail(c, 404, 404, "英雄不存在")
		return
	}
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"list": saved})
}
