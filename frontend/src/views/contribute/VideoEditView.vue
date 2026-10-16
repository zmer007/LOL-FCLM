<template>
  <div class="contribute-page">
    <RouterLink :to="backLink" class="back">← 返回对线页</RouterLink>

    <div v-if="loading" class="state">加载中…</div>

    <div v-else-if="done" class="done-box">
      <div class="done-icon">✓</div>
      <h2>{{ editId ? '修改已提交，等待重新审核' : '视频已提交，等待审核' }}</h2>
      <p>审核通过后将出现在{{ heroTitle }}的对线页「相关内容」中</p>
      <div class="done-ops">
        <button type="button" class="btn-primary" @click="resetForm">{{ editId ? '关闭' : '再传一个' }}</button>
        <RouterLink :to="backLink" class="btn-ghost-link">返回对线页</RouterLink>
      </div>
    </div>

    <template v-else>
      <div class="page-head">
        <h1>{{ editId ? '编辑视频' : '上传视频' }}</h1>
        <p class="sub">投稿至 <em>{{ heroTitle }}</em> 的对线页 · 仅支持 MP4（H.264），大小不超过 200MB</p>
      </div>

      <!-- 视频文件 -->
      <section class="card">
        <div class="field-label">视频文件 <i>*</i></div>
        <div
          class="drop-zone"
          :class="{ dragging }"
          @click="fileInput?.click()"
          @dragover.prevent="dragging = true"
          @dragleave.prevent="dragging = false"
          @drop.prevent="onDrop"
        >
          <input
            ref="fileInput"
            type="file"
            accept="video/mp4"
            hidden
            @change="onFileChange"
          />
          <template v-if="!videoFile">
            <span v-if="existingUrl" class="existing">
              已有视频，不更换则保留：<a :href="existingUrl" target="_blank" @click.stop>查看当前视频</a>
            </span>
            <span v-else class="dz-main">点击选择或拖拽视频到此处</span>
            <span class="dz-sub">MP4 / H.264 · ≤200MB</span>
          </template>
          <template v-else>
            <span class="dz-main">{{ videoFile.name }}</span>
            <span class="dz-sub">{{ (videoFile.size / 1024 / 1024).toFixed(1) }} MB · 时长 {{ durationSec }}s</span>
            <button type="button" class="dz-remove" @click.stop="clearVideo">移除</button>
          </template>
        </div>
        <div v-if="fileError" class="field-error">{{ fileError }}</div>
      </section>

      <!-- 标题 -->
      <section class="card">
        <div class="field-label">标题 <i>*</i></div>
        <input v-model="form.title" class="text-input" type="text" maxlength="120" placeholder="一句话讲清这个视频的内容" />
      </section>

      <!-- 对位英雄 & 段位 -->
      <section class="card">
        <div class="opp-rank-row">
          <div class="opp-rank-col">
            <div class="field-label">对位英雄 <i>*</i></div>
            <OpponentPicker
              :heroes="heroes"
              v-model="form.opponentHeroId"
              :exclude-id="heroId"
              :touched="submitted"
            />
          </div>
          <div class="opp-rank-col rank-col">
            <div class="field-label">段位 <i>*</i></div>
            <RankPicker v-model="form.rank" :touched="submitted" />
          </div>
        </div>
      </section>

      <!-- 推荐装备 & 符文 -->
      <section class="card">
        <BuildEditor
          v-model:note="form.buildNote"
          :items="form.items"
          :runes="form.runes"
          :asset-map="assetMap"
          :trees="trees"
          :shards="shards"
          :touched="submitted"
        />
      </section>

      <!-- 简介 -->
      <section class="card">
        <div class="field-label">视频简介</div>
        <textarea v-model="form.body" class="textarea" rows="4" maxlength="10000" placeholder="补充这个视频的要点（可选）"></textarea>
      </section>

      <div v-if="submitError" class="submit-error">{{ submitError }}</div>
      <div class="submit-bar">
        <button type="button" class="btn-primary" :disabled="saving" @click="submit">
          {{ saving ? '提交中…' : editId ? '保存修改' : '提交上传' }}
        </button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { listHeroes, resolveHero, type Hero } from '../../api/hero'
import { getAssetMap, type GameAsset } from '../../api/asset'
import {
  emptyItems,
  emptyRunes,
  getRuneTrees,
  type BuildItems,
  type BuildRunes,
  type RuneTree,
  type Shard,
} from '../../api/build'
import { getContent, publishContent, updateContent, type ContentMeta } from '../../api/content'
import OpponentPicker from '../../components/OpponentPicker.vue'
import RankPicker from '../../components/RankPicker.vue'
import BuildEditor from '../../components/build/BuildEditor.vue'

const MAX_MB = 200
const route = useRoute()
const router = useRouter()

const loading = ref(true)
const saving = ref(false)
const submitted = ref(false)
const done = ref(false)
const dragging = ref(false)
const fileError = ref('')
const submitError = ref('')
const fileInput = ref<HTMLInputElement | null>(null)

const heroes = ref<Hero[]>([])
const assetMap = ref<Map<number, GameAsset>>(new Map())
const trees = ref<RuneTree[]>([])
const shards = ref<Shard[][]>([])

const heroId = ref(0)
const heroTitle = ref('')
const editId = computed(() => Number(route.query.id) || 0)
const existingUrl = ref('')

const videoFile = ref<File | null>(null)
const durationSec = ref(0)

const form = ref<{
  title: string
  body: string
  opponentHeroId: number
  rank: string
  items: BuildItems
  runes: BuildRunes
  buildNote: string
}>({
  title: '',
  body: '',
  opponentHeroId: 0,
  rank: '',
  items: emptyItems(),
  runes: emptyRunes(),
  buildNote: '',
})

// 是否配置了任意装备或符文（思路随装备符文必填）
function hasBuild() {
  return (
    Object.values(form.value.items).some((slots) => slots.some((s: number | null) => s !== null)) ||
    form.value.runes.primary !== 0 ||
    form.value.runes.secondary !== 0
  )
}

const backLink = computed(() => ({ path: '/matchup', query: { champion: heroTitle.value } }))

function blankForm() {
  form.value = {
    title: '',
    body: '',
    opponentHeroId: 0,
    rank: '',
    items: emptyItems(),
    runes: emptyRunes(),
    buildNote: '',
  }
  videoFile.value = null
  durationSec.value = 0
  existingUrl.value = ''
  fileError.value = ''
}

function readDuration(file: File): Promise<number> {
  return new Promise((resolve) => {
    const video = document.createElement('video')
    video.preload = 'metadata'
    const url = URL.createObjectURL(file)
    video.onloadedmetadata = () => {
      const d = Math.round(video.duration)
      URL.revokeObjectURL(url)
      resolve(Number.isFinite(d) ? d : 0)
    }
    video.onerror = () => {
      URL.revokeObjectURL(url)
      resolve(0)
    }
    video.src = url
  })
}

async function pickFile(file: File) {
  fileError.value = ''
  const isMp4 = file.type === 'video/mp4' || /\.mp4$/i.test(file.name)
  if (!isMp4) {
    fileError.value = '仅支持 MP4 格式视频'
    return
  }
  if (file.size > MAX_MB * 1024 * 1024) {
    fileError.value = `视频大小不能超过 ${MAX_MB}MB`
    return
  }
  videoFile.value = file
  durationSec.value = await readDuration(file)
}

function onFileChange(e: Event) {
  const f = (e.target as HTMLInputElement).files?.[0]
  if (f) pickFile(f)
}
function onDrop(e: DragEvent) {
  dragging.value = false
  const f = e.dataTransfer?.files?.[0]
  if (f) pickFile(f)
}
function clearVideo() {
  videoFile.value = null
  durationSec.value = 0
  if (fileInput.value) fileInput.value.value = ''
}

function validate(): string {
  if (!form.value.title.trim()) return '请填写标题'
  if (!form.value.opponentHeroId) return '请从候选中选择对位英雄'
  if (!form.value.rank) return '请选择段位'
  if (hasBuild() && !form.value.buildNote.trim()) return '请填写这套装备符文的思路'
  if (!editId.value && !videoFile.value) return '请上传视频文件'
  return ''
}

async function submit() {
  submitted.value = true
  submitError.value = ''
  const err = validate()
  if (err) {
    submitError.value = err
    return
  }
  const meta: ContentMeta = {
    heroId: heroId.value,
    opponentHeroId: form.value.opponentHeroId,
    rank: form.value.rank,
    kind: 'video',
    title: form.value.title.trim(),
    body: form.value.body,
    durationSec: durationSec.value,
    items: form.value.items,
    runes: form.value.runes,
    buildNote: form.value.buildNote.trim(),
  }
  saving.value = true
  try {
    if (editId.value) {
      await updateContent(editId.value, { meta, video: videoFile.value })
    } else {
      await publishContent({ meta, video: videoFile.value })
    }
    done.value = true
  } catch (e) {
    submitError.value = e instanceof Error ? e.message : '提交失败'
  } finally {
    saving.value = false
  }
}

function resetForm() {
  if (editId.value) {
    router.push(backLink.value)
    return
  }
  done.value = false
  blankForm()
}

onMounted(async () => {
  try {
    const [heroList, assets, runeData] = await Promise.all([
      listHeroes(),
      getAssetMap(),
      getRuneTrees(),
    ])
    heroes.value = heroList.list
    assetMap.value = assets
    trees.value = runeData.trees
    shards.value = runeData.shards

    if (editId.value) {
      const c = await getContent(editId.value)
      heroId.value = c.heroId
      heroTitle.value = c.heroTitle
      form.value = {
        title: c.title,
        body: c.body,
        opponentHeroId: c.opponentId,
        rank: c.rank ?? '',
        items: c.items,
        runes: c.runes,
        buildNote: c.buildNote ?? '',
      }
      durationSec.value = c.durationSec
      existingUrl.value = c.videoUrl
    } else {
      const hero = resolveHero(heroList.list, String(route.query.champion ?? ''))
      if (!hero) {
        router.replace('/')
        return
      }
      heroId.value = hero.id
      heroTitle.value = hero.title
    }
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.contribute-page {
  max-width: 760px;
  margin: 0 auto;
  padding: 24px 20px 80px;
}
.back {
  display: inline-block;
  color: var(--mute);
  font-size: 13px;
  text-decoration: none;
  margin-bottom: 16px;
}
.back:hover {
  color: var(--gold);
}
.state {
  text-align: center;
  color: var(--mute);
  padding: 80px 0;
}
.page-head h1 {
  font-size: 22px;
  color: var(--gold-bright);
  margin: 0 0 6px;
}
.sub {
  font-size: 13px;
  color: var(--faint);
  margin: 0 0 20px;
}
.sub em {
  font-style: normal;
  color: var(--gold);
}
.card {
  background: var(--panel);
  border: 1px solid var(--line-soft);
  border-radius: 12px;
  padding: 18px 20px;
  margin-bottom: 16px;
}
.field-label {
  font-size: 13.5px;
  font-weight: 700;
  color: var(--gold-bright);
  margin-bottom: 10px;
}
.field-label i {
  font-style: normal;
  color: var(--red);
}
.field-label .hint {
  font-weight: 400;
  font-size: 12px;
  color: var(--faint);
}
.opp-rank-row {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}
.opp-rank-col {
  flex: 1;
  min-width: 0;
}
.opp-rank-col.rank-col {
  flex: 0 0 170px;
}
.text-input,
.textarea {
  width: 100%;
  font-family: inherit;
  font-size: 13.5px;
  background: #0a1520;
  border: 1px solid var(--line);
  border-radius: 8px;
  color: var(--gold-bright);
  padding: 10px 12px;
  box-sizing: border-box;
}
.text-input:focus,
.textarea:focus {
  outline: none;
  border-color: var(--gold);
}
.textarea {
  resize: vertical;
}
.drop-zone {
  border: 1.5px dashed var(--line);
  border-radius: 10px;
  padding: 34px 20px;
  text-align: center;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 8px;
  transition: border-color 0.15s;
}
.drop-zone:hover,
.drop-zone.dragging {
  border-color: var(--gold);
}
.dz-main {
  font-size: 14px;
  color: var(--gold-bright);
}
.dz-sub,
.existing {
  font-size: 12.5px;
  color: var(--faint);
}
.existing a {
  color: var(--cyan);
}
.dz-remove {
  align-self: center;
  margin-top: 4px;
  background: none;
  border: 1px solid var(--red);
  color: var(--red);
  border-radius: 6px;
  font-size: 12px;
  padding: 3px 12px;
  cursor: pointer;
}
.field-error {
  margin-top: 8px;
  font-size: 12.5px;
  color: var(--red);
}
.submit-error {
  color: var(--red);
  font-size: 13px;
  text-align: center;
  margin-bottom: 12px;
}
.submit-bar {
  text-align: center;
  margin-top: 8px;
}
.btn-primary {
  font-family: inherit;
  font-size: 14px;
  font-weight: 700;
  background: var(--gold);
  color: #0a1428;
  border: none;
  border-radius: 9px;
  padding: 11px 44px;
  cursor: pointer;
}
.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.done-box {
  text-align: center;
  padding: 70px 20px;
}
.done-icon {
  width: 64px;
  height: 64px;
  margin: 0 auto 20px;
  border-radius: 50%;
  background: rgba(200, 170, 110, 0.15);
  border: 2px solid var(--gold);
  color: var(--gold);
  font-size: 30px;
  line-height: 60px;
}
.done-box h2 {
  font-size: 19px;
  color: var(--gold-bright);
  margin: 0 0 10px;
}
.done-box p {
  font-size: 13px;
  color: var(--faint);
  margin: 0 0 26px;
}
.done-ops {
  display: flex;
  gap: 14px;
  justify-content: center;
  align-items: center;
}
.done-ops .btn-primary {
  padding: 9px 26px;
}
.btn-ghost-link {
  font-size: 13.5px;
  color: var(--gold-bright);
  text-decoration: none;
  border: 1px solid var(--gold);
  border-radius: 9px;
  padding: 9px 22px;
}
</style>
