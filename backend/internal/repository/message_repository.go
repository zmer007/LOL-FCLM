package repository

import (
	"gorm.io/gorm"

	"roadside-league/backend/internal/model"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create 新增一条留言。
func (r *MessageRepository) Create(m *model.Message) error {
	return r.db.Create(m).Error
}

// ListAll 管理员查看所有留言，按创建时间倒序。
func (r *MessageRepository) ListAll() ([]model.Message, error) {
	var list []model.Message
	err := r.db.Order("created_at DESC").Find(&list).Error
	return list, err
}

// ListByUser 某用户查看自己的留言，倒序。
func (r *MessageRepository) ListByUser(userID int64) ([]model.Message, error) {
	var list []model.Message
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&list).Error
	return list, err
}
