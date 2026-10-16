import { http } from './http'

export interface Message {
  id: number
  userId: number
  username: string
  content: string
  isRead: boolean
  createdAt: string
}

export interface MessageListResult {
  list: Message[]
}

export function sendMessage(content: string) {
  return http<Message>({ method: 'POST', url: '/messages', data: { content } })
}

export function listMyMessages() {
  return http<MessageListResult>({ method: 'GET', url: '/messages/mine' })
}

export function listAdminMessages() {
  return http<MessageListResult>({ method: 'GET', url: '/admin/messages' })
}
