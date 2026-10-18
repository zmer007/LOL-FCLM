package ddragon

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Meta 离线数据包元数据，用于校验数据完整性与 ID 一致性。
type Meta struct {
	Version     string `json:"version"`
	DirName     string `json:"dirName"`
	GeneratedAt string `json:"generatedAt"`
	Lang        string `json:"lang"`
	Counts      struct {
		Champions int `json:"champions"`
		Items     int `json:"items"`
		Summoners int `json:"summoners"`
		RuneTrees int `json:"runeTrees"`
		Runes     int `json:"runes"`
	} `json:"counts"`
	Heroes    []MetaRef `json:"heroes"`
	Items     []MetaRef `json:"items"`
	Spells    []MetaRef `json:"spells"`
	Runes     []MetaRef `json:"runes"`
	Checksums map[string]string `json:"checksums"`
}

// MetaRef 数据项的 ID 映射，用于跨版本比对避免 ID 错乱。
type MetaRef struct {
	ID   string `json:"id"`
	Key  string `json:"key,omitempty"`
	Name string `json:"name"`
}

// LocalReader 从本地 assets 目录读取 Data Dragon 数据。
type LocalReader struct {
	root string // assets 根目录，如 /app/assets
	meta *Meta
}

// NewLocalReader 创建本地读取器并加载 meta.json。
func NewLocalReader(root string) (*LocalReader, error) {
	r := &LocalReader{root: root}
	if err := r.loadMeta(); err != nil {
		return nil, err
	}
	return r, nil
}

// Version 返回当前数据包版本。
func (r *LocalReader) Version() string {
	if r.meta != nil {
		return r.meta.Version
	}
	return ""
}

// Meta 返回元数据。
func (r *LocalReader) Meta() *Meta { return r.meta }

// loadMeta 定位版本目录并加载 meta.json。
// 优先使用 ASSETS_VERSION 指定的目录，否则取最新（按目录名排序）。
func (r *LocalReader) loadMeta() error {
	versionDir := os.Getenv("ASSETS_VERSION")
	if versionDir == "" {
		dirs, err := os.ReadDir(r.root)
		if err != nil {
			return fmt.Errorf("读取 assets 目录失败: %w", err)
		}
		var names []string
		for _, d := range dirs {
			if d.IsDir() {
				names = append(names, d.Name())
			}
		}
		if len(names) == 0 {
			return fmt.Errorf("assets 目录 %s 下没有任何版本数据包", r.root)
		}
		sort.Sort(sort.Reverse(sort.StringSlice(names)))
		versionDir = names[0]
	}
	metaPath := filepath.Join(r.root, versionDir, "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return fmt.Errorf("读取 meta.json 失败: %w", err)
	}
	var meta Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("解析 meta.json 失败: %w", err)
	}
	// 校验 JSON 文件校验和
	for rel, want := range meta.Checksums {
		full := filepath.Join(r.root, versionDir, filepath.FromSlash(rel))
		got, err := sha256File(full)
		if err != nil {
			return fmt.Errorf("校验 %s 失败: %w", rel, err)
		}
		if got != want {
			return fmt.Errorf("校验和不匹配: %s (期望 %s，实际 %s)", rel, want, got)
		}
	}
	r.meta = &meta
	return nil
}

func (r *LocalReader) versionDir() string {
	return filepath.Join(r.root, r.meta.DirName)
}

func (r *LocalReader) readJSON(relPath string, out any) error {
	full := filepath.Join(r.versionDir(), filepath.FromSlash(relPath))
	data, err := os.ReadFile(full)
	if err != nil {
		return fmt.Errorf("读取 %s 失败: %w", relPath, err)
	}
	return json.Unmarshal(data, out)
}

// Champions 读取英雄数据。
func (r *LocalReader) Champions() (*ChampionList, error) {
	var list ChampionList
	if err := r.readJSON("data/champion.json", &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// Items 读取装备数据。
func (r *LocalReader) Items() (*ItemList, error) {
	var list ItemList
	if err := r.readJSON("data/item.json", &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// Summoners 读取召唤师技能数据。
func (r *LocalReader) Summoners() (*SummonerList, error) {
	var list SummonerList
	if err := r.readJSON("data/summoner.json", &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// Runes 读取符文数据。
func (r *LocalReader) Runes() ([]RuneTree, error) {
	var list []RuneTree
	if err := r.readJSON("data/runesReforged.json", &list); err != nil {
		return nil, err
	}
	return list, nil
}

// OpenIcon 打开版本相关图标流（heroes/items/spells）。
// remotePath 形如 img/champion/Ahri.png。
func (r *LocalReader) OpenIcon(remotePath string) (io.ReadCloser, error) {
	return r.open(filepath.Join("img", filepath.FromSlash(remotePath)))
}

// OpenIconGlobal 打开全局图标流（符文图标位于 img/perk-images/）。
func (r *LocalReader) OpenIconGlobal(remotePath string) (io.ReadCloser, error) {
	// remotePath 形如 perk-images/Styles/7201_Default.png
	return r.open(filepath.Join("img", filepath.FromSlash(remotePath)))
}

// OpenShardIcon 打开属性碎片图标流。
func (r *LocalReader) OpenShardIcon(file string) (io.ReadCloser, error) {
	return r.open(filepath.Join("img", "statmods", file+".png"))
}

func (r *LocalReader) open(relPath string) (io.ReadCloser, error) {
	full := filepath.Join(r.versionDir(), filepath.FromSlash(relPath))
	f, err := os.Open(full)
	if err != nil {
		return nil, fmt.Errorf("打开图标 %s 失败: %w", relPath, err)
	}
	return f, nil
}

// IconRelPath 返回图标在 assets 中的相对路径，供直接复制使用。
func (r *LocalReader) IconRelPath(remotePath string) string {
	return filepath.ToSlash(filepath.Join("img", remotePath))
}

// shardRelPath 返回属性碎片相对路径。
func (r *LocalReader) ShardRelPath(file string) string {
	return filepath.ToSlash(filepath.Join("img", "statmods", file+".png"))
}

// sha256File 计算文件 SHA256。
func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// VersionDirName 把版本号转为目录名（16.20.1 → 16-20-1）。
func VersionDirName(version string) string {
	return strings.ReplaceAll(version, ".", "-")
}
