<template>
  <div v-if="current" class="build-panel">
    <!-- 多方案数字 tab -->
    <div class="plan-tabs">
      <button
        v-for="p in plans"
        :key="p.id"
        type="button"
        class="plan-tab"
        :class="{ active: activeId === p.id }"
        @click="activeId = p.id"
      >
        {{ p.seq }}
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
      </div>
    </div>

    <div v-if="current.vs" class="vs-line">适用于对抗：<em>{{ current.vs }}</em></div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { BuildPlan, RuneTree, Shard } from '../../api/build'
import type { GameAsset } from '../../api/asset'
import ItemSlots from './ItemSlots.vue'
import RunePage from './RunePage.vue'

const props = defineProps<{
  plans: BuildPlan[]
  assetMap: Map<number, GameAsset>
  trees: RuneTree[]
  shards: Shard[][]
}>()

const activeId = ref(props.plans[0]?.id ?? 0)
watch(
  () => props.plans,
  (list) => {
    if (!list.some((p) => p.id === activeId.value)) activeId.value = list[0]?.id ?? 0
  },
)

const current = computed(() => props.plans.find((p) => p.id === activeId.value))
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
.vs-line {
  margin-top: 14px;
  font-size: 12.5px;
  color: var(--faint);
}
.vs-line em {
  font-style: normal;
  color: var(--cyan);
}
</style>
