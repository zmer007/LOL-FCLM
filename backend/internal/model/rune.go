package model

import "time"

// RuneTree 符文系（精密/主宰/巫术/坚决/启迪），数据来自 Data Dragon runesReforged.json。
// ID 直接使用官方系数字 id（8000…），不自增。
type RuneTree struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Key       string    `gorm:"size:32" json:"key"`
	Name      string    `gorm:"size:16" json:"name"`
	Icon      string    `gorm:"size:255" json:"icon"`
	Sort      int       `json:"sort"`
	Runes     []Rune    `gorm:"foreignKey:TreeID;constraint:-" json:"runes"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (RuneTree) TableName() string { return "rune_trees" }

// Rune 单个符文。Row=0 为基石，1-3 为基石下三小排；IsKeystone 标记基石。
// ID 直接使用官方符文数字 id，不自增。
type Rune struct {
	ID         int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	TreeID     int64     `gorm:"index;not null" json:"treeId"`
	Name       string    `gorm:"size:48" json:"name"`
	Icon       string    `gorm:"size:255" json:"icon"`
	Row        int       `gorm:"column:slot_row" json:"row"`
	IsKeystone bool      `json:"isKeystone"`
	Sort       int       `json:"sort"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (Rune) TableName() string { return "runes" }
