package repository

import (
	"gorm.io/gorm"

	"roadside-league/backend/internal/model"
)

type HeroRepository struct {
	db *gorm.DB
}

func NewHeroRepository(db *gorm.DB) *HeroRepository {
	return &HeroRepository{db: db}
}

func (r *HeroRepository) Create(h *model.Hero) error {
	return r.db.Create(h).Error
}

func (r *HeroRepository) List(lane string) ([]model.Hero, error) {
	q := r.db.Order("id ASC")
	if lane != "" {
		q = q.Where("line = ?", lane)
	}
	var heroes []model.Hero
	err := q.Find(&heroes).Error
	return heroes, err
}

func (r *HeroRepository) FindByDDragonID(ddID string) (*model.Hero, error) {
	var h model.Hero
	err := r.db.Where("ddragon_id = ?", ddID).First(&h).Error
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// UpsertDDragon 按 Data Dragon id 插入或更新英雄；保留管理员维护的 Lane 不变。
func (r *HeroRepository) UpsertDDragon(h *model.Hero) error {
	var existing model.Hero
	err := r.db.Where("ddragon_id = ?", h.DDragonID).First(&existing).Error
	if err == nil {
		return r.db.Model(&existing).Updates(map[string]any{
			"name":        h.Name,
			"title":       h.Title,
			"ddragon_key": h.DDragonKey,
			"avatar":      h.Avatar,
			"aliases":     h.Aliases,
			"roles":       h.Roles,
		}).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return r.db.Create(h).Error
}

// UpdateMetaByDDragon 按英文 id 更新别名与官方定位；英雄不存在或内容无变化时不写库。
func (r *HeroRepository) UpdateMetaByDDragon(ddID, aliases, roles string) error {
	return r.db.Model(&model.Hero{}).
		Where("ddragon_id = ? AND (aliases <> ? OR roles <> ?)", ddID, aliases, roles).
		Updates(map[string]any{"aliases": aliases, "roles": roles}).Error
}

// UpdateAliasesByDDragon 只更新内置别名（启动时拉取官方定位失败的降级路径）。
func (r *HeroRepository) UpdateAliasesByDDragon(ddID, aliases string) error {
	return r.db.Model(&model.Hero{}).
		Where("ddragon_id = ? AND aliases <> ?", ddID, aliases).
		Update("aliases", aliases).Error
}

func (r *HeroRepository) UpdateAvatar(id int64, avatar string) error {
	res := r.db.Model(&model.Hero{}).Where("id = ?", id).Update("avatar", avatar)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
