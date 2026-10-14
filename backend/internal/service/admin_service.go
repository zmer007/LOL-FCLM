package service

import (
	"errors"
	"io"

	"gorm.io/gorm"

	"roadside-league/backend/internal/model"
	"roadside-league/backend/internal/repository"
	"roadside-league/backend/internal/storage"
)

var (
	ErrHeroExists  = errors.New("hero already exists")
	ErrAssetExists = errors.New("asset already exists")
)

// 各类目允许的子分类
var validSubCategory = map[string]map[string]bool{
	model.AssetCategoryItem:  {"攻击": true, "法术": true, "防御": true, "移动": true, "消耗": true},
	model.AssetCategorySkill: {"召唤师技能": true},
	model.AssetCategoryRune:  {"精密": true, "主宰": true, "巫术": true, "坚决": true, "启迪": true},
}

var validLane = map[string]bool{
	model.LaneTop: true, model.LaneJungle: true, model.LaneMid: true,
	model.LaneBottom: true, model.LaneSupport: true,
}

type AdminService struct {
	heroes *repository.HeroRepository
	assets *repository.AssetRepository
	store  *storage.LocalStore
}

func NewAdminService(heroes *repository.HeroRepository, assets *repository.AssetRepository, store *storage.LocalStore) *AdminService {
	return &AdminService{heroes: heroes, assets: assets, store: store}
}

// ---------- 英雄 ----------

func (s *AdminService) ListHeroes(lane string) ([]model.Hero, error) {
	return s.heroes.List(lane)
}

func (s *AdminService) CreateHero(name, lane string) (*model.Hero, error) {
	if name == "" {
		return nil, &ValidationError{Msg: "英雄名称不能为空"}
	}
	if !validLane[lane] {
		return nil, &ValidationError{Msg: "分路不合法"}
	}
	h := &model.Hero{Name: name, Lane: lane}
	if err := s.heroes.Create(h); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrHeroExists
		}
		return nil, err
	}
	return h, nil
}

func (s *AdminService) UpdateHeroAvatar(id int64, filename, contentType string, src io.Reader, size int64) error {
	url, err := s.store.SaveIcon(filename, contentType, src, size)
	if err != nil {
		return err
	}
	if err := s.heroes.UpdateAvatar(id, url); err != nil {
		return err
	}
	return nil
}

// ---------- 素材 ----------

func (s *AdminService) ListAssets(category, subCategory string) ([]model.GameAsset, error) {
	if category != "" && category != model.AssetCategoryItem &&
		category != model.AssetCategorySkill && category != model.AssetCategoryRune {
		return nil, &ValidationError{Msg: "素材分类不合法"}
	}
	return s.assets.List(category, subCategory)
}

func (s *AdminService) CreateAsset(category, subCategory, name string) (*model.GameAsset, error) {
	subs, catOK := validSubCategory[category]
	if !catOK {
		return nil, &ValidationError{Msg: "素材分类不合法"}
	}
	if category == model.AssetCategoryItem || category == model.AssetCategoryRune {
		if !subs[subCategory] {
			return nil, &ValidationError{Msg: "子分类不合法"}
		}
	}
	if category == model.AssetCategorySkill && subCategory != "" && !subs[subCategory] {
		return nil, &ValidationError{Msg: "子分类不合法"}
	}
	if name == "" {
		return nil, &ValidationError{Msg: "素材名称不能为空"}
	}
	a := &model.GameAsset{Category: category, SubCategory: subCategory, Name: name}
	if err := s.assets.Create(a); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrAssetExists
		}
		return nil, err
	}
	return a, nil
}

func (s *AdminService) UpdateAssetIcon(id int64, filename, contentType string, src io.Reader, size int64) error {
	url, err := s.store.SaveIcon(filename, contentType, src, size)
	if err != nil {
		return err
	}
	return s.assets.UpdateIcon(id, url)
}

func (s *AdminService) DeleteAsset(id int64) error {
	return s.assets.Delete(id)
}
