<template>
  <div class="matchup-page">
    <RouterLink to="/" class="back">← 返回首页</RouterLink>

    <!-- 加载中 -->
    <div v-if="loading" class="state">加载中…</div>

    <!-- 未找到英雄 -->
    <div v-else-if="!hero" class="state">
      未找到英雄「{{ queryName }}」
      <RouterLink to="/" class="back-inline">返回首页重新搜索</RouterLink>
    </div>

    <template v-else>
      <!-- 英雄横幅 -->
      <div class="hero-banner">
        <div class="hero-portrait">
          <img :src="hero.avatar" :alt="hero.title" />
        </div>
        <div class="hero-info">
          <h1>{{ hero.title }}</h1>
          <p class="hero-name">{{ hero.name }}</p>
          <div class="role-tags">
            <span v-for="r in hero.roles" :key="r" class="tag">{{ r }}</span>
          </div>
        </div>
        <div class="banner-actions">
          <button type="button" class="favorite" @click="toast('收藏功能即将上线')">☆ 收藏</button>
        </div>
      </div>

      <!-- 推荐装备 & 符文（提取自已通过内容） -->
      <section class="section">
        <h2 class="section-title">推荐装备 &amp; 符文</h2>
        <BuildPanel
          v-if="builds.length"
          :builds="builds"
          :asset-map="assetMap"
          :trees="trees"
          :shards="shards"
        />
        <div v-else class="muted-box">暂无推荐方案，欢迎投稿时补充装备与符文</div>
      </section>

      <!-- 相关内容 -->
      <section class="section">
        <div class="section-head">
          <h2 class="section-title">相关内容</h2>
          <div class="head-right">
            <div class="filter-tabs">
              <button
                v-for="t in CONTENT_TYPES"
                :key="t.value"
                type="button"
                :class="{ active: typeFilter === t.value }"
                @click="switchType(t.value)"
              >
                {{ t.label }}
              </button>
            </div>
            <input
              v-model="oppQuery"
              class="opp-search"
              type="text"
              placeholder="搜索对位英雄"
            />
            <div class="post-buttons">
              <button type="button" class="btn-ghost" @click="goUpload('video')">上传视频</button>
              <button type="button" class="btn-primary" @click="goUpload('article')">发布图文</button>
            </div>
          </div>
        </div>

        <!-- 九宫格 -->
        <div v-if="contents.length" class="video-grid">
          <div
            v-for="c in contents"
            :key="c.id"
            class="video-card"
            @click="openContent(c)"
          >
            <div class="video-thumb" :class="[c.kind, badgeClass(c)]">
              <span class="video-badge" :class="badgeClass(c)">
                {{ c.perspective === 'self' ? '操作你' : '被操作' }}
              </span>
              <span v-if="c.rank" class="rank-badge">{{ c.rank }}</span>
              <div class="cover-duo">
                <img :src="c.heroAvatar" alt="" />
                <img :src="c.opponentAvatar" alt="" />
              </div>
              <span v-if="c.kind === 'video'" class="play"></span>
              <svg v-else class="doc-ic" width="34" height="34" viewBox="0 0 24 24" fill="none"
                stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="3" width="18" height="18" rx="2" />
                <circle cx="8.5" cy="8.5" r="1.5" />
                <path d="M21 15l-5-5L5 21" />
              </svg>
              <span class="dur">
                {{ c.kind === 'video' ? formatDur(c.durationSec) : `${c.images.length} 图` }}
              </span>
            </div>
            <div class="video-title">{{ c.title }}</div>
            <div class="video-meta">{{ c.authorName }} · {{ c.views }} {{ c.kind === 'video' ? '播放' : '阅读' }}</div>
          </div>
        </div>
        <div v-else-if="!gridLoading" class="empty">暂无相关内容</div>
        <div v-else class="empty">加载中…</div>

        <!-- 分页 -->
        <div v-if="totalPages > 1" class="pager">
          <button type="button" :disabled="page === 1" @click="changePage(page - 1)">上一页</button>
          <button
            v-for="p in totalPages"
            :key="p"
            type="button"
            :class="{ on: p === page }"
            @click="changePage(p)"
          >
            {{ p }}
          </button>
          <button type="button" :disabled="page === totalPages" @click="changePage(page + 1)">下一页</button>
        </div>
      </section>
    </template>

    <!-- 轻提示 -->
    <transition name="toast">
      <div v-if="toastMsg" class="toast">{{ toastMsg }}</div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { listHeroes, resolveHero, type Hero } from '../api/hero'
import type { GameAsset } from '../api/asset'
import { getAssetMap } from '../api/asset'
import { getRuneTrees, type RuneTree, type Shard } from '../api/build'
import {
  getHeroBuilds,
  listContents,
  type Content,
  type ContentKind,
  type HeroBuild,
} from '../api/content'
import BuildPanel from '../components/build/BuildPanel.vue'

const PAGE_SIZE = 9

const CONTENT_TYPES = [
  { label: '全部', value: '' },
  { label: '视频', value: 'video' },
  { label: '图文', value: 'article' },
] as const
type TypeFilter = '' | ContentKind

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const gridLoading = ref(false)
const hero = ref<Hero | null>(null)
const assetMap = ref<Map<number, GameAsset>>(new Map())
const trees = ref<RuneTree[]>([])
const shards = ref<Shard[][]>([])
const builds = ref<HeroBuild[]>([])

const typeFilter = ref<TypeFilter>('')
const oppQuery = ref('')
const oppId = ref<number | undefined>(undefined)
const contents = ref<Content[]>([])
const page = ref(1)
const total = ref(0)
const totalPages = ref(0)

const queryName = String(route.query.champion ?? '')

let oppTimer: ReturnType<typeof setTimeout> | undefined
let heroCache: Hero[] = []

function badgeClass(c: Content) {
  return c.perspective === 'self' ? 'pro' : 'con'
}

function formatDur(sec: number) {
  if (!sec) return ''
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return `${m}:${String(s).padStart(2, '0')}`
}

async function loadContents() {
  if (!hero.value) return
  gridLoading.value = true
  try {
    const res = await listContents({
      heroId: hero.value.id,
      opponentId: oppId.value,
      kind: typeFilter.value,
      page: page.value,
    })
    contents.value = res.list
    total.value = res.total
    totalPages.value = Math.max(1, Math.ceil(res.total / PAGE_SIZE))
  } finally {
    gridLoading.value = false
  }
}

function switchType(t: TypeFilter) {
  typeFilter.value = t
  page.value = 1
  loadContents()
}
function changePage(p: number) {
  page.value = p
  loadContents()
  window.scrollTo({ top: 0 })
}

// 对位英雄搜索：关键词解析为英雄 id（解析不到则不过滤）
watch(oppQuery, (q) => {
  clearTimeout(oppTimer)
  oppTimer = setTimeout(() => {
    const key = q.trim()
    const next = key ? resolveHero(heroCache, key)?.id : undefined
    if (next !== oppId.value) {
      oppId.value = next
      page.value = 1
      loadContents()
    }
  }, 300)
})

function goUpload(kind: ContentKind) {
  if (!hero.value) return
  router.push({ path: `/upload/${kind}`, query: { champion: hero.value.title } })
}
function openContent(c: Content) {
  // 带上当前对线英雄：从「被操作」内容详情返回时仍回到本英雄对线页
  router.push({ path: `/${c.kind}/${c.id}`, query: { champion: hero.value?.title ?? '' } })
}

onMounted(async () => {
  try {
    const [heroRes, assets, runeData] = await Promise.all([
      listHeroes(),
      getAssetMap(),
      getRuneTrees(),
    ])
    assetMap.value = assets
    trees.value = runeData.trees
    shards.value = runeData.shards
    heroCache = heroRes.list

    const matched = resolveHero(heroRes.list, queryName)
    if (matched) {
      hero.value = matched
      const [buildRes] = await Promise.all([
        getHeroBuilds(matched.id),
        loadContents(),
      ])
      builds.value = buildRes.list
    }
  } finally {
    loading.value = false
  }
})

const toastMsg = ref('')
let toastTimer: ReturnType<typeof setTimeout> | undefined
function toast(msg: string) {
  toastMsg.value = msg
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => (toastMsg.value = ''), 2000)
}
</script>

<style scoped>
.matchup-page {
  max-width: 1080px;
  margin: 0 auto;
  padding: 30px 20px 60px;
}
.back {
  display: inline-block;
  color: var(--mute);
  font-size: 13.5px;
  text-decoration: none;
  margin-bottom: 18px;
}
.back:hover {
  color: var(--gold);
}
.state {
  text-align: center;
  color: var(--mute);
  font-size: 14px;
  padding: 80px 0;
}
.back-inline {
  display: block;
  margin-top: 14px;
  color: var(--gold);
  text-decoration: none;
}

/* 英雄横幅 */
.hero-banner {
  display: flex;
  align-items: center;
  gap: 22px;
  background: linear-gradient(135deg, rgba(30, 35, 40, 0.7), rgba(9, 20, 35, 0.5));
  border: 1px solid var(--line);
  border-radius: 14px;
  padding: 22px 26px;
  margin-bottom: 30px;
}
.hero-portrait {
  width: 96px;
  height: 96px;
  border-radius: 14px;
  overflow: hidden;
  border: 2px solid var(--gold);
  flex-shrink: 0;
}
.hero-portrait img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.hero-info {
  flex: 1;
}
.hero-info h1 {
  margin: 0 0 4px;
  font-size: 26px;
  color: var(--gold-bright);
}
.hero-name {
  margin: 0 0 10px;
  font-size: 13px;
  color: var(--faint);
}
.role-tags {
  display: flex;
  gap: 8px;
}
.tag {
  font-size: 12px;
  color: var(--gold);
  border: 1px solid var(--gold);
  border-radius: 999px;
  padding: 2px 12px;
}
.banner-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: flex-end;
}
.favorite {
  font-family: inherit;
  font-size: 13px;
  color: var(--mute);
  background: transparent;
  border: none;
  cursor: pointer;
}
.favorite:hover {
  color: var(--gold);
}

/* 区块 */
.section {
  margin-bottom: 38px;
}
.section-title {
  font-size: 18px;
  color: var(--gold-bright);
  margin: 0 0 18px;
}
.muted-box {
  background: var(--panel);
  border: 1px dashed var(--line);
  border-radius: 12px;
  padding: 40px;
  text-align: center;
  color: var(--faint);
  font-size: 13.5px;
}
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  flex-wrap: wrap;
  margin-bottom: 18px;
}
.section-head .section-title {
  margin: 0;
}
.head-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.filter-tabs {
  display: flex;
  border: 1px solid var(--line);
  border-radius: 8px;
  overflow: hidden;
}
.filter-tabs button {
  font-family: inherit;
  font-size: 12.5px;
  background: var(--panel);
  color: var(--mute);
  border: none;
  padding: 7px 16px;
  cursor: pointer;
}
.filter-tabs button.active {
  background: var(--gold);
  color: #0a1428;
  font-weight: 700;
}
.opp-search {
  font-family: inherit;
  font-size: 13px;
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 8px;
  color: var(--gold-bright);
  padding: 7px 12px;
  width: 150px;
}
.opp-search:focus {
  outline: none;
  border-color: var(--gold);
}
.post-buttons {
  display: flex;
  gap: 10px;
}
.btn-ghost,
.btn-primary {
  font-family: inherit;
  font-size: 13px;
  border-radius: 8px;
  padding: 7px 16px;
  cursor: pointer;
}
.btn-ghost {
  background: transparent;
  color: var(--gold-bright);
  border: 1px solid var(--gold);
}
.btn-primary {
  background: var(--gold);
  color: #0a1428;
  border: 1px solid var(--gold);
  font-weight: 700;
}

/* 九宫格 */
.video-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 18px;
}
.video-card {
  cursor: pointer;
  min-width: 0;
}
.video-thumb {
  position: relative;
  aspect-ratio: 16 / 9;
  background: #0a1520;
  border: 1px solid var(--line);
  border-radius: 10px;
  overflow: hidden;
  display: grid;
  place-items: center;
}
.video-card:hover .video-thumb {
  border-color: var(--gold);
}
.video-thumb.pro {
  border-color: rgba(10, 200, 185, 0.55);
}
.video-thumb.con {
  border-color: rgba(232, 64, 87, 0.55);
}
.video-badge {
  position: absolute;
  left: 0;
  top: 0;
  z-index: 3;
  font-size: 11px;
  font-weight: 700;
  line-height: 1;
  padding: 5px 10px;
  border-radius: 10px 0 8px 0;
  color: #04141a;
}
.video-badge.pro {
  background: var(--cyan);
}
.video-badge.con {
  background: var(--red);
  color: var(--gold-bright);
}
.rank-badge {
  position: absolute;
  right: 0;
  top: 0;
  z-index: 3;
  font-size: 11px;
  font-weight: 700;
  line-height: 1;
  padding: 5px 10px;
  border-radius: 0 10px 0 8px;
  background: rgba(10, 20, 30, 0.82);
  color: var(--gold-bright);
  border: 1px solid var(--gold);
  border-top: none;
  border-right: none;
}
.cover-duo {
  position: absolute;
  inset: 0;
  display: flex;
  opacity: 0.55;
}
.cover-duo img {
  width: 50%;
  height: 100%;
  object-fit: cover;
}
.play {
  position: relative;
  z-index: 2;
  width: 0;
  height: 0;
  border-left: 16px solid var(--faint);
  border-top: 10px solid transparent;
  border-bottom: 10px solid transparent;
}
.video-card:hover .play {
  border-left-color: var(--gold);
}
.doc-ic {
  position: relative;
  z-index: 2;
  color: var(--faint);
}
.video-card:hover .doc-ic {
  color: var(--gold);
}
.dur {
  position: absolute;
  right: 8px;
  bottom: 8px;
  z-index: 3;
  font-size: 11px;
  color: var(--mute);
  background: rgba(7, 14, 22, 0.75);
  border-radius: 4px;
  padding: 1px 6px;
}
.video-title {
  margin-top: 9px;
  font-size: 13px;
  color: var(--gold-bright);
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.video-meta {
  margin-top: 4px;
  font-size: 11.5px;
  color: var(--faint);
}

/* 分页 */
.pager {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-top: 26px;
}
.pager button {
  font-family: inherit;
  font-size: 12.5px;
  cursor: pointer;
  background: none;
  border: 1px solid var(--line);
  border-radius: 8px;
  color: var(--faint);
  min-width: 30px;
  height: 28px;
  padding: 0 8px;
}
.pager button:hover {
  color: var(--mute);
}
.pager button.on {
  border-color: var(--gold);
  color: var(--gold);
}
.pager button:disabled {
  opacity: 0.35;
  cursor: default;
}

.empty {
  background: var(--panel);
  border: 1px dashed var(--line);
  border-radius: 12px;
  padding: 60px;
  text-align: center;
  color: var(--faint);
  font-size: 14px;
}

/* toast */
.toast {
  position: fixed;
  left: 50%;
  bottom: 48px;
  transform: translateX(-50%);
  background: rgba(10, 20, 35, 0.95);
  border: 1px solid var(--gold);
  color: var(--gold-bright);
  font-size: 13px;
  padding: 10px 22px;
  border-radius: 999px;
  z-index: 100;
}
.toast-enter-active,
.toast-leave-active {
  transition: opacity 0.2s;
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
}
</style>
