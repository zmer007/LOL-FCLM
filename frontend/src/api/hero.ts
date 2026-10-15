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
