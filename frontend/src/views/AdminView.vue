<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useUserStore } from '../stores/user'
import {
  type AdminUser,
  type AssetCategory,
  type GameAsset,
  type Hero,
  type SyncState,
  addUser,
  deleteUser,
  getSyncStatus,
  listAssets,
  listHeroes,
  listUsers,
  resetUserPassword,
  setUserRole,
  setUserStatus,
  triggerSync,
} from '../api/admin'

const userStore = useUserStore()
const isAdmin = ref(false)
const checking = ref(true)

type TabKey = 'users' | 'hero' | AssetCategory
const tabs: { key: TabKey; label: string }[] = [
  { key: 'users', label: '用户管理' },
  { key: 'hero', label: '英雄头像' },
  { key: 'item', label: '装备图标' },
  { key: 'skill', label: '召唤师技能' },
  { key: 'rune', label: '符文图标' },
]
const activeTab = ref<TabKey>('users')

// 装备子分类（与后端同步归类一致）
const ITEM_SUBS = ['攻击', '法术', '防御', '移动', '消耗', '其他']

const heroes = ref<Hero[]>([])
const assets = ref<GameAsset[]>([])
const users = ref<AdminUser[]>([])
const loading = ref(false)
const errorMsg = ref('')

// 用户管理表单
const newUsername = ref('')
const newPassword = ref('')
const newRole = ref<'user' | 'admin'>('user')
const currentUserId = ref<number>(0)

// Data Dragon 同步状态
const sync = ref<SyncState | null>(null)
const itemSubFilter = ref('全部')
let pollTimer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  const u = await userStore.loadMe()
  isAdmin.value = u?.role === 'admin'
  currentUserId.value = u?.id ?? 0
  checking.value = false
  if (isAdmin.value) {
    loadData()
    refreshSync()
  }
})

onBeforeUnmount(() => stopPolling())

async function loadData() {
  loading.value = true
  errorMsg.value = ''
  try {
    if (activeTab.value === 'hero') {
      heroes.value = (await listHeroes()).list
    } else if (activeTab.value === 'users') {
      users.value = (await listUsers()).list
    } else {
      assets.value = (await listAssets(activeTab.value as AssetCategory)).list
    }
  } catch (e) {
    errorMsg.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

function switchTab(key: TabKey) {
  activeTab.value = key
  errorMsg.value = ''
  itemSubFilter.value = '全部'
  loadData()
}

const visibleAssets = () =>
  itemSubFilter.value === '全部'
    ? assets.value
    : assets.value.filter((a) => a.subCategory === itemSubFilter.value)

// ---------- Data Dragon 同步 ----------

async function refreshSync() {
  try {
    sync.value = await getSyncStatus()
    if (sync.value.running) startPolling()
  } catch {
    // 状态拉取失败不打扰用户
  }
}

function startPolling() {
  if (pollTimer) return
  pollTimer = setInterval(async () => {
    try {
      sync.value = await getSyncStatus()
      if (!sync.value.running) {
        stopPolling()
        await loadData() // 同步结束后刷新当前网格
      }
    } catch {
      stopPolling()
    }
  }, 2000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

async function startSync() {
  errorMsg.value = ''
  try {
    sync.value = await triggerSync(false)
    startPolling()
  } catch (e) {
    errorMsg.value = (e as Error).message
  }
}

const progressText = () => {
  if (!sync.value) return ''
  const s = sync.value
  if (s.running && s.stageTotal > 0) {
    return `${s.stage} ${s.stageDone}/${s.stageTotal}`
  }
  if (s.running) return s.stage
  return s.message
}

// ---------- 用户管理 ----------

async function addNewUser() {
  errorMsg.value = ''
  const username = newUsername.value.trim()
  if (!username) {
    errorMsg.value = '请输入用户名'
    return
  }
  if (!newPassword.value) {
    errorMsg.value = '请输入初始密码'
    return
  }
  try {
    await addUser(username, newPassword.value, newRole.value)
    newUsername.value = ''
    newPassword.value = ''
    newRole.value = 'user'
    await loadData()
  } catch (e) {
    errorMsg.value = (e as Error).message
  }
}

async function toggleUserStatus(u: AdminUser) {
  errorMsg.value = ''
  const next = u.status === 'active' ? 'disabled' : 'active'
  if (next === 'disabled' && !window.confirm(`确定禁用用户「${u.username}」吗？禁用后该账号立即无法使用。`)) return
  try {
    await setUserStatus(u.id, next)
    await loadData()
  } catch (e) {
    errorMsg.value = (e as Error).message
  }
}

async function changeUserRole(u: AdminUser, role: 'user' | 'admin') {
  if (role === u.role) return
  errorMsg.value = ''
  try {
    await setUserRole(u.id, role)
    await loadData()
  } catch (e) {
    errorMsg.value = (e as Error).message
  }
}

async function resetPassword(u: AdminUser) {
  const pwd = window.prompt(`为「${u.username}」设置新密码（6-32 个字符）：`)
  if (pwd === null) return
  errorMsg.value = ''
  try {
    await resetUserPassword(u.id, pwd)
    window.alert('密码已重置')
  } catch (e) {
    errorMsg.value = (e as Error).message
  }
}

async function removeUser(u: AdminUser) {
  if (!window.confirm(`确定删除用户「${u.username}」吗？该操作不可恢复。`)) return
  errorMsg.value = ''
  try {
    await deleteUser(u.id)
    await loadData()
  } catch (e) {
    errorMsg.value = (e as Error).message
  }
}
</script>

<template>
  <main class="admin">
    <h1 class="admin-title">管理后台</h1>

    <p v-if="checking" class="admin-tip">权限校验中…</p>
    <section v-else-if="!isAdmin" class="admin-denied">
      <p>403 · 需要管理员权限</p>
      <router-link to="/">返回首页</router-link>
    </section>

    <template v-else>
      <div class="tabs">
        <button
          v-for="t in tabs"
          :key="t.key"
          class="tab"
          :class="{ on: activeTab === t.key }"
          @click="switchTab(t.key)"
        >
          {{ t.label }}
        </button>
      </div>

      <p v-if="errorMsg" class="admin-error">{{ errorMsg }}</p>

      <!-- 用户管理面板 -->
      <section v-if="activeTab === 'users'" class="panel">
        <div class="create-row">
          <input v-model="newUsername" class="txt" type="text" maxlength="20" placeholder="用户名（3-20 字符）"
                 @keyup.enter="addNewUser" />
          <input v-model="newPassword" class="txt" type="text" maxlength="32" placeholder="初始密码（6-32 字符）"
                 @keyup.enter="addNewUser" />
          <select v-model="newRole" class="sel">
            <option value="user">普通用户</option>
            <option value="admin">管理员</option>
          </select>
          <button class="btn" @click="addNewUser">添加用户</button>
        </div>

        <div v-if="loading" class="admin-tip">加载中…</div>
        <table v-else class="utable">
          <thead>
            <tr>
              <th>ID</th>
              <th>用户名</th>
              <th>角色</th>
              <th>状态</th>
              <th>注册时间</th>
              <th class="col-actions">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in users" :key="u.id">
              <td>{{ u.id }}</td>
              <td class="u-name">{{ u.username }}</td>
              <td>
                <select
                  class="mini-sel"
                  :value="u.role"
                  :disabled="u.id === currentUserId"
                  @change="changeUserRole(u, ($event.target as HTMLSelectElement).value as 'user' | 'admin')"
                >
                  <option value="user">普通用户</option>
                  <option value="admin">管理员</option>
                </select>
              </td>
              <td>
                <span class="badge" :class="u.status === 'active' ? 'badge-active' : 'badge-disabled'">
                  {{ u.status === 'active' ? '正常' : '已禁用' }}
                </span>
              </td>
              <td class="u-time">{{ new Date(u.createdAt).toLocaleString('zh-CN', { hour12: false }) }}</td>
              <td class="col-actions">
                <button
                  class="link-btn"
                  :class="u.status === 'active' ? 'text-warn' : 'text-ok'"
                  :disabled="u.id === currentUserId"
                  @click="toggleUserStatus(u)"
                >
                  {{ u.status === 'active' ? '禁用' : '启用' }}
                </button>
                <button class="link-btn" @click="resetPassword(u)">重置密码</button>
                <button
                  class="link-btn text-danger"
                  :disabled="u.id === currentUserId"
                  @click="removeUser(u)"
                >
                  删除
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <p v-if="!loading && users.length === 0" class="admin-tip">还没有用户</p>
      </section>

      <!-- 数据面板（英雄 / 装备 / 召唤师技能 / 符文，均由 Data Dragon 同步） -->
      <section v-else class="panel">
        <div class="sync-bar">
          <div class="sync-info">
            <span class="sync-version">
              数据版本：{{ sync?.version ? sync.version : '未同步' }}
            </span>
            <span v-if="sync?.finishedAt" class="sync-time">最近更新：{{ sync.finishedAt }}</span>
            <span v-if="sync" class="sync-msg" :class="'msg-' + sync.status">{{ progressText() }}</span>
          </div>
          <button class="btn" :disabled="sync?.running" @click="startSync">
            {{ sync?.running ? '更新中…' : '更新数据' }}
          </button>
        </div>

        <!-- 装备子分类筛选 -->
        <div v-if="activeTab === 'item'" class="subfilters">
          <button
            class="chip"
            :class="{ on: itemSubFilter === '全部' }"
            @click="itemSubFilter = '全部'"
          >全部</button>
          <button
            v-for="s in ITEM_SUBS"
            :key="s"
            class="chip"
            :class="{ on: itemSubFilter === s }"
            @click="itemSubFilter = s"
          >{{ s }}</button>
        </div>

        <div v-if="loading" class="admin-tip">加载中…</div>

        <!-- 英雄只读网格 -->
        <div v-else-if="activeTab === 'hero'" class="grid">
          <div v-for="h in heroes" :key="h.id" class="cell">
            <div class="thumb hero-thumb">
              <img v-if="h.avatar" :src="h.avatar" :alt="h.name" />
              <span v-else class="placeholder">暂无图标</span>
            </div>
            <div class="cell-name">{{ h.name }}</div>
            <div class="cell-sub">{{ h.title || '—' }}</div>
          </div>
          <p v-if="heroes.length === 0" class="admin-tip full-row">尚未同步，点击右上角「更新数据」</p>
        </div>

        <!-- 素材只读网格 -->
        <div v-else class="grid">
          <div v-for="a in visibleAssets()" :key="a.id" class="cell">
            <div class="thumb asset-thumb">
              <img v-if="a.icon" :src="a.icon" :alt="a.name" />
              <span v-else class="placeholder">暂无图标</span>
            </div>
            <div class="cell-name">{{ a.name }}</div>
            <div class="cell-sub">{{ a.subCategory || '—' }}</div>
          </div>
          <p v-if="visibleAssets().length === 0" class="admin-tip full-row">
            {{ assets.length === 0 ? '尚未同步，点击右上角「更新数据」' : '该分类下暂无素材' }}
          </p>
        </div>
      </section>
    </template>
  </main>
</template>

<style scoped>
.admin {
  max-width: 900px;
  margin: 0 auto;
  padding: 32px 24px 60px;
}
.admin-title {
  font-size: 20px;
  color: var(--gold-bright);
  margin-bottom: 22px;
  display: flex;
  align-items: center;
  gap: 10px;
}
.admin-title::before {
  content: '';
  width: 3px;
  height: 16px;
  background: var(--gold);
  border-radius: 2px;
}
.admin-tip {
  color: var(--faint);
  font-size: 13px;
  padding: 20px 0;
}
.full-row {
  grid-column: 1 / -1;
  text-align: center;
}
.admin-denied {
  background: var(--panel);
  border: 1px solid var(--line-soft);
  border-radius: 16px;
  padding: 48px 24px;
  text-align: center;
}
.admin-denied p {
  color: var(--red);
  font-size: 15px;
  margin-bottom: 14px;
}
.admin-denied a {
  color: var(--gold);
  font-size: 13px;
  text-decoration: none;
}

.tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 22px;
  flex-wrap: wrap;
}
.tab {
  font-family: inherit;
  font-size: 12.5px;
  cursor: pointer;
  background: none;
  border: 1px solid var(--line);
  border-radius: 999px;
  color: var(--faint);
  padding: 6px 20px;
}
.tab:hover {
  color: var(--mute);
}
.tab.on {
  border-color: var(--gold);
  color: var(--gold);
}

.panel {
  background: var(--panel);
  border: 1px solid var(--line-soft);
  border-radius: 16px;
  padding: 22px 24px 28px;
}
.create-row {
  display: flex;
  gap: 10px;
  margin-bottom: 24px;
}
.txt {
  flex: 1;
  height: 38px;
  padding: 0 12px;
  font-size: 13.5px;
  font-family: inherit;
  color: var(--gold-bright);
  background: var(--bg);
  border: 1px solid var(--line);
  border-radius: 8px;
  outline: none;
}
.txt:focus {
  border-color: var(--gold);
}
.sel {
  height: 38px;
  padding: 0 10px;
  font-size: 13.5px;
  font-family: inherit;
  color: var(--gold-bright);
  background: var(--bg);
  border: 1px solid var(--line);
  border-radius: 8px;
  outline: none;
}
.btn {
  height: 38px;
  padding: 0 20px;
  font-size: 13.5px;
  font-family: inherit;
  color: #0a1520;
  background: var(--gold);
  border: none;
  border-radius: 8px;
  font-weight: 700;
  cursor: pointer;
}
.btn:hover:not(:disabled) {
  background: #d8bc82;
}
.btn:disabled {
  opacity: 0.6;
  cursor: default;
}
.admin-error {
  font-size: 12.5px;
  color: var(--red);
  margin-bottom: 14px;
}

/* Data Dragon 同步条 */
.sync-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 16px;
  margin-bottom: 20px;
  background: var(--bg);
  border: 1px solid var(--line);
  border-radius: 10px;
}
.sync-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}
.sync-version {
  font-size: 13px;
  color: var(--gold-bright);
}
.sync-time {
  font-size: 11.5px;
  color: var(--faint);
}
.sync-msg {
  font-size: 12px;
  color: var(--mute);
}
.sync-msg.msg-success {
  color: var(--cyan);
}
.sync-msg.msg-skipped {
  color: var(--faint);
}
.sync-msg.msg-failed {
  color: var(--red);
}

.subfilters {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 20px;
}
.chip {
  font-family: inherit;
  font-size: 12px;
  cursor: pointer;
  background: none;
  border: 1px solid var(--line);
  border-radius: 999px;
  color: var(--faint);
  padding: 4px 14px;
}
.chip.on {
  border-color: var(--gold);
  color: var(--gold);
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(96px, 1fr));
  gap: 22px 14px;
}
.cell {
  text-align: center;
}
.thumb {
  display: grid;
  place-items: center;
  margin: 0 auto;
  border-radius: 12px;
  border: 1px solid var(--line);
  background: linear-gradient(135deg, #1c3350, #132438);
  overflow: hidden;
}
.hero-thumb {
  width: 72px;
  height: 72px;
}
.asset-thumb {
  width: 56px;
  height: 56px;
  border-radius: 10px;
}
.thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.placeholder {
  font-size: 11px;
  color: var(--faint);
  padding: 0 6px;
}
.cell-name {
  margin-top: 8px;
  font-size: 12.5px;
  color: var(--gold-bright);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.cell-sub {
  margin-top: 3px;
  font-size: 11px;
  color: var(--faint);
}

/* 用户管理表格 */
.utable {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.utable th {
  text-align: left;
  font-weight: 500;
  font-size: 12px;
  color: var(--faint);
  padding: 8px 10px;
  border-bottom: 1px solid var(--line);
  white-space: nowrap;
}
.utable td {
  padding: 11px 10px;
  border-bottom: 1px solid var(--line-soft);
  color: var(--gold-bright);
  white-space: nowrap;
}
.utable .u-name {
  font-weight: 500;
}
.utable .u-time {
  font-size: 12px;
  color: var(--faint);
}
.col-actions {
  text-align: right;
}
.mini-sel {
  height: 30px;
  padding: 0 8px;
  font-size: 12.5px;
  font-family: inherit;
  color: var(--gold-bright);
  background: var(--bg);
  border: 1px solid var(--line);
  border-radius: 6px;
  outline: none;
}
.mini-sel:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.badge {
  display: inline-block;
  font-size: 11.5px;
  border-radius: 999px;
  padding: 2px 12px;
}
.badge-active {
  color: var(--cyan);
  border: 1px solid rgba(10, 200, 185, 0.5);
}
.badge-disabled {
  color: var(--faint);
  border: 1px solid var(--line);
}
.link-btn {
  font-family: inherit;
  font-size: 12.5px;
  background: none;
  border: none;
  padding: 2px 8px;
  cursor: pointer;
  color: var(--mute);
}
.link-btn:hover:not(:disabled) {
  color: var(--gold-bright);
  text-decoration: underline;
}
.link-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
.link-btn.text-warn {
  color: #d89a4a;
}
.link-btn.text-ok {
  color: var(--cyan);
}
.link-btn.text-danger {
  color: var(--red);
}
</style>
