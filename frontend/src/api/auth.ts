import { http } from './http'

export interface UserVO {
  id: number
  username: string
  email: string
  avatar: string
  role: 'user' | 'admin'
  createdAt: string
}

export interface AuthData {
  token: string
  user: UserVO
}

export function sendCode(email: string) {
  return http<{ msg: string }>({
    method: 'POST',
    url: '/auth/send-code',
    data: { email },
  })
}

export function register(username: string, email: string, password: string, code: string) {
  return http<AuthData>({
    method: 'POST',
    url: '/auth/register',
    data: { username, email, password, code },
  })
}

export function login(account: string, password: string) {
  return http<AuthData>({
    method: 'POST',
    url: '/auth/login',
    data: { account, password },
  })
}

export function fetchMe() {
  return http<{ user: UserVO }>({ method: 'GET', url: '/me' })
}
