<template>
  <div class="rank-picker">
    <button
      type="button"
      class="rank-trigger"
      :class="{ invalid: touched && !modelValue }"
      @click="open = !open"
    >
      <span :class="{ ph: !modelValue }">{{ modelValue || '选择段位（可选）' }}</span>
      <em class="caret">▾</em>
    </button>
    <div v-if="open" class="rank-dropdown">
      <div
        v-for="r in ranks"
        :key="r"
        class="rank-opt"
        :class="{ on: r === modelValue }"
        @mousedown.prevent="choose(r)"
      >
        {{ r }}
      </div>
      <div v-if="modelValue" class="rank-clear" @mousedown.prevent="choose('')">清除选择</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  modelValue: string
  touched?: boolean
}>()
const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>()

const open = ref(false)

const ranks = [
  '黑铁',
  '青铜',
  '白银',
  '黄金',
  '铂金',
  '翡翠',
  '钻石',
  '大师',
  '宗师',
  '王者',
]

function choose(r: string) {
  emit('update:modelValue', r)
  open.value = false
}
</script>

<style scoped>
.rank-picker {
  position: relative;
}
.rank-trigger {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-family: inherit;
  font-size: 13.5px;
  background: #0a1520;
  border: 1px solid var(--line);
  border-radius: 8px;
  color: var(--gold-bright);
  padding: 10px 12px;
  cursor: pointer;
}
.rank-trigger:focus {
  outline: none;
  border-color: var(--gold);
}
.rank-trigger.invalid {
  border-color: var(--red);
}
.rank-trigger .ph {
  color: var(--faint);
}
.caret {
  font-style: normal;
  color: var(--faint);
  font-size: 12px;
}
.rank-dropdown {
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
.rank-opt {
  padding: 9px 12px;
  font-size: 13px;
  color: var(--gold-bright);
  cursor: pointer;
}
.rank-opt:hover,
.rank-opt.on {
  background: rgba(200, 170, 110, 0.12);
}
.rank-clear {
  padding: 9px 12px;
  font-size: 12.5px;
  color: var(--faint);
  cursor: pointer;
  border-top: 1px solid var(--line-soft);
}
.rank-clear:hover {
  color: var(--red);
}
</style>
