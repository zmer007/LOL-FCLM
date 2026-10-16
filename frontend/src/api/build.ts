import { http } from './http'

// ---------- 结构化符文 ----------

export interface Rune {
  id: number
  treeId: number
  name: string
  icon: string
  row: number // 0=基石，1-3=三小排
  isKeystone: boolean
  sort: number
}

export interface RuneTree {
  id: number
  key: string
  name: string
  icon: string
  sort: number
  runes: Rune[]
}

export interface Shard {
  name: string
  icon: string
}

export interface RuneData {
  trees: RuneTree[]
  shards: Shard[][] // 3 行 × 3 列
}

// ---------- 装备符文（内嵌在视频/图文投稿中） ----------

export interface BuildItems {
  start: (number | null)[] // 出门装 2
  summs: (number | null)[] // 召唤师技能 2
  mid: (number | null)[] // 中期过渡 5
  core: (number | null)[] // 核心装备 6
  final: (number | null)[] // 后期神装 7
}

export interface BuildRunes {
  primary: number // 主系 id
  primaryPicks: number[] // [基石, 排1, 排2, 排3]
  secondary: number // 副系 id
  secondaryPicks: (number | null)[] // 三排选两排，null=该排不选
  shards: number[] // 三行各选列索引 0-2，-1=未选
}

export function getRuneTrees() {
  return http<RuneData>({ method: 'GET', url: '/rune-trees' })
}

// 生成一套空装备/符文（投稿页新增时用）。
export function emptyItems(): BuildItems {
  return {
    start: [null, null],
    summs: [null, null],
    mid: [null, null, null, null, null],
    core: [null, null, null, null, null, null],
    final: [null, null, null, null, null, null, null],
  }
}

export function emptyRunes(): BuildRunes {
  return { primary: 0, primaryPicks: [0, 0, 0, 0], secondary: 0, secondaryPicks: [null, null, null], shards: [-1, -1, -1] }
}
