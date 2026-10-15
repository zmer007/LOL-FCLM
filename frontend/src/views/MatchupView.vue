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

      <!-- 推荐装备 & 符文 -->
      <section class="section">
        <h2 class="section-title">推荐装备 &amp; 符文</h2>
        <BuildPanel v-if="plans.length" :plans="plans" :asset-map="assetMap" :trees="trees" :shards="shards" />
        <div v-else class="muted-box">暂无官方推荐方案</div>
      </section>

      <!-- 相关内容（视频/图文，里程碑4接通数据） -->
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
                @click="typeFilter = t.value"
              >
                {{ t.label }}
              </button>
            </div>
            <input v-model="oppQuery" class="opp-search" type="text" placeholder="搜索对位英雄" />
            <div class="post-buttons">
              <button type="button" class="btn-ghost" @click="toast('视频投稿功能即将上线')">上传视频</button>
              <button type="button" class="btn-primary" @click="toast('图文投稿功能即将上线')">发布图文</button>
            </div>
          </div>
        </div>

        <!-- 九宫格：本期暂无数据 -->
        <div class="empty">暂无相关内容</div>
      </section>
    </template>

    <!-- 轻提示 -->
    <transition name="toast">
      <div v-if="toastMsg" class="toast">{{ toastMsg }}</div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { listHeroes, type Hero } from '../api/hero'
import type { GameAsset } from '../api/asset'
import { getAssetMap } from '../api/asset'
import { getHeroPlans, getRuneTrees, type BuildPlan, type RuneTree, type Shard } from '../api/build'
import BuildPanel from '../components/build/BuildPanel.vue'

const CONTENT_TYPES = [
  { label: '全部', value: '' },
  { label: '视频', value: 'video' },
  { label: '图文', value: 'article' },
] as const

const route = useRoute()

const loading = ref(true)
const hero = ref<Hero | null>(null)
const assetMap = ref<Map<number, GameAsset>>(new Map())
const trees = ref<RuneTree[]>([])
const shards = ref<Shard[][]>([])
const plans = ref<BuildPlan[]>([])

const typeFilter = ref<(typeof CONTENT_TYPES)[number]['value']>('')
const oppQuery = ref('')

const queryName = String(route.query.champion ?? '')

// 精确（短名/称号/别名）→ 包含 的两级匹配
function resolveHero(list: Hero[], q: string): Hero | undefined {
  const key = q.trim().toLowerCase()
  if (!key) return undefined
  const eq = (s: string) => s.toLowerCase() === key
  const hit = list.find(
    (h) => eq(h.title) || eq(h.name) || h.aliases.some(eq),
  )
  if (hit) return hit
  return list.find(
    (h) =>
      h.title.toLowerCase().includes(key) ||
      h.name.toLowerCase().includes(key) ||
      h.aliases.some((a) => a.toLowerCase().includes(key)),
  )
}

onMounted(async () => {
  try {
    const [{ list: heroes }, assets, runeData] = await Promise.all([
      listHeroes(),
      getAssetMap(),
      getRuneTrees(),
    ])
    assetMap.value = assets
    trees.value = runeData.trees
    shards.value = runeData.shards

    const matched = resolveHero(heroes, queryName)
    if (matched) {
      hero.value = matched
      plans.value = (await getHeroPlans(matched.id)).list
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
