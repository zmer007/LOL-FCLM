<template>
  <div class="opp-picker">
    <input
      v-model="keyword"
      type="text"
      class="opp-input"
      :class="{ invalid: touched && !modelValue }"
      placeholder="输入对位英雄名称 / 别名，从候选中选择"
      @focus="open = true"
      @input="onInput"
    />
    <div v-if="open" class="opp-dropdown">
      <div
        v-for="h in candidates"
        :key="h.id"
        class="opp-opt"
        :class="{ on: h.id === modelValue }"
        @mousedown.prevent="choose(h)"
      >
        <img :src="h.avatar" :alt="h.title" />
        <span>{{ h.title }}</span>
        <em>{{ h.name }}</em>
      </div>
      <div v-if="!candidates.length" class="opp-empty">无匹配英雄，请更换关键词</div>
    </div>
    <span v-if="selected" class="opp-sel">已选：{{ selected.title }}（{{ selected.name }}）</span>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Hero } from '../api/hero'

const props = defineProps<{
  heroes: Hero[]
  modelValue: number
  excludeId?: number
  touched?: boolean
}>()
const emit = defineEmits<{ (e: 'update:modelValue', id: number): void }>()

const keyword = ref('')
const open = ref(false)

const selected = computed(() => props.heroes.find((h) => h.id === props.modelValue))

const candidates = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  let list = props.heroes.filter((h) => h.id !== props.excludeId)
  if (q) {
    list = list.filter(
      (h) =>
        h.title.toLowerCase().includes(q) ||
        h.name.toLowerCase().includes(q) ||
        h.aliases.some((a) => a.toLowerCase().includes(q)),
    )
  }
  return list.slice(0, 20)
})

function onInput() {
  open.value = true
  // 输入变化与已选项不一致时清空选择，强制从候选重新选择
  if (props.modelValue && selected.value) {
    const q = keyword.value.trim()
    if (q && selected.value.title !== q) emit('update:modelValue', 0)
  }
}
function choose(h: Hero) {
  keyword.value = h.title
  open.value = false
  emit('update:modelValue', h.id)
}
</script>

<style scoped>
.opp-picker {
  position: relative;
}
.opp-input {
  width: 100%;
  font-family: inherit;
  font-size: 13.5px;
  background: #0a1520;
  border: 1px solid var(--line);
  border-radius: 8px;
  color: var(--gold-bright);
  padding: 10px 12px;
}
.opp-input:focus {
  outline: none;
  border-color: var(--gold);
}
.opp-input.invalid {
  border-color: var(--red);
}
.opp-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  max-height: 260px;
  overflow-y: auto;
  background: var(--panel-2, #132438);
  border: 1px solid var(--gold);
  border-radius: 8px;
  z-index: 40;
}
.opp-opt {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 12px;
  cursor: pointer;
}
.opp-opt:hover,
.opp-opt.on {
  background: rgba(200, 170, 110, 0.12);
}
.opp-opt img {
  width: 28px;
  height: 28px;
  border-radius: 6px;
}
.opp-opt span {
  font-size: 13px;
  color: var(--gold-bright);
}
.opp-opt em {
  font-style: normal;
  font-size: 11.5px;
  color: var(--faint);
}
.opp-empty {
  padding: 16px;
  text-align: center;
  font-size: 12.5px;
  color: var(--faint);
}
.opp-sel {
  display: inline-block;
  margin-top: 6px;
  font-size: 12px;
  color: var(--cyan);
}
</style>
