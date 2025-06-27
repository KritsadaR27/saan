'use client'

import { create } from 'zustand'
import { chatApi } from '@/lib/api'
import { io, Socket } from 'socket.io-client'

export interface Message {
  id: string
  chat_id: string
  content: string
  sender_type: 'customer' | 'ai' | 'agent'
  sender_id: string
  metadata?: Record<string, any>
  created_at: string
}

export interface Chat {
  id: string
  customer_id: string
  customer_name: string
  status: 'active' | 'closed' | 'waiting'
  last_message?: string
  created_at: string
  updated_at: string
}

interface ChatStore {
  // State
  messages: Message[]
  currentChat: Chat | null
  socket: Socket | null
  isConnecting: boolean
  isConnected: boolean

  // Actions
  setMessages: (messages: Message[]) => void
  addMessage: (message: Message) => void
  setCurrentChat: (chat: Chat | null) => void
  connectWebSocket: () => void
  disconnectWebSocket: () => void
  sendMessage: (content: string) => Promise<void>
  createChat: (customerName: string) => Promise<Chat | null>
}

export const useChatStore = create<ChatStore>((set, get) => ({
  // Initial state
  messages: [],
  currentChat: null,
  socket: null,
  isConnecting: false,
  isConnected: false,

  // Actions
  setMessages: (messages) => set({ messages }),
  
  addMessage: (message) => set((state) => ({
    messages: [...state.messages, message]
  })),

  setCurrentChat: (chat) => set({ currentChat: chat }),

  connectWebSocket: () => {
    const { socket, isConnected } = get()
    
    if (socket && isConnected) return

    set({ isConnecting: true })

    const newSocket = io(process.env.NEXT_PUBLIC_CHAT_API_URL || 'http://localhost:8001', {
      transports: ['websocket']
    })

    newSocket.on('connect', () => {
      set({ isConnected: true, isConnecting: false })
      console.log('Connected to chat service')
    })

    newSocket.on('disconnect', () => {
      set({ isConnected: false })
      console.log('Disconnected from chat service')
    })

    newSocket.on('message', (message: Message) => {
      get().addMessage(message)
    })

    newSocket.on('connect_error', (error) => {
      set({ isConnecting: false, isConnected: false })
      console.error('Connection error:', error)
    })

    set({ socket: newSocket })
  },

  disconnectWebSocket: () => {
    const { socket } = get()
    if (socket) {
      socket.disconnect()
      set({ socket: null, isConnected: false })
    }
  },

  sendMessage: async (content: string) => {
    const { currentChat, socket } = get()
    
    if (!currentChat) {
      // Create a new chat first
      const newChat = await get().createChat('Anonymous User')
      if (!newChat) return
    }

    const chat = get().currentChat
    if (!chat) return

    try {
      const response = await chatApi.post(`/chats/${chat.id}/messages`, {
        content,
        sender_type: 'customer',
        sender_id: 'user_' + Date.now(),
      })

      const message = response.data
      get().addMessage(message)

      // Also send via WebSocket for real-time updates
      if (socket && socket.connected) {
        socket.emit('send_message', {
          chat_id: chat.id,
          content,
          sender_type: 'customer',
        })
      }
    } catch (error) {
      console.error('Failed to send message:', error)
    }
  },

  createChat: async (customerName: string): Promise<Chat | null> => {
    try {
      const response = await chatApi.post('/chats', {
        customer_id: 'user_' + Date.now(),
        customer_name: customerName,
      })

      const chat = response.data
      set({ currentChat: chat })
      
      // Join the chat room via WebSocket
      const { socket } = get()
      if (socket && socket.connected) {
        socket.emit('join_chat', { chat_id: chat.id })
      }

      return chat
    } catch (error) {
      console.error('Failed to create chat:', error)
      return null
    }
  },
}))
