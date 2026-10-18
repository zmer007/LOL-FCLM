package model

import "time"

// Favorite 用户收藏的内容。
type Favorite struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"index:idx_user_content,unique;not null" json:"userId"`
	ContentID int64     `gorm:"index:idx_user_content,unique;not null" json:"contentId"`
	CreatedAt time.Time `json:"createdAt"`
}

func (Favorite) TableName() string { return "favorites" }
