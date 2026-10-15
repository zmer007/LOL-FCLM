package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"roadside-league/backend/internal/model"
)

type RuneRepository struct {
	db *gorm.DB
}

func NewRuneRepository(db *gorm.DB) *RuneRepository {
	return &RuneRepository{db: db}
}

// UpsertTree 按官方系 id 插入或更新符文系；冲突时只更新业务字段，保留 created_at。
func (r *RuneRepository) UpsertTree(t *model.RuneTree) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"key", "name", "icon", "sort", "updated_at"}),
	}).Create(t).Error
}

// UpsertRune 按官方符文 id 插入或更新符文；冲突时只更新业务字段，保留 created_at。
func (r *RuneRepository) UpsertRune(rn *model.Rune) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"tree_id", "name", "icon", "slot_row", "is_keystone", "sort", "updated_at"}),
	}).Create(rn).Error
}

// ListTreesWithRunes 返回全部符文系（按 sort）及其符文（按排次、排内顺序）。
func (r *RuneRepository) ListTreesWithRunes() ([]model.RuneTree, error) {
	var trees []model.RuneTree
	err := r.db.Order("sort ASC").
		Preload("Runes", func(db *gorm.DB) *gorm.DB {
			return db.Order("slot_row ASC, sort ASC")
		}).
		Find(&trees).Error
	return trees, err
}
