package model

import "time"

// Message 普通用户给管理员的留言，仅管理员可见，不公开展示。
type Message struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null;index" json:"userId"`
	Username  string    `gorm:"size:32;not null" json:"username"`
	Content   string    `gorm:"size:500;not null" json:"content"`
	IsRead    bool      `gorm:"not null;default:false" json:"isRead"`
	CreatedAt time.Time `json:"createdAt"`
}
