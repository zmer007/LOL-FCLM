package repository

import (
	"gorm.io/gorm"

	"roadside-league/backend/internal/model"
)

type AssetRepository struct {
	db *gorm.DB
}

func NewAssetRepository(db *gorm.DB) *AssetRepository {
	return &AssetRepository{db: db}
}

func (r *AssetRepository) Create(a *model.GameAsset) error {
	return r.db.Create(a).Error
}

func (r *AssetRepository) List(category, subCategory string) ([]model.GameAsset, error) {
	q := r.db.Order("category ASC, sub_category ASC, id DESC")
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if subCategory != "" {
		q = q.Where("sub_category = ?", subCategory)
	}
	var assets []model.GameAsset
	err := q.Find(&assets).Error
	return assets, err
}

func (r *AssetRepository) UpdateIcon(id int64, icon string) error {
	res := r.db.Model(&model.GameAsset{}).Where("id = ?", id).Update("icon", icon)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AssetRepository) Delete(id int64) error {
	res := r.db.Delete(&model.GameAsset{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
