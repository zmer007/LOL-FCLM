import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchMe, login as apiLogin, register as apiRegister, type UserVO } from '../api/auth'
import { TOKEN_KEY } from '../api/http'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem(TOKEN_KEY) || '')
  const user = ref<UserVO | null>(null)

  function setSession(t: string, u: UserVO) {
    token.value = t
    user.value = u
    localStorage.setItem(TOKEN_KEY, t)
  }

  async function login(username: string, password: string) {
    const data = await apiLogin(username, password)
    setSession(data.token, data.user)
  }

  async function register(username: string, password: string) {
    const data = await apiRegister(username, password)
    setSession(data.token, data.user)
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

  return { token, user, login, register, loadMe, logout, setSession }
})
