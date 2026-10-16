package handler

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"strconv"

	"github.com/gin-gonic/gin"

	"roadside-league/backend/internal/middleware"
	"roadside-league/backend/internal/response"
	"roadside-league/backend/internal/service"
)

const (
	contentPageSize = 9
	reviewPageSize  = 10
)

type ContentHandler struct {
	svc *service.ContentService
}

func NewContentHandler(svc *service.ContentService) *ContentHandler {
	return &ContentHandler{svc: svc}
}

// ---------- 公开 ----------

// List GET /api/contents?heroId=&opponentId=&kind=&page=
func (h *ContentHandler) List(c *gin.Context) {
	heroID, _ := strconv.ParseInt(c.Query("heroId"), 10, 64)
	if heroID == 0 {
		response.Fail(c, 400, 400, "缺少英雄参数")
		return
	}
	opponentID, _ := strconv.ParseInt(c.Query("opponentId"), 10, 64)
	kind := c.Query("kind")
	page := atoiDefault(c.Query("page"), 1)

	pg, vos, err := h.svc.PublicList(heroID, opponentID, kind, page, contentPageSize)
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"list": vos, "total": pg.Total, "page": page, "size": contentPageSize})
}

// Detail GET /api/contents/:id
func (h *ContentHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, 400, "内容 ID 不合法")
		return
	}
	uid := middleware.UserID(c)
	isAdmin := middleware.IsAdmin(c)
	vo, err := h.svc.PublicGet(id, uid, isAdmin)
	if errors.Is(err, service.ErrContentNotFound) {
		response.Fail(c, 404, 404, "内容不存在或尚未通过审核")
		return
	}
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, vo)
}

// HeroBuilds GET /api/heroes/:heroId/builds
func (h *ContentHandler) HeroBuilds(c *gin.Context) {
	heroID, err := strconv.ParseInt(c.Param("heroId"), 10, 64)
	if err != nil {
		response.Fail(c, 400, 400, "英雄 ID 不合法")
		return
	}
	list, err := h.svc.HeroBuilds(heroID)
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"list": list})
}

// ---------- 投稿 / 我的 ----------

func (h *ContentHandler) parseSubmit(c *gin.Context) (service.ContentInput, service.FileSet, error) {
	var in service.ContentInput
	if raw := c.PostForm("meta"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &in); err != nil {
			return in, service.FileSet{}, err
		}
	}
	fs := service.FileSet{}
	if fh, err := c.FormFile("video"); err == nil {
		fs.Video = fh
	}
	if form, err := c.MultipartForm(); err == nil && form != nil {
		fs.Images = form.File["images"]
	}
	return in, fs, nil
}

// Create POST /api/contents
func (h *ContentHandler) Create(c *gin.Context) {
	in, fs, err := h.parseSubmit(c)
	if err != nil {
		response.Fail(c, 400, 400, "提交内容格式错误")
		return
	}
	created, err := h.svc.Create(in, fs, middleware.UserID(c))
	h.writeSaveResult(c, err, created)
}

// Update PUT /api/contents/:id
func (h *ContentHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, 400, "内容 ID 不合法")
		return
	}
	in, fs, err := h.parseSubmit(c)
	if err != nil {
		response.Fail(c, 400, 400, "提交内容格式错误")
		return
	}
	updated, err := h.svc.Update(id, middleware.UserID(c), middleware.IsAdmin(c), in, fs)
	h.writeSaveResult(c, err, updated)
}

func (h *ContentHandler) writeSaveResult(c *gin.Context, err error, content any) {
	if ve := asValidation(err); ve != nil {
		response.Fail(c, 400, 400, ve.Msg)
		return
	}
	if errors.Is(err, service.ErrContentNotFound) {
		response.Fail(c, 404, 404, "内容不存在")
		return
	}
	if errors.Is(err, service.ErrContentForbidden) {
		response.Fail(c, 403, 403, "只能操作自己发布的内容")
		return
	}
	if err != nil {
		response.Fail(c, 400, 400, "文件上传失败，请检查格式与大小后重试")
		return
	}
	response.OK(c, content)
}

// Remove DELETE /api/contents/:id
func (h *ContentHandler) Remove(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, 400, "内容 ID 不合法")
		return
	}
	err = h.svc.Delete(id, middleware.UserID(c), middleware.IsAdmin(c))
	if errors.Is(err, service.ErrContentNotFound) {
		response.Fail(c, 404, 404, "内容不存在")
		return
	}
	if errors.Is(err, service.ErrContentForbidden) {
		response.Fail(c, 403, 403, "只能删除自己发布的内容")
		return
	}
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"id": id})
}

// Mine GET /api/my/contents?status=
func (h *ContentHandler) Mine(c *gin.Context) {
	list, err := h.svc.MyList(middleware.UserID(c), c.Query("status"))
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"list": list})
}

// ---------- 管理员审批 ----------

// AdminList GET /api/admin/contents?status=&page=
func (h *ContentHandler) AdminList(c *gin.Context) {
	page := atoiDefault(c.Query("page"), 1)
	pg, vos, err := h.svc.ReviewList(c.Query("status"), page, reviewPageSize)
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"list": vos, "total": pg.Total, "page": page, "size": reviewPageSize})
}

// Approve POST /api/admin/contents/:id/approve
func (h *ContentHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, 400, "内容 ID 不合法")
		return
	}
	if err := h.svc.Approve(id); err != nil {
		h.writeReviewError(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

// Reject POST /api/admin/contents/:id/reject
func (h *ContentHandler) Reject(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, 400, "内容 ID 不合法")
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&body)
	if err := h.svc.Reject(id, body.Reason); err != nil {
		h.writeReviewError(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (h *ContentHandler) writeReviewError(c *gin.Context, err error) {
	if ve := asValidation(err); ve != nil {
		response.Fail(c, 400, 400, ve.Msg)
		return
	}
	if errors.Is(err, service.ErrContentNotFound) {
		response.Fail(c, 404, 404, "内容不存在")
		return
	}
	response.Fail(c, 500, 500, "服务器开小差了")
}

func atoiDefault(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	return def
}

// 确保 multipart.FileHeader 类型被引用（parseSubmit 的 FileSet 字段类型）。
var _ = func(_ *multipart.FileHeader) {}
