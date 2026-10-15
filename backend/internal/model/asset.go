package model

import "time"

// 素材分类
const (
	AssetCategoryItem  = "item"  // 装备
	AssetCategorySkill = "skill" // 召唤师技能
	AssetCategoryRune  = "rune"  // 符文

	// 装备子分类
	ItemSubOther = "其他"
)

// GameAsset 游戏素材图标：装备 / 召唤师技能 / 符文，数据来自 Data Dragon。
//
//	唯一键：category + name
//	DDragonID：装备为物品 id、召唤师技能为 SummonerXxx、符文为数字 id 字符串
//	SubCategory 含义随分类变化：
//	  item  → 攻击/法术/防御/移动/消耗/其他（按官方 tags 启发式归类）
//	  skill → 召唤师技能
//	  rune  → 精密/主宰/巫术/坚决/启迪（符文系）
type GameAsset struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Category    string    `gorm:"size:16;not null;index:idx_asset_name,unique" json:"category"`
	SubCategory string    `gorm:"size:32" json:"subCategory"`
	Name        string    `gorm:"size:48;not null;index:idx_asset_name,unique" json:"name"`
	DDragonID   string    `gorm:"column:ddragon_id;size:32" json:"ddragonId"`
	Version     string    `gorm:"size:16" json:"version"`
	Icon        string    `gorm:"size:255" json:"icon"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
