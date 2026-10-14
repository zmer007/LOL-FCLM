<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useUserStore } from '../stores/user'
import {
  type AdminUser,
  type AssetCategory,
  type GameAsset,
  type Hero,
  addUser,
  createAsset,
  createHero,
  deleteAsset,
  deleteUser,
  listAssets,
  listHeroes,
  listUsers,
  resetUserPassword,
  setUserRole,
  setUserStatus,
  uploadAssetIcon,
  uploadHeroAvatar,
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

const LANES = ['上单', '打野', '中单', '下路', '辅助']
const SUBS: Record<Exclude<TabKey, 'hero' | 'users'>, string[]> = {
  item: ['攻击', '法术', '防御', '移动', '消耗'],
  skill: [],
  rune: ['精密', '主宰', '巫术', '坚决', '启迪'],
}

const heroes = ref<Hero[]>([])
const assets = ref<GameAsset[]>([])
const users = ref<AdminUser[]>([])
const loading = ref(false)
const busyId = ref<number | null>(null)
const errorMsg = ref('')

// 新增英雄表单
const heroName = ref('')
const heroLane = ref(LANES[0])
// 新增素材表单
const assetName = ref('')
const assetSub = ref('')
// 新增用户表单
const newUsername = ref('')
const newPassword = ref('')
const newRole = ref<'user' | 'admin'>('user')
const currentUserId = ref<number>(0)

onMounted(async () => {
  const u = await userStore.loadMe()
  isAdmin.value = u?.role === 'admin'
  currentUserId.value = u?.id ?? 0
  checking.value = false
  if (isAdmin.value) loadData()
})

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
  assetName.value = ''
  if (key !== 'hero' && key !== 'users') assetSub.value = SUBS[key][0] || ''
  loadData()
}

async function addHero() {
  errorMsg.value = ''
  const name = heroName.value.trim()
  if (!name) {
    errorMsg.value = '请输入英雄名称'
    return
  }
  try {
    await createHero(name, heroLane.value)
    heroName.value = ''
    await loadData()
  } catch (e) {
    errorMsg.value = (e as Error).message
  }
}

async function addAsset() {
  errorMsg.value = ''
  const name = assetName.value.trim()
  if (!name) {
    errorMsg.value = '请输入素材名称'
    return
  }
  const category = activeTab.value as AssetCategory
  const sub = category === 'skill' ? '召唤师技能' : assetSub.value
  try {
    await createAsset(category, sub, name)
    assetName.value = ''
    await loadData()
  } catch (e) {
    errorMsg.value = (e as Error).message
  }
}

async function onHeroFile(hero: Hero, file: File | undefined) {
  if (!file) return
  busyId.value = hero.id
  errorMsg.value = ''
  try {
    await uploadHeroAvatar(hero.id, file)
    await loadData()
  } catch (e) {
    errorMsg.value = (e as Error).message
  } finally {
    busyId.value = null
  }
}

async function onAssetFile(asset: GameAsset, file: File | undefined) {
  if (!file) return
  busyId.value = asset.id
  errorMsg.value = ''
  try {
    await uploadAssetIcon(asset.id, file)
    await loadData()
  } catch (e) {
    errorMsg.value = (e as Error).message
  } finally {
    busyId.value = null
  }
}

async function removeAsset(asset: GameAsset) {
  if (!window.confirm(`确定删除素材「${asset.name}」吗？`)) return
  errorMsg.value = ''
  try {
    await deleteAsset(asset.id)
    await loadData()
  } catch (e) {
    errorMsg.value = (e as Error).message
  }
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

      <!-- 英雄面板 -->
      <section v-else-if="activeTab === 'hero'" class="panel">
        <div class="create-row">
          <input v-model="heroName" class="txt" type="text" maxlength="20" placeholder="新英雄名称"
                 @keyup.enter="addHero" />
          <select v-model="heroLane" class="sel">
            <option v-for="l in LANES" :key="l" :value="l">{{ l }}</option>
          </select>
          <button class="btn" @click="addHero">新增英雄</button>
        </div>

        <div v-if="loading" class="admin-tip">加载中…</div>
        <div v-else-if="heroes.length === 0" class="admin-tip">还没有英雄，先新增一个吧</div>
        <div v-else class="grid">
          <div v-for="h in heroes" :key="h.id" class="cell">
            <label class="thumb hero-thumb" :class="{ busy: busyId === h.id }">
              <img v-if="h.avatar" :src="h.avatar" :alt="h.name" />
              <span v-else class="placeholder">点击上传</span>
              <input type="file" accept="image/png,image/jpeg,image/webp,image/gif,image/svg+xml"
                     @change="onHeroFile(h, ($event.target as HTMLInputElement).files?.[0])" />
            </label>
            <div class="cell-name">{{ h.name }}</div>
            <div class="cell-sub">{{ h.lane }}</div>
          </div>
        </div>
      </section>

      <!-- 素材面板 -->
      <section v-else class="panel">
        <div class="create-row">
          <input v-model="assetName" class="txt" type="text" maxlength="48" placeholder="新素材名称"
                 @keyup.enter="addAsset" />
          <select v-if="activeTab !== 'skill'" v-model="assetSub" class="sel">
            <option v-for="s in SUBS[activeTab as Exclude<TabKey, 'hero' | 'users'>]" :key="s" :value="s">{{ s }}</option>
          </select>
          <button class="btn" @click="addAsset">新增素材</button>
        </div>

        <div v-if="loading" class="admin-tip">加载中…</div>
        <div v-else-if="assets.length === 0" class="admin-tip">还没有素材，先新增一个吧</div>
        <div v-else class="grid">
          <div v-for="a in assets" :key="a.id" class="cell">
            <div class="asset-wrap">
              <label class="thumb asset-thumb" :class="{ busy: busyId === a.id }">
                <img v-if="a.icon" :src="a.icon" :alt="a.name" />
                <span v-else class="placeholder">上传图标</span>
                <input type="file" accept="image/png,image/jpeg,image/webp,image/gif,image/svg+xml"
                       @change="onAssetFile(a, ($event.target as HTMLInputElement).files?.[0])" />
              </label>
              <button class="cell-del" title="删除" @click="removeAsset(a)">×</button>
            </div>
            <div class="cell-name">{{ a.name }}</div>
            <div class="cell-sub">{{ a.subCategory || '—' }}</div>
          </div>
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
.btn:hover {
  background: #d8bc82;
}
.admin-error {
  font-size: 12.5px;
  color: var(--red);
  margin-bottom: 14px;
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
  position: relative;
  display: grid;
  place-items: center;
  margin: 0 auto;
  border-radius: 12px;
  border: 1px solid var(--line);
  background: linear-gradient(135deg, #1c3350, #132438);
  cursor: pointer;
  overflow: hidden;
  transition: border-color 0.15s;
}
.thumb:hover {
  border-color: var(--gold);
}
.thumb.busy {
  opacity: 0.5;
  pointer-events: none;
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
.thumb input {
  display: none;
}
.asset-wrap {
  position: relative;
  width: 56px;
  margin: 0 auto;
}
.cell-del {
  position: absolute;
  top: -8px;
  right: -10px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 1px solid var(--red);
  background: var(--panel);
  color: var(--red);
  font-size: 13px;
  line-height: 1;
  cursor: pointer;
  display: grid;
  place-items: center;
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
