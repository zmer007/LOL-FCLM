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
