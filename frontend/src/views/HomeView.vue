<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { type Hero, listHeroes } from '../api/hero'

const router = useRouter()
const keyword = ref('')
const heroes = ref<Hero[]>([])

onMounted(async () => {
  try {
    heroes.value = (await listHeroes()).list
  } catch {
    // 英雄列表加载失败不影响搜索
  }
})

// 关键字过滤（短名/称号/别名）；默认无输入时不展示任何英雄
const filteredHeroes = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  if (!q) return []
  return heroes.value.filter(
    (h) =>
      h.title.toLowerCase().includes(q) ||
      h.name.toLowerCase().includes(q) ||
      h.aliases.some((a) => a.toLowerCase().includes(q)),
  )
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

// 点击英雄名进入该英雄的对线页；优先用短名（阿狸），其次全名
function enterHero(h: Hero) {
  router.push({ path: '/matchup', query: { champion: h.title || h.name } })
}

// 复制 QQ 群号，成功后短暂显示「已复制」
const copied = ref('')
async function copyText(text: string, key: string) {
  try {
    await navigator.clipboard.writeText(text)
  } catch {
    // 剪贴板 API 不可用时降级为临时 textarea 复制
    const ta = document.createElement('textarea')
    ta.value = text
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    ta.remove()
  }
  copied.value = key
  setTimeout(() => {
    if (copied.value === key) copied.value = ''
  }, 1500)
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

    <!-- 英雄结果区：固定预留两排空间避免页面跳动，默认无匹配时为空；纯文字展示，头像加载慢不用 -->
    <div class="hero-wall">
      <button
        v-for="h in filteredHeroes"
        :key="h.id"
        type="button"
        class="hero-cell"
        :title="h.name"
        @click="enterHero(h)"
      >
        <span class="hero-name">{{ h.title || h.name }}</span>
      </button>
    </div>

    <!-- 站点交流联系方式 -->
    <div class="contact">
      <p>
        站点 QQ 交流群：962481115
        <button
          type="button"
          class="copy-btn"
          :title="copied === 'site' ? '已复制' : '复制'"
          @click="copyText('962481115', 'site')"
        >
          <!-- 复制成功时切换为对勾图标 -->
          <svg v-if="copied === 'site'" viewBox="0 0 16 16" class="copy-icon">
            <path d="M3 8.5 L6.5 12 L13 4.5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
          <svg v-else viewBox="0 0 16 16" class="copy-icon">
            <rect x="5.5" y="5.5" width="8" height="8" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.6" />
            <path d="M10.5 3.5 H3.5 A1.5 1.5 0 0 0 2 5 V11.5" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
          </svg>
        </button>
      </p>
      <p>
        抖音花田错，QQ群：657601118
        <button
          type="button"
          class="copy-btn"
          :title="copied === 'douyin' ? '已复制' : '复制'"
          @click="copyText('657601118', 'douyin')"
        >
          <svg v-if="copied === 'douyin'" viewBox="0 0 16 16" class="copy-icon">
            <path d="M3 8.5 L6.5 12 L13 4.5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
          <svg v-else viewBox="0 0 16 16" class="copy-icon">
            <rect x="5.5" y="5.5" width="8" height="8" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.6" />
            <path d="M10.5 3.5 H3.5 A1.5 1.5 0 0 0 2 5 V11.5" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
          </svg>
        </button>
      </p>
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

/* 英雄名称墙：固定两排高度预留空间（默认空也占位，避免输入时页面跳动），超出滚动 */
.hero-wall {
  width: 100%;
  max-width: 708px;
  margin-top: 26px;
  display: grid;
  grid-template-columns: repeat(auto-fill, 56px);
  grid-auto-rows: min-content;
  justify-content: space-between;
  gap: 16px 8px;
  height: 140px;
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
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  border: 1px solid var(--line);
  background: linear-gradient(135deg, #1c3350, #132438);
  transition: border-color 0.15s, transform 0.15s;
}
.hero-cell:hover {
  border-color: var(--gold);
  transform: translateY(-2px);
}
/* 名称文字：居中显示，过长自动换行 */
.hero-name {
  padding: 0 2px;
  font-size: 12px;
  line-height: 1.25;
  text-align: center;
  color: var(--mute);
  transition: color 0.15s;
}
.hero-cell:hover .hero-name {
  color: var(--gold-bright);
}

/* 站点联系方式：头像墙下方朴素纯文本 */
.contact {
  margin-top: 18px;
  text-align: center;
}
.contact p {
  margin: 4px 0;
  font-size: 13px;
  color: var(--mute);
}
/* 复制按钮：与文本同排的极小图标按钮 */
.copy-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  vertical-align: -2px;
  width: 18px;
  height: 18px;
  padding: 0;
  margin-left: 6px;
  color: var(--faint);
  background: none;
  border: 0;
  border-radius: 4px;
  cursor: pointer;
  transition: color 0.15s;
}
.copy-btn:hover {
  color: var(--gold-bright);
}
.copy-icon {
  width: 13px;
  height: 13px;
}
</style>
