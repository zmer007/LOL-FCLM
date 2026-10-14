package model

import "time"

// 分路取值
const (
	LaneTop     = "上单"
	LaneJungle  = "打野"
	LaneMid     = "中单"
	LaneBottom  = "下路"
	LaneSupport = "辅助"
)

// Hero 英雄。Avatar 为图标相对 URL（/media/icons/...），由管理员上传。
type Hero struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:32;uniqueIndex;not null" json:"name"`
	Lane      string    `gorm:"size:16;not null;index" json:"lane"`
	Avatar    string    `gorm:"size:255" json:"avatar"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
