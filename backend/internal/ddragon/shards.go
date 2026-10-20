package ddragon

// 属性碎片（stat shards）固定 3 行每行 3 选 1。
// Riot Data Dragon 桶的 statmods 图标返回 403，改用 Community Dragon 镜像（同为 Riot 资产）。
// 9 个槽去重后为 6 张图。

// ShardDef 一个碎片槽：File 为 CDragon 图标文件名（不含扩展名），Name 为中文名。
type ShardDef struct {
	File string
	Name string
}

// shard 文件名常量（来自 Community Dragon statmods 目录）
const (
	shardAdaptive      = "statmodsadaptiveforceicon" // 适应之力
	shardAttack        = "statmodsattackspeedicon"   // 攻击速度
	shardHaste         = "statmodscdrscalingicon"    // 技能急速
	shardMoveSpeed     = "statmodsmovementspeedicon" // 移动速度
	shardHealthFlat    = "statmodshealthplusicon"    // 成长生命值（心+向上箭头）
	shardTenacity      = "statmodstenacityicon"      // 韧性
	shardHealthScaling = "statmodshealthscalingicon" // 生命值（实心爱心）
)

// ShardGrid 三行碎片定义，行内顺序即游戏内列顺序（列索引 0-2）。
// 依据 2026 赛季（Patch 26.x）游戏内实际布局：
//
//			Row 1 (进攻): 适应之力 / 攻击速度 / 技能急速
//		Row 2 (灵活): 适应之力 / 移动速度 / 生命值
//	  Row 3 (防御): 生命值 / 韧性 / 成长生命值
var ShardGrid = [3][3]ShardDef{
	{{shardAdaptive, "适应之力"}, {shardAttack, "攻击速度"}, {shardHaste, "技能急速"}},
	{{shardAdaptive, "适应之力"}, {shardMoveSpeed, "移动速度"}, {shardHealthScaling, "生命值"}},
	{{shardHealthScaling, "生命值"}, {shardTenacity, "韧性"}, {shardHealthFlat, "成长生命值"}},
}

// ShardFiles 返回去重后的碎片图标文件名（同步下载用）。
func ShardFiles() []string {
	seen := map[string]bool{}
	out := make([]string, 0, 6)
	for _, row := range ShardGrid {
		for _, d := range row {
			if !seen[d.File] {
				seen[d.File] = true
				out = append(out, d.File)
			}
		}
	}
	return out
}

// ShardIconURL 返回属性碎片图标的对外 URL，直接从 assets 目录经 /ddragon 静态路由提供。
func ShardIconURL(file string) string {
	return "/ddragon/statmods/" + file
}
