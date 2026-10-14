<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

// 对线页占位：里程碑 3 将按 design/对线页面.html 完整实现
const route = useRoute()
const champion = computed(() => (route.query.champion as string) || '')
</script>

<template>
  <main class="matchup-placeholder">
    <p class="ph-name">{{ champion }}</p>
    <p class="ph-tip">对线页建设中（里程碑 3）</p>
  </main>
</template>

<style scoped>
.matchup-placeholder {
  text-align: center;
  padding-top: 24vh;
}
.ph-name {
  font-size: 28px;
  color: var(--gold-bright);
  letter-spacing: 4px;
}
.ph-tip {
  margin-top: 14px;
  font-size: 13px;
  color: var(--faint);
}
</style>
