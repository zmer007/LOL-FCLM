<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'

const props = defineProps<{ mode: 'login' | 'register' }>()

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const username = ref('')
const password = ref('')
const errorMsg = ref('')
const submitting = ref(false)

const isLogin = computed(() => props.mode === 'login')
const title = computed(() => (isLogin.value ? '登录' : '注册'))
const submitText = computed(() => (isLogin.value ? '登录' : '注册并登录'))

async function onSubmit() {
  errorMsg.value = ''
  submitting.value = true
  try {
    if (isLogin.value) {
      await userStore.login(username.value, password.value)
    } else {
      await userStore.register(username.value, password.value)
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
        <span>用户名</span>
        <input v-model="username" type="text" maxlength="20" autocomplete="username"
               placeholder="3-20 个字符" />
      </label>

      <label class="auth-field">
        <span>密码</span>
        <input v-model="password" type="password" maxlength="32"
               :autocomplete="isLogin ? 'current-password' : 'new-password'"
               placeholder="6-32 个字符" />
      </label>

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
