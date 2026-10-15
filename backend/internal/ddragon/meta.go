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
