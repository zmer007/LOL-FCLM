import { http } from './http'

export interface GameAsset {
  id: number
  category: 'item' | 'skill' | 'rune'
  subCategory: string
  name: string
  ddragonId: string
  version: string
  icon: string
}

export function listAssets(category = '', subCategory = '') {
  return http<{ list: GameAsset[] }>({
    method: 'GET',
    url: '/assets',
    params: { category, subCategory },
  })
}

// 素材字典缓存：方案只存素材 id，展示/制作器统一用它解析图标与名称。
let assetCache: Map<number, GameAsset> | null = null

export async function getAssetMap(force = false): Promise<Map<number, GameAsset>> {
  if (!force && assetCache) return assetCache
  const { list } = await listAssets()
  assetCache = new Map(list.map((a) => [a.id, a]))
  return assetCache
}
