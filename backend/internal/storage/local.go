// Package storage 负责用户上传文件的落盘与访问路径。
package storage

import (
	"errors"
	"io"
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
var allowedExt = map[string]string{
	".png":  ".png",
	".jpg":  ".jpg",
	".jpeg": ".jpg",
	".webp": ".webp",
	".gif":  ".gif",
	".svg":  ".svg",
}

type LocalStore struct {
	root      string // 文件根目录（容器内 /data/media）
	urlPrefix string // 对外 URL 前缀
	maxBytes  int64
}

func NewLocalStore(root string, maxImageMB int) *LocalStore {
	return &LocalStore{
		root:      root,
		urlPrefix: "/media",
		maxBytes:  int64(maxImageMB) * 1024 * 1024,
	}
}

// SaveIcon 保存上传的图标文件，返回可访问的相对 URL（/media/icons/yyyymm/uuid.ext）。
func (s *LocalStore) SaveIcon(filename string, contentType string, src io.Reader, size int64) (string, error) {
	if size <= 0 || size > s.maxBytes {
		return "", invalidImage("文件大小需大于 0 且不超过上传上限")
	}

	ext := strings.ToLower(filepath.Ext(filename))
	savedExt, ok := allowedExt[ext]
	if !ok {
		return "", invalidImage("仅支持 png/jpg/jpeg/webp/gif/svg 格式")
	}

	// SVG 是文本，无法用 http.DetectContentType 判定，单独放行；其余校验真实文件头
	if ext != ".svg" {
		head := make([]byte, 512)
		n, _ := io.ReadFull(src, head)
		detected := http.DetectContentType(head[:n])
		if !strings.HasPrefix(detected, "image/") {
			return "", invalidImage("文件内容不是有效图片")
		}
		src = io.MultiReader(strings.NewReader(string(head[:n])), src)
	}

	month := time.Now().Format("200601")
	dir := filepath.Join(s.root, "icons", month)
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

	return s.urlPrefix + "/icons/" + month + "/" + name, nil
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
	return allowedExt[strings.ToLower(filepath.Ext(name))] != ""
}
