<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const keyword = ref('')

function onSearch() {
  // 与设计稿一致：即使未输入也进入对线页（参数可为空）
  router.push({ path: '/matchup', query: { champion: keyword.value.trim() } })
}
</script>

<template>
  <main class="center">
    <div class="big-logo">
      <span class="logo-left">路边</span>
      <!-- 盾牌笑脸 logo：盾形 + 弯月笑眼 + 微笑 -->
      <svg class="logo-svg" viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M24 4 L40 10 V24 C40 34 33 41 24 44 C15 41 8 34 8 24 V10 Z"
              fill="none" stroke="#C8AA6E" stroke-width="2.5" stroke-linejoin="round" />
        <!-- 弯月笑眼（左金右青，暗合对线双方） -->
        <path d="M14 21 Q18 16 22 21" stroke="#F0E6D2" stroke-width="2.8" stroke-linecap="round" />
        <path d="M26 21 Q30 16 34 21" stroke="#0AC8B9" stroke-width="2.8" stroke-linecap="round" />
        <!-- 微笑 -->
        <path d="M16 30 Q24 38 32 30" stroke="#C8AA6E" stroke-width="2.8" stroke-linecap="round" />
      </svg>
      <span class="logo-right"><i>联盟</i></span>
    </div>

    <form class="search" @submit.prevent="onSearch">
      <span class="icon">🔍</span>
      <input v-model="keyword" type="text" placeholder="输入英雄名称" autofocus />
      <button type="submit">进入</button>
    </form>
  </main>
</template>

<style scoped>
/* 搜索区偏上，类似百度；高度扣除 62px 导航栏 */
.center {
  min-height: calc(100vh - 62px);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 14vh 24px 24px;
}

.big-logo {
  font-size: 42px;
  font-weight: 800;
  letter-spacing: 6px;
  color: var(--gold-bright);
  margin-bottom: 34px;
  display: flex;
  align-items: center;
  gap: 10px;
}
/* 视觉居中：两侧字符字面留白不同，非对称补偿 */
.logo-left {
  margin-right: -8px;
}
.logo-right {
  margin-left: -2px;
}
.big-logo i {
  font-style: normal;
  color: var(--gold);
}
.logo-svg {
  width: 46px;
  height: 46px;
  flex-shrink: 0;
}

.search {
  width: 100%;
  max-width: 520px;
  height: 58px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 10px 0 22px;
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 12px;
  transition: border-color 0.15s;
}
.search:focus-within {
  border-color: var(--gold);
}
.search .icon {
  color: var(--faint);
  font-size: 17px;
}
.search input {
  flex: 1;
  min-width: 0;
  background: none;
  border: 0;
  outline: none;
  color: var(--gold-bright);
  font-size: 16px;
  font-family: inherit;
}
.search input::placeholder {
  color: var(--faint);
}
.search button {
  height: 42px;
  padding: 0 26px;
  border: 0;
  border-radius: 9px;
  background: var(--gold);
  color: #0a1428;
  font-size: 14px;
  font-weight: 700;
  font-family: inherit;
  cursor: pointer;
}
.search button:hover {
  background: var(--gold-bright);
}
</style>
