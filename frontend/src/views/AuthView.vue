<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const account = ref('')
const password = ref('')
const errorMsg = ref('')
const submitting = ref(false)

async function onSubmit() {
  errorMsg.value = ''
  submitting.value = true
  try {
    await userStore.login(account.value, password.value)
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
      <h1 class="auth-title">登录</h1>

      <label class="auth-field">
        <span>用户名 / 邮箱</span>
        <input v-model="account" type="text" maxlength="20" autocomplete="username"
               placeholder="请输入账号" />
      </label>

      <label class="auth-field">
        <span>密码</span>
        <input v-model="password" type="password" maxlength="32"
               autocomplete="current-password" placeholder="6-32 个字符" />
      </label>

      <p v-if="errorMsg" class="auth-error">{{ errorMsg }}</p>

      <button class="auth-submit" type="submit" :disabled="submitting">
        {{ submitting ? '请稍候…' : '登录' }}
      </button>

      <p class="auth-switch">本站采用邀请制，账号由管理员创建</p>
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
