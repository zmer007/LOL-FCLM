package repository

import (
	"gorm.io/gorm"

	"roadside-league/backend/internal/model"
)

type BuildPlanRepository struct {
	db *gorm.DB
}

func NewBuildPlanRepository(db *gorm.DB) *BuildPlanRepository {
	return &BuildPlanRepository{db: db}
}

// ListByHero 返回某英雄的全部方案，按序号升序。
func (r *BuildPlanRepository) ListByHero(heroID int64) ([]model.BuildPlan, error) {
	var plans []model.BuildPlan
	err := r.db.Where("hero_id = ?", heroID).Order("seq ASC").Find(&plans).Error
	return plans, err
}

// ReplaceAll 用新方案整体替换某英雄的方案（事务内先删后插）；空切片表示清空。
// 调用方需已设置好每个方案的 Seq/AuthorID/HeroID。
func (r *BuildPlanRepository) ReplaceAll(heroID int64, plans []model.BuildPlan) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("hero_id = ?", heroID).Delete(&model.BuildPlan{}).Error; err != nil {
			return err
		}
		if len(plans) == 0 {
			return nil
		}
		return tx.Create(&plans).Error
	})
}
