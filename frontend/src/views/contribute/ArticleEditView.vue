<template>
  <div class="contribute-page">
    <RouterLink :to="backLink" class="back">← 返回对线页</RouterLink>

    <div v-if="loading" class="state">加载中…</div>

    <div v-else-if="done" class="done-box">
      <div class="done-icon">✓</div>
      <h2>{{ editId ? '修改已提交，等待重新审核' : '图文已发布，等待审核' }}</h2>
      <p>审核通过后将出现在{{ heroTitle }}的对线页「相关内容」中</p>
      <div class="done-ops">
        <button type="button" class="btn-primary" @click="resetForm">{{ editId ? '关闭' : '再写一篇' }}</button>
        <RouterLink :to="backLink" class="btn-ghost-link">返回对线页</RouterLink>
      </div>
    </div>

    <template v-else>
      <div class="page-head">
        <h1>{{ editId ? '编辑图文' : '发布图文' }}</h1>
        <p class="sub">投稿至 <em>{{ heroTitle }}</em> 的对线页</p>
      </div>

      <!-- 标题 -->
      <section class="card">
        <div class="field-label">标题 <i>*</i></div>
        <input v-model="form.title" class="text-input" type="text" maxlength="120" placeholder="一句话讲清这篇图文的内容" />
      </section>

      <!-- 原地址链接 -->
      <section class="card">
        <div class="field-label">原地址链接 <span class="hint">（可选，如转载来源地址）</span></div>
        <input v-model="form.sourceUrl" class="text-input" type="url" maxlength="500" placeholder="https://..." />
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

      <!-- 正文 -->
      <section class="card">
        <div class="field-label">正文 <i>*</i></div>
        <textarea v-model="form.body" class="textarea" rows="10" maxlength="10000" placeholder="写下对线思路、连招细节、团战打法…"></textarea>
      </section>

      <!-- 配图 -->
      <section class="card">
        <div class="field-label">配图 <span class="hint">最多 9 张，可选</span></div>
        <div class="img-grid">
          <div v-for="(img, i) in keptImages" :key="'old' + i" class="img-cell">
            <img :src="img" alt="配图" />
            <button type="button" class="img-del" @click="removeKept(i)">×</button>
          </div>
          <div v-for="(url, i) in previewUrls" :key="'new' + i" class="img-cell">
            <img :src="url" alt="配图" />
            <button type="button" class="img-del" @click="removeNew(i)">×</button>
          </div>
          <button
            v-if="totalImages < 9"
            type="button"
            class="img-add"
            @click="fileInput?.click()"
          >
            ＋<span>添加图片</span>
          </button>
        </div>
        <input
          ref="fileInput"
          type="file"
          accept="image/png,image/jpeg,image/gif,image/webp"
          multiple
          hidden
          @change="onFileChange"
        />
        <div v-if="fileError" class="field-error">{{ fileError }}</div>
      </section>

      <div v-if="submitError" class="submit-error">{{ submitError }}</div>
      <div class="submit-bar">
        <button type="button" class="btn-primary" :disabled="saving" @click="submit">
          {{ saving ? '提交中…' : editId ? '保存修改' : '发布图文' }}
        </button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
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

const MAX_IMAGES = 9
const route = useRoute()
const router = useRouter()

const loading = ref(true)
const saving = ref(false)
const submitted = ref(false)
const done = ref(false)
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

const keptImages = ref<string[]>([])
const newImages = ref<File[]>([])
const previewUrls = ref<string[]>([])

const form = ref<{
  title: string
  body: string
  opponentHeroId: number
  rank: string
  items: BuildItems
  runes: BuildRunes
  buildNote: string
  sourceUrl: string
}>({
  title: '',
  body: '',
  opponentHeroId: 0,
  rank: '',
  items: emptyItems(),
  runes: emptyRunes(),
  buildNote: '',
  sourceUrl: '',
})

// 是否配置了任意装备或符文（思路随装备符文必填）
function hasBuild() {
  return (
    Object.values(form.value.items).some((slots) => slots.some((s: number | null) => s !== null)) ||
    form.value.runes.primary !== 0 ||
    form.value.runes.secondary !== 0
  )
}

const totalImages = computed(() => keptImages.value.length + newImages.value.length)
const backLink = computed(() => ({ path: '/matchup', query: { champion: heroTitle.value } }))

function isImage(f: File) {
  return ['image/png', 'image/jpeg', 'image/gif', 'image/webp'].includes(f.type)
}

function onFileChange(e: Event) {
  fileError.value = ''
  const files = Array.from((e.target as HTMLInputElement).files ?? [])
  const room = MAX_IMAGES - totalImages.value
  if (files.length > room) {
    fileError.value = `最多 ${MAX_IMAGES} 张配图，还能添加 ${room} 张`
  }
  for (const f of files.slice(0, room)) {
    if (!isImage(f)) {
      fileError.value = '仅支持 PNG / JPG / GIF / WEBP 格式图片'
      continue
    }
    newImages.value.push(f)
    previewUrls.value.push(URL.createObjectURL(f))
  }
  if (fileInput.value) fileInput.value.value = ''
}

function removeKept(i: number) {
  keptImages.value.splice(i, 1)
}
function removeNew(i: number) {
  URL.revokeObjectURL(previewUrls.value[i])
  newImages.value.splice(i, 1)
  previewUrls.value.splice(i, 1)
}

function validate(): string {
  if (!form.value.title.trim()) return '请填写标题'
  if (!form.value.opponentHeroId) return '请从候选中选择对位英雄'
  if (!form.value.rank) return '请选择段位'
  if (hasBuild() && !form.value.buildNote.trim()) return '请填写这套装备符文的思路'
  if (!form.value.body.trim()) return '请填写正文'
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
    kind: 'article',
    title: form.value.title.trim(),
    body: form.value.body,
    durationSec: 0,
    items: form.value.items,
    runes: form.value.runes,
    buildNote: form.value.buildNote.trim(),
    keptImages: keptImages.value,
    sourceUrl: form.value.sourceUrl.trim(),
  }
  saving.value = true
  try {
    if (editId.value) {
      await updateContent(editId.value, { meta, images: newImages.value })
    } else {
      await publishContent({ meta, images: newImages.value })
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
  form.value = {
    title: '',
    body: '',
    opponentHeroId: 0,
    rank: '',
    items: emptyItems(),
    runes: emptyRunes(),
    buildNote: '',
    sourceUrl: '',
  }
  keptImages.value = []
  clearNewImages()
  fileError.value = ''
}

function clearNewImages() {
  previewUrls.value.forEach((u) => URL.revokeObjectURL(u))
  newImages.value = []
  previewUrls.value = []
}

onBeforeUnmount(clearNewImages)

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
        sourceUrl: c.sourceUrl ?? '',
      }
      keptImages.value = [...c.images]
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
  line-height: 1.7;
}

/* 配图 */
.img-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 10px;
}
.img-cell {
  position: relative;
  aspect-ratio: 1;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--line);
}
.img-cell img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.img-del {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: none;
  background: rgba(0, 0, 0, 0.65);
  color: #fff;
  font-size: 13px;
  line-height: 1;
  cursor: pointer;
  padding: 0;
}
.img-del:hover {
  background: var(--red);
}
.img-add {
  aspect-ratio: 1;
  border: 1.5px dashed var(--line);
  border-radius: 8px;
  background: transparent;
  color: var(--faint);
  font-size: 22px;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
}
.img-add span {
  font-size: 12px;
}
.img-add:hover {
  border-color: var(--gold);
  color: var(--gold);
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
