<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { listMyFavorites, toggleFavorite } from '../api/favorite'
import type { Content } from '../api/content'

const myFavorites = ref<Content[]>([])
const loading = ref(true)

function detailPath(c: Content) {
  return `/${c.kind === 'video' ? 'video' : 'article'}/${c.id}`
}

async function loadFavorites() {
  loading.value = true
  try {
    const res = await listMyFavorites()
    myFavorites.value = res.list
  } catch (e: any) {
    alert(e?.message || '加载收藏失败')
  } finally {
    loading.value = false
  }
}

async function removeFavorite(c: Content) {
  try {
    await toggleFavorite(c.id)
    myFavorites.value = myFavorites.value.filter((f) => f.id !== c.id)
  } catch (e: any) {
    alert(e?.message || '取消收藏失败')
  }
}

onMounted(loadFavorites)
</script>

<template>
  <main class="favorites">
    <div class="fav-head">
      <RouterLink to="/me" class="fav-back">← 返回</RouterLink>
      <h1 class="fav-title">我的收藏</h1>
    </div>

    <div v-if="loading" class="fav-tip">加载中…</div>
    <div v-else-if="!myFavorites.length" class="fav-tip">
      还没有收藏内容，去对线页收藏喜欢的攻略吧
    </div>
    <div v-else class="fav-list">
      <div v-for="c in myFavorites" :key="c.id" class="fav-row">
        <RouterLink :to="detailPath(c)" class="fav-main">
          <span class="fav-kind" :class="c.kind">{{ c.kind === 'video' ? '视频' : '图文' }}</span>
          <span class="fav-title-text">{{ c.title }}</span>
          <span class="fav-vs">{{ c.heroTitle }} vs {{ c.opponentTitle }}</span>
        </RouterLink>
        <div class="fav-right">
          <span class="fav-time">{{ c.updatedAt }}</span>
          <button type="button" class="fav-unfav" @click="removeFavorite(c)">取消收藏</button>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
.favorites {
  max-width: 1080px;
  margin: 0 auto;
  padding: 28px 24px 80px;
}
.fav-head {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}
.fav-back {
  font-size: 13px;
  color: var(--mute);
  text-decoration: none;
}
.fav-back:hover {
  color: var(--gold);
}
.fav-title {
  font-size: 22px;
  color: var(--gold-bright);
  margin: 0;
}
.fav-tip {
  font-size: 13px;
  color: var(--faint);
  background: var(--panel);
  border: 1px dashed var(--line);
  border-radius: 12px;
  padding: 48px;
  text-align: center;
}
.fav-list {
  background: var(--panel);
  border: 1px solid var(--line-soft);
  border-radius: 14px;
  overflow: hidden;
}
.fav-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 14px 18px;
  border-bottom: 1px solid var(--line-soft);
}
.fav-row:last-child {
  border-bottom: none;
}
.fav-main {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  text-decoration: none;
  flex: 1;
}
.fav-kind {
  flex-shrink: 0;
  font-size: 11px;
  border-radius: 4px;
  padding: 2px 8px;
  border: 1px solid var(--line);
  color: var(--faint);
}
.fav-kind.video {
  color: var(--cyan);
  border-color: rgba(10, 200, 185, 0.5);
}
.fav-title-text {
  font-size: 13.5px;
  color: var(--gold-bright);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.fav-vs {
  flex-shrink: 0;
  font-size: 11.5px;
  color: var(--faint);
}
.fav-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}
.fav-time {
  font-size: 11.5px;
  color: var(--faint);
}
.fav-unfav {
  font-family: inherit;
  font-size: 12px;
  color: var(--faint);
  background: none;
  border: 1px solid var(--line);
  border-radius: 999px;
  padding: 2px 12px;
  cursor: pointer;
}
.fav-unfav:hover {
  color: var(--red);
  border-color: var(--red);
}
</style>
