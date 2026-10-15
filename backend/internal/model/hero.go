package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

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
//	DDragonID 为英文 id（如 Ahri），Name 为中文称号，Title 为英雄短名。
//	Lane（分路）官方数据中没有，默认空，由管理员后续维护。
//	Aliases 为社区常见别名（逗号分隔，如 铁男,铁皮人,大锤），仅用于搜索。
//	Avatar 为本地图标 URL（/media/icons/heroes/...）。
type Hero struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string    `gorm:"size:32;uniqueIndex;not null" json:"name"`
	Title      string    `gorm:"size:32" json:"title"`
	Lane       string    `gorm:"size:16;not null;default:'';index" json:"lane"`
	DDragonID  string    `gorm:"column:ddragon_id;size:32;uniqueIndex" json:"ddragonId"`
	DDragonKey string    `gorm:"column:ddragon_key;size:16" json:"ddragonKey"`
	Aliases    string    `gorm:"size:512;not null;default:''" json:"-"`
	AliasList  []string  `gorm:"-" json:"aliases"`
	Roles      string    `gorm:"size:64;not null;default:''" json:"-"`
	RoleList   []string  `gorm:"-" json:"roles"` // 官方定位：战士/坦克/刺客/法师/射手/辅助（来自 Data Dragon tags）
	Avatar     string    `gorm:"size:255" json:"avatar"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// TableName 显式指定表名，避免 GORM 默认复数把 Hero 变成 heros。
func (Hero) TableName() string { return "heroes" }

// AfterFind 查询后把逗号分隔的字段拆成列表输出。
func (h *Hero) AfterFind(_ *gorm.DB) error {
	h.AliasList = splitList(h.Aliases)
	h.RoleList = splitList(h.Roles)
	return nil
}

func splitList(s string) []string {
	out := make([]string, 0)
	for a := range strings.SplitSeq(s, ",") {
		if a = strings.TrimSpace(a); a != "" {
			out = append(out, a)
		}
	}
	return out
}

// JoinList 把字符串列表规范化为逗号分隔字符串（别名/定位/分路通用）。
func JoinList(items []string) string {
	clean := make([]string, 0, len(items))
	for _, a := range items {
		if a = strings.TrimSpace(a); a != "" {
			clean = append(clean, a)
		}
	}
	return strings.Join(clean, ",")
}
