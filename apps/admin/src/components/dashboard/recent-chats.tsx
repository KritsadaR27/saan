'use client'

import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { MessageCircle } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { th } from 'date-fns/locale'

interface Chat {
  id: string
  customer_name: string
  status: string
  last_message: string
  updated_at: string
}

export function RecentChats() {
  const { data: chats, isLoading } = useQuery<Chat[]>({
    queryKey: ['recent-chats'],
    queryFn: async () => {
      // Mock data - replace with actual API call
      return [
        {
          id: '1',
          customer_name: 'นายสมชาย ใจดี',
          status: 'active',
          last_message: 'ขอดูสินค้าใหม่หน่อยครับ',
          updated_at: new Date(Date.now() - 1000 * 60 * 2).toISOString()
        },
        {
          id: '2',
          customer_name: 'นางสาวมาลี สวยงาม',
          status: 'waiting',
          last_message: 'เมื่อไหร่จะส่งสินค้าคะ',
          updated_at: new Date(Date.now() - 1000 * 60 * 10).toISOString()
        },
        {
          id: '3',
          customer_name: 'นายวิทย์ เก่งกาจ',
          status: 'active',
          last_message: 'ต้องการสั่งซื้อเพิ่มครับ',
          updated_at: new Date(Date.now() - 1000 * 60 * 25).toISOString()
        }
      ]
    }
  })

  const getStatusColor = (status: string) => {
    const colors = {
      active: 'bg-green-100 text-green-800',
      waiting: 'bg-yellow-100 text-yellow-800',
      closed: 'bg-gray-100 text-gray-800'
    }
    return colors[status as keyof typeof colors] || 'bg-gray-100 text-gray-800'
  }

  const getStatusText = (status: string) => {
    const statusMap = {
      active: 'กำลังแชท',
      waiting: 'รอตอบกลับ',
      closed: 'ปิดแล้ว'
    }
    return statusMap[status as keyof typeof statusMap] || status
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <MessageCircle className="w-5 h-5" />
          <span>แชทล่าสุด</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="space-y-4">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="animate-pulse">
                <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                <div className="h-3 bg-gray-200 rounded w-1/2"></div>
              </div>
            ))}
          </div>
        ) : (
          <div className="space-y-4">
            {chats?.map((chat) => (
              <div key={chat.id} className="flex items-start space-x-3 p-3 border rounded-lg hover:bg-gray-50 cursor-pointer">
                <div className="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">
                  <MessageCircle className="w-5 h-5 text-blue-600" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between">
                    <p className="font-medium text-gray-900 truncate">
                      {chat.customer_name}
                    </p>
                    <Badge className={getStatusColor(chat.status)}>
                      {getStatusText(chat.status)}
                    </Badge>
                  </div>
                  <p className="text-sm text-gray-600 truncate mt-1">
                    {chat.last_message}
                  </p>
                  <p className="text-xs text-gray-500 mt-1">
                    {formatDistanceToNow(new Date(chat.updated_at), { 
                      addSuffix: true, 
                      locale: th 
                    })}
                  </p>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
