<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()

onMounted(() => {
  userStore.loadMe()
})

function onLogout() {
  userStore.logout()
  router.replace('/')
}
</script>

<template>
  <main class="me">
    <section class="me-card">
      <div class="me-avatar">{{ userStore.user?.username?.charAt(0) || '?' }}</div>
      <div class="me-info">
        <div class="me-name">{{ userStore.user?.username || '…' }}</div>
        <div class="me-meta">
          <span class="me-role">{{ userStore.user?.role === 'admin' ? '管理员' : '普通用户' }}</span>
          <span v-if="userStore.user">
            注册于 {{ new Date(userStore.user.createdAt).toLocaleDateString('zh-CN') }}
          </span>
        </div>
      </div>
      <button class="me-logout" type="button" @click="onLogout">退出登录</button>
    </section>
  </main>
</template>

<style scoped>
.me {
  max-width: 720px;
  margin: 0 auto;
  padding: 40px 24px;
}
.me-card {
  display: flex;
  align-items: center;
  gap: 18px;
  background: var(--panel);
  border: 1px solid var(--line-soft);
  border-radius: 16px;
  padding: 26px 28px;
}
.me-avatar {
  width: 64px;
  height: 64px;
  flex-shrink: 0;
  border-radius: 50%;
  border: 1px solid var(--gold);
  background: var(--panel-2);
  color: var(--gold);
  display: grid;
  place-items: center;
  font-size: 24px;
}
.me-info {
  flex: 1;
}
.me-name {
  font-size: 18px;
  color: var(--gold-bright);
}
.me-meta {
  margin-top: 6px;
  display: flex;
  gap: 12px;
  align-items: center;
  font-size: 12.5px;
  color: var(--faint);
}
.me-role {
  color: var(--gold);
  border: 1px solid var(--gold);
  border-radius: 999px;
  padding: 1px 10px;
  font-size: 11.5px;
}
.me-logout {
  font-family: inherit;
  font-size: 12.5px;
  color: var(--mute);
  background: none;
  border: 1px solid var(--line);
  border-radius: 999px;
  padding: 6px 18px;
  cursor: pointer;
}
.me-logout:hover {
  color: var(--red);
  border-color: var(--red);
}
</style>
