package model

import "time"

// 素材分类
const (
	AssetCategoryItem  = "item"  // 装备
	AssetCategorySkill = "skill" // 召唤师技能
	AssetCategoryRune  = "rune"  // 符文
)

// GameAsset 游戏素材图标：装备 / 召唤师技能 / 符文。
// 同一分类下名称唯一；SubCategory 含义随分类变化：
//   item  → 攻击/法术/防御/移动/消耗
//   skill → 召唤师技能（可空）
//   rune  → 精密/主宰/巫术/坚决/启迪（符文系）
type GameAsset struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Category    string    `gorm:"size:16;not null;index:idx_asset_name,unique" json:"category"`
	SubCategory string    `gorm:"size:32" json:"subCategory"`
	Name        string    `gorm:"size:48;not null;index:idx_asset_name,unique" json:"name"`
	Icon        string    `gorm:"size:255" json:"icon"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
