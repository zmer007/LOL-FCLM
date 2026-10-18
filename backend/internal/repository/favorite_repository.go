package repository

import (
	"gorm.io/gorm"

	"roadside-league/backend/internal/model"
)

type FavoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

// Create 新增收藏记录。
func (r *FavoriteRepository) Create(f *model.Favorite) error {
	return r.db.Create(f).Error
}

// Delete 删除某用户对某内容的收藏。
func (r *FavoriteRepository) Delete(userID, contentID int64) error {
	return r.db.Where("user_id = ? AND content_id = ?", userID, contentID).Delete(&model.Favorite{}).Error
}

// Exists 判断某用户是否已收藏某内容。
func (r *FavoriteRepository) Exists(userID, contentID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.Favorite{}).Where("user_id = ? AND content_id = ?", userID, contentID).Count(&count).Error
	return count > 0, err
}

// ListByUser 列出某用户收藏的所有内容 ID，按收藏时间倒序。
func (r *FavoriteRepository) ListByUser(userID int64) ([]model.Favorite, error) {
	var list []model.Favorite
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&list).Error
	return list, err
}

// ExistsBatch 批量判断某用户是否收藏了给定内容 ID 集合，返回已收藏的 ID 集合。
func (r *FavoriteRepository) ExistsBatch(userID int64, contentIDs []int64) (map[int64]struct{}, error) {
	if len(contentIDs) == 0 {
		return map[int64]struct{}{}, nil
	}
	var favs []model.Favorite
	err := r.db.Where("user_id = ? AND content_id IN ?", userID, contentIDs).Find(&favs).Error
	if err != nil {
		return nil, err
	}
	out := make(map[int64]struct{}, len(favs))
	for _, f := range favs {
		out[f.ContentID] = struct{}{}
	}
	return out, nil
}
