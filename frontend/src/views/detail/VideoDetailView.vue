<template>
  <div class="detail-page">
    <RouterLink :to="backLink" class="back">← 返回对线页</RouterLink>

    <div v-if="loading" class="state">加载中…</div>
    <div v-else-if="!content" class="state">
      内容不存在或尚未通过审核
      <RouterLink to="/" class="back-inline">返回首页</RouterLink>
    </div>

    <template v-else>
      <!-- 审核状态提示（作者/管理员查看未过审内容时） -->
      <div v-if="content.reviewing" class="status-strip pending">
        审核视图：以下为{{ content.status === 'rejected' ? '被拒绝' : '待审核' }}的新版本，通过后对外展示；当前公开页仍为旧版
      </div>
      <div v-else-if="content.status !== 'approved'" class="status-strip" :class="content.status">
        <template v-if="content.status === 'pending'">该内容正在审核中，通过后其他用户可见</template>
        <template v-else>
          审核未通过<span v-if="content.rejectReason">：{{ content.rejectReason }}</span>
        </template>
      </div>

      <AdminAuditBar
        :content-id="content.id"
        :status="content.status"
        :reject-reason="content.rejectReason"
        @done="reload"
      />

      <div class="player">
        <video
          :src="content.videoUrl"
          controls
          controlslist="nodownload"
          disablepictureinpicture
          preload="metadata"
          :poster="''"
          @contextmenu.prevent
        ></video>
      </div>

      <div class="title-row">
        <h1 class="v-title">{{ content.title }}</h1>
        <RouterLink v-if="content.canEdit" class="edit-top" :to="editLink">✎ 编辑</RouterLink>
      </div>

      <div class="v-meta">
        <span class="up">UP主：{{ content.authorName }}</span>
        <span>播放量 {{ content.views }}</span>
        <span>{{ content.createdAt }}</span>
      </div>

      <div v-if="content.body" class="v-desc">
        <p v-for="(p, i) in bodyParas" :key="i">{{ p }}</p>
      </div>

      <div v-if="hasBuild" class="v-build">
        <BuildBlock
          :items="content.items"
          :runes="content.runes"
          :build-note="content.buildNote"
          :asset-map="assetMap"
          :trees="trees"
          :shards="shards"
        />
      </div>

      <div class="v-actions">
        <button type="button" class="act" @click="onLike">▲ 点赞</button>
        <button type="button" class="act" @click="onFavorite">☆ 收藏</button>
        <button type="button" class="act" @click="onDownload">⬇ 下载</button>
        <button type="button" class="act" @click="onShare">分享</button>
      </div>
    </template>

    <transition name="toast">
      <div v-if="toastMsg" class="toast">{{ toastMsg }}</div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { getContent, type Content } from '../../api/content'
import { getAssetMap, type GameAsset } from '../../api/asset'
import { getRuneTrees, type RuneTree, type Shard } from '../../api/build'
import AdminAuditBar from '../../components/AdminAuditBar.vue'
import BuildBlock from '../../components/build/BuildBlock.vue'
import { requireAuth } from '../../utils/auth'

const route = useRoute()
const id = Number(route.params.id) || 0

const loading = ref(true)
const content = ref<Content | null>(null)
const assetMap = ref<Map<number, GameAsset>>(new Map())
const trees = ref<RuneTree[]>([])
const shards = ref<Shard[][]>([])

// 是否配置了任意装备或符文（决定是否展示装备符文区）
const hasBuild = computed(() => {
  const c = content.value
  if (!c) return false
  return (
    Object.values(c.items).some((slots) => slots.some((s) => s !== null)) ||
    c.runes.primary !== 0 ||
    c.runes.secondary !== 0
  )
})

const backLink = computed(() => ({
  path: '/matchup',
  // 优先使用来源对线页携带的英雄（被操作视角返回对方英雄对线页的场景）
  query: { champion: String(route.query.champion ?? content.value?.heroTitle ?? '') },
}))
const editLink = computed(() => ({ path: '/upload/video', query: { id } }))
const bodyParas = computed(() => content.value?.body.split(/\r?\n/).filter((s) => s.trim()) ?? [])

const toastMsg = ref('')
let toastTimer: ReturnType<typeof setTimeout> | undefined
function toast(msg: string) {
  toastMsg.value = msg
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => (toastMsg.value = ''), 2000)
}

function onLike() {
  if (!requireAuth()) return
  toast('点赞成功')
}
function onFavorite() {
  if (!requireAuth()) return
  toast('收藏成功')
}
function onDownload() {
  if (!requireAuth()) return
  const url = content.value?.videoUrl
  if (!url) return
  const a = document.createElement('a')
  a.href = url
  a.download = ''
  a.target = '_blank'
  a.rel = 'noopener'
  document.body.appendChild(a)
  a.click()
  a.remove()
  toast('开始下载视频')
}
async function onShare() {
  try {
    await navigator.clipboard.writeText(window.location.href)
    toast('链接已复制，去分享吧')
  } catch {
    toast('分享链接：' + window.location.href)
  }
}

async function reload() {
  try {
    content.value = await getContent(id)
  } catch {
    content.value = null
  }
}

onMounted(async () => {
  try {
    const [assets, runeData] = await Promise.all([getAssetMap(), getRuneTrees()])
    assetMap.value = assets
    trees.value = runeData.trees
    shards.value = runeData.shards
    await reload()
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.detail-page {
  max-width: 900px;
  margin: 0 auto;
  padding: 24px 24px 60px;
}
.back {
  display: inline-block;
  color: var(--faint);
  font-size: 13px;
  text-decoration: none;
  margin-bottom: 14px;
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

/* 审核状态条 */
.status-strip {
  border-radius: 10px;
  padding: 10px 16px;
  font-size: 13px;
  margin-bottom: 16px;
  border: 1px solid;
}
.status-strip.pending {
  color: var(--gold);
  border-color: var(--gold);
  background: rgba(200, 170, 110, 0.08);
}
.status-strip.rejected {
  color: var(--red);
  border-color: var(--red);
  background: rgba(232, 64, 87, 0.08);
}

/* 播放器 */
.player {
  aspect-ratio: 16 / 9;
  background: #000;
  border: 1px solid var(--line-soft);
  border-radius: 14px;
  overflow: hidden;
}
.player video {
  width: 100%;
  height: 100%;
  display: block;
}

/* 标题行 */
.title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-top: 18px;
}
.v-title {
  flex: 1;
  font-size: 21px;
  font-weight: 800;
  line-height: 1.4;
  color: var(--gold-bright);
  margin: 0;
}
.edit-top {
  flex-shrink: 0;
  margin-top: 2px;
  font-size: 12.5px;
  text-decoration: none;
  border: 1px solid var(--gold);
  border-radius: 999px;
  color: var(--gold);
  padding: 6px 16px;
}
.edit-top:hover {
  background: rgba(200, 170, 110, 0.1);
}

.v-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
  margin-top: 12px;
  font-size: 12.5px;
  color: var(--faint);
}
.v-meta .up {
  color: var(--mute);
}

/* 简介 */
.v-desc {
  margin-top: 18px;
  font-size: 13px;
  color: var(--mute);
  line-height: 1.9;
}
.v-desc p {
  margin: 0 0 8px;
}
.v-build {
  margin-top: 22px;
}

/* 操作栏 */
.v-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 18px;
  padding: 14px 0;
  border-top: 1px solid var(--line-soft);
  border-bottom: 1px solid var(--line-soft);
}
.act {
  font-family: inherit;
  font-size: 13px;
  cursor: pointer;
  background: var(--panel-2);
  border: 1px solid var(--line);
  border-radius: 999px;
  color: var(--mute);
  padding: 7px 20px;
}
.act:hover {
  border-color: var(--gold);
  color: var(--gold);
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
