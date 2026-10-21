import axios, { type AxiosRequestConfig } from 'axios'

// JWT 在 localStorage 的键名，store 与拦截器共用
export const TOKEN_KEY = 'lol_token'

export interface ApiResult<T> {
  code: number
  msg: string
  data: T
}

const instance = axios.create({
  baseURL: '/api',
  timeout: 15000,
})

// 请求拦截：自动注入 Bearer token
instance.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 错误拦截：401 清除登录态并跳登录页
// （直接操作 localStorage 与 location，避免与 Pinia store 循环依赖）
instance.interceptors.response.use(
  (resp) => resp,
  (error) => {
    const status = error.response?.status
    if (status === 401) {
      localStorage.removeItem(TOKEN_KEY)
      const path = window.location.pathname
      if (path !== '/login') {
        const redirect = encodeURIComponent(path + window.location.search)
        window.location.href = `/login?redirect=${redirect}`
        return new Promise(() => {}) // 中断后续处理，等待页面跳转
      }
    }
    const msg = error.response?.data?.msg || error.message || '网络异常，请稍后重试'
    return Promise.reject(new Error(msg))
  },
)

// 类型化请求函数：统一解包 Result，成功返回值即后端 data；code!=0 抛 Error(msg)
export async function http<T = unknown>(config: AxiosRequestConfig): Promise<T> {
  const resp = await instance.request<ApiResult<T>>(config)
  const body = resp.data
  if (body && typeof body === 'object' && 'code' in body) {
    if (body.code === 0) return body.data
    throw new Error(body.msg || '请求失败')
  }
  return body as unknown as T
}
