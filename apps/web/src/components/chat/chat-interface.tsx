'use client'

import { useState, useEffect, useRef } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card } from '@/components/ui/card'
import { Send, MessageCircle } from 'lucide-react'
import { useChatStore } from '@/store/chat-store'
import { cn } from '@/lib/utils'

export function ChatInterface() {
  const [message, setMessage] = useState('')
  const [isConnected, setIsConnected] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  
  const { messages, sendMessage, connectWebSocket, isConnecting } = useChatStore()

  useEffect(() => {
    connectWebSocket()
    setIsConnected(true)
  }, [connectWebSocket])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const handleSendMessage = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!message.trim()) return

    await sendMessage(message)
    setMessage('')
  }

  return (
    <div className="flex flex-col h-96">
      {/* Connection Status */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center space-x-2">
          <MessageCircle className="w-5 h-5 text-blue-600" />
          <span className="font-medium">แชท</span>
        </div>
        <div className={cn(
          "flex items-center space-x-2",
          isConnected ? "text-green-600" : "text-red-600"
        )}>
          <div className={cn(
            "w-2 h-2 rounded-full",
            isConnected ? "bg-green-600" : "bg-red-600"
          )} />
          <span className="text-sm">
            {isConnected ? 'เชื่อมต่อแล้ว' : 'ไม่ได้เชื่อมต่อ'}
          </span>
        </div>
      </div>

      {/* Messages Area */}
      <Card className="flex-1 p-4 overflow-y-auto mb-4">
        <div className="space-y-4">
          {messages.length === 0 ? (
            <div className="text-center text-gray-500 py-8">
              <MessageCircle className="w-12 h-12 mx-auto mb-4 text-gray-300" />
              <p>เริ่มต้นการสนทนาเพื่อสั่งซื้อสินค้า</p>
              <p className="text-sm mt-2">พิมพ์ข้อความเพื่อเริ่มแชท</p>
            </div>
          ) : (
            messages.map((msg) => (
              <div
                key={msg.id}
                className={cn(
                  "flex",
                  msg.sender_type === 'customer' ? "justify-end" : "justify-start"
                )}
              >
                <div
                  className={cn(
                    "max-w-xs lg:max-w-md px-4 py-2 rounded-lg",
                    msg.sender_type === 'customer'
                      ? "bg-blue-600 text-white"
                      : "bg-gray-100 text-gray-900"
                  )}
                >
                  <p className="text-sm">{msg.content}</p>
                  <p className="text-xs opacity-70 mt-1">
                    {new Date(msg.created_at).toLocaleTimeString('th-TH')}
                  </p>
                </div>
              </div>
            ))
          )}
          <div ref={messagesEndRef} />
        </div>
      </Card>

      {/* Message Input */}
      <form onSubmit={handleSendMessage} className="flex space-x-2">
        <Input
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          placeholder="พิมพ์ข้อความ..."
          disabled={isConnecting || !isConnected}
          className="flex-1"
        />
        <Button
          type="submit"
          disabled={!message.trim() || isConnecting || !isConnected}
          className="px-4"
        >
          <Send className="w-4 h-4" />
        </Button>
      </form>
    </div>
  )
}
