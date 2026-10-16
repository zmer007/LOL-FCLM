// Package model 定义 GORM 数据模型。
package model

import "time"

// 用户状态
const (
	UserStatusActive   = "active"
	UserStatusDisabled = "disabled"
)

// User 站点用户。Role 取值：user / admin；Status 取值：active / disabled。
type User struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username      string    `gorm:"size:32;uniqueIndex;not null" json:"username"`
	Email         string    `gorm:"size:128;uniqueIndex" json:"email"`
	EmailVerified bool      `gorm:"not null;default:false" json:"emailVerified"`
	PasswordHash  string    `gorm:"size:100;not null" json:"-"`
	Avatar        string    `gorm:"size:255" json:"avatar"`
	Role          string    `gorm:"size:16;not null;default:user;index" json:"role"`
	Status        string    `gorm:"size:16;not null;default:active;index" json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
