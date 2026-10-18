<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { listMyContents, type Content } from '../api/content'
import { sendMessage, listMyMessages, type Message } from '../api/message'
import { listMyFavorites } from '../api/favorite'

const router = useRouter()
const userStore = useUserStore()
const isAdmin = computed(() => userStore.user?.role === 'admin')

const myContents = ref<Content[]>([])
const loading = ref(true)

// 收藏数量
const favCount = ref(0)

const pendingList = computed(() => myContents.value.filter((c) => c.status === 'pending'))
const rejectedList = computed(() => myContents.value.filter((c) => c.status === 'rejected'))
const reviewing = computed(() => [...pendingList.value, ...rejectedList.value])

// 留言功能
const msgContent = ref('')
const msgSubmitting = ref(false)
const myMessages = ref<Message[]>([])
const msgLoading = ref(false)

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

async function submitMessage() {
  const text = msgContent.value.trim()
  if (!text) return
  if (text.length > 500) {
    alert('留言内容不能超过 500 字')
    return
  }
  msgSubmitting.value = true
  try {
    await sendMessage(text)
    msgContent.value = ''
    await loadMessages()
  } catch (e: any) {
    alert(e?.message || '留言提交失败')
  } finally {
    msgSubmitting.value = false
  }
}

async function loadMessages() {
  msgLoading.value = true
  try {
    const res = await listMyMessages()
    myMessages.value = res.list
  } finally {
    msgLoading.value = false
  }
}

async function loadFavorites() {
  try {
    const res = await listMyFavorites()
    favCount.value = res.list.length
  } catch {
    favCount.value = 0
  }
}

onMounted(async () => {
  try {
    await userStore.loadMe()
    if (isAdmin.value) {
      myContents.value = (await listMyContents()).list
    }
    await Promise.all([loadMessages(), loadFavorites()])
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
      <div class="me-stats">
        <RouterLink to="/favorites" class="me-stat">
          <b>{{ favCount }}</b>
          <span>收藏</span>
        </RouterLink>
      </div>
      <div class="me-actions">
        <RouterLink v-if="userStore.user?.role === 'admin'" class="me-admin" to="/admin">
          管理后台
        </RouterLink>
        <button class="me-logout" type="button" @click="onLogout">退出登录</button>
      </div>
    </section>

    <!-- 审批中（待审核 + 被拒绝）—— 仅管理员 -->
    <section v-if="isAdmin && reviewing.length" class="block">
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

    <!-- 我的投稿（全部）—— 仅管理员 -->
    <section v-if="isAdmin" class="block">
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

    <!-- 给管理员留言 -->
    <section class="block">
      <h2 class="block-title">给管理员留言</h2>
      <div class="msg-box">
        <textarea
          v-model="msgContent"
          class="msg-textarea"
          placeholder="有什么想对管理员说的？（最多 500 字，留言仅管理员可见）"
          maxlength="500"
          rows="4"
        />
        <div class="msg-foot">
          <span class="msg-count">{{ msgContent.length }}/500</span>
          <button
            type="button"
            class="msg-submit"
            :disabled="!msgContent.trim() || msgSubmitting"
            @click="submitMessage"
          >{{ msgSubmitting ? '提交中…' : '提交留言' }}</button>
        </div>
      </div>
      <div class="msg-history">
        <h3 class="msg-history-title">我的留言记录</h3>
        <div v-if="msgLoading" class="c-tip">加载中…</div>
        <div v-else-if="!myMessages.length" class="c-tip">暂无留言</div>
        <div v-else class="msg-list">
          <div v-for="m in myMessages" :key="m.id" class="msg-item">
            <div class="msg-item-content">{{ m.content }}</div>
            <div class="msg-item-meta">
              <span class="msg-item-status" :class="{ read: m.isRead }">
                {{ m.isRead ? '已读' : '未读' }}
              </span>
              <span class="msg-item-time">{{ m.createdAt }}</span>
            </div>
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
  flex-shrink: 0;
}
.me-stats {
  display: flex;
  flex-shrink: 0;
}
.me-stat {
  text-align: center;
  padding: 0 22px;
  border-left: 1px solid var(--line-soft);
  text-decoration: none;
  color: inherit;
}
.me-stat b {
  display: block;
  font-size: 15px;
  color: var(--gold-bright);
  font-weight: 600;
  transition: color 0.15s;
}
.me-stat:hover b {
  color: var(--gold);
}
.me-stat span {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: var(--faint);
  transition: color 0.15s;
}
.me-stat:hover span {
  color: var(--mute);
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

/* 留言 */
.msg-box {
  background: var(--panel);
  border: 1px solid var(--line-soft);
  border-radius: 14px;
  padding: 16px;
}
.msg-textarea {
  width: 100%;
  box-sizing: border-box;
  background: var(--panel-2);
  border: 1px solid var(--line);
  border-radius: 10px;
  padding: 12px 14px;
  color: var(--text);
  font-family: inherit;
  font-size: 13.5px;
  line-height: 1.6;
  resize: vertical;
}
.msg-textarea:focus {
  outline: none;
  border-color: var(--gold);
}
.msg-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 10px;
}
.msg-count {
  font-size: 12px;
  color: var(--faint);
}
.msg-submit {
  font-family: inherit;
  font-size: 13px;
  color: #0a1428;
  background: var(--gold);
  border: none;
  border-radius: 999px;
  padding: 7px 22px;
  cursor: pointer;
  font-weight: 700;
}
.msg-submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.msg-history {
  margin-top: 22px;
}
.msg-history-title {
  font-size: 14px;
  color: var(--gold-bright);
  margin: 0 0 12px;
}
.msg-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.msg-item {
  background: var(--panel);
  border: 1px solid var(--line-soft);
  border-radius: 12px;
  padding: 12px 16px;
}
.msg-item-content {
  font-size: 13.5px;
  color: var(--text);
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}
.msg-item-meta {
  margin-top: 8px;
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 11.5px;
  color: var(--faint);
}
.msg-item-status {
  border: 1px solid var(--line);
  border-radius: 999px;
  padding: 1px 10px;
  color: var(--faint);
}
.msg-item-status.read {
  color: var(--cyan);
  border-color: rgba(10, 200, 185, 0.5);
}
</style>
