<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'

const props = defineProps<{ mode: 'login' | 'register' }>()

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const account = ref('')
const email = ref('')
const password = ref('')
const code = ref('')
const errorMsg = ref('')
const submitting = ref(false)
const sendingCode = ref(false)
const countdown = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

const isLogin = computed(() => props.mode === 'login')
const title = computed(() => (isLogin.value ? '登录' : '注册'))
const submitText = computed(() => (isLogin.value ? '登录' : '注册并登录'))
const codeBtnText = computed(() =>
  countdown.value > 0 ? `${countdown.value}s 后重发` : sendingCode.value ? '发送中…' : '发送验证码'
)

function startCountdown() {
  countdown.value = 60
  timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0 && timer) {
      clearInterval(timer)
      timer = null
    }
  }, 1000)
}

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

async function onSendCode() {
  if (sendingCode.value || countdown.value > 0) return
  if (!email.value) {
    errorMsg.value = '请先填写邮箱'
    return
  }
  errorMsg.value = ''
  sendingCode.value = true
  try {
    await userStore.sendCode(email.value)
    startCountdown()
  } catch (e) {
    errorMsg.value = (e as Error).message
  } finally {
    sendingCode.value = false
  }
}

async function onSubmit() {
  errorMsg.value = ''
  submitting.value = true
  try {
    if (isLogin.value) {
      await userStore.login(account.value, password.value)
    } else {
      await userStore.register(account.value, email.value, password.value, code.value)
    }
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    router.replace(redirect)
  } catch (e) {
    errorMsg.value = (e as Error).message
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <main class="auth">
    <form class="auth-card" @submit.prevent="onSubmit">
      <h1 class="auth-title">{{ title }}</h1>

      <label class="auth-field">
        <span>{{ isLogin ? '用户名 / 邮箱' : '用户名' }}</span>
        <input v-model="account" type="text" maxlength="20" autocomplete="username"
               placeholder="3-20 个字符" />
      </label>

      <label v-if="!isLogin" class="auth-field">
        <span>邮箱</span>
        <input v-model="email" type="email" maxlength="128" autocomplete="email"
               placeholder="用于接收验证码" />
      </label>

      <label class="auth-field">
        <span>密码</span>
        <input v-model="password" type="password" maxlength="32"
               :autocomplete="isLogin ? 'current-password' : 'new-password'"
               placeholder="6-32 个字符" />
      </label>

      <div v-if="!isLogin" class="auth-field auth-code-row">
        <label class="auth-code-input">
          <span>验证码</span>
          <input v-model="code" type="text" maxlength="6" inputmode="numeric"
                 placeholder="6 位数字" />
        </label>
        <button type="button" class="auth-code-btn" :disabled="sendingCode || countdown > 0"
                @click="onSendCode">
          {{ codeBtnText }}
        </button>
      </div>

      <p v-if="errorMsg" class="auth-error">{{ errorMsg }}</p>

      <button class="auth-submit" type="submit" :disabled="submitting">
        {{ submitting ? '请稍候…' : submitText }}
      </button>

      <p v-if="isLogin" class="auth-switch">
        还没有账号？<router-link to="/register">立即注册</router-link>
      </p>
      <p v-else class="auth-switch">
        已有账号？<router-link to="/login">去登录</router-link>
      </p>
    </form>
  </main>
</template>

<style scoped>
.auth {
  display: flex;
  justify-content: center;
  padding-top: 10vh;
}
.auth-card {
  width: 360px;
  background: var(--panel);
  border: 1px solid var(--line-soft);
  border-radius: 16px;
  padding: 32px 30px 26px;
}
.auth-title {
  font-size: 20px;
  color: var(--gold-bright);
  text-align: center;
  margin-bottom: 24px;
  letter-spacing: 4px;
}
.auth-field {
  display: block;
  margin-bottom: 18px;
}
.auth-field span {
  display: block;
  font-size: 12.5px;
  color: var(--mute);
  margin-bottom: 7px;
}
.auth-field input {
  width: 100%;
  height: 40px;
  padding: 0 12px;
  font-size: 14px;
  font-family: inherit;
  color: var(--gold-bright);
  background: var(--bg);
  border: 1px solid var(--line);
  border-radius: 8px;
  outline: none;
}
.auth-field input:focus {
  border-color: var(--gold);
}
.auth-code-row {
  display: flex;
  gap: 10px;
  align-items: flex-end;
}
.auth-code-input {
  flex: 1;
  margin-bottom: 0;
}
.auth-code-btn {
  height: 40px;
  padding: 0 14px;
  font-size: 12.5px;
  font-family: inherit;
  color: var(--gold-bright);
  background: transparent;
  border: 1px solid var(--gold);
  border-radius: 8px;
  cursor: pointer;
  white-space: nowrap;
}
.auth-code-btn:hover:not(:disabled) {
  background: rgba(212, 175, 55, 0.12);
}
.auth-code-btn:disabled {
  opacity: 0.5;
  cursor: default;
}
.auth-error {
  font-size: 12.5px;
  color: var(--red);
  margin-bottom: 12px;
}
.auth-submit {
  width: 100%;
  height: 42px;
  margin-top: 4px;
  font-size: 14px;
  font-family: inherit;
  letter-spacing: 4px;
  color: #0a1520;
  background: var(--gold);
  border: none;
  border-radius: 999px;
  cursor: pointer;
}
.auth-submit:hover {
  background: #d8bc82;
}
.auth-submit:disabled {
  opacity: 0.6;
  cursor: default;
}
.auth-switch {
  margin-top: 18px;
  text-align: center;
  font-size: 12.5px;
  color: var(--faint);
}
.auth-switch a {
  color: var(--gold);
  text-decoration: none;
}
</style>
