import { http } from './http'

// 获取当前离线部署的 Data Dragon 数据版本
export function getDDragonVersion() {
  return http<{ version: string }>({ method: 'GET', url: '/ddragon-version' })
}

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

// ---------- 内容批量导入 / 导出 ----------

export interface ExportFileInfo {
  name: string
  size: number
}

export interface ExportResult {
  dir: string
  files: ExportFileInfo[]
}

// exportContents 按 ID 导出内容（含媒体文件，分卷 zip），返回文件列表。
export function exportContents(ids: number[] = []) {
  return http<ExportResult>({
    method: 'POST',
    url: '/admin/contents/export',
    data: { ids },
  })
}

// downloadExportFile 下载导出的单个文件。
export function downloadExportFile(dir: string, file: string) {
  return http<Blob>({
    method: 'GET',
    url: `/admin/exports/${encodeURIComponent(dir)}/${encodeURIComponent(file)}`,
    responseType: 'blob',
  })
}

// importContents 上传 manifest.json + media-*.zip（多文件）批量恢复内容与媒体。
export function importContents(files: File[], onProgress?: (pct: number) => void) {
  const form = new FormData()
  for (const f of files) {
    form.append('files', f)
  }
  return http<{ imported: number }>({
    method: 'POST',
    url: '/admin/contents/import',
    data: form,
    timeout: 0,
    onUploadProgress: (e) => {
      if (onProgress && e.total) {
        onProgress(Math.round((e.loaded / e.total) * 100))
      }
    },
  })
}
