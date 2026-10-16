import { http } from './http'

export interface Hero {
  id: number
  name: string // 英雄称号（如 九尾妖狐）
  title: string // 短名（如 阿狸）
  lane: string
  ddragonId: string
  aliases: string[] // 社区常见别名（如 狐狸）
  roles: string[] // 官方定位：战士/坦克/刺客/法师/射手/辅助
  avatar: string
}

export function listHeroes(lane = '') {
  return http<{ list: Hero[] }>({ method: 'GET', url: '/heroes', params: { lane } })
}

// 两级匹配：先精确（短名/称号/别名），再包含匹配
export function resolveHero(list: Hero[], q: string): Hero | undefined {
  const key = q.trim().toLowerCase()
  if (!key) return undefined
  const eq = (s: string) => s.toLowerCase() === key
  const hit = list.find((h) => eq(h.title) || eq(h.name) || h.aliases.some(eq))
  if (hit) return hit
  return list.find(
    (h) =>
      h.title.toLowerCase().includes(key) ||
      h.name.toLowerCase().includes(key) ||
      h.aliases.some((a) => a.toLowerCase().includes(key)),
  )
}
