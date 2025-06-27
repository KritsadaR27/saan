import { BaseEntity } from './index'

// Chat Types
export interface Chat extends BaseEntity {
  customer_id: string
  customer_name: string
  status: ChatStatus
  last_message?: string
  agent_id?: string
  metadata?: Record<string, any>
}

export interface Message extends BaseEntity {
  chat_id: string
  content: string
  sender_type: MessageSenderType
  sender_id: string
  message_type: MessageType
  metadata?: Record<string, any>
  is_ai_response?: boolean
}

export type ChatStatus = 'active' | 'waiting' | 'closed' | 'escalated'
export type MessageSenderType = 'customer' | 'ai' | 'agent' | 'system'
export type MessageType = 'text' | 'image' | 'file' | 'order' | 'system'

// Chat API Request/Response Types
export interface CreateChatRequest {
  customer_id: string
  customer_name: string
  metadata?: Record<string, any>
}

export interface SendMessageRequest {
  content: string
  sender_type: MessageSenderType
  sender_id: string
  message_type?: MessageType
  metadata?: Record<string, any>
}

export interface ChatListResponse {
  chats: Chat[]
  total: number
}

export interface MessageListResponse {
  messages: Message[]
  total: number
}
