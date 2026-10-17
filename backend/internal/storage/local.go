// Package storage 负责用户上传文件的落盘与访问路径。
package storage

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidImage = errors.New("invalid image file")

// InvalidImageError 同时满足 errors.Is(err, ErrInvalidImage) 与纯中文提示。
type InvalidImageError struct{ Msg string }

func (e *InvalidImageError) Error() string { return e.Msg }
func (e *InvalidImageError) Is(target error) bool {
	return target == ErrInvalidImage
}

func invalidImage(msg string) error {
	return &InvalidImageError{Msg: msg}
}

// 允许的图片扩展名 → 落盘扩展名
var allowedImageExt = map[string]string{
	".png":  ".png",
	".jpg":  ".jpg",
	".jpeg": ".jpg",
	".webp": ".webp",
	".gif":  ".gif",
	".svg":  ".svg",
}

var allowedVideoExt = map[string]string{
	".mp4":  ".mp4",
	".mkv":  ".mkv",
	".mov":  ".mov",
	".avi":  ".avi",
	".webm": ".webm",
	".flv":  ".flv",
	".wmv":  ".wmv",
	".m4v":  ".m4v",
	".ts":   ".ts",
}

type LocalStore struct {
	root          string // 文件根目录（容器内 /data/media）
	urlPrefix     string // 对外 URL 前缀
	maxImageBytes int64
	maxVideoBytes int64
}

func NewLocalStore(root string, maxImageMB, maxVideoMB int) *LocalStore {
	return &LocalStore{
		root:          root,
		urlPrefix:     "/media",
		maxImageBytes: int64(maxImageMB) * 1024 * 1024,
		maxVideoBytes: int64(maxVideoMB) * 1024 * 1024,
	}
}

// saveFile 通用用户上传落盘：category 一级目录 + yyyymm 二级目录 + uuid 命名。
// sniff 为 nil 时跳过文件头校验；否则读取前 512 字节校验真实类型。
func (s *LocalStore) saveFile(
	category, filename string, src io.Reader, size, maxBytes int64,
	allowExt map[string]string, errMsg string, sniff func(detected string) bool,
) (string, error) {
	if size <= 0 || size > maxBytes {
		return "", invalidImage("文件大小需大于 0 且不超过上传上限")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	savedExt, ok := allowExt[ext]
	if !ok {
		return "", invalidImage(errMsg)
	}

	if sniff != nil {
		head := make([]byte, 512)
		n, _ := io.ReadFull(src, head)
		detected := http.DetectContentType(head[:n])
		if !sniff(detected) {
			return "", invalidImage(errMsg)
		}
		src = io.MultiReader(strings.NewReader(string(head[:n])), src)
	}

	month := time.Now().Format("200601")
	dir := filepath.Join(s.root, category, month)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := uuid.NewString() + savedExt
	full := filepath.Join(dir, name)

	f, err := os.Create(full)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, src); err != nil {
		return "", err
	}
	return s.urlPrefix + "/" + category + "/" + month + "/" + name, nil
}

// SaveIcon 保存上传的图标文件，返回 /media/icons/yyyymm/uuid.ext。
func (s *LocalStore) SaveIcon(filename string, contentType string, src io.Reader, size int64) (string, error) {
	return s.saveFile("icons", filename, src, size, s.maxImageBytes, allowedImageExt,
		"仅支持 png/jpg/jpeg/webp/gif/svg 格式", func(detected string) bool {
			// SVG 为文本，DetectContentType 判为 text/xml 或 image/svg+xml，一并放行
			return strings.HasPrefix(detected, "image/") || strings.Contains(detected, "xml")
		})
}

// SaveArticleImage 保存图文插图，目录 /media/images/yyyymm/。
func (s *LocalStore) SaveArticleImage(filename string, src io.Reader, size int64) (string, error) {
	return s.saveFile("images", filename, src, size, s.maxImageBytes, allowedImageExt,
		"仅支持 png/jpg/jpeg/webp/gif 图片", func(detected string) bool {
			return strings.HasPrefix(detected, "image/")
		})
}

// SaveVideo 保存投稿视频（multipart 流式落盘），目录 /media/videos/yyyymm/。
// 支持 mp4/mkv/mov/avi/webm/flv/wmv/m4v/ts 等常见格式。
func (s *LocalStore) SaveVideo(fh *multipart.FileHeader) (string, error) {
	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	return s.saveFile("videos", fh.Filename, src, fh.Size, s.maxVideoBytes, allowedVideoExt,
		"仅支持 mp4/mkv/mov/avi/webm/flv/wmv/m4v/ts 格式视频", func(detected string) bool {
			// Go 内置 http.DetectContentType 仅能识别 mp4/webm/ogg，其余格式会被判为
			// application/octet-stream；扩展名已在上层白名单校验，此处宽松放行。
			return strings.HasPrefix(detected, "video/") || detected == "application/octet-stream"
		})
}

// SaveIconFile 保存服务端从可信来源（Data Dragon CDN）下载的图标，
// 按固定目录与文件名落盘以保证幂等；文件已存在时跳过下载。
// dir 为 icons 下的分组（heroes/items/spells/runes）。
// 返回：本地 URL、是否新写入。
func (s *LocalStore) SaveIconFile(dir, filename string, src io.Reader) (string, bool, error) {
	if !safeIconName(filename) {
		return "", false, ErrInvalidImage
	}
	rel := filepath.Join("icons", dir, filename)
	full := filepath.Join(s.root, rel)
	url := s.urlPrefix + "/" + filepath.ToSlash(rel)

	if _, err := os.Stat(full); err == nil {
		return url, false, nil // 已存在，复用
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return "", false, err
	}
	f, err := os.Create(full)
	if err != nil {
		return "", false, err
	}
	defer f.Close()
	if _, err := io.Copy(f, src); err != nil {
		return "", false, err
	}
	return url, true, nil
}

// IconURL 返回图标在固定目录/文件名下的对外 URL。
func (s *LocalStore) IconURL(dir, filename string) string {
	return s.urlPrefix + "/" + filepath.ToSlash(filepath.Join("icons", dir, filename))
}

// IconExists 判断固定路径图标是否已存在（同步前用于跳过下载）。
func (s *LocalStore) IconExists(dir, filename string) bool {
	if !safeIconName(filename) {
		return false
	}
	_, err := os.Stat(filepath.Join(s.root, "icons", dir, filename))
	return err == nil
}

// safeIconName 防止路径穿越，只允许单层文件名。
func safeIconName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	if filepath.Base(name) != name {
		return false
	}
	return allowedImageExt[strings.ToLower(filepath.Ext(name))] != ""
}
