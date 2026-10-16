<template>
  <div class="build-card">
    <div class="build-left">
      <div class="block-title">推荐装备</div>
      <ItemSlots :items="items" :asset-map="assetMap" />
    </div>
    <div class="build-right">
      <div class="block-title">推荐符文</div>
      <RunePage :runes="runes" :trees="trees" :shards="shards" />
      <p v-if="buildNote" class="build-note">{{ buildNote }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { GameAsset } from '../../api/asset'
import type { BuildItems, BuildRunes, RuneTree, Shard } from '../../api/build'
import ItemSlots from './ItemSlots.vue'
import RunePage from './RunePage.vue'

defineProps<{
  items: BuildItems
  runes: BuildRunes
  buildNote: string
  assetMap: Map<number, GameAsset>
  trees: RuneTree[]
  shards: Shard[][]
}>()
</script>

<style scoped>
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
/* 展示态符文区整体左移，与对线页对齐 */
.build-right :deep(.rune-page) {
  margin-left: -30px;
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
@media (max-width: 720px) {
  .build-card {
    flex-direction: column;
  }
}
</style>
