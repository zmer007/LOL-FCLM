package ddragon

// RoleMap Data Dragon 英文定位标签 → 中文定位。
var RoleMap = map[string]string{
	"Fighter":  "战士",
	"Tank":     "坦克",
	"Mage":     "法师",
	"Assassin": "刺客",
	"Marksman": "射手",
	"Support":  "辅助",
}

// RolesFromTags 把 Data Dragon tags 转为去重后的中文定位列表。
func RolesFromTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	seen := map[string]bool{}
	for _, t := range tags {
		r, ok := RoleMap[t]
		if !ok || seen[r] {
			continue
		}
		seen[r] = true
		out = append(out, r)
	}
	return out
}

// runeTreeOrder 符文系的游戏内标准展示顺序：精密/主宰/巫术/坚决/启迪。
var runeTreeOrder = map[int64]int{
	8000: 0, // 精密
	8100: 1, // 主宰
	8200: 2, // 巫术
	8400: 3, // 坚决
	8300: 4, // 启迪
}

// TreeSort 返回符文系排序权重（未知系排末尾）。
func TreeSort(id int64) int {
	if n, ok := runeTreeOrder[id]; ok {
		return n
	}
	return 99
}
