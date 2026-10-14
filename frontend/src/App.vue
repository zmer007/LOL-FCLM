<script setup lang="ts">
import { onMounted } from 'vue'
import AppNav from './components/AppNav.vue'
import { useUserStore } from './stores/user'

// 启动时若存在 token，拉取用户信息恢复登录态（失败会在 store 内清除）
const userStore = useUserStore()
onMounted(() => {
  userStore.loadMe()
})
</script>

<template>
  <AppNav />
  <router-view />
</template>
