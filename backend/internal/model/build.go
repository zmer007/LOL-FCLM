package model

// 装备五分区槽位数（与设计稿/投稿页严格一致）。
const (
	BuildStageStartSlots = 2 // 出门装
	BuildStageSummsSlots = 2 // 召唤师技能（与出门装并排）
	BuildStageMidSlots   = 5 // 中期过渡
	BuildStageCoreSlots  = 6 // 核心装备
	BuildStageFinalSlots = 7 // 后期神装
)

// BuildItems 装备方案。每个切片为对应分区的槽位，元素是 GameAsset.id，空槽用 nil。
// Summs 槽只允许 category=skill 的素材，其余分区只允许 category=item。
type BuildItems struct {
	Start []*int64 `json:"start"`
	Summs []*int64 `json:"summs"`
	Mid   []*int64 `json:"mid"`
	Core  []*int64 `json:"core"`
	Final []*int64 `json:"final"`
}

// BuildRunes 符文方案。
//
//	Primary/Secondary 为符文系 id；两者不能相同。
//	PrimaryPicks 长度 4 = [基石, 第1排, 第2排, 第3排]，存官方符文 id（0 表示未选）。
//	SecondaryPicks 长度 3 = 副系三排，每排可选一个或不选（nil=该排不选；规则上 3 排选 2 排）。
//	Shards 长度 3，每行选中 0-2 的列索引（-1 表示未选）。
type BuildRunes struct {
	Primary        int64    `json:"primary"`
	PrimaryPicks   []int64  `json:"primaryPicks"`
	Secondary      int64    `json:"secondary"`
	SecondaryPicks []*int64 `json:"secondaryPicks"`
	Shards         []int    `json:"shards"`
}

// IsEmpty 判断整套装备符文是否都未配置（对线页提取方案时过滤）。
func (b *BuildItems) IsEmpty() bool {
	for _, slots := range [][]*int64{b.Start, b.Summs, b.Mid, b.Core, b.Final} {
		for _, p := range slots {
			if p != nil {
				return false
			}
		}
	}
	return true
}

func (r *BuildRunes) IsEmpty() bool {
	return r.Primary == 0 && r.Secondary == 0
}
