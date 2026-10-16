<template>
  <div class="rune-wrap">
    <div v-if="editable || runes.primary || runes.secondary" class="rune-page">
      <!-- 主系：系选择器 + 基石 + 三排 -->
      <div class="rune-path primary">
        <div v-if="editable" class="path-picker-col">
          <span class="picker-label">主系</span>
          <div class="pick-row">
            <button
              v-for="t in trees"
              :key="'p' + t.id"
              type="button"
              class="pick"
              :class="{ on: runes.primary === t.id, dis: runes.secondary === t.id }"
              :title="t.name"
              @click="setPrimary(t.id)"
            >
              <img :src="t.icon" :alt="t.name" />
            </button>
          </div>
        </div>
        <div class="rune-head">{{ primaryTree ? primaryTree.name : '主系符文' }}</div>
        <template v-if="primaryTree">
          <div
            v-for="row in [0, 1, 2, 3]"
            :key="row"
            class="rune-candidates"
            :class="{ keystone: row === 0 }"
          >
            <button
              v-for="rn in runesOf(primaryTree, row)"
              :key="rn.id"
              type="button"
              class="rune-slot"
              :class="{ on: runes.primaryPicks[row] === rn.id }"
              :title="rn.name"
              @click="pickPrimary(row, rn.id)"
            >
              <img :src="rn.icon" :alt="rn.name" />
            </button>
          </div>
        </template>
        <template v-else>
          <div
            v-for="row in [0, 1, 2, 3]"
            :key="'ph' + row"
            class="rune-candidates"
            :class="{ keystone: row === 0 }"
          >
            <div
              v-for="i in (row === 0 ? 4 : 3)"
              :key="i"
              class="rune-slot ph"
            ></div>
          </div>
        </template>
      </div>

      <!-- 副系：系选择器 + 三排选两排 -->
      <div class="rune-path secondary">
        <div v-if="editable" class="path-picker-col">
          <span class="picker-label">副系</span>
          <div class="pick-row">
            <button
              v-for="t in trees"
              :key="'s' + t.id"
              type="button"
              class="pick cyan"
              :class="{ on: runes.secondary === t.id, dis: runes.primary === t.id }"
              :title="t.name"
              @click="setSecondary(t.id)"
            >
              <img :src="t.icon" :alt="t.name" />
            </button>
          </div>
        </div>
        <div class="rune-head">{{ secondaryTree ? secondaryTree.name : '副系符文' }}</div>
        <template v-if="secondaryTree">
          <div
            v-for="slotIdx in [0, 1, 2]"
            :key="slotIdx"
            class="rune-candidates"
          >
            <button
              v-for="rn in runesOf(secondaryTree, slotIdx + 1)"
              :key="rn.id"
              type="button"
              class="rune-slot"
              :class="{ on: runes.secondaryPicks[slotIdx] === rn.id }"
              :title="rn.name"
              @click="pickSecondary(slotIdx, rn.id)"
            >
              <img :src="rn.icon" :alt="rn.name" />
            </button>
          </div>
        </template>
        <template v-else>
          <div
            v-for="slotIdx in [0, 1, 2]"
            :key="'ph' + slotIdx"
            class="rune-candidates"
          >
            <div
              v-for="i in 3"
              :key="i"
              class="rune-slot ph"
            ></div>
          </div>
        </template>
      </div>

      <!-- 属性碎片：3 行每行 3 选 1 -->
      <div class="shards">
        <div class="rune-head">属性碎片</div>
        <div v-for="(row, ri) in shards" :key="ri" class="shard-row">
          <button
            v-for="(sh, ci) in row"
            :key="ci"
            type="button"
            class="shard"
            :class="{ on: runes.shards[ri] === ci }"
            :title="sh.name"
            @click="pickShard(ri, ci)"
          >
            <img :src="sh.icon" :alt="sh.name" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { BuildRunes, RuneTree, Shard } from '../../api/build'

const props = withDefaults(
  defineProps<{
    runes: BuildRunes
    trees: RuneTree[]
    shards: Shard[][]
    editable?: boolean
  }>(),
  { editable: false },
)

function runesOf(tree: RuneTree, row: number): RuneTree['runes'] {
  return tree.runes.filter((r) => r.row === row)
}

const primaryTree = computed<RuneTree | undefined>(() =>
  props.trees.find((t) => t.id === props.runes.primary),
)
const secondaryTree = computed<RuneTree | undefined>(() =>
  props.trees.find((t) => t.id === props.runes.secondary),
)

function setPrimary(id: number) {
  if (!props.editable || id === props.runes.secondary) return
  props.runes.primary = id
  props.runes.primaryPicks = [0, 0, 0, 0]
}
function setSecondary(id: number) {
  if (!props.editable || id === props.runes.primary) return
  props.runes.secondary = id
  props.runes.secondaryPicks = [null, null, null]
}
function pickPrimary(row: number, id: number) {
  if (!props.editable) return
  props.runes.primaryPicks[row] = props.runes.primaryPicks[row] === id ? 0 : id
}
function pickSecondary(slotIdx: number, id: number) {
  if (!props.editable) return
  const picks = props.runes.secondaryPicks
  if (picks[slotIdx] === id) {
    picks[slotIdx] = null
    return
  }
  // 已选两排再选新排时，替换最早选中的一排
  const chosen = picks.map((v, i) => (v != null ? i : -1)).filter((i) => i >= 0)
  if (chosen.length >= 2 && picks[slotIdx] == null) {
    picks[chosen[0]] = null
  }
  picks[slotIdx] = id
}
function pickShard(row: number, col: number) {
  if (!props.editable) return
  props.runes.shards[row] = props.runes.shards[row] === col ? -1 : col
}
</script>

<style scoped>
.path-picker-col {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  gap: 6px;
  margin-bottom: 8px;
}
.picker-label {
  font-size: 12px;
  color: var(--faint);
}
.pick-row {
  display: flex;
  flex-wrap: nowrap;
  justify-content: center;
  gap: 6px;
}
.pick {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  padding: 0;
  background: #0a1520;
  border: 1.5px solid #2a3f55;
  opacity: 0.55;
  cursor: pointer;
  overflow: hidden;
}
.pick img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.pick:hover {
  opacity: 0.9;
}
.pick.on {
  opacity: 1;
  border-color: var(--gold);
  box-shadow: 0 0 8px rgba(200, 170, 110, 0.4);
}
.pick.cyan.on {
  border-color: var(--cyan);
  box-shadow: 0 0 8px rgba(10, 200, 185, 0.4);
}
.pick.dis {
  opacity: 0.25;
  cursor: not-allowed;
}

.rune-page {
  display: flex;
  justify-content: flex-start;
  align-items: stretch;
  gap: 26px;
}
.rune-path {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}
.rune-path .rune-head + .rune-candidates {
  margin-top: auto;
}
.shards .rune-head + .shard-row {
  margin-top: auto;
}
.rune-head {
  font-size: 12px;
  color: var(--faint);
  margin-bottom: 8px;
  white-space: nowrap;
}
.rune-path.secondary .rune-head {
  margin-bottom: 8px;
}
.rune-candidates {
  display: flex;
  gap: 12px;
}
.rune-slot {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  background: #0a1520;
  border: 1.5px solid #2a3f55;
  opacity: 0.4;
  padding: 0;
  cursor: default;
  overflow: hidden;
}
.rune-candidates.keystone .rune-slot {
  width: 42px;
  height: 42px;
}
.rune-slot img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.rune-slot.on {
  opacity: 1;
  border-color: var(--gold);
  box-shadow: 0 0 10px rgba(200, 170, 110, 0.35);
}
.rune-path.secondary .rune-slot.on {
  border-color: var(--cyan);
  box-shadow: 0 0 10px rgba(10, 200, 185, 0.35);
}
.shards {
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  gap: 18px;
  padding-left: 24px;
  border-left: 1px solid var(--line-soft);
}
.shards .rune-head {
  margin-bottom: 8px;
}
.shard-row {
  display: flex;
  gap: 10px;
}
.shard {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: #0a1520;
  border: 1.5px solid #2a3f55;
  opacity: 0.4;
  padding: 0;
  cursor: default;
  overflow: hidden;
}
.shard img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.shard.on {
  opacity: 1;
  border-color: var(--gold);
  box-shadow: 0 0 8px rgba(200, 170, 110, 0.35);
}
.rune-empty {
  text-align: center;
  font-size: 12.5px;
  color: var(--faint);
  padding: 10px 0;
}
.rune-slot.ph {
  opacity: 0.18;
  border-style: dashed;
}
</style>
