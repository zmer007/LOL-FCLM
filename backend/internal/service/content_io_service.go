package service

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"roadside-league/backend/internal/model"
)

// ExportContent 单条导出记录（含英雄/作者名称，便于人工核对）。
type ExportContent struct {
	Kind             string           `json:"kind"`
	HeroID           int64            `json:"heroId"`
	HeroName         string           `json:"heroName"`
	OpponentHeroID   int64            `json:"opponentHeroId"`
	OpponentHeroName string           `json:"opponentHeroName"`
	AuthorID         int64            `json:"authorId"`
	AuthorName       string           `json:"authorName"`
	Rank             string           `json:"rank"`
	Title            string           `json:"title"`
	Body             string           `json:"body"`
	DurationSec      int              `json:"durationSec"`
	VideoURL         string           `json:"videoUrl"`
	Images           []string         `json:"images"`
	Items            model.BuildItems `json:"items"`
	Runes            model.BuildRunes `json:"runes"`
	BuildNote        string           `json:"buildNote"`
	GameVersion      string           `json:"gameVersion"`
	SourceURL        string           `json:"sourceUrl"`
	Status           string           `json:"status"`
}

// ExportFile 导出清单结构（v2：含媒体索引）。
type ExportFile struct {
	Version    int               `json:"version"`
	ExportedAt string            `json:"exportedAt"`
	Contents   []ExportContent   `json:"contents"`
	MediaIndex map[string]string `json:"mediaIndex,omitempty"` // relPath -> zip 文件名
}

// ExportFileInfo 单个导出文件信息。
type ExportFileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// ExportResult 导出结果。
type ExportResult struct {
	Dir   string           `json:"dir"`
	Files []ExportFileInfo `json:"files"`
}

const maxZipBytes = 2 * 1024 * 1024 * 1024 // 单卷 2GB

// ExportsRoot 导出目录根（放在 media 卷内，保证持久化）。
func (s *ContentService) ExportsRoot() string {
	return filepath.Join(s.store.MediaRoot(), "exports")
}

// buildExportContents 组装导出内容列表（复用英雄/作者名称解析）。
func (s *ContentService) buildExportContents(list []model.Content) []ExportContent {
	heroIDs := make([]int64, 0, len(list)*2)
	userIDs := make([]int64, 0, len(list))
	for _, c := range list {
		heroIDs = append(heroIDs, c.HeroID, c.OpponentHeroID)
		userIDs = append(userIDs, c.AuthorID)
	}
	heroMap, _ := s.heroes.FindByIDs(heroIDs)
	userMap, _ := s.users.FindByIDs(userIDs)

	heroName := func(id int64) string {
		if h, ok := heroMap[id]; ok {
			return h.Name
		}
		return ""
	}
	userName := func(id int64) string {
		if u, ok := userMap[id]; ok {
			return u.Username
		}
		return ""
	}

	out := make([]ExportContent, 0, len(list))
	for _, c := range list {
		out = append(out, ExportContent{
			Kind: c.Kind, HeroID: c.HeroID, HeroName: heroName(c.HeroID),
			OpponentHeroID: c.OpponentHeroID, OpponentHeroName: heroName(c.OpponentHeroID),
			AuthorID: c.AuthorID, AuthorName: userName(c.AuthorID),
			Rank: c.Rank, Title: c.Title, Body: c.Body, DurationSec: c.DurationSec,
			VideoURL: c.VideoURL, Images: c.Images, Items: c.Items, Runes: c.Runes,
			BuildNote: c.BuildNote, GameVersion: c.GameVersion, SourceURL: c.SourceURL, Status: c.Status,
		})
	}
	return out
}

// collectMedia 收集内容引用的全部媒体文件，去重并检查磁盘存在性。
func (s *ContentService) collectMedia(list []model.Content) []string {
	seen := map[string]bool{}
	var rels []string
	add := func(url string) {
		rel := s.store.URLToRelPath(url)
		if rel == "" || seen[rel] {
			return
		}
		if _, err := os.Stat(s.store.AbsPath(rel)); err != nil {
			return // 文件不存在则跳过
		}
		seen[rel] = true
		rels = append(rels, rel)
	}
	for _, c := range list {
		add(c.VideoURL)
		for _, img := range c.Images {
			add(img)
		}
	}
	sort.Strings(rels)
	return rels
}

// partitionMedia 按体积将媒体文件分卷，每卷不超过 maxZipBytes。
func partitionMedia(rels []string, absFn func(string) string) [][]string {
	var groups [][]string
	var cur []string
	var curSize int64
	for _, rel := range rels {
		fi, err := os.Stat(absFn(rel))
		if err != nil {
			continue
		}
		// 单个文件超过上限则独占一卷
		if len(cur) > 0 && curSize+fi.Size() > maxZipBytes {
			groups = append(groups, cur)
			cur = nil
			curSize = 0
		}
		cur = append(cur, rel)
		curSize += fi.Size()
	}
	if len(cur) > 0 {
		groups = append(groups, cur)
	}
	return groups
}

// ExportSelected 按 ID 导出内容（ids 为空则导出全部），生成 manifest.json + 分卷 zip。
func (s *ContentService) ExportSelected(ids []int64) (*ExportResult, error) {
	var list []model.Content
	var err error
	if len(ids) == 0 {
		list, err = s.contents.ListAll()
	} else {
		list, err = s.contents.FindByIDs(ids)
	}
	if err != nil {
		return nil, err
	}

	// 组装内容数据
	contents := s.buildExportContents(list)

	// 收集媒体文件并分卷
	mediaRels := s.collectMedia(list)
	groups := partitionMedia(mediaRels, s.store.AbsPath)

	// 创建导出目录
	ts := time.Now().Format("20060102-150405")
	dir := filepath.Join(s.ExportsRoot(), ts)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	mediaIndex := map[string]string{}
	var files []ExportFileInfo

	// 写入分卷 zip
	for i, group := range groups {
		zipName := fmt.Sprintf("media-%03d.zip", i+1)
		zipPath := filepath.Join(dir, zipName)
		if err := writeZip(zipPath, group, s.store.AbsPath); err != nil {
			return nil, err
		}
		fi, _ := os.Stat(zipPath)
		files = append(files, ExportFileInfo{Name: zipName, Size: fi.Size()})
		for _, rel := range group {
			mediaIndex[rel] = zipName
		}
	}

	// 写入 manifest.json
	manifest := ExportFile{
		Version:    2,
		ExportedAt: time.Now().Format("2006-01-02 15:04:05"),
		Contents:   contents,
		MediaIndex: mediaIndex,
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	manifestPath := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestBytes, 0o644); err != nil {
		return nil, err
	}
	fi, _ := os.Stat(manifestPath)
	files = append([]ExportFileInfo{{Name: "manifest.json", Size: fi.Size()}}, files...)

	return &ExportResult{Dir: ts, Files: files}, nil
}

// writeZip 将媒体文件打包为 zip，存储时保留相对路径。
func writeZip(zipPath string, rels []string, absFn func(string) string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	defer zw.Close()

	for _, rel := range rels {
		abs := absFn(rel)
		src, err := os.Open(abs)
		if err != nil {
			return err
		}
		w, err := zw.Create(rel)
		if err != nil {
			src.Close()
			return err
		}
		if _, err := io.Copy(w, src); err != nil {
			src.Close()
			return err
		}
		src.Close()
	}
	return nil
}

// ExportAll 导出全部内容为 JSON 字节流（保留旧接口兼容）。
func (s *ContentService) ExportAll() ([]byte, error) {
	list, err := s.contents.ListAll()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(ExportFile{
		Version:    1,
		ExportedAt: time.Now().Format("2006-01-02 15:04:05"),
		Contents:   s.buildExportContents(list),
	}, "", "  ")
}

// Import 从 JSON 字节流批量导入内容（v1 单文件导入，保留兼容）。
func (s *ContentService) Import(data []byte, authorID int64) (int, error) {
	var file ExportFile
	if err := json.Unmarshal(data, &file); err != nil {
		return 0, &ValidationError{Msg: "JSON 格式错误：" + err.Error()}
	}
	return s.importFromExportFile(&file, authorID)
}

// ImportFromFiles 从多文件导入：manifest.json（内容清单）+ media-*.zip（媒体文件）。
// files 为文件名 -> 内容；支持 v2 分卷格式，也兼容单个旧版 JSON 文件。
func (s *ContentService) ImportFromFiles(files map[string][]byte, authorID int64) (int, error) {
	// 定位 manifest.json；若没有则找任意 .json（兼容旧版单文件导出）
	var manifestName string
	for name := range files {
		if strings.EqualFold(name, "manifest.json") {
			manifestName = name
			break
		}
	}
	if manifestName == "" {
		for name := range files {
			if strings.HasSuffix(strings.ToLower(name), ".json") {
				manifestName = name
				break
			}
		}
	}
	if manifestName == "" {
		return 0, &ValidationError{Msg: "未找到 manifest.json 或 JSON 导出文件"}
	}

	var file ExportFile
	if err := json.Unmarshal(files[manifestName], &file); err != nil {
		return 0, &ValidationError{Msg: "JSON 格式错误：" + err.Error()}
	}

	// 解压所有 zip，恢复媒体文件
	for name, data := range files {
		if !strings.HasSuffix(strings.ToLower(name), ".zip") {
			continue
		}
		if err := s.extractZipMedia(data); err != nil {
			return 0, &ValidationError{Msg: "解压媒体文件失败：" + err.Error()}
		}
	}

	return s.importFromExportFile(&file, authorID)
}

// extractZipMedia 解压 zip 中的媒体文件到媒体根目录。
func (s *ContentService) extractZipMedia(data []byte) error {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		rel := filepath.ToSlash(zf.Name)
		if strings.Contains(rel, "..") {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		err = s.store.SaveMediaFile(rel, rc)
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// importFromExportFile 从解析后的 ExportFile 导入内容。
func (s *ContentService) importFromExportFile(file *ExportFile, authorID int64) (int, error) {
	if len(file.Contents) == 0 {
		return 0, &ValidationError{Msg: "导出文件中没有内容"}
	}

	allHeroes, err := s.heroes.List("")
	if err != nil {
		return 0, err
	}
	byID := map[int64]*model.Hero{}
	byName := map[string]*model.Hero{}
	for i := range allHeroes {
		h := &allHeroes[i]
		byID[h.ID] = h
		byName[h.Name] = h
	}
	resolveHero := func(id int64, name string) (*model.Hero, error) {
		if h, ok := byID[id]; ok {
			return h, nil
		}
		if name != "" {
			if h, ok := byName[name]; ok {
				return h, nil
			}
		}
		return nil, &ValidationError{Msg: "英雄不存在：" + name}
	}

	allUsers, err := s.users.List()
	if err != nil {
		return 0, err
	}
	userByName := map[string]int64{}
	for _, u := range allUsers {
		userByName[u.Username] = u.ID
	}

	count := 0
	for i, ec := range file.Contents {
		hero, err := resolveHero(ec.HeroID, ec.HeroName)
		if err != nil {
			return count, &ValidationError{Msg: "第" + strconv.Itoa(i+1) + "条：" + err.Error()}
		}
		opp, err := resolveHero(ec.OpponentHeroID, ec.OpponentHeroName)
		if err != nil {
			return count, &ValidationError{Msg: "第" + strconv.Itoa(i+1) + "条：" + err.Error()}
		}
		if hero.ID == opp.ID {
			return count, &ValidationError{Msg: "第" + strconv.Itoa(i+1) + "条：对位英雄不能与主英雄相同"}
		}

		author := authorID
		if ec.AuthorName != "" {
			if id, ok := userByName[ec.AuthorName]; ok {
				author = id
			}
		}

		in := ContentInput{
			Kind: ec.Kind, HeroID: hero.ID, OpponentHeroID: opp.ID, Rank: ec.Rank,
			Title: ec.Title, Body: ec.Body, DurationSec: ec.DurationSec,
			Items: ec.Items, Runes: ec.Runes, BuildNote: ec.BuildNote, SourceURL: ec.SourceURL,
		}
		if err := s.validateInput(&in); err != nil {
			return count, &ValidationError{Msg: "第" + strconv.Itoa(i+1) + "条：" + err.Error()}
		}

		gv := ec.GameVersion
		if gv == "" {
			gv = s.currentGameVersion()
		}
		status := ec.Status
		if status != model.ContentApproved && status != model.ContentRejected {
			status = model.ContentApproved
		}

		c := &model.Content{
			AuthorID: author, HeroID: hero.ID, OpponentHeroID: opp.ID,
			Rank: ec.Rank, Kind: ec.Kind, Title: ec.Title, Body: ec.Body,
			DurationSec: ec.DurationSec, VideoURL: ec.VideoURL, Images: ec.Images,
			Items: ec.Items, Runes: ec.Runes, BuildNote: ec.BuildNote, SourceURL: ec.SourceURL,
			GameVersion: gv, Status: status,
		}
		if status == model.ContentApproved {
			c.Approved = &model.ContentSnapshot{
				Title: c.Title, Body: c.Body, VideoURL: c.VideoURL, DurationSec: c.DurationSec,
				Images: c.Images, OpponentHeroID: c.OpponentHeroID, Rank: c.Rank,
				Items: c.Items, Runes: c.Runes, BuildNote: c.BuildNote, GameVersion: gv,
				SourceURL: c.SourceURL,
			}
		}
		if err := s.contents.Create(c); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
