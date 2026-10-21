<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { type Hero, listHeroes } from '../api/hero'

const router = useRouter()
const keyword = ref('')
const heroes = ref<Hero[]>([])

// 官方定位筛选项（顺序与 Data Dragon 六大定位一致），'' 表示全部
const ROLE_FILTERS = ['', '战士', '坦克', '刺客', '法师', '射手', '辅助'] as const
const activeRole = ref<(typeof ROLE_FILTERS)[number]>('')

onMounted(async () => {
  try {
    heroes.value = (await listHeroes()).list
  } catch {
    // 头像墙加载失败不影响搜索
  }
})

// 关键字（短名/称号/别名）+ 官方定位双重过滤
const filteredHeroes = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  return heroes.value.filter((h) => {
    if (activeRole.value && !h.roles.includes(activeRole.value)) return false
    if (!q) return true
    return (
      h.title.toLowerCase().includes(q) ||
      h.name.toLowerCase().includes(q) ||
      h.aliases.some((a) => a.toLowerCase().includes(q))
    )
  })
})

function onSearch() {
  // 回车：有精确/首个匹配时直接进入，无输入则进入空对线页
  const q = keyword.value.trim()
  const target = q ? filteredHeroes.value[0] : undefined
  router.push({
    path: '/matchup',
    query: { champion: target ? target.title || target.name : q },
  })
}

// 点击头像进入该英雄的对线页；优先用短名（阿狸），其次全名
function enterHero(h: Hero) {
  router.push({ path: '/matchup', query: { champion: h.title || h.name } })
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
      <span class="logo-right"><i>小站</i></span>
    </div>

    <form class="search" @submit.prevent="onSearch">
      <span class="icon">🔍</span>
      <input v-model="keyword" type="text" placeholder="输入英雄名称" autofocus />
    </form>

    <!-- 官方定位筛选：默认全部，点选后与关键字叠加过滤 -->
    <div class="role-bar">
      <button
        v-for="r in ROLE_FILTERS"
        :key="r"
        type="button"
        class="role-chip"
        :class="{ active: activeRole === r }"
        @click="activeRole = r"
      >
        {{ r || '全部' }}
      </button>
    </div>

    <!-- 英雄头像墙：随输入/定位实时过滤，固定三排高度，其余滚动查看 -->
    <div v-if="filteredHeroes.length" class="hero-wall">
      <button
        v-for="h in filteredHeroes"
        :key="h.id"
        type="button"
        class="hero-cell"
        :title="h.name"
        @click="enterHero(h)"
      >
        <span class="hero-avatar">
          <img v-if="h.avatar" :src="h.avatar" :alt="h.title || h.name" loading="lazy" />
        </span>
      </button>
    </div>
  </main>
</template>

<style scoped>
/* 搜索区偏上，类似百度；高度扣除 62px 导航栏 */
.center {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 10vh 24px 40px;
}

.big-logo {
  font-size: 42px;
  font-weight: 800;
  letter-spacing: 6px;
  color: var(--gold-bright);
  margin-bottom: 30px;
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
  padding: 0 22px;
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

/* 官方定位筛选条 */
.role-bar {
  margin-top: 22px;
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px;
}
.role-chip {
  font-family: inherit;
  font-size: 13.5px;
  color: var(--mute);
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 999px;
  padding: 6px 18px;
  cursor: pointer;
  transition: all 0.15s;
}
.role-chip:hover {
  color: var(--gold-bright);
  border-color: var(--gold);
}
.role-chip.active {
  color: #0a1428;
  background: var(--gold);
  border-color: var(--gold);
  font-weight: 700;
}

/* 英雄头像墙：三排固定高度（56*3 + 间距 16*2），纵向滚动 */
.hero-wall {
  width: 100%;
  max-width: 860px;
  margin-top: 26px;
  display: grid;
  grid-template-columns: repeat(auto-fill, 56px);
  grid-auto-rows: min-content;
  justify-content: space-between;
  gap: 16px 8px;
  max-height: 208px;
  overflow-y: auto;
  padding: 4px 10px 8px 2px;
  scrollbar-width: thin;
  scrollbar-color: var(--line) transparent;
}
.hero-wall::-webkit-scrollbar {
  width: 6px;
}
.hero-wall::-webkit-scrollbar-thumb {
  background: var(--line);
  border-radius: 3px;
}
.hero-cell {
  font-family: inherit;
  background: none;
  border: 0;
  padding: 0;
  cursor: pointer;
  width: 56px;
}
.hero-avatar {
  display: block;
  width: 56px;
  height: 56px;
  border-radius: 10px;
  border: 1px solid var(--line);
  background: linear-gradient(135deg, #1c3350, #132438);
  overflow: hidden;
  transition: border-color 0.15s, transform 0.15s;
}
.hero-cell:hover .hero-avatar {
  border-color: var(--gold);
  transform: translateY(-2px);
}
.hero-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
</style>
