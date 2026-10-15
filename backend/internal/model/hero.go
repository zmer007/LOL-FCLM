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

// Hero 英雄。数据来自 Data Dragon：
//
//	DDragonID 为英文 id（如 Ahri），Name 为中文名，Title 为称号。
//	Lane（分路）官方数据中没有，默认空，由管理员后续维护。
//	Avatar 为本地图标 URL（/media/icons/heroes/...）。
type Hero struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string    `gorm:"size:32;uniqueIndex;not null" json:"name"`
	Title      string    `gorm:"size:32" json:"title"`
	Lane       string    `gorm:"size:16;not null;default:'';index" json:"lane"`
	DDragonID  string    `gorm:"column:ddragon_id;size:32;uniqueIndex" json:"ddragonId"`
	DDragonKey string    `gorm:"column:ddragon_key;size:16" json:"ddragonKey"`
	Avatar     string    `gorm:"size:255" json:"avatar"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// TableName 显式指定表名，避免 GORM 默认复数把 Hero 变成 heros。
func (Hero) TableName() string { return "heroes" }
