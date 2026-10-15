import { http } from './http'

export interface Hero {
  id: number
  name: string
  title: string
  lane: string
  ddragonId: string
  ddragonKey: string
  avatar: string
  createdAt: string
  updatedAt: string
}

export type AssetCategory = 'item' | 'skill' | 'rune'

export interface GameAsset {
  id: number
  category: AssetCategory
  subCategory: string
  name: string
  ddragonId: string
  version: string
  icon: string
}

// ---------- 英雄 ----------

export function listHeroes(lane = '') {
  return http<{ list: Hero[] }>({ method: 'GET', url: '/admin/heroes', params: { lane } })
}

export function createHero(name: string, lane: string) {
  return http<{ hero: Hero }>({
    method: 'POST',
    url: '/admin/heroes',
    data: { name, lane },
  })
}

export function uploadHeroAvatar(id: number, file: File) {
  const form = new FormData()
  form.append('icon', file)
  return http<{ id: number }>({
    method: 'POST',
    url: `/admin/heroes/${id}/avatar`,
    data: form,
  })
}

// ---------- 素材 ----------

export function listAssets(category: AssetCategory, subCategory = '') {
  return http<{ list: GameAsset[] }>({
    method: 'GET',
    url: '/admin/assets',
    params: { category, subCategory },
  })
}

export function createAsset(category: AssetCategory, subCategory: string, name: string) {
  return http<{ asset: GameAsset }>({
    method: 'POST',
    url: '/admin/assets',
    data: { category, subCategory, name },
  })
}

export function uploadAssetIcon(id: number, file: File) {
  const form = new FormData()
  form.append('icon', file)
  return http<{ id: number }>({
    method: 'POST',
    url: `/admin/assets/${id}/icon`,
    data: form,
  })
}

export function deleteAsset(id: number) {
  return http<{ id: number }>({ method: 'DELETE', url: `/admin/assets/${id}` })
}

// ---------- 用户管理 ----------

export interface AdminUser {
  id: number
  username: string
  role: 'user' | 'admin'
  status: 'active' | 'disabled'
  createdAt: string
}

export function listUsers() {
  return http<{ list: AdminUser[] }>({ method: 'GET', url: '/admin/users' })
}

export function addUser(username: string, password: string, role: 'user' | 'admin') {
  return http<{ user: AdminUser }>({
    method: 'POST',
    url: '/admin/users',
    data: { username, password, role },
  })
}

export function setUserStatus(id: number, status: 'active' | 'disabled') {
  return http<{ ok: boolean }>({
    method: 'PATCH',
    url: `/admin/users/${id}/status`,
    data: { status },
  })
}

export function setUserRole(id: number, role: 'user' | 'admin') {
  return http<{ ok: boolean }>({
    method: 'PATCH',
    url: `/admin/users/${id}/role`,
    data: { role },
  })
}

export function resetUserPassword(id: number, password: string) {
  return http<{ ok: boolean }>({
    method: 'POST',
    url: `/admin/users/${id}/password`,
    data: { password },
  })
}

export function deleteUser(id: number) {
  return http<{ ok: boolean }>({ method: 'DELETE', url: `/admin/users/${id}` })
}

// ---------- Data Dragon 同步 ----------

export interface SyncState {
  running: boolean
  status: 'idle' | 'running' | 'success' | 'skipped' | 'failed'
  version: string
  stage: string
  total: number
  done: number
  stageTotal: number
  stageDone: number
  failed: number
  downloaded: number
  message: string
  startedAt?: string
  finishedAt?: string
}

export function triggerSync(force = false) {
  return http<SyncState>({ method: 'POST', url: '/admin/sync/ddragon', data: { force } })
}

export function getSyncStatus() {
  return http<SyncState>({ method: 'GET', url: '/admin/sync/status' })
}
