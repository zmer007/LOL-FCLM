<template>
  <div class="editor-page">
    <!-- 非管理员 -->
    <div v-if="checked && !isAdmin" class="forbid">
      <h2>403</h2>
      <p>仅管理员可制作推荐装备 &amp; 符文方案</p>
      <RouterLink to="/" class="back-link">返回首页</RouterLink>
    </div>

    <template v-else-if="checked">
      <div class="page-head">
        <div class="page-title">装备 &amp; 符文</div>
        <div class="page-sub">为英雄制作对线页展示的「推荐装备 &amp; 符文」，支持多套方案</div>
      </div>

      <!-- 英雄选择 -->
      <section class="section hero-pick">
        <span class="hp-label">选择英雄</span>
        <div class="hp-box">
          <input
            v-model="heroKeyword"
            class="hp-input"
            type="text"
            placeholder="输入英雄名称 / 别名（如 阿狸、铁男）"
            @focus="heroOpen = true"
          />
          <div v-if="heroOpen" class="hp-dropdown">
            <div
              v-for="h in filteredHeroes"
              :key="h.id"
              class="hp-opt"
              @mousedown.prevent="selectHero(h)"
            >
              <img :src="h.avatar" :alt="h.title" />
              <span>{{ h.title }}</span>
              <em>{{ h.name }}</em>
            </div>
            <div v-if="!filteredHeroes.length" class="hp-empty">无匹配英雄</div>
          </div>
        </div>
        <span v-if="hero" class="hp-current">当前：{{ hero.title }}（{{ hero.name }}）</span>
      </section>

      <template v-if="hero">
        <!-- 方案栏 -->
        <div class="plans">
          <button
            v-for="(_, i) in plans"
            :key="i"
            type="button"
            class="plan-tab"
            :class="{ on: activeIdx === i }"
            @click="activeIdx = i"
          >
            {{ i + 1 }}
            <span v-if="plans.length > 1" class="del" @click.stop="removePlan(i)">×</span>
          </button>
          <button type="button" class="plan-add" title="新增方案" @click="addPlan">+</button>

          <div class="plan-ops">
            <button type="button" class="btn btn-ghost" @click="isPreview = !isPreview">
              {{ isPreview ? '返回编辑' : '预览效果' }}
            </button>
            <button type="button" class="btn btn-ghost" @click="clearCurrent">清空当前</button>
            <button type="button" class="btn btn-primary" :disabled="saving" @click="save">
              {{ saving ? '保存中…' : '保存方案' }}
            </button>
          </div>
          <div v-if="tip" class="save-tip" :class="{ warn: tipWarn }">{{ tip }}</div>
        </div>

        <div class="editor-grid">
          <!-- 左：素材库（预览态隐藏） -->
          <section v-if="!isPreview" class="section library">
            <div class="section-title">素材库</div>
            <div class="filters">
              <button
                v-for="c in LIB_CATS"
                :key="c"
                type="button"
                class="filter"
                :class="{ on: libCat === c }"
                @click="setLibCat(c)"
              >
                {{ c }}
              </button>
            </div>
            <div class="item-pool">
              <div
                v-for="a in poolAssets"
                :key="a.id"
                class="pool-item"
                :class="{ summ: a.category === 'skill' }"
                :title="a.name"
                @click="fillAsset(a.id)"
              >
                <div class="pool-icon"><img :src="a.icon" :alt="a.name" /></div>
                <div class="pool-name">{{ a.name }}</div>
              </div>
            </div>
            <div class="lib-hint">
              点击右侧空格选中，再点此处物品填入；<br />
              拖动已填格子可互换；点击已填格子可移除。
            </div>
          </section>

          <!-- 右：制作区 -->
          <div class="workbench">
            <section class="section">
              <div class="section-title">推荐装备</div>
              <ItemSlots
                :items="current.items"
                :asset-map="assetMap"
                :active-slot="activeSlot"
                :editable="!isPreview"
                @pick="onPick"
                @remove="onRemove"
                @swap="onSwap"
              />
            </section>

            <section class="section">
              <div class="section-title">推荐符文</div>
              <RunePage
                :runes="current.runes"
                :trees="trees"
                :shards="shards"
                :editable="!isPreview"
              />
              <input
                v-model="current.vs"
                type="text"
                class="vs-input"
                placeholder="适用于对抗：输入对位英雄，多个用顿号分隔"
              />
            </section>
          </div>
        </div>
      </template>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { listHeroes, type Hero } from '../api/hero'
import { getAssetMap, type GameAsset } from '../api/asset'
import {
  emptyItems,
  emptyRunes,
  getHeroPlans,
  getRuneTrees,
  saveHeroPlans,
  type BuildItems,
  type PlanInput,
  type RuneTree,
  type Shard,
} from '../api/build'
import { useUserStore } from '../stores/user'
import ItemSlots from '../components/build/ItemSlots.vue'
import RunePage from '../components/build/RunePage.vue'

type StageKey = keyof BuildItems

const LIB_CATS = ['全部', '攻击', '法术', '防御', '移动', '消耗', '其他', '召唤师技能'] as const

const route = useRoute()
const userStore = useUserStore()

const checked = ref(false)
const isAdmin = computed(() => userStore.user?.role === 'admin')

const heroes = ref<Hero[]>([])
const hero = ref<Hero | null>(null)
const heroKeyword = ref('')
const heroOpen = ref(false)

const assetMap = ref<Map<number, GameAsset>>(new Map())
const trees = ref<RuneTree[]>([])
const shards = ref<Shard[][]>([])

const plans = ref<PlanInput[]>([])
const activeIdx = ref(0)
const activeSlot = ref<{ stage: StageKey; index: number } | null>(null)
const libCat = ref<(typeof LIB_CATS)[number]>('全部')
const isPreview = ref(false)
const saving = ref(false)
const tip = ref('')
const tipWarn = ref(false)
let tipTimer: ReturnType<typeof setTimeout> | undefined

const current = computed<PlanInput>(() => plans.value[activeIdx.value])

const filteredHeroes = computed(() => {
  const q = heroKeyword.value.trim().toLowerCase()
  const all = heroes.value
  if (!q) return all.slice(0, 30)
  return all
    .filter(
      (h) =>
        h.title.toLowerCase().includes(q) ||
        h.name.toLowerCase().includes(q) ||
        h.aliases.some((a) => a.toLowerCase().includes(q)),
    )
    .slice(0, 30)
})

// 素材库：召唤师技能槽强制技能分类，其余按装备分类
const poolAssets = computed<GameAsset[]>(() => {
  const all = [...assetMap.value.values()]
  if (activeSlot.value?.stage === 'summs') {
    return all.filter((a) => a.category === 'skill')
  }
  if (libCat.value === '全部') return all.filter((a) => a.category === 'item')
  if (libCat.value === '召唤师技能') return all.filter((a) => a.category === 'skill')
  return all.filter((a) => a.category === 'item' && a.subCategory === libCat.value)
})

function showTip(msg: string, warn = false) {
  tip.value = msg
  tipWarn.value = warn
  clearTimeout(tipTimer)
  tipTimer = setTimeout(() => (tip.value = ''), 2600)
}

function blankPlan(): PlanInput {
  return { vs: '', items: emptyItems(), runes: emptyRunes() }
}

async function selectHero(h: Hero) {
  hero.value = h
  heroKeyword.value = h.title
  heroOpen.value = false
  activeSlot.value = null
  libCat.value = '全部'
  try {
    const list = await getHeroPlans(h.id)
    plans.value = list.list.length
      ? list.list.map((p) => ({ vs: p.vs, items: p.items, runes: p.runes }))
      : [blankPlan()]
  } catch {
    plans.value = [blankPlan()]
  }
  activeIdx.value = 0
}

function addPlan() {
  if (plans.value.length >= 20) {
    showTip('最多 20 套方案', true)
    return
  }
  plans.value.push(blankPlan())
  activeIdx.value = plans.value.length - 1
  isPreview.value = false
}
function removePlan(i: number) {
  plans.value.splice(i, 1)
  if (activeIdx.value >= plans.value.length) activeIdx.value = plans.value.length - 1
}
function clearCurrent() {
  plans.value[activeIdx.value] = blankPlan()
  activeSlot.value = null
}

function setLibCat(c: (typeof LIB_CATS)[number]) {
  // 当前目标槽是召唤师技能时，不允许切到装备分类
  if (activeSlot.value?.stage === 'summs' && c !== '召唤师技能') return
  libCat.value = c
}
function onPick(p: { stage: StageKey; index: number }) {
  activeSlot.value = p
  libCat.value = p.stage === 'summs' ? '召唤师技能' : libCat.value === '召唤师技能' ? '全部' : libCat.value
}
function onRemove(p: { stage: StageKey; index: number }) {
  current.value.items[p.stage][p.index] = null
}
function onSwap(p: {
  fromStage: StageKey
  fromIndex: number
  toStage: StageKey
  toIndex: number
}) {
  const summLike = (s: StageKey) => s === 'summs'
  if (summLike(p.fromStage) !== summLike(p.toStage)) {
    showTip('装备与召唤师技能槽不能互换', true)
    return
  }
  const items = current.value.items
  const tmp = items[p.fromStage][p.fromIndex]
  items[p.fromStage][p.fromIndex] = items[p.toStage][p.toIndex]
  items[p.toStage][p.toIndex] = tmp
}
function fillAsset(id: number) {
  if (!activeSlot.value) {
    showTip('请先点击右侧一个装备空格', true)
    return
  }
  current.value.items[activeSlot.value.stage][activeSlot.value.index] = id
}

async function save() {
  if (!hero.value) return
  saving.value = true
  try {
    await saveHeroPlans(hero.value.id, plans.value)
    showTip('已保存')
  } catch (e) {
    showTip(e instanceof Error ? e.message : '保存失败', true)
  } finally {
    saving.value = false
  }
}

function resolveInitial(list: Hero[], q: string): Hero | undefined {
  const key = q.trim().toLowerCase()
  if (!key) return undefined
  const eq = (s: string) => s.toLowerCase() === key
  return (
    list.find((h) => eq(h.title) || eq(h.name) || h.aliases.some(eq)) ||
    list.find(
      (h) =>
        h.title.toLowerCase().includes(key) ||
        h.name.toLowerCase().includes(key) ||
        h.aliases.some((a) => a.toLowerCase().includes(key)),
    )
  )
}

onMounted(async () => {
  await userStore.loadMe()
  checked.value = true
  if (!isAdmin.value) return

  const [heroList, assets, runeData] = await Promise.all([
    listHeroes(),
    getAssetMap(),
    getRuneTrees(),
  ])
  heroes.value = heroList.list
  assetMap.value = assets
  trees.value = runeData.trees
  shards.value = runeData.shards
  plans.value = [blankPlan()]

  const initial = resolveInitial(heroList.list, String(route.query.champion ?? ''))
  if (initial) await selectHero(initial)
})
</script>

<style scoped>
.editor-page {
  max-width: 1180px;
  margin: 0 auto;
  padding: 0 24px 60px;
}
.forbid {
  text-align: center;
  padding: 100px 0;
  color: var(--mute);
}
.forbid h2 {
  font-size: 48px;
  color: var(--gold);
  margin: 0 0 12px;
}
.back-link {
  display: inline-block;
  margin-top: 16px;
  color: var(--gold);
}
.page-head {
  margin: 26px 0 20px;
}
.page-title {
  font-size: 24px;
  font-weight: 800;
  color: var(--gold-bright);
}
.page-sub {
  margin-top: 6px;
  font-size: 13px;
  color: var(--faint);
}

.section {
  background: var(--panel);
  border: 1px solid var(--line-soft);
  border-radius: 16px;
  padding: 22px 24px 24px;
}
.section-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--gold-bright);
  margin-bottom: 18px;
}

/* 英雄选择 */
.hero-pick {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 20px;
  position: relative;
}
.hp-label {
  font-size: 13.5px;
  color: var(--gold-bright);
  font-weight: 700;
}
.hp-box {
  position: relative;
  width: 300px;
}
.hp-input {
  width: 100%;
  font-family: inherit;
  font-size: 13px;
  background: #0a1520;
  border: 1px solid var(--line);
  border-radius: 8px;
  color: var(--gold-bright);
  padding: 8px 12px;
}
.hp-input:focus {
  outline: none;
  border-color: var(--gold);
}
.hp-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  max-height: 280px;
  overflow-y: auto;
  background: var(--panel-2, #132438);
  border: 1px solid var(--gold);
  border-radius: 8px;
  z-index: 30;
}
.hp-opt {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 12px;
  cursor: pointer;
}
.hp-opt:hover {
  background: rgba(200, 170, 110, 0.12);
}
.hp-opt img {
  width: 28px;
  height: 28px;
  border-radius: 6px;
}
.hp-opt span {
  font-size: 13px;
  color: var(--gold-bright);
}
.hp-opt em {
  font-style: normal;
  font-size: 11.5px;
  color: var(--faint);
}
.hp-empty {
  padding: 16px;
  text-align: center;
  font-size: 12.5px;
  color: var(--faint);
}
.hp-current {
  font-size: 12.5px;
  color: var(--cyan);
}

/* 方案栏 */
.plans {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 20px;
}
.plan-tab {
  position: relative;
  width: 32px;
  height: 30px;
  font-family: inherit;
  font-size: 13px;
  cursor: pointer;
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 8px;
  color: var(--faint);
}
.plan-tab.on {
  border-color: var(--gold);
  color: var(--gold);
}
.plan-tab .del {
  position: absolute;
  top: -7px;
  right: -7px;
  width: 15px;
  height: 15px;
  border-radius: 50%;
  background: var(--bg);
  border: 1px solid var(--line);
  color: var(--faint);
  font-size: 10px;
  line-height: 13px;
  cursor: pointer;
}
.plan-tab .del:hover {
  color: var(--red);
  border-color: var(--red);
}
.plan-add {
  width: 32px;
  height: 30px;
  font-size: 16px;
  background: none;
  border: 1px dashed var(--line);
  border-radius: 8px;
  color: var(--faint);
  cursor: pointer;
}
.plan-add:hover {
  border-color: var(--gold);
  color: var(--gold);
}
.plan-ops {
  margin-left: auto;
  display: flex;
  gap: 10px;
}
.btn {
  height: 32px;
  padding: 0 18px;
  font-family: inherit;
  font-size: 13px;
  cursor: pointer;
  border-radius: 8px;
}
.btn-primary {
  background: var(--gold);
  border: none;
  color: #0a1428;
  font-weight: 700;
}
.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.btn-ghost {
  background: none;
  border: 1px solid var(--line);
  color: var(--mute);
}
.btn-ghost:hover {
  border-color: var(--gold);
  color: var(--gold);
}
.save-tip {
  width: 100%;
  font-size: 12px;
  color: var(--cyan);
  order: 5;
}
.save-tip.warn {
  color: var(--red);
}

/* 编辑布局 */
.editor-grid {
  display: grid;
  grid-template-columns: 296px 1fr;
  gap: 22px;
  align-items: start;
}
.library {
  position: sticky;
  top: 86px;
}
.filters {
  display: flex;
  gap: 7px;
  flex-wrap: wrap;
  margin-bottom: 18px;
}
.filter {
  font-family: inherit;
  font-size: 12px;
  cursor: pointer;
  background: none;
  border: 1px solid var(--line);
  border-radius: 999px;
  color: var(--faint);
  padding: 4px 13px;
}
.filter.on {
  border-color: var(--gold);
  color: var(--gold);
}
.item-pool {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(66px, 1fr));
  gap: 16px 10px;
  max-height: 520px;
  overflow-y: auto;
  padding-right: 4px;
}
.pool-item {
  text-align: center;
  cursor: pointer;
}
.pool-icon {
  width: 52px;
  height: 52px;
  margin: 0 auto;
  border-radius: 10px;
  border: 1px solid var(--line);
  background: #0a1520;
  overflow: hidden;
  transition: border-color 0.15s, transform 0.15s;
}
.pool-item:hover .pool-icon {
  border-color: var(--gold);
  transform: translateY(-2px);
}
.pool-item.summ:hover .pool-icon {
  border-color: var(--cyan);
}
.pool-icon img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.pool-name {
  margin-top: 7px;
  font-size: 11.5px;
  color: var(--mute);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.lib-hint {
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid var(--line-soft);
  font-size: 12px;
  color: var(--faint);
  line-height: 1.8;
}

.workbench {
  display: flex;
  flex-direction: column;
  gap: 22px;
}
.vs-input {
  width: 100%;
  margin-top: 20px;
  font-family: inherit;
  font-size: 13px;
  background: #0a1520;
  border: 1px solid var(--line);
  border-radius: 8px;
  color: var(--gold-bright);
  padding: 9px 12px;
}
.vs-input:focus {
  outline: none;
  border-color: var(--gold);
}
</style>
