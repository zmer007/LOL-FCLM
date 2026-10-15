<template>
  <div class="item-slots">
    <!-- 出门装 + 召唤师技能并排 -->
    <div class="stage-row">
      <div v-for="key in TOP_STAGES" :key="key" class="item-stage">
        <div class="stage-label">{{ STAGE_LABEL[key] }}</div>
        <div class="items">
          <template v-for="(id, idx) in items[key]" :key="idx">
            <div
              class="item"
              :class="{ empty: id == null, active: isActive(key, idx) }"
              :draggable="editable && id != null"
              :title="id != null ? nameOf(id) : ''"
              @click="onSlotClick(key, idx)"
              @dragstart="onDragStart(key, idx)"
              @dragover.prevent="onDragOver"
              @drop.prevent="onDrop(key, idx)"
            >
              <img v-if="id != null && iconOf(id)" :src="iconOf(id)" :alt="nameOf(id)" draggable="false" />
              <span v-else-if="editable" class="plus">＋</span>
            </div>
          </template>
        </div>
      </div>
    </div>

    <!-- 其余三个分区各占一行 -->
    <div v-for="key in OTHER_STAGES" :key="key" class="item-stage">
      <div class="stage-label">{{ STAGE_LABEL[key] }}</div>
      <div class="items">
        <template v-for="(id, idx) in items[key]" :key="idx">
          <div
            class="item"
            :class="{ empty: id == null, active: isActive(key, idx) }"
            :draggable="editable && id != null"
            :title="id != null ? nameOf(id) : ''"
            @click="onSlotClick(key, idx)"
            @dragstart="onDragStart(key, idx)"
            @dragover.prevent="onDragOver"
            @drop.prevent="onDrop(key, idx)"
          >
            <img v-if="id != null && iconOf(id)" :src="iconOf(id)" :alt="nameOf(id)" draggable="false" />
            <span v-else-if="editable" class="plus">＋</span>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { BuildItems } from '../../api/build'
import type { GameAsset } from '../../api/asset'

type StageKey = keyof BuildItems

const props = defineProps<{
  items: BuildItems
  assetMap: Map<number, GameAsset>
  editable?: boolean
  activeSlot?: { stage: StageKey; index: number } | null
}>()

const emit = defineEmits<{
  (e: 'pick', payload: { stage: StageKey; index: number }): void
  (e: 'remove', payload: { stage: StageKey; index: number }): void
  (e: 'swap', payload: { fromStage: StageKey; fromIndex: number; toStage: StageKey; toIndex: number }): void
}>()

const STAGE_LABEL: Record<StageKey, string> = {
  start: '出门装',
  summs: '召唤师技能',
  mid: '中期过渡',
  core: '核心装备',
  final: '后期神装',
}
const TOP_STAGES: StageKey[] = ['start', 'summs']
const OTHER_STAGES: StageKey[] = ['mid', 'core', 'final']

function assetOf(id: number | null): GameAsset | undefined {
  return id == null ? undefined : props.assetMap.get(id)
}
function iconOf(id: number | null): string {
  return assetOf(id)?.icon ?? ''
}
function nameOf(id: number | null): string {
  return assetOf(id)?.name ?? ''
}
function isActive(stage: StageKey, index: number): boolean {
  return props.activeSlot?.stage === stage && props.activeSlot?.index === index
}

function onSlotClick(stage: StageKey, index: number) {
  if (!props.editable) return
  if (props.items[stage][index] == null) {
    emit('pick', { stage, index })
  } else {
    // 已填槽：再次点击已激活的槽 → 移除；否则切换为当前目标
    if (isActive(stage, index)) {
      emit('remove', { stage, index })
    } else {
      emit('pick', { stage, index })
    }
  }
}

let dragging: { stage: StageKey; index: number } | null = null
function onDragOver() {
  // 仅用于配合 .prevent 允许 drop
}
function onDragStart(stage: StageKey, index: number) {
  dragging = { stage, index }
}
function onDrop(stage: StageKey, index: number) {
  if (!dragging) return
  emit('swap', {
    fromStage: dragging.stage,
    fromIndex: dragging.index,
    toStage: stage,
    toIndex: index,
  })
  dragging = null
}
</script>

<style scoped>
.item-stage {
  margin-bottom: 16px;
}
.item-stage:last-child {
  margin-bottom: 0;
}
.stage-row {
  display: flex;
  gap: 28px;
}
.stage-label {
  font-size: 12px;
  color: var(--faint);
  margin-bottom: 8px;
}
.items {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.item {
  width: 48px;
  height: 48px;
  border-radius: 9px;
  background: #0a1520;
  border: 1px solid var(--line);
  display: grid;
  place-items: center;
  overflow: hidden;
}
.item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 8px;
}
.item.empty {
  background: #0a1520;
}
.item .plus {
  font-size: 16px;
  color: var(--faint);
}
.item.active {
  border-color: var(--gold);
  box-shadow: 0 0 8px rgba(200, 170, 110, 0.4);
}
.item[draggable='true'] {
  cursor: grab;
}
</style>
