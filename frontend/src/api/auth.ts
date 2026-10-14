import { http } from './http'

export interface UserVO {
  id: number
  username: string
  avatar: string
  role: 'user' | 'admin'
  createdAt: string
}

export interface AuthData {
  token: string
  user: UserVO
}

export function register(username: string, password: string) {
  return http<AuthData>({
    method: 'POST',
    url: '/auth/register',
    data: { username, password },
  })
}

export function login(username: string, password: string) {
  return http<AuthData>({
    method: 'POST',
    url: '/auth/login',
    data: { username, password },
  })
}

export function fetchMe() {
  return http<{ user: UserVO }>({ method: 'GET', url: '/me' })
}
