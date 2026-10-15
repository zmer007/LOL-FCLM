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
		}).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return r.db.Create(h).Error
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
