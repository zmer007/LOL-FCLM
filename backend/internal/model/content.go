package model

import "time"

const (
	ContentPending  = "pending"  // 待审批
	ContentApproved = "approved" // 已通过（对外可见）
	ContentRejected = "rejected" // 已拒绝
)

const (
	KindVideo   = "video"
	KindArticle = "article"
)

// ContentSnapshot 内容最近一次通过审批的快照。
// 作者编辑后 Content 实时字段变化并回到 pending，但快照保留 → 复审期间对外仍展示旧版。
type ContentSnapshot struct {
	Title          string     `json:"title"`
	Body           string     `json:"body"`
	VideoURL       string     `json:"videoUrl"`
	DurationSec    int        `json:"durationSec"`
	Images         []string   `json:"images"`
	OpponentHeroID int64      `json:"opponentHeroId"`
	Rank           string     `json:"rank"`
	Items          BuildItems `json:"items"`
	Runes          BuildRunes `json:"runes"`
	BuildNote      string     `json:"buildNote"` // 这套装备符文的思路
	GameVersion    string     `json:"gameVersion"` // 创作时的游戏版本（Data Dragon 版本号）
	ApprovedAt     time.Time  `json:"approvedAt"`
}

// Content 用户投稿的视频/图文内容。
type Content struct {
	ID             int64            `gorm:"primaryKey;autoIncrement" json:"id"`
	AuthorID       int64            `gorm:"index;not null" json:"authorId"`
	HeroID         int64            `gorm:"index;not null" json:"heroId"`            // 主视角英雄
	OpponentHeroID int64            `gorm:"not null" json:"opponentHeroId"`          // 对位英雄
	Rank           string           `gorm:"size:16;not null;default:''" json:"rank"` // 段位（可选）
	Kind           string           `gorm:"size:16;index;not null" json:"kind"`      // video | article
	Title          string           `gorm:"size:120;not null" json:"title"`
	Body           string           `gorm:"type:text" json:"body"` // 视频简介/图文正文
	VideoURL       string           `gorm:"size:255" json:"videoUrl"`
	DurationSec    int              `json:"durationSec"`
	Images         []string         `gorm:"type:json;serializer:json" json:"images"`
	Items          BuildItems       `gorm:"type:json;serializer:json" json:"items"`
	Runes          BuildRunes       `gorm:"type:json;serializer:json" json:"runes"`
	BuildNote      string           `gorm:"size:500;not null;default:''" json:"buildNote"` // 装备符文思路，配置了装备/符文时必填
	GameVersion    string           `gorm:"size:16;not null;default:''" json:"gameVersion"` // 创作时的游戏版本（Data Dragon 版本号）
	Status         string           `gorm:"size:16;index;not null;default:'pending'" json:"status"`
	RejectReason   string           `gorm:"size:255" json:"rejectReason"`
	Approved       *ContentSnapshot `gorm:"type:json;serializer:json" json:"approved,omitempty"`
	Views          int              `json:"views"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`
}

func (Content) TableName() string { return "contents" }
