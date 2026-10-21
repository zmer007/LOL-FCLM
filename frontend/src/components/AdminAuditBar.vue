<template>
  <div v-if="visible" class="audit-bar">
    <div class="audit-info">
      <span class="audit-title">管理员审核</span>
      <span class="audit-state" :class="status">
        {{ status === 'pending' ? '该内容正在等待审核' : '该内容已被拒绝' }}
      </span>
      <span v-if="status === 'rejected' && rejectReason" class="audit-reason">
        拒绝理由：{{ rejectReason }}
      </span>
    </div>
    <div class="audit-ops">
      <button type="button" class="btn-approve" :disabled="busy" @click="onApprove">✓ 通过</button>
      <button type="button" class="btn-reject" :disabled="busy" @click="onReject">× 拒绝</button>
    </div>
    <p v-if="error" class="audit-error">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { approveContent, rejectContent, type ContentStatus } from '../api/content'
import { useUserStore } from '../stores/user'

const props = defineProps<{
  contentId: number
  status: ContentStatus
  rejectReason?: string
}>()
const emit = defineEmits<{ done: [] }>()

const userStore = useUserStore()
const isAdmin = computed(() => userStore.user?.role === 'admin')
const visible = computed(() => isAdmin.value && props.status !== 'approved')

const busy = ref(false)
const error = ref('')

async function onApprove() {
  if (!window.confirm('确认通过该内容？通过后将立即对外展示当前版本。')) return
  error.value = ''
  busy.value = true
  try {
    await approveContent(props.contentId)
    emit('done')
  } catch (e) {
    error.value = e instanceof Error ? e.message : '操作失败'
  } finally {
    busy.value = false
  }
}

async function onReject() {
  const reason = window.prompt('请填写拒绝理由（将通知作者）：', props.rejectReason || '')
  if (reason === null) return
  if (!reason.trim()) {
    error.value = '拒绝理由不能为空'
    return
  }
  error.value = ''
  busy.value = true
  try {
    await rejectContent(props.contentId, reason.trim())
    emit('done')
  } catch (e) {
    error.value = e instanceof Error ? e.message : '操作失败'
  } finally {
    busy.value = false
  }
}
</script>

<style scoped>
.audit-bar {
  border: 1px solid var(--gold);
  background: rgba(200, 170, 110, 0.08);
  border-radius: 12px;
  padding: 14px 18px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}
.audit-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  font-size: 13px;
}
.audit-title {
  font-weight: 700;
  color: var(--gold-bright);
  border: 1px solid var(--gold);
  border-radius: 999px;
  padding: 2px 12px;
  font-size: 12px;
}
.audit-state {
  color: var(--gold);
}
.audit-state.rejected {
  color: var(--red);
}
.audit-reason {
  color: var(--red);
  font-size: 12.5px;
}
.audit-ops {
  display: flex;
  gap: 10px;
}
.btn-approve,
.btn-reject {
  font-family: inherit;
  font-size: 13px;
  border-radius: 8px;
  padding: 8px 22px;
  cursor: pointer;
}
.btn-approve {
  background: var(--gold);
  color: #0a1428;
  border: 1px solid var(--gold);
  font-weight: 700;
}
.btn-reject {
  background: transparent;
  color: var(--red);
  border: 1px solid var(--red);
}
.btn-approve:disabled,
.btn-reject:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
.audit-error {
  width: 100%;
  margin: 4px 0 0;
  font-size: 12.5px;
  color: var(--red);
}
</style>
