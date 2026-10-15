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

// ---------- 装备符文方案 ----------

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
  shards: number[] // 三行各选列索引 0-2
}

export interface BuildPlan {
  id: number
  heroId: number
  seq: number
  authorId: number
  vs: string
  items: BuildItems
  runes: BuildRunes
  createdAt: string
  updatedAt: string
}

// 提交给后端的单套方案（不含 id/序号/作者）。
export type PlanInput = Pick<BuildPlan, 'vs' | 'items' | 'runes'>

export function getRuneTrees() {
  return http<RuneData>({ method: 'GET', url: '/rune-trees' })
}

export function getHeroPlans(heroId: number) {
  return http<{ list: BuildPlan[] }>({ method: 'GET', url: `/heroes/${heroId}/plans` })
}

// 整英雄批量替换方案（空数组=清空），仅管理员。
export function saveHeroPlans(heroId: number, plans: PlanInput[]) {
  return http<{ list: BuildPlan[] }>({
    method: 'PUT',
    url: `/admin/heroes/${heroId}/plans`,
    data: { plans },
  })
}

// 生成一套空方案（制作器新增时用）。
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
