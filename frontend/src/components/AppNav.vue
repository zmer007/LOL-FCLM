<script setup lang="ts">
import { computed } from 'vue'
import { useUserStore } from '../stores/user'

// 全站导航：首页 / 关于 + 右侧常驻「我的」头像。
// 浏览无需登录；未登录点击头像进入 /me，由路由守卫跳转登录页。
const userStore = useUserStore()
const isLoggedIn = computed(() => !!userStore.token)
</script>

<template>
  <header class="app-nav">
    <nav class="app-nav-links">
      <router-link to="/">首页</router-link>
      <router-link to="/about">关于</router-link>
    </nav>
    <router-link class="app-nav-avatar" to="/me" title="我的">
      {{ isLoggedIn ? (userStore.user?.username?.charAt(0) || '我') : '我' }}
    </router-link>
  </header>
</template>

<style scoped>
.app-nav {
  position: sticky;
  top: 0;
  z-index: 10;
  height: 62px;
  padding: 0 40px;
  display: flex;
  align-items: center;
  gap: 36px;
  border-bottom: 1px solid var(--line-soft);
  background: rgba(7, 14, 22, 0.92);
}
.app-nav-links {
  display: flex;
  gap: 28px;
  font-size: 14px;
}
.app-nav-links a {
  color: var(--mute);
  text-decoration: none;
}
.app-nav-links a.router-link-exact-active {
  color: var(--gold-bright);
}
.app-nav-avatar {
  margin-left: auto;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1px solid var(--gold);
  background: var(--panel);
  color: var(--gold);
  display: grid;
  place-items: center;
  font-size: 13px;
  text-decoration: none;
}
</style>
