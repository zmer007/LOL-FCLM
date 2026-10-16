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
import {
  approveContent,
  listReviewContents,
  rejectContent,
  type Content,
} from '../api/content'
import { getAssetMap } from '../api/asset'
import { getRuneTrees, type RuneTree, type Shard } from '../api/build'
import ItemSlots from '../components/build/ItemSlots.vue'
import RunePage from '../components/build/RunePage.vue'

const userStore = useUserStore()
const isAdmin = ref(false)
const checking = ref(true)

type TabKey = 'review' | 'users' | 'hero' | AssetCategory
const tabs: { key: TabKey; label: string }[] = [
  { key: 'review', label: '内容审批' },
  { key: 'users', label: '用户管理' },
  { key: 'hero', label: '英雄头像' },
  { key: 'item', label: '装备图标' },
  { key: 'skill', label: '召唤师技能' },
  { key: 'rune', label: '符文图标' },
]
const activeTab = ref<TabKey>('review')

// 内容审批
const reviewList = ref<Content[]>([])
const reviewStatus = ref<'pending' | 'rejected'>('pending')
const reviewPage = ref(1)
const reviewTotalPages = ref(1)
const reviewPreview = ref<Content | null>(null)
const reviewAssetMap = ref<Map<number, GameAsset>>(new Map())
const reviewTrees = ref<RuneTree[]>([])
const reviewShards = ref<Shard[][]>([])
let reviewMetaLoaded = false

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
    if (activeTab.value === 'review') {
      await loadReview()
    } else if (activeTab.value === 'hero') {
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

// ---------- 内容审批 ----------

async function loadReview() {
  const res = await listReviewContents(reviewStatus.value, reviewPage.value)
  reviewList.value = res.list
  reviewTotalPages.value = Math.max(1, Math.ceil(res.total / res.size))
}

function switchReviewStatus(s: 'pending' | 'rejected') {
  reviewStatus.value = s
  reviewPage.value = 1
  loadData()
}

async function openPreview(c: Content) {
  reviewPreview.value = c
  if (!reviewMetaLoaded) {
    reviewMetaLoaded = true
    try {
      const [assets, runeData] = await Promise.all([getAssetMap(), getRuneTrees()])
      reviewAssetMap.value = assets
      reviewTrees.value = runeData.trees
      reviewShards.value = runeData.shards
    } catch {
      reviewMetaLoaded = false
    }
  }
}

async function doApprove(c: Content) {
  errorMsg.value = ''
  try {
    await approveContent(c.id)
    reviewPreview.value = null
    await loadData()
  } catch (e) {
    errorMsg.value = (e as Error).message
  }
}

async function doReject(c: Content) {
  const reason = window.prompt('请填写拒绝理由（将通知投稿者）：', c.rejectReason || '')
  if (reason === null) return
  if (!reason.trim()) {
    errorMsg.value = '拒绝理由不能为空'
    return
  }
  errorMsg.value = ''
  try {
    await rejectContent(c.id, reason.trim())
    reviewPreview.value = null
    await loadData()
  } catch (e) {
    errorMsg.value = (e as Error).message
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

      <!-- 内容审批面板 -->
      <section v-if="activeTab === 'review'" class="panel">
        <div class="review-sub">
          <button
            class="chip"
            :class="{ on: reviewStatus === 'pending' }"
            @click="switchReviewStatus('pending')"
          >待审核</button>
          <button
            class="chip"
            :class="{ on: reviewStatus === 'rejected' }"
            @click="switchReviewStatus('rejected')"
          >已拒绝</button>
        </div>

        <div v-if="loading" class="admin-tip">加载中…</div>
        <table v-else-if="reviewList.length" class="utable">
          <thead>
            <tr>
              <th>类型</th>
              <th>标题</th>
              <th>投稿者</th>
              <th>对线</th>
              <th>提交时间</th>
              <th class="col-actions">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="c in reviewList" :key="c.id">
              <td>{{ c.kind === 'video' ? '视频' : '图文' }}</td>
              <td class="u-name">
                {{ c.title }}
                <span v-if="c.status === 'rejected' && c.rejectReason" class="reject-reason">
                  （{{ c.rejectReason }}）
                </span>
              </td>
              <td>{{ c.authorName }}</td>
              <td class="u-time">{{ c.heroTitle }} vs {{ c.opponentTitle }}</td>
              <td class="u-time">{{ c.createdAt }}</td>
              <td class="col-actions">
                <button class="link-btn" @click="openPreview(c)">预览</button>
                <button v-if="c.status === 'pending'" class="link-btn text-ok" @click="doApprove(c)">通过</button>
                <button class="link-btn text-danger" @click="doReject(c)">拒绝</button>
              </td>
            </tr>
          </tbody>
        </table>
        <p v-else class="admin-tip">
          {{ reviewStatus === 'pending' ? '没有待审核的内容' : '没有被拒绝的内容' }}
        </p>

        <div v-if="reviewTotalPages > 1" class="pager">
          <button :disabled="reviewPage === 1" @click="reviewPage--; loadData()">上一页</button>
          <button
            v-for="p in reviewTotalPages"
            :key="p"
            :class="{ on: p === reviewPage }"
            @click="reviewPage = p; loadData()"
          >{{ p }}</button>
          <button :disabled="reviewPage === reviewTotalPages" @click="reviewPage++; loadData()">下一页</button>
        </div>
      </section>

      <!-- 审批预览弹层 -->
      <div v-if="reviewPreview" class="preview-mask" @click.self="reviewPreview = null">
        <div class="preview-dialog">
          <div class="preview-head">
            <span>内容预览 · {{ reviewPreview.kind === 'video' ? '视频' : '图文' }}</span>
            <button class="preview-close" @click="reviewPreview = null">×</button>
          </div>
          <div class="preview-body">
            <h3 class="pv-title">{{ reviewPreview.title }}</h3>
            <div class="pv-meta">
              {{ reviewPreview.authorName }} · {{ reviewPreview.heroTitle }} vs {{ reviewPreview.opponentTitle }}
              · {{ reviewPreview.createdAt }}
            </div>

            <video
              v-if="reviewPreview.kind === 'video' && reviewPreview.videoUrl"
              class="pv-video"
              :src="reviewPreview.videoUrl"
              controls
              controlslist="nodownload"
              disablepictureinpicture
              preload="metadata"
              @contextmenu.prevent
            ></video>

            <p v-if="reviewPreview.body" class="pv-text">{{ reviewPreview.body }}</p>

            <div v-if="reviewPreview.images.length" class="pv-imgs">
              <a v-for="(img, i) in reviewPreview.images" :key="i" :href="img" target="_blank">
                <img :src="img" :alt="`配图 ${i + 1}`" />
              </a>
            </div>

            <div v-if="!reviewPreview.items.start.every((s) => s === null) || reviewPreview.runes.primary" class="pv-build">
              <div class="pv-block-title">推荐装备</div>
              <ItemSlots :items="reviewPreview.items" :asset-map="reviewAssetMap" />
              <div class="pv-block-title pv-block-2">推荐符文</div>
              <RunePage :runes="reviewPreview.runes" :trees="reviewTrees" :shards="reviewShards" />
              <div v-if="reviewPreview.buildNote" class="pv-block-title pv-block-2">装备符文思路</div>
              <p v-if="reviewPreview.buildNote" class="pv-note">{{ reviewPreview.buildNote }}</p>
            </div>
          </div>
          <div v-if="reviewPreview.status === 'pending'" class="preview-foot">
            <button class="pv-btn reject" @click="doReject(reviewPreview)">拒绝</button>
            <button class="pv-btn approve" @click="doApprove(reviewPreview)">通过</button>
          </div>
        </div>
      </div>

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
.reject-reason {
  color: var(--red);
  font-size: 11.5px;
}

/* 审批子筛选 / 分页 */
.review-sub {
  display: flex;
  gap: 8px;
  margin-bottom: 18px;
}
.pager {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-top: 22px;
}
.pager button {
  font-family: inherit;
  font-size: 12.5px;
  cursor: pointer;
  background: none;
  border: 1px solid var(--line);
  border-radius: 8px;
  color: var(--faint);
  min-width: 30px;
  height: 28px;
  padding: 0 8px;
}
.pager button.on {
  border-color: var(--gold);
  color: var(--gold);
}
.pager button:disabled {
  opacity: 0.35;
  cursor: default;
}

/* 审批预览弹层 */
.preview-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  display: grid;
  place-items: center;
  z-index: 90;
  padding: 24px;
}
.preview-dialog {
  width: min(720px, 94vw);
  max-height: 86vh;
  display: flex;
  flex-direction: column;
  background: var(--panel-2);
  border: 1px solid var(--gold);
  border-radius: 14px;
  overflow: hidden;
}
.preview-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  font-size: 14px;
  font-weight: 700;
  color: var(--gold-bright);
  border-bottom: 1px solid var(--line);
}
.preview-close {
  background: none;
  border: none;
  color: var(--faint);
  font-size: 20px;
  cursor: pointer;
  line-height: 1;
}
.preview-close:hover {
  color: var(--gold);
}
.preview-body {
  padding: 18px 20px;
  overflow-y: auto;
}
.pv-title {
  margin: 0 0 8px;
  font-size: 17px;
  color: var(--gold-bright);
}
.pv-meta {
  font-size: 12px;
  color: var(--faint);
  margin-bottom: 14px;
}
.pv-video {
  width: 100%;
  aspect-ratio: 16 / 9;
  background: #000;
  border-radius: 10px;
  margin-bottom: 14px;
}
.pv-text {
  font-size: 13.5px;
  color: var(--mute);
  line-height: 1.9;
  white-space: pre-wrap;
  margin: 0 0 14px;
}
.pv-imgs {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-bottom: 14px;
}
.pv-imgs a {
  display: block;
  aspect-ratio: 16 / 9;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--line);
}
.pv-imgs img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.pv-build {
  border-top: 1px solid var(--line);
  padding-top: 14px;
}
.pv-block-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--gold-bright);
  margin-bottom: 12px;
}
.pv-block-2 {
  margin-top: 20px;
}
.pv-note {
  margin: 0;
  font-size: 13px;
  line-height: 1.7;
  color: var(--mute);
  white-space: pre-wrap;
  word-break: break-word;
}
.preview-foot {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 14px 20px;
  border-top: 1px solid var(--line);
}
.pv-btn {
  font-family: inherit;
  font-size: 13px;
  border-radius: 8px;
  padding: 8px 26px;
  cursor: pointer;
}
.pv-btn.reject {
  background: none;
  border: 1px solid var(--red);
  color: var(--red);
}
.pv-btn.approve {
  background: var(--gold);
  border: 1px solid var(--gold);
  color: #0a1428;
  font-weight: 700;
}
</style>
