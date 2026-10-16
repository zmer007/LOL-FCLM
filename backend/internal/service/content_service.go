package service

import (
	"errors"
	"mime/multipart"
	"strings"
	"unicode/utf8"

	"roadside-league/backend/internal/model"
	"roadside-league/backend/internal/repository"
	"roadside-league/backend/internal/storage"
)

var (
	ErrContentNotFound  = errors.New("content not found")
	ErrContentForbidden = errors.New("content forbidden")
)

// ContentInput 投稿/编辑提交的表单业务字段（文件在 handler 单独传）。
type ContentInput struct {
	HeroID         int64            `json:"heroId"`
	OpponentHeroID int64            `json:"opponentHeroId"`
	Rank           string           `json:"rank"`
	Kind           string           `json:"kind"`
	Title          string           `json:"title"`
	Body           string           `json:"body"`
	DurationSec    int              `json:"durationSec"`
	Items          model.BuildItems `json:"items"`
	Runes          model.BuildRunes `json:"runes"`
	BuildNote      string           `json:"buildNote"`  // 装备符文思路
	KeptImages     []string         `json:"keptImages"` // 图文编辑时保留的旧图 URL
}

// ContentVO 对外/管理统一视图。公开场景字段来自 Approved 快照。
type ContentVO struct {
	ID             int64            `json:"id"`
	AuthorID       int64            `json:"authorId"`
	AuthorName     string           `json:"authorName"`
	Kind           string           `json:"kind"`
	Status         string           `json:"status"`
	RejectReason   string           `json:"rejectReason"`
	HeroID         int64            `json:"heroId"`
	HeroName       string           `json:"heroName"`
	HeroTitle      string           `json:"heroTitle"`
	HeroAvatar     string           `json:"heroAvatar"`
	OpponentID     int64            `json:"opponentId"`
	OpponentName   string           `json:"opponentName"`
	OpponentTitle  string           `json:"opponentTitle"`
	OpponentAvatar string           `json:"opponentAvatar"`
	Rank           string           `json:"rank"`
	Perspective    string           `json:"perspective"` // 相对查询英雄：self=操作你 / opponent=被操作
	CanEdit        bool             `json:"canEdit"`
	Title          string           `json:"title"`
	Body           string           `json:"body"`
	VideoURL       string           `json:"videoUrl"`
	DurationSec    int              `json:"durationSec"`
	Images         []string         `json:"images"`
	Items          model.BuildItems `json:"items"`
	Runes          model.BuildRunes `json:"runes"`
	BuildNote      string           `json:"buildNote"`
	Views          int              `json:"views"`
	CreatedAt      string           `json:"createdAt"`
	UpdatedAt      string           `json:"updatedAt"`
	Reviewing      bool             `json:"reviewing"` // 管理员审核视图：展示待审实时版本而非公开快照
}

// BuildVO 对线页提取的装备符文方案（来源于已通过内容）。
type BuildVO struct {
	ContentID int64            `json:"contentId"`
	Kind      string           `json:"kind"`
	Title     string           `json:"title"`
	Vs        string           `json:"vs"` // 对位英雄短名
	Items     model.BuildItems `json:"items"`
	Runes     model.BuildRunes `json:"runes"`
	BuildNote string           `json:"buildNote"`
}

type ContentService struct {
	contents *repository.ContentRepository
	heroes   *repository.HeroRepository
	users    *repository.UserRepository
	assets   *repository.AssetRepository
	runes    *repository.RuneRepository
	store    *storage.LocalStore
}

func NewContentService(
	contents *repository.ContentRepository,
	heroes *repository.HeroRepository,
	users *repository.UserRepository,
	assets *repository.AssetRepository,
	runes *repository.RuneRepository,
	store *storage.LocalStore,
) *ContentService {
	return &ContentService{
		contents: contents, heroes: heroes, users: users, assets: assets, runes: runes, store: store,
	}
}

// ---------- 投稿 / 编辑 ----------

// FileSet 投稿携带的文件：单个视频或多张图文配图。
type FileSet struct {
	Video  *multipart.FileHeader
	Images []*multipart.FileHeader
}

func (s *ContentService) saveFiles(fs FileSet) (videoURL string, images []string, err error) {
	if fs.Video != nil {
		u, e := s.store.SaveVideo(fs.Video)
		if e != nil {
			return "", nil, e
		}
		videoURL = u
	}
	for _, fh := range fs.Images {
		f, e := fh.Open()
		if e != nil {
			return "", nil, e
		}
		u, e := s.store.SaveArticleImage(fh.Filename, f, fh.Size)
		f.Close()
		if e != nil {
			return "", nil, e
		}
		images = append(images, u)
	}
	return videoURL, images, nil
}

func (s *ContentService) Create(in ContentInput, fs FileSet, authorID int64) (*model.Content, error) {
	if err := s.validateInput(&in); err != nil {
		return nil, err
	}
	videoURL, images, err := s.saveFiles(fs)
	if err != nil {
		return nil, err
	}
	if in.Kind == model.KindVideo && videoURL == "" {
		return nil, &ValidationError{Msg: "请上传视频文件"}
	}
	if in.Kind == model.KindArticle && len(images) > 9 {
		return nil, &ValidationError{Msg: "配图最多 9 张"}
	}

	c := &model.Content{
		AuthorID: authorID, HeroID: in.HeroID, OpponentHeroID: in.OpponentHeroID,
		Rank: in.Rank,
		Kind: in.Kind, Title: in.Title, Body: in.Body, DurationSec: in.DurationSec,
		VideoURL: videoURL, Images: images, Items: in.Items, Runes: in.Runes, BuildNote: in.BuildNote,
		Status: model.ContentPending,
	}
	if err := s.contents.Create(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *ContentService) Update(id, uid int64, isAdmin bool, in ContentInput, fs FileSet) (*model.Content, error) {
	c, err := s.contents.Get(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrContentNotFound
	}
	if err != nil {
		return nil, err
	}
	if c.AuthorID != uid && !isAdmin {
		return nil, ErrContentForbidden
	}
	if err := s.validateInput(&in); err != nil {
		return nil, err
	}

	videoURL := c.VideoURL
	if fs.Video != nil {
		u, e := s.store.SaveVideo(fs.Video)
		if e != nil {
			return nil, e
		}
		videoURL = u
	}
	images := in.KeptImages
	for _, fh := range fs.Images {
		f, e := fh.Open()
		if e != nil {
			return nil, e
		}
		u, e := s.store.SaveArticleImage(fh.Filename, f, fh.Size)
		f.Close()
		if e != nil {
			return nil, e
		}
		images = append(images, u)
	}
	if in.Kind == model.KindVideo && videoURL == "" {
		return nil, &ValidationError{Msg: "请上传视频文件"}
	}
	if in.Kind == model.KindArticle && len(images) > 9 {
		return nil, &ValidationError{Msg: "配图最多 9 张"}
	}

	c.HeroID = in.HeroID
	c.OpponentHeroID = in.OpponentHeroID
	c.Rank = in.Rank
	c.Kind = in.Kind
	c.Title = in.Title
	c.Body = in.Body
	c.DurationSec = in.DurationSec
	c.Items = in.Items
	c.Runes = in.Runes
	c.BuildNote = in.BuildNote
	c.VideoURL = videoURL
	c.Images = images
	c.Status = model.ContentPending // 重新审核
	c.RejectReason = ""
	// Approved 快照不动 → 复审通过前对外仍展示旧版
	if err := s.contents.Update(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *ContentService) Delete(id, uid int64, isAdmin bool) error {
	c, err := s.contents.Get(id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrContentNotFound
	}
	if err != nil {
		return err
	}
	if c.AuthorID != uid && !isAdmin {
		return ErrContentForbidden
	}
	return s.contents.Delete(id)
}

// ---------- 审批 ----------

func (s *ContentService) Approve(id int64) error {
	c, err := s.contents.Get(id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrContentNotFound
	}
	if err != nil {
		return err
	}
	c.Approved = &model.ContentSnapshot{
		Title: c.Title, Body: c.Body, VideoURL: c.VideoURL, DurationSec: c.DurationSec,
		Images: c.Images, OpponentHeroID: c.OpponentHeroID, Rank: c.Rank,
		Items: c.Items, Runes: c.Runes, BuildNote: c.BuildNote,
	}
	c.Status = model.ContentApproved
	c.RejectReason = ""
	return s.contents.Update(c)
}

func (s *ContentService) Reject(id int64, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return &ValidationError{Msg: "请填写拒绝理由"}
	}
	if utf8.RuneCountInString(reason) > 255 {
		return &ValidationError{Msg: "拒绝理由不能超过 255 字"}
	}
	c, err := s.contents.Get(id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrContentNotFound
	}
	if err != nil {
		return err
	}
	c.Status = model.ContentRejected
	c.RejectReason = reason
	return s.contents.Update(c)
}

// ---------- 查询与视图组装 ----------

func (s *ContentService) enrich(list []model.Content, queryHeroID, viewerUID int64, isAdmin, useSnapshot bool) ([]ContentVO, error) {
	userIDs := map[int64]struct{}{}
	heroIDs := map[int64]struct{}{}
	for _, c := range list {
		userIDs[c.AuthorID] = struct{}{}
		heroIDs[c.HeroID] = struct{}{}
		heroIDs[c.OpponentHeroID] = struct{}{}
	}
	uids := keys(userIDs)
	hids := keys(heroIDs)
	userMap, err := s.users.FindByIDs(uids)
	if err != nil {
		return nil, err
	}
	heroMap, err := s.heroes.FindByIDs(hids)
	if err != nil {
		return nil, err
	}

	out := make([]ContentVO, 0, len(list))
	for _, c := range list {
		vo := ContentVO{
			ID: c.ID, AuthorID: c.AuthorID, Kind: c.Kind, Status: c.Status,
			RejectReason: c.RejectReason, HeroID: c.HeroID, Views: c.Views,
			CreatedAt: c.CreatedAt.Format("2006-01-02 15:04"), UpdatedAt: c.UpdatedAt.Format("2006-01-02 15:04"),
			CanEdit: c.AuthorID == viewerUID || isAdmin,
		}
		if u := userMap[c.AuthorID]; u != nil {
			vo.AuthorName = u.Username
		}
		fillHero := func(id int64) (name, title, avatar string, oid int64) {
			if h := heroMap[id]; h != nil {
				return h.Name, h.Title, h.Avatar, h.ID
			}
			return
		}
		vo.HeroName, vo.HeroTitle, vo.HeroAvatar, _ = fillHero(c.HeroID)
		if queryHeroID != 0 {
			if c.HeroID == queryHeroID {
				vo.Perspective = "self"
			} else {
				vo.Perspective = "opponent"
			}
		}

		// 公开视图取已通过快照；管理/本人视图取实时字段
		snap := c.Approved
		if !useSnapshot || snap == nil {
			vo.Title, vo.Body, vo.VideoURL, vo.DurationSec = c.Title, c.Body, c.VideoURL, c.DurationSec
			vo.Images, vo.Items, vo.Runes, vo.BuildNote = c.Images, c.Items, c.Runes, c.BuildNote
			vo.Rank = c.Rank
			vo.OpponentName, vo.OpponentTitle, vo.OpponentAvatar, vo.OpponentID = fillHero(c.OpponentHeroID)
		} else {
			vo.Title, vo.Body, vo.VideoURL, vo.DurationSec = snap.Title, snap.Body, snap.VideoURL, snap.DurationSec
			vo.Images, vo.Items, vo.Runes, vo.BuildNote = snap.Images, snap.Items, snap.Runes, snap.BuildNote
			vo.Rank = snap.Rank
			vo.OpponentName, vo.OpponentTitle, vo.OpponentAvatar, vo.OpponentID = fillHero(snap.OpponentHeroID)
		}
		out = append(out, vo)
	}
	return out, nil
}

func keys(m map[int64]struct{}) []int64 {
	out := make([]int64, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func (s *ContentService) PublicList(heroID, opponentHeroID int64, kind string, page, size int) (*repository.ContentPage, []ContentVO, error) {
	pg, err := s.contents.ListPublic(heroID, opponentHeroID, kind, page, size)
	if err != nil {
		return nil, nil, err
	}
	vos, err := s.enrich(pg.List, heroID, 0, false, true)
	return pg, vos, err
}

// PublicGet 详情：
// 已通过内容对所有人展示最近通过快照（重审期间旧版继续可见）并计数；
// 从未通过的内容仅作者/管理员可见实时草稿；
// 管理员查看任何未过审内容时一律展示待审实时版本（Reviewing=true），供审核页观看与操作。
func (s *ContentService) PublicGet(id, viewerUID int64, isAdmin bool) (*ContentVO, error) {
	c, err := s.contents.Get(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrContentNotFound
	}
	if err != nil {
		return nil, err
	}
	useSnapshot := true
	if c.Approved == nil {
		// 从未通过：只有作者本人或管理员能看实时内容
		if c.AuthorID != viewerUID && !isAdmin {
			return nil, ErrContentNotFound
		}
		useSnapshot = false
	} else if isAdmin && c.Status != model.ContentApproved {
		// 管理员审核视图：重审/被拒期间看实时待审版本，不增加公开播放计数
		useSnapshot = false
	} else if c.Status == model.ContentApproved {
		_ = s.contents.IncrViews(id)
		c.Views++
	}
	vos, err := s.enrich([]model.Content{*c}, 0, viewerUID, isAdmin, useSnapshot)
	if err != nil || len(vos) == 0 {
		return nil, err
	}
	vos[0].Reviewing = isAdmin && c.Status != model.ContentApproved
	return &vos[0], nil
}

func (s *ContentService) MyList(uid int64, status string) ([]ContentVO, error) {
	list, err := s.contents.ListByAuthor(uid, status)
	if err != nil {
		return nil, err
	}
	return s.enrich(list, 0, uid, false, false)
}

func (s *ContentService) ReviewList(status string, page, size int) (*repository.ContentPage, []ContentVO, error) {
	pg, err := s.contents.ListForReview(status, page, size)
	if err != nil {
		return nil, nil, err
	}
	vos, err := s.enrich(pg.List, 0, 0, true, false)
	return pg, vos, err
}

// HeroBuilds 从某英雄主视角的已通过内容提取装备符文方案（装备或符文非空才算）。
func (s *ContentService) HeroBuilds(heroID int64) ([]BuildVO, error) {
	list, err := s.contents.ListApprovedByHero(heroID)
	if err != nil {
		return nil, err
	}
	idSet := map[int64]struct{}{}
	for _, c := range list {
		if c.Approved != nil {
			idSet[c.Approved.OpponentHeroID] = struct{}{}
		}
	}
	heroMap, _ := s.heroes.FindByIDs(keys(idSet))

	out := make([]BuildVO, 0, len(list))
	for _, c := range list {
		snap := c.Approved
		if snap == nil {
			continue
		}
		if snap.Items.IsEmpty() && snap.Runes.IsEmpty() {
			continue
		}
		vs := ""
		if h := heroMap[snap.OpponentHeroID]; h != nil {
			vs = h.Title
		}
		out = append(out, BuildVO{
			ContentID: c.ID, Kind: c.Kind, Title: snap.Title, Vs: vs,
			Items: snap.Items, Runes: snap.Runes, BuildNote: snap.BuildNote,
		})
	}
	return out, nil
}

// ---------- 校验 ----------

func (s *ContentService) validateInput(in *ContentInput) error {
	if in.Kind != model.KindVideo && in.Kind != model.KindArticle {
		return &ValidationError{Msg: "内容类型不合法"}
	}
	in.Title = strings.TrimSpace(in.Title)
	if utf8.RuneCountInString(in.Title) == 0 {
		return &ValidationError{Msg: "请填写标题"}
	}
	if utf8.RuneCountInString(in.Title) > 120 {
		return &ValidationError{Msg: "标题不能超过 120 字"}
	}
	if in.HeroID == 0 {
		return &ValidationError{Msg: "缺少主英雄信息"}
	}
	if in.OpponentHeroID == 0 {
		return &ValidationError{Msg: "请选择对位英雄"}
	}
	if in.HeroID == in.OpponentHeroID {
		return &ValidationError{Msg: "对位英雄不能与主英雄相同"}
	}
	heroOK1, err := s.heroes.Exists(in.HeroID)
	if err != nil {
		return err
	}
	heroOK2, err := s.heroes.Exists(in.OpponentHeroID)
	if err != nil {
		return err
	}
	if !heroOK1 || !heroOK2 {
		return &ValidationError{Msg: "英雄不存在"}
	}
	if utf8.RuneCountInString(in.Body) > 10000 {
		return &ValidationError{Msg: "正文不能超过 10000 字"}
	}
	if in.DurationSec < 0 {
		return &ValidationError{Msg: "视频时长不合法"}
	}
	if err := s.validateItems(in.Items); err != nil {
		return err
	}
	if err := s.validateRunes(in.Runes); err != nil {
		return err
	}
	in.BuildNote = strings.TrimSpace(in.BuildNote)
	if !in.Items.IsEmpty() || !in.Runes.IsEmpty() {
		if in.BuildNote == "" {
			return &ValidationError{Msg: "请填写这套装备符文的思路"}
		}
	}
	if utf8.RuneCountInString(in.BuildNote) > 500 {
		return &ValidationError{Msg: "装备符文思路不能超过 500 字"}
	}
	in.Rank = strings.TrimSpace(in.Rank)
	if in.Rank == "" {
		return &ValidationError{Msg: "请选择段位"}
	}
	if !model.ValidRank(in.Rank) {
		return &ValidationError{Msg: "段位不合法"}
	}
	return nil
}

func (s *ContentService) validateItems(items model.BuildItems) error {
	if len(items.Start) != model.BuildStageStartSlots ||
		len(items.Summs) != model.BuildStageSummsSlots ||
		len(items.Mid) != model.BuildStageMidSlots ||
		len(items.Core) != model.BuildStageCoreSlots ||
		len(items.Final) != model.BuildStageFinalSlots {
		return &ValidationError{Msg: "装备分区数据不完整"}
	}
	idSet := map[int64]struct{}{}
	for _, slots := range [][]*int64{items.Start, items.Summs, items.Mid, items.Core, items.Final} {
		for _, p := range slots {
			if p != nil {
				idSet[*p] = struct{}{}
			}
		}
	}
	if len(idSet) == 0 {
		return nil
	}
	assetByID, err := s.assets.FindByIDs(keys(idSet))
	if err != nil {
		return err
	}
	checkCategory := func(slots []*int64, want, stage string) error {
		for _, p := range slots {
			if p == nil {
				continue
			}
			a, ok := assetByID[*p]
			if !ok {
				return &ValidationError{Msg: "引用了不存在的装备素材"}
			}
			if a.Category != want {
				return &ValidationError{Msg: stage + "槽位素材类型不正确"}
			}
		}
		return nil
	}
	for _, st := range []struct {
		slots []*int64
		want  string
		name  string
	}{
		{items.Start, model.AssetCategoryItem, "出门装"},
		{items.Mid, model.AssetCategoryItem, "中期过渡"},
		{items.Core, model.AssetCategoryItem, "核心装备"},
		{items.Final, model.AssetCategoryItem, "后期神装"},
		{items.Summs, model.AssetCategorySkill, "召唤师技能"},
	} {
		if err := checkCategory(st.slots, st.want, st.name); err != nil {
			return err
		}
	}
	return nil
}

func (s *ContentService) validateRunes(r model.BuildRunes) error {
	if r.Primary == 0 && r.Secondary == 0 {
		return nil // 符文整体可留空
	}
	trees, err := s.runes.ListTreesWithRunes()
	if err != nil {
		return err
	}
	treeSet := map[int64]bool{}
	runeByID := map[int64]model.Rune{}
	for _, t := range trees {
		treeSet[t.ID] = true
		for _, rn := range t.Runes {
			runeByID[rn.ID] = rn
		}
	}
	if !treeSet[r.Primary] || !treeSet[r.Secondary] {
		return &ValidationError{Msg: "符文系不合法"}
	}
	if r.Primary == r.Secondary {
		return &ValidationError{Msg: "主系和副系符文不能相同"}
	}
	if len(r.PrimaryPicks) != 4 || len(r.SecondaryPicks) != 3 || len(r.Shards) != 3 {
		return &ValidationError{Msg: "符文选择数据不完整"}
	}
	for i, rid := range r.PrimaryPicks {
		if rid == 0 {
			return &ValidationError{Msg: "主系符文未选满"}
		}
		rn, ok := runeByID[rid]
		if !ok || rn.TreeID != r.Primary || rn.Row != i {
			return &ValidationError{Msg: "主系符文不合法"}
		}
	}
	picked := 0
	for i, p := range r.SecondaryPicks {
		if p == nil {
			continue
		}
		picked++
		rn, ok := runeByID[*p]
		if !ok || rn.TreeID != r.Secondary || rn.Row != i+1 {
			return &ValidationError{Msg: "副系符文不合法"}
		}
	}
	if picked != 2 {
		return &ValidationError{Msg: "副系符文须从三排中选择两排"}
	}
	for _, sh := range r.Shards {
		if sh < -1 || sh > 2 {
			return &ValidationError{Msg: "属性碎片不合法"}
		}
	}
	return nil
}
