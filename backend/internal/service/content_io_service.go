package service

import (
	"encoding/json"
	"strconv"
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

// ExportFile 导出文件结构。
type ExportFile struct {
	Version    int             `json:"version"`
	ExportedAt string          `json:"exportedAt"`
	Contents   []ExportContent `json:"contents"`
}

// ExportAll 导出全部内容为 JSON 字节流。
func (s *ContentService) ExportAll() ([]byte, error) {
	list, err := s.contents.ListAll()
	if err != nil {
		return nil, err
	}

	// 批量解析英雄与作者名称。
	heroIDs := make([]int64, 0)
	userIDs := make([]int64, 0)
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

	return json.MarshalIndent(ExportFile{
		Version:    1,
		ExportedAt: time.Now().Format("2006-01-02 15:04:05"),
		Contents:   out,
	}, "", "  ")
}

// Import 从 JSON 字节流批量导入内容；authorID 为执行导入的管理员（原作者找不到时回退到该管理员）。
// 返回成功导入的条数。
func (s *ContentService) Import(data []byte, authorID int64) (int, error) {
	var file ExportFile
	if err := json.Unmarshal(data, &file); err != nil {
		return 0, &ValidationError{Msg: "JSON 格式错误：" + err.Error()}
	}
	if len(file.Contents) == 0 {
		return 0, &ValidationError{Msg: "导出文件中没有内容"}
	}

	// 加载全部英雄，按 ID 与名称建索引（ID 优先，名称兜底）。
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

	// 加载全部用户，按用户名匹配原作者。
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

		// 作者：优先原作者用户名，找不到则用导入管理员。
		author := authorID
		if ec.AuthorName != "" {
			if id, ok := userByName[ec.AuthorName]; ok {
				author = id
			}
		}

		// 校验装备/符文引用合法性（复用投稿校验逻辑）。
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
			status = model.ContentApproved // 导入的是已发布内容，默认已通过
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
