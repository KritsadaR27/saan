'use client'

import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { TrendingUp, TrendingDown, Users, ShoppingBag, Package, Truck } from 'lucide-react'

interface StatsData {
  totalOrders: number
  totalRevenue: number
  activeChats: number
  lowStockItems: number
  ordersGrowth: number
  revenueGrowth: number
}

export function DashboardStats() {
  const { data: stats, isLoading } = useQuery<StatsData>({
    queryKey: ['dashboard-stats'],
    queryFn: async () => {
      // Mock data - replace with actual API call
      return {
        totalOrders: 1234,
        totalRevenue: 567890,
        activeChats: 42,
        lowStockItems: 8,
        ordersGrowth: 12.5,
        revenueGrowth: 8.3
      }
    }
  })

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {[...Array(4)].map((_, i) => (
          <Card key={i} className="animate-pulse">
            <CardHeader className="pb-2">
              <div className="h-4 bg-gray-200 rounded w-3/4"></div>
            </CardHeader>
            <CardContent>
              <div className="h-8 bg-gray-200 rounded w-1/2"></div>
            </CardContent>
          </Card>
        ))}
      </div>
    )
  }

  const statCards = [
    {
      title: 'คำสั่งซื้อทั้งหมด',
      value: stats?.totalOrders.toLocaleString() || '0',
      growth: stats?.ordersGrowth || 0,
      icon: ShoppingBag,
      color: 'text-blue-600'
    },
    {
      title: 'รายได้รวม',
      value: `฿${stats?.totalRevenue.toLocaleString() || '0'}`,
      growth: stats?.revenueGrowth || 0,
      icon: TrendingUp,
      color: 'text-green-600'
    },
    {
      title: 'แชทที่ใช้งานอยู่',
      value: stats?.activeChats.toString() || '0',
      growth: 0,
      icon: Users,
      color: 'text-purple-600'
    },
    {
      title: 'สินค้าใกล้หมด',
      value: stats?.lowStockItems.toString() || '0',
      growth: 0,
      icon: Package,
      color: 'text-red-600'
    }
  ]

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      {statCards.map((stat, index) => (
        <Card key={index}>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">
              {stat.title}
            </CardTitle>
            <stat.icon className={`h-4 w-4 ${stat.color}`} />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-gray-900">{stat.value}</div>
            {stat.growth !== 0 && (
              <div className="flex items-center text-xs text-gray-600 mt-1">
                {stat.growth > 0 ? (
                  <TrendingUp className="h-3 w-3 text-green-500 mr-1" />
                ) : (
                  <TrendingDown className="h-3 w-3 text-red-500 mr-1" />
                )}
                <span className={stat.growth > 0 ? 'text-green-600' : 'text-red-600'}>
                  {Math.abs(stat.growth)}%
                </span>
                <span className="ml-1">จากเดือนที่แล้ว</span>
              </div>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
