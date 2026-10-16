<template>
  <div class="build-editor">
    <div class="be-title">推荐装备</div>
    <ItemSlots
      :items="items"
      :asset-map="assetMap"
      :active-slot="activeSlot"
      editable
      @pick="openPicker"
      @remove="onRemove"
      @swap="onSwap"
    />

    <div class="be-title be-title-2">推荐符文</div>
    <RunePage :runes="runes" :trees="trees" :shards="shards" editable />

    <!-- 素材选择弹窗：点装备/技能空槽后弹出 -->
    <div v-if="pickerOpen" class="picker-mask" @click.self="pickerOpen = false">
      <div class="picker-dialog">
        <div class="picker-head">
          <span>{{ activeSlot?.stage === 'summs' ? '选择召唤师技能' : '选择装备' }}</span>
          <button type="button" class="picker-close" @click="pickerOpen = false">×</button>
        </div>
        <AssetLibrary :active-slot="activeSlot" @fill="onFill" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { GameAsset } from '../../api/asset'
import type { BuildItems, BuildRunes, RuneTree, Shard } from '../../api/build'
import ItemSlots from './ItemSlots.vue'
import RunePage from './RunePage.vue'
import AssetLibrary from './AssetLibrary.vue'

type StageKey = keyof BuildItems

const props = defineProps<{
  items: BuildItems
  runes: BuildRunes
  assetMap: Map<number, GameAsset>
  trees: RuneTree[]
  shards: Shard[][]
}>()

const activeSlot = ref<{ stage: StageKey; index: number } | null>(null)
const pickerOpen = ref(false)

function openPicker(slot: { stage: StageKey; index: number }) {
  activeSlot.value = slot
  pickerOpen.value = true
}
function onFill(id: number) {
  if (activeSlot.value) {
    props.items[activeSlot.value.stage][activeSlot.value.index] = id
    pickerOpen.value = false
  }
}
function onRemove(p: { stage: StageKey; index: number }) {
  props.items[p.stage][p.index] = null
}
function onSwap(p: {
  fromStage: StageKey
  fromIndex: number
  toStage: StageKey
  toIndex: number
}) {
  if ((p.fromStage === 'summs') !== (p.toStage === 'summs')) return
  const a = props.items[p.fromStage][p.fromIndex]
  props.items[p.fromStage][p.fromIndex] = props.items[p.toStage][p.toIndex]
  props.items[p.toStage][p.toIndex] = a
}
</script>

<style scoped>
.be-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--gold-bright);
  margin-bottom: 14px;
}
.be-title-2 {
  margin-top: 26px;
}
.picker-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: grid;
  place-items: center;
  z-index: 80;
}
.picker-dialog {
  width: min(560px, 92vw);
  max-height: 80vh;
  overflow-y: auto;
  background: var(--panel-2, #132438);
  border: 1px solid var(--gold);
  border-radius: 14px;
  padding: 20px 22px;
}
.picker-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 14px;
  font-weight: 700;
  color: var(--gold-bright);
  margin-bottom: 16px;
}
.picker-close {
  background: none;
  border: none;
  color: var(--faint);
  font-size: 20px;
  cursor: pointer;
  line-height: 1;
}
.picker-close:hover {
  color: var(--gold);
}
</style>
