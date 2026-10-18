import { http } from './http'
import type { Content } from './content'

// 切换收藏状态，返回切换后是否已收藏。
export function toggleFavorite(contentId: number) {
  return http<{ favorited: boolean }>({
    method: 'POST',
    url: `/favorites/${contentId}`,
  })
}

// 获取当前用户的收藏列表。
export async function listMyFavorites() {
  const res = await http<{ list: Content[] }>({ method: 'GET', url: '/favorites/mine' })
  return { list: res.list }
}
