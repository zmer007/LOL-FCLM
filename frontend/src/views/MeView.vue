<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { listMyContents, type Content } from '../api/content'

const router = useRouter()
const userStore = useUserStore()
const isAdmin = computed(() => userStore.user?.role === 'admin')

const myContents = ref<Content[]>([])
const loading = ref(true)

const pendingList = computed(() => myContents.value.filter((c) => c.status === 'pending'))
const rejectedList = computed(() => myContents.value.filter((c) => c.status === 'rejected'))
const reviewing = computed(() => [...pendingList.value, ...rejectedList.value])

function editPath(c: Content) {
  return { path: `/upload/${c.kind === 'video' ? 'video' : 'article'}`, query: { id: c.id } }
}
function detailPath(c: Content) {
  return `/${c.kind === 'video' ? 'video' : 'article'}/${c.id}`
}
function statusLabel(s: Content['status']) {
  return s === 'approved' ? '已通过' : s === 'pending' ? '审核中' : '已拒绝'
}

// 展开/收起拒绝理由（一次只展开一条）
const openReasonId = ref<number | null>(null)
function toggleReason(id: number) {
  openReasonId.value = openReasonId.value === id ? null : id
}

onMounted(async () => {
  try {
    await userStore.loadMe()
    myContents.value = (await listMyContents()).list
  } finally {
    loading.value = false
  }
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
      <div class="me-actions">
        <RouterLink v-if="userStore.user?.role === 'admin'" class="me-admin" to="/admin">
          管理后台
        </RouterLink>
        <button class="me-logout" type="button" @click="onLogout">退出登录</button>
      </div>
    </section>

    <!-- 审批中（待审核 + 被拒绝） -->
    <section v-if="reviewing.length" class="block">
      <h2 class="block-title">审批中</h2>
      <div class="c-list">
        <div v-for="c in reviewing" :key="c.id" class="c-row">
          <RouterLink :to="editPath(c)" class="c-main">
            <span class="c-kind" :class="c.kind">{{ c.kind === 'video' ? '视频' : '图文' }}</span>
            <span class="c-title">{{ c.title }}</span>
            <span class="c-vs">{{ c.heroTitle }} vs {{ c.opponentTitle }}</span>
          </RouterLink>
          <div class="c-right">
            <span class="c-status" :class="c.status">{{ statusLabel(c.status) }}</span>
            <button
              v-if="c.status === 'rejected' && c.rejectReason"
              type="button"
              class="c-reason"
              :class="{ open: openReasonId === c.id }"
              :title="openReasonId === c.id ? '点击收起' : '点击查看完整理由'"
              @click="toggleReason(c.id)"
            >{{ c.rejectReason }}</button>
            <RouterLink v-if="isAdmin" :to="detailPath(c)" class="c-audit">审核</RouterLink>
            <RouterLink :to="editPath(c)" class="c-edit">编辑</RouterLink>
          </div>
        </div>
      </div>
    </section>

    <!-- 我的投稿（全部） -->
    <section class="block">
      <h2 class="block-title">我的投稿</h2>
      <div v-if="loading" class="c-tip">加载中…</div>
      <div v-else-if="!myContents.length" class="c-tip">还没有投稿，去英雄对线页上传视频或发布图文吧</div>
      <div v-else class="c-list">
        <div v-for="c in myContents" :key="c.id" class="c-row">
          <RouterLink :to="detailPath(c)" class="c-main">
            <span class="c-kind" :class="c.kind">{{ c.kind === 'video' ? '视频' : '图文' }}</span>
            <span class="c-title">{{ c.title }}</span>
            <span class="c-vs">{{ c.heroTitle }} vs {{ c.opponentTitle }}</span>
          </RouterLink>
          <div class="c-right">
            <span class="c-status" :class="c.status">{{ statusLabel(c.status) }}</span>
            <button
              v-if="c.status === 'rejected' && c.rejectReason"
              type="button"
              class="c-reason"
              :class="{ open: openReasonId === c.id }"
              :title="openReasonId === c.id ? '点击收起' : '点击查看完整理由'"
              @click="toggleReason(c.id)"
            >{{ c.rejectReason }}</button>
            <span class="c-time">{{ c.updatedAt }}</span>
            <RouterLink :to="editPath(c)" class="c-edit">编辑</RouterLink>
          </div>
        </div>
      </div>
    </section>
  </main>
</template>

<style scoped>
.me {
  max-width: 820px;
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
.me-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}
.me-admin {
  font-size: 12.5px;
  color: var(--gold);
  text-decoration: none;
  border: 1px solid var(--gold);
  border-radius: 999px;
  padding: 6px 18px;
}
.me-admin:hover {
  background: rgba(200, 170, 110, 0.12);
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

/* 内容区块 */
.block {
  margin-top: 30px;
}
.block-title {
  font-size: 16px;
  color: var(--gold-bright);
  margin: 0 0 14px;
  display: flex;
  align-items: center;
  gap: 10px;
}
.block-title::before {
  content: '';
  width: 3px;
  height: 14px;
  background: var(--gold);
  border-radius: 2px;
}
.c-tip {
  font-size: 13px;
  color: var(--faint);
  background: var(--panel);
  border: 1px dashed var(--line);
  border-radius: 12px;
  padding: 28px;
  text-align: center;
}
.c-list {
  background: var(--panel);
  border: 1px solid var(--line-soft);
  border-radius: 14px;
  overflow: hidden;
}
.c-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 14px 18px;
  border-bottom: 1px solid var(--line-soft);
}
.c-row:last-child {
  border-bottom: none;
}
.c-main {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  text-decoration: none;
  flex: 1;
}
.c-kind {
  flex-shrink: 0;
  font-size: 11px;
  border-radius: 4px;
  padding: 2px 8px;
  border: 1px solid var(--line);
  color: var(--faint);
}
.c-kind.video {
  color: var(--cyan);
  border-color: rgba(10, 200, 185, 0.5);
}
.c-title {
  font-size: 13.5px;
  color: var(--gold-bright);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.c-vs {
  flex-shrink: 0;
  font-size: 11.5px;
  color: var(--faint);
}
.c-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}
.c-status {
  font-size: 11.5px;
  border-radius: 999px;
  padding: 2px 12px;
}
.c-status.approved {
  color: var(--cyan);
  border: 1px solid rgba(10, 200, 185, 0.5);
}
.c-status.pending {
  color: var(--gold);
  border: 1px solid var(--gold);
}
.c-status.rejected {
  color: var(--red);
  border: 1px solid var(--red);
}
.c-right {
  flex-wrap: wrap;
}
.c-reason {
  font-family: inherit;
  font-size: 11.5px;
  color: var(--red);
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  text-align: left;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.c-reason:hover {
  text-decoration: underline;
}
.c-reason.open {
  flex-basis: 100%;
  max-width: none;
  white-space: normal;
  text-align: right;
  line-height: 1.6;
}
.c-time {
  font-size: 11.5px;
  color: var(--faint);
}
.c-audit {
  font-size: 12px;
  color: #0a1428;
  background: var(--gold);
  border-radius: 999px;
  padding: 2px 12px;
  text-decoration: none;
  font-weight: 700;
}
.c-edit {
  font-size: 12px;
  color: var(--gold);
  text-decoration: none;
}
.c-edit:hover {
  text-decoration: underline;
}
</style>
