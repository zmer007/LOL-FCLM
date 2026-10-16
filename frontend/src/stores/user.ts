import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchMe, login as apiLogin, register as apiRegister, sendCode as apiSendCode, type UserVO } from '../api/auth'
import { TOKEN_KEY } from '../api/http'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem(TOKEN_KEY) || '')
  const user = ref<UserVO | null>(null)

  function setSession(t: string, u: UserVO) {
    token.value = t
    user.value = u
    localStorage.setItem(TOKEN_KEY, t)
  }

  async function login(account: string, password: string) {
    const data = await apiLogin(account, password)
    setSession(data.token, data.user)
  }

  async function register(username: string, email: string, password: string, code: string) {
    const data = await apiRegister(username, email, password, code)
    setSession(data.token, data.user)
  }

  async function sendCode(email: string) {
    return apiSendCode(email)
  }

  async function loadMe(): Promise<UserVO | null> {
    if (!token.value) return null
    if (user.value) return user.value
    try {
      const data = await fetchMe()
      user.value = data.user
      return data.user
    } catch {
      logout()
      return null
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
  }

  return { token, user, login, register, sendCode, loadMe, logout, setSession }
})
