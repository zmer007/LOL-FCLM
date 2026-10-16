import { TOKEN_KEY } from '../api/http'

// 判断是否已登录（与路由守卫使用同一真相源：localStorage token）
export function isLoggedIn(): boolean {
  return !!localStorage.getItem(TOKEN_KEY)
}

// 未登录则跳转登录页并记录回跳地址；已登录返回 true
export function requireAuth(): boolean {
  if (isLoggedIn()) return true
  const redirect = encodeURIComponent(window.location.pathname + window.location.search)
  window.location.href = `/login?redirect=${redirect}`
  return false
}
