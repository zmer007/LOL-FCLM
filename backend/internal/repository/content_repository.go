package repository

import (
	"gorm.io/gorm"

	"roadside-league/backend/internal/model"
)

type ContentRepository struct {
	db *gorm.DB
}

func NewContentRepository(db *gorm.DB) *ContentRepository {
	return &ContentRepository{db: db}
}

func (r *ContentRepository) Create(c *model.Content) error {
	return r.db.Create(c).Error
}

func (r *ContentRepository) Update(c *model.Content) error {
	return r.db.Save(c).Error
}

func (r *ContentRepository) Get(id int64) (*model.Content, error) {
	var c model.Content
	if err := r.db.Where("id = ?", id).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ContentRepository) Delete(id int64) error {
	return r.db.Where("id = ?", id).Delete(&model.Content{}).Error
}

// ContentPage 分页结果。
type ContentPage struct {
	List  []model.Content
	Total int64
}

// ListPublic 对线页九宫格：存在已通过快照的内容（approved IS NOT NULL；
// 已通过内容被编辑重审或修订被拒时，旧快照仍对外可见），且（主视角是该英雄 或 对位是该英雄）。
// opponentHeroID>0 时再按对位英雄过滤（对线页对位搜索）。
func (r *ContentRepository) ListPublic(heroID, opponentHeroID int64, kind string, page, size int) (*ContentPage, error) {
	q := r.db.Model(&model.Content{}).
		Where("approved IS NOT NULL").
		Where("hero_id = ? OR opponent_hero_id = ?", heroID, heroID)
	if opponentHeroID > 0 {
		q = q.Where("opponent_hero_id = ?", opponentHeroID)
	}
	if kind == model.KindVideo || kind == model.KindArticle {
		q = q.Where("kind = ?", kind)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []model.Content
	err := q.Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return &ContentPage{List: list, Total: total}, nil
}

// ListByAuthor 我的整理；status 为空时返回全部状态。
func (r *ContentRepository) ListByAuthor(authorID int64, status string) ([]model.Content, error) {
	q := r.db.Where("author_id = ?", authorID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var list []model.Content
	err := q.Order("updated_at DESC").Find(&list).Error
	return list, err
}

// ListForReview 管理后台审批列表；status 为空时返回非通过内容（pending+rejected）。
func (r *ContentRepository) ListForReview(status string, page, size int) (*ContentPage, error) {
	q := r.db.Model(&model.Content{})
	switch status {
	case model.ContentPending, model.ContentApproved, model.ContentRejected:
		q = q.Where("status = ?", status)
	default:
		q = q.Where("status <> ?", model.ContentApproved)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []model.Content
	err := q.Order("updated_at DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return &ContentPage{List: list, Total: total}, err
}

// ListApprovedByHero 聚合某英雄主视角的有已通过快照的内容（提取推荐装备符文用）。
func (r *ContentRepository) ListApprovedByHero(heroID int64) ([]model.Content, error) {
	var list []model.Content
	err := r.db.Where("approved IS NOT NULL AND hero_id = ?", heroID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *ContentRepository) IncrViews(id int64) error {
	return r.db.Model(&model.Content{}).Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + 1")).Error
}

// ListAll 获取全部内容（按 ID 升序，用于导出备份）。
func (r *ContentRepository) ListAll() ([]model.Content, error) {
	var list []model.Content
	err := r.db.Order("id ASC").Find(&list).Error
	return list, err
}

// FindByIDs 按 ID 批量获取内容。
func (r *ContentRepository) FindByIDs(ids []int64) ([]model.Content, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []model.Content
	err := r.db.Where("id IN ?", ids).Order("id ASC").Find(&list).Error
	return list, err
}
