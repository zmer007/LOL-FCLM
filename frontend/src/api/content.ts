import { http } from './http'
import type { BuildItems, BuildRunes } from './build'

export type ContentKind = 'video' | 'article'
export type ContentStatus = 'pending' | 'approved' | 'rejected'

export interface Content {
  id: number
  authorId: number
  authorName: string
  kind: ContentKind
  status: ContentStatus
  rejectReason: string
  heroId: number
  heroName: string
  heroTitle: string
  heroAvatar: string
  opponentId: number
  opponentName: string
  opponentTitle: string
  opponentAvatar: string
  rank: string
  perspective: 'self' | 'opponent' | ''
  canEdit: boolean
  title: string
  body: string
  videoUrl: string
  durationSec: number
  images: string[]
  items: BuildItems
  runes: BuildRunes
  buildNote: string
  gameVersion: string
  views: number
  createdAt: string
  updatedAt: string
  reviewing?: boolean
}

export interface ContentListResult {
  list: Content[]
  total: number
  page: number
  size: number
}

// 对线页从已通过内容提取的装备符文方案。
export interface HeroBuild {
  contentId: number
  kind: ContentKind
  title: string
  vs: string
  items: BuildItems
  runes: BuildRunes
  buildNote: string
  gameVersion: string
}

export interface ContentMeta {
  heroId: number
  opponentHeroId: number
  rank: string
  kind: ContentKind
  title: string
  body: string
  durationSec: number
  items: BuildItems
  runes: BuildRunes
  buildNote: string
  keptImages?: string[]
}

// 投稿/编辑提交载荷：meta JSON + 可选视频 + 多图。
export interface SubmitPayload {
  meta: ContentMeta
  video?: File | null
  images?: File[]
}

// 后端无配图时 images 序列化为 null，统一兜底为空数组，避免模板取下标崩溃
function normalize(c: Content): Content {
  return { ...c, images: c.images ?? [] }
}

function toFormData(payload: SubmitPayload): FormData {
  const fd = new FormData()
  fd.append('meta', JSON.stringify(payload.meta))
  if (payload.video) fd.append('video', payload.video)
  payload.images?.forEach((f) => fd.append('images', f))
  return fd
}

export async function listContents(params: {
  heroId: number
  opponentId?: number
  kind?: '' | ContentKind
  page?: number
}) {
  const res = await http<ContentListResult>({ method: 'GET', url: '/contents', params })
  return { ...res, list: res.list.map(normalize) }
}

export async function getContent(id: number) {
  return normalize(await http<Content>({ method: 'GET', url: `/contents/${id}` }))
}

export async function listMyContents(status = '') {
  const res = await http<{ list: Content[] }>({
    method: 'GET',
    url: '/my/contents',
    params: { status },
  })
  return { list: res.list.map(normalize) }
}

export function getHeroBuilds(heroId: number) {
  return http<{ list: HeroBuild[] }>({ method: 'GET', url: `/heroes/${heroId}/builds` })
}

export function publishContent(payload: SubmitPayload) {
  return http<Content>({
    method: 'POST',
    url: '/contents',
    data: toFormData(payload),
    timeout: 600000, // 视频最大 200MB，放宽上传超时
  })
}

export function updateContent(id: number, payload: SubmitPayload) {
  return http<Content>({
    method: 'PUT',
    url: `/contents/${id}`,
    data: toFormData(payload),
    timeout: 600000,
  })
}

export function deleteContent(id: number) {
  return http<{ id: number }>({ method: 'DELETE', url: `/contents/${id}` })
}

// ---------- 管理员审批 ----------

export async function listReviewContents(status = 'pending', page = 1) {
  const res = await http<ContentListResult>({
    method: 'GET',
    url: '/admin/contents',
    params: { status, page },
  })
  return { ...res, list: res.list.map(normalize) }
}

export function approveContent(id: number) {
  return http<{ id: number }>({ method: 'POST', url: `/admin/contents/${id}/approve` })
}

export function rejectContent(id: number, reason: string) {
  return http<{ id: number }>({
    method: 'POST',
    url: `/admin/contents/${id}/reject`,
    data: { reason },
  })
}
