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
	q := r.db.Order("sub_category ASC, id ASC")
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

// UpsertDDragon 按 category+name 插入或更新素材（同一分类下 Riot 名称唯一）。
func (r *AssetRepository) UpsertDDragon(a *model.GameAsset) error {
	var existing model.GameAsset
	err := r.db.Where("category = ? AND name = ?", a.Category, a.Name).First(&existing).Error
	if err == nil {
		return r.db.Model(&existing).Updates(map[string]any{
			"sub_category": a.SubCategory,
			"ddragon_id":   a.DDragonID,
			"version":      a.Version,
			"icon":         a.Icon,
		}).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return r.db.Create(a).Error
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
