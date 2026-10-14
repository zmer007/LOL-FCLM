package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"roadside-league/backend/internal/repository"
	"roadside-league/backend/internal/response"
	"roadside-league/backend/internal/service"
	"roadside-league/backend/internal/storage"
)

type AdminHandler struct {
	admin *service.AdminService
}

func NewAdminHandler(admin *service.AdminService) *AdminHandler {
	return &AdminHandler{admin: admin}
}

// ---------- 英雄 ----------

func (h *AdminHandler) ListHeroes(c *gin.Context) {
	heroes, err := h.admin.ListHeroes(c.Query("lane"))
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"list": heroes})
}

type createHeroRequest struct {
	Name string `json:"name"`
	Lane string `json:"lane"`
}

func (h *AdminHandler) CreateHero(c *gin.Context) {
	var req createHeroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, 400, "请求参数格式错误")
		return
	}
	h2, err := h.admin.CreateHero(req.Name, req.Lane)
	if ve := asValidation(err); ve != nil {
		response.Fail(c, 400, 400, ve.Msg)
		return
	}
	if errors.Is(err, service.ErrHeroExists) {
		response.Fail(c, 409, 409, "英雄已存在")
		return
	}
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"hero": h2})
}

func (h *AdminHandler) UploadHeroAvatar(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, 400, "英雄 ID 不合法")
		return
	}
	if err := h.saveUpload(c, func(filename, ct string, src fileReader, size int64) error {
		return h.admin.UpdateHeroAvatar(id, filename, ct, src, size)
	}); err != nil {
		return
	}
	response.OK(c, gin.H{"id": id})
}

// ---------- 素材 ----------

func (h *AdminHandler) ListAssets(c *gin.Context) {
	assets, err := h.admin.ListAssets(c.Query("category"), c.Query("subCategory"))
	if ve := asValidation(err); ve != nil {
		response.Fail(c, 400, 400, ve.Msg)
		return
	}
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"list": assets})
}

type createAssetRequest struct {
	Category    string `json:"category"`
	SubCategory string `json:"subCategory"`
	Name        string `json:"name"`
}

func (h *AdminHandler) CreateAsset(c *gin.Context) {
	var req createAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, 400, "请求参数格式错误")
		return
	}
	a, err := h.admin.CreateAsset(req.Category, req.SubCategory, req.Name)
	if ve := asValidation(err); ve != nil {
		response.Fail(c, 400, 400, ve.Msg)
		return
	}
	if errors.Is(err, service.ErrAssetExists) {
		response.Fail(c, 409, 409, "该分类下已存在同名素材")
		return
	}
	if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"asset": a})
}

func (h *AdminHandler) UploadAssetIcon(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, 400, "素材 ID 不合法")
		return
	}
	if err := h.saveUpload(c, func(filename, ct string, src fileReader, size int64) error {
		return h.admin.UpdateAssetIcon(id, filename, ct, src, size)
	}); err != nil {
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (h *AdminHandler) DeleteAsset(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 400, 400, "素材 ID 不合法")
		return
	}
	if err := h.admin.DeleteAsset(id); errors.Is(err, repository.ErrNotFound) {
		response.Fail(c, 404, 404, "素材不存在")
		return
	} else if err != nil {
		response.Fail(c, 500, 500, "服务器开小差了")
		return
	}
	response.OK(c, gin.H{"id": id})
}

// ---------- 上传公共处理 ----------

type fileReader = interface {
	Read(p []byte) (int, error)
}

// saveUpload 解析 multipart 中的 icon 文件并交给回调落盘，统一错误响应。
func (h *AdminHandler) saveUpload(c *gin.Context, save func(filename, contentType string, src fileReader, size int64) error) error {
	fh, err := c.FormFile("icon")
	if err != nil {
		response.Fail(c, 400, 400, "请选择要上传的图标文件")
		return err
	}
	f, err := fh.Open()
	if err != nil {
		response.Fail(c, 400, 400, "读取上传文件失败")
		return err
	}
	defer f.Close()

	if err := save(fh.Filename, fh.Header.Get("Content-Type"), f, fh.Size); err != nil {
		if errors.Is(err, storage.ErrInvalidImage) {
			response.Fail(c, 400, 400, err.Error())
			return err
		}
		if errors.Is(err, repository.ErrNotFound) {
			response.Fail(c, 404, 404, "目标对象不存在")
			return err
		}
		response.Fail(c, 500, 500, "图标保存失败")
		return err
	}
	return nil
}

func asValidation(err error) *service.ValidationError {
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		return ve
	}
	return nil
}
