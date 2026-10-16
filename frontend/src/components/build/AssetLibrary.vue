<template>
  <div class="asset-library">
    <div class="filters">
      <button
        v-for="c in cats"
        :key="c"
        type="button"
        class="filter"
        :class="{ on: activeCat === c, disabled: forceSkill && c !== '召唤师技能' }"
        @click="setCat(c)"
      >
        {{ c }}
      </button>
    </div>
    <div class="item-pool">
      <div
        v-for="a in poolAssets"
        :key="a.id"
        class="pool-item"
        :class="{ summ: a.category === 'skill', dim: !activeSlot }"
        :title="a.name"
        @click="emit('fill', a.id)"
      >
        <div class="pool-icon"><img :src="a.icon" :alt="a.name" /></div>
        <div class="pool-name">{{ a.name }}</div>
      </div>
    </div>
    <div class="lib-hint">{{ activeSlot ? '点击上方物品填入选中的槽位；拖动已填格子可互换。' : '请先点击右侧一个装备空格' }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { getAssetMap, type GameAsset } from '../../api/asset'
import type { BuildItems } from '../../api/build'

type StageKey = keyof BuildItems

const props = defineProps<{
  activeSlot: { stage: StageKey; index: number } | null
}>()
const emit = defineEmits<{ (e: 'fill', id: number): void }>()

const cats = ['全部', '攻击', '法术', '防御', '移动', '消耗', '其他', '召唤师技能'] as const
type Cat = (typeof cats)[number]

const assets = ref<GameAsset[]>([])
const activeCat = ref<Cat>('全部')

getAssetMap().then((m) => (assets.value = [...m.values()]))

const forceSkill = computed(() => props.activeSlot?.stage === 'summs')

// 选中召唤师技能槽时自动切到技能分类
watch(
  () => props.activeSlot?.stage,
  (stage) => {
    if (stage === 'summs') activeCat.value = '召唤师技能'
    else if (activeCat.value === '召唤师技能') activeCat.value = '全部'
  },
)

function setCat(c: Cat) {
  if (forceSkill.value && c !== '召唤师技能') return
  activeCat.value = c
}

const poolAssets = computed<GameAsset[]>(() => {
  if (forceSkill.value) return assets.value.filter((a) => a.category === 'skill')
  if (activeCat.value === '全部') return assets.value.filter((a) => a.category === 'item')
  if (activeCat.value === '召唤师技能') return assets.value.filter((a) => a.category === 'skill')
  return assets.value.filter((a) => a.category === 'item' && a.subCategory === activeCat.value)
})
</script>

<style scoped>
.filters {
  display: flex;
  gap: 7px;
  flex-wrap: wrap;
  margin-bottom: 16px;
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
.filter.disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
.item-pool {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(64px, 1fr));
  gap: 14px 10px;
  max-height: 480px;
  overflow-y: auto;
  padding-right: 4px;
}
.pool-item {
  text-align: center;
  cursor: pointer;
}
.pool-item.dim {
  opacity: 0.55;
}
.pool-icon {
  width: 50px;
  height: 50px;
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
  margin-top: 6px;
  font-size: 11.5px;
  color: var(--mute);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.lib-hint {
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid var(--line-soft);
  font-size: 12px;
  color: var(--faint);
}
</style>
