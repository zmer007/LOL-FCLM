<template>
  <div v-if="current" class="build-panel">
    <!-- 多方案数字 tab -->
    <div class="plan-tabs">
      <button
        v-for="(p, i) in builds"
        :key="p.contentId"
        type="button"
        class="plan-tab"
        :class="{ active: activeIdx === i }"
        @click="activeIdx = i"
      >
        {{ i + 1 }}
      </button>
    </div>

    <div class="build-card">
      <div class="build-left">
        <div class="block-title">推荐装备</div>
        <ItemSlots :items="current.items" :asset-map="assetMap" />
      </div>
      <div class="build-right">
        <div class="block-title">推荐符文</div>
        <RunePage :runes="current.runes" :trees="trees" :shards="shards" />
        <p v-if="current.buildNote" class="build-note">{{ current.buildNote }}</p>
      </div>
    </div>

    <div class="vs-line">
      <RouterLink class="src-link" :to="detailPath">
        {{ current.title }}<em v-if="current.vs"> · 适用于对抗：{{ current.vs }}</em>
      </RouterLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import type { RuneTree, Shard } from '../../api/build'
import type { HeroBuild } from '../../api/content'
import type { GameAsset } from '../../api/asset'
import ItemSlots from './ItemSlots.vue'
import RunePage from './RunePage.vue'

const props = defineProps<{
  builds: HeroBuild[]
  assetMap: Map<number, GameAsset>
  trees: RuneTree[]
  shards: Shard[][]
}>()

const activeIdx = ref(0)
watch(
  () => props.builds,
  (list) => {
    if (activeIdx.value >= list.length) activeIdx.value = 0
  },
)

const current = computed(() => props.builds[activeIdx.value])
const detailPath = computed(() =>
  current.value
    ? `/${current.value.kind === 'video' ? 'video' : 'article'}/${current.value.contentId}`
    : '/',
)
</script>

<style scoped>
.plan-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}
.plan-tab {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: var(--panel);
  border: 1px solid var(--line);
  color: var(--mute);
  font-size: 13px;
  font-family: inherit;
  cursor: pointer;
}
.plan-tab.active {
  background: var(--gold);
  color: #0a1428;
  border-color: var(--gold);
  font-weight: 700;
}
.build-card {
  display: flex;
  gap: 30px;
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 22px;
}
.build-left {
  flex: 1;
  min-width: 0;
}
.build-right {
  flex: 1;
  min-width: 0;
}
.block-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--gold-bright);
  margin-bottom: 14px;
}
.build-note {
  margin: 14px 0 0;
  font-size: 12.5px;
  line-height: 1.7;
  color: var(--mute);
  white-space: pre-wrap;
  word-break: break-word;
}
.vs-line {
  margin-top: 14px;
  font-size: 12.5px;
}
.src-link {
  color: var(--faint);
  text-decoration: none;
}
.src-link:hover {
  color: var(--gold);
}
.src-link em {
  font-style: normal;
  color: var(--cyan);
}

/* 展示态符文区整体左移 */
.build-right :deep(.rune-page) {
  margin-left: -30px;
}
</style>
