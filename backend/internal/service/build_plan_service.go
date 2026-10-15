package service

import (
	"errors"
	"strconv"
	"unicode/utf8"

	"roadside-league/backend/internal/model"
	"roadside-league/backend/internal/repository"
)

// ErrPlanHeroNotFound 方案所属英雄不存在。
var ErrPlanHeroNotFound = errors.New("hero not found")

// PlanInput 管理员提交的单套方案（不含 id/序号/作者，由服务端补齐）。
type PlanInput struct {
	Vs    string           `json:"vs"`
	Items model.BuildItems `json:"items"`
	Runes model.BuildRunes `json:"runes"`
}

type BuildPlanService struct {
	heroes *repository.HeroRepository
	assets *repository.AssetRepository
	runes  *repository.RuneRepository
	plans  *repository.BuildPlanRepository
}

func NewBuildPlanService(
	heroes *repository.HeroRepository,
	assets *repository.AssetRepository,
	runes *repository.RuneRepository,
	plans *repository.BuildPlanRepository,
) *BuildPlanService {
	return &BuildPlanService{heroes: heroes, assets: assets, runes: runes, plans: plans}
}

// ListByHero 返回某英雄的方案。
func (s *BuildPlanService) ListByHero(heroID int64) ([]model.BuildPlan, error) {
	return s.plans.ListByHero(heroID)
}

// ReplaceByHero 校验后整体替换某英雄的方案，序号按数组顺序重排为 1..n，返回保存后的列表。
func (s *BuildPlanService) ReplaceByHero(heroID, authorID int64, inputs []PlanInput) ([]model.BuildPlan, error) {
	exists, err := s.heroes.Exists(heroID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrPlanHeroNotFound
	}
	if len(inputs) > 20 {
		return nil, &ValidationError{Msg: "方案数量不能超过 20 套"}
	}

	// 符文参照：系集合 + 符文索引
	trees, err := s.runes.ListTreesWithRunes()
	if err != nil {
		return nil, err
	}
	treeSet := map[int64]bool{}
	runeByID := map[int64]model.Rune{}
	for _, t := range trees {
		treeSet[t.ID] = true
		for _, rn := range t.Runes {
			runeByID[rn.ID] = rn
		}
	}

	// 收集所有被引用的素材 id（装备槽 + 召唤师技能槽）
	idSet := map[int64]struct{}{}
	for _, in := range inputs {
		for _, slots := range [][]*int64{in.Items.Start, in.Items.Summs, in.Items.Mid, in.Items.Core, in.Items.Final} {
			for _, p := range slots {
				if p != nil {
					idSet[*p] = struct{}{}
				}
			}
		}
	}
	ids := make([]int64, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	assetByID, err := s.assets.FindByIDs(ids)
	if err != nil {
		return nil, err
	}

	plans := make([]model.BuildPlan, 0, len(inputs))
	for idx, in := range inputs {
		label := func() string {
			if len(inputs) > 1 {
				return "第" + strconv.Itoa(idx+1) + "套方案"
			}
			return "方案"
		}()

		vs := in.Vs
		if utf8.RuneCountInString(vs) > 255 {
			return nil, &ValidationError{Msg: label + "的适用对抗文本过长"}
		}

		items, err := s.validateItems(label, in.Items, assetByID)
		if err != nil {
			return nil, err
		}
		runes, err := s.validateRunes(label, in.Runes, treeSet, runeByID)
		if err != nil {
			return nil, err
		}

		plans = append(plans, model.BuildPlan{
			HeroID: heroID, Seq: idx + 1, AuthorID: authorID,
			Vs: vs, Items: items, Runes: runes,
		})
	}

	if err := s.plans.ReplaceAll(heroID, plans); err != nil {
		return nil, err
	}
	return s.plans.ListByHero(heroID)
}

func (s *BuildPlanService) validateItems(label string, items model.BuildItems, assetByID map[int64]*model.GameAsset) (model.BuildItems, error) {
	checkLen := func(slots []*int64, n int, stage string) error {
		if len(slots) != n {
			return &ValidationError{Msg: label + "的" + stage + "必须为 " + strconv.Itoa(n) + " 个槽位"}
		}
		return nil
	}
	if err := checkLen(items.Start, model.BuildStageStartSlots, "出门装"); err != nil {
		return items, err
	}
	if err := checkLen(items.Summs, model.BuildStageSummsSlots, "召唤师技能"); err != nil {
		return items, err
	}
	if err := checkLen(items.Mid, model.BuildStageMidSlots, "中期过渡"); err != nil {
		return items, err
	}
	if err := checkLen(items.Core, model.BuildStageCoreSlots, "核心装备"); err != nil {
		return items, err
	}
	if err := checkLen(items.Final, model.BuildStageFinalSlots, "后期神装"); err != nil {
		return items, err
	}

	// 装备槽只能放装备，召唤师技能槽只能放召唤师技能
	checkCategory := func(slots []*int64, want, stage string) error {
		for _, p := range slots {
			if p == nil {
				continue
			}
			a, ok := assetByID[*p]
			if !ok {
				return &ValidationError{Msg: label + "引用了不存在的素材"}
			}
			if a.Category != want {
				return &ValidationError{Msg: label + "的" + stage + "槽位素材类型不正确"}
			}
		}
		return nil
	}
	for _, stage := range []struct {
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
		if err := checkCategory(stage.slots, stage.want, stage.name); err != nil {
			return items, err
		}
	}
	return items, nil
}

func (s *BuildPlanService) validateRunes(
	label string,
	r model.BuildRunes,
	treeSet map[int64]bool,
	runeByID map[int64]model.Rune,
) (model.BuildRunes, error) {
	// 主副系都未选：视为本套不配置符文，归一化为零值
	if r.Primary == 0 && r.Secondary == 0 {
		return model.BuildRunes{}, nil
	}
	if !treeSet[r.Primary] || !treeSet[r.Secondary] {
		return r, &ValidationError{Msg: label + "的符文系不合法"}
	}
	if r.Primary == r.Secondary {
		return r, &ValidationError{Msg: label + "的主系和副系不能相同"}
	}
	if len(r.PrimaryPicks) != 4 {
		return r, &ValidationError{Msg: label + "的主系符文选择不完整"}
	}
	if len(r.SecondaryPicks) != 3 {
		return r, &ValidationError{Msg: label + "的副系符文选择不完整"}
	}
	if len(r.Shards) != 3 {
		return r, &ValidationError{Msg: label + "的属性碎片选择不完整"}
	}

	// 主系：第 0 个为基石(row=0)，其后 row=1..3
	for i, rid := range r.PrimaryPicks {
		if rid == 0 {
			return r, &ValidationError{Msg: label + "的主系符文未选满"}
		}
		rn, ok := runeByID[rid]
		if !ok || rn.TreeID != r.Primary || rn.Row != i {
			return r, &ValidationError{Msg: label + "的主系符文不合法"}
		}
	}
	// 副系：三排选两排，nil=该排不选；选中的须属于副系且 row=i+1
	pickedSecondary := 0
	for i, p := range r.SecondaryPicks {
		if p == nil {
			continue
		}
		pickedSecondary++
		rn, ok := runeByID[*p]
		if !ok || rn.TreeID != r.Secondary || rn.Row != i+1 {
			return r, &ValidationError{Msg: label + "的副系符文不合法"}
		}
	}
	if pickedSecondary != 2 {
		return r, &ValidationError{Msg: label + "的副系须从三排中选择两排"}
	}
	for _, sh := range r.Shards {
		if sh < 0 || sh > 2 {
			return r, &ValidationError{Msg: label + "的属性碎片不合法"}
		}
	}
	return r, nil
}
