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
	q := r.db.Order("id DESC")
	if lane != "" {
		q = q.Where("line = ?", lane)
	}
	var heroes []model.Hero
	err := q.Find(&heroes).Error
	return heroes, err
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
