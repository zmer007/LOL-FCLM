package model

// 英雄联盟段位（2026 赛季顺序，由低到高）。
var Ranks = []string{
	"黑铁",
	"青铜",
	"白银",
	"黄金",
	"铂金",
	"翡翠",
	"钻石",
	"大师",
	"宗师",
	"王者",
}

// ValidRank 判断给定段位是否在合法列表中（空串视为未选，合法）。
func ValidRank(rank string) bool {
	if rank == "" {
		return true
	}
	for _, r := range Ranks {
		if r == rank {
			return true
		}
	}
	return false
}
