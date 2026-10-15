package model

import "time"

// Setting 键值型系统配置（如 Data Dragon 最近同步版本与时间）。
type Setting struct {
	Key       string    `gorm:"primaryKey;size:64" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// 设置项键名
const (
	SettingDDragonVersion  = "ddragon_version"
	SettingDDragonLastSync = "ddragon_last_sync"
)
