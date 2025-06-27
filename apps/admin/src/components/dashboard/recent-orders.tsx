'use client'

import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { formatDistanceToNow } from 'date-fns'
import { th } from 'date-fns/locale'

interface Order {
  id: string
  customer_name: string
  status: string
  total_amount: number
  created_at: string
}

export function RecentOrders() {
  const { data: orders, isLoading } = useQuery<Order[]>({
    queryKey: ['recent-orders'],
    queryFn: async () => {
      // Mock data - replace with actual API call
      return [
        {
          id: '1',
          customer_name: 'นายสมชาย ใจดี',
          status: 'pending',
          total_amount: 1250,
          created_at: new Date(Date.now() - 1000 * 60 * 5).toISOString()
        },
        {
          id: '2',
          customer_name: 'นางสาวมาลี สวยงาม',
          status: 'confirmed',
          total_amount: 890,
          created_at: new Date(Date.now() - 1000 * 60 * 15).toISOString()
        },
        {
          id: '3',
          customer_name: 'นายวิทย์ เก่งกาจ',
          status: 'processing',
          total_amount: 2100,
          created_at: new Date(Date.now() - 1000 * 60 * 30).toISOString()
        }
      ]
    }
  })

  const getStatusColor = (status: string) => {
    const colors = {
      pending: 'bg-yellow-100 text-yellow-800',
      confirmed: 'bg-blue-100 text-blue-800',
      processing: 'bg-purple-100 text-purple-800',
      shipped: 'bg-green-100 text-green-800',
      delivered: 'bg-green-100 text-green-800',
      cancelled: 'bg-red-100 text-red-800'
    }
    return colors[status as keyof typeof colors] || 'bg-gray-100 text-gray-800'
  }

  const getStatusText = (status: string) => {
    const statusMap = {
      pending: 'รอดำเนินการ',
      confirmed: 'ยืนยันแล้ว',
      processing: 'กำลังจัดเตรียม',
      shipped: 'จัดส่งแล้ว',
      delivered: 'ส่งแล้ว',
      cancelled: 'ยกเลิก'
    }
    return statusMap[status as keyof typeof statusMap] || status
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>คำสั่งซื้อล่าสุด</CardTitle>
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
            {orders?.map((order) => (
              <div key={order.id} className="flex items-center justify-between p-3 border rounded-lg">
                <div className="space-y-1">
                  <p className="font-medium text-gray-900">{order.customer_name}</p>
                  <div className="flex items-center space-x-2">
                    <Badge className={getStatusColor(order.status)}>
                      {getStatusText(order.status)}
                    </Badge>
                    <span className="text-sm text-gray-500">
                      {formatDistanceToNow(new Date(order.created_at), { 
                        addSuffix: true, 
                        locale: th 
                      })}
                    </span>
                  </div>
                </div>
                <div className="text-right">
                  <p className="font-semibold text-gray-900">
                    ฿{order.total_amount.toLocaleString()}
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
