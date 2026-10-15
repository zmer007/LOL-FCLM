package ddragon

// 属性碎片（stat shards）固定 3 行每行 3 选 1。
// Riot Data Dragon 桶的 statmods 图标返回 403，改用 Community Dragon 镜像（同为 Riot 资产）。
// 9 个槽去重后为 6 张图。

// ShardDef 一个碎片槽：File 为 CDragon 图标文件名（不含扩展名），Name 为中文名。
type ShardDef struct {
	File string
	Name string
}

// shard 文件名常量
const (
	shardAdaptive = "statmodsadaptiveforceicon" // 适应之力
	shardAttack   = "statmodsattackspeedicon"   // 攻击速度
	shardHaste    = "statmodscdrscalingicon"    // 技能急速
	shardArmor    = "statmodsarmoricon"         // 护甲
	shardMagicRes = "statmodsmagicresicon"      // 魔法抗性
	shardHealth   = "statmodshealthscalingicon" // 生命值
)

// ShardGrid 三行碎片定义，行内顺序即游戏内列顺序（列索引 0-2）。
var ShardGrid = [3][3]ShardDef{
	{{shardAdaptive, "适应之力"}, {shardAttack, "攻击速度"}, {shardHaste, "技能急速"}},
	{{shardAdaptive, "适应之力"}, {shardArmor, "护甲"}, {shardMagicRes, "魔法抗性"}},
	{{shardHealth, "生命值"}, {shardArmor, "护甲"}, {shardMagicRes, "魔法抗性"}},
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
