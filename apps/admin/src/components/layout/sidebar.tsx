'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'
import {
  BarChart3,
  MessageSquare,
  ShoppingBag,
  Package,
  Truck,
  DollarSign,
  Settings,
  Users,
  Home
} from 'lucide-react'

const navigation = [
  { name: 'แดชบอร์ด', href: '/', icon: Home },
  { name: 'แชท', href: '/chats', icon: MessageSquare },
  { name: 'คำสั่งซื้อ', href: '/orders', icon: ShoppingBag },
  { name: 'คลังสินค้า', href: '/inventory', icon: Package },
  { name: 'จัดส่ง', href: '/delivery', icon: Truck },
  { name: 'การเงิน', href: '/finance', icon: DollarSign },
  { name: 'ลูกค้า', href: '/customers', icon: Users },
  { name: 'รายงาน', href: '/reports', icon: BarChart3 },
  { name: 'ตั้งค่า', href: '/settings', icon: Settings },
]

export function Sidebar() {
  const pathname = usePathname()

  return (
    <div className="flex flex-col h-screen bg-white shadow-sm border-r">
      {/* Logo */}
      <div className="flex items-center h-16 px-6 border-b">
        <div className="flex items-center space-x-2">
          <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
            <span className="text-white font-bold text-sm">สาน</span>
          </div>
          <span className="text-xl font-bold text-gray-900">Admin</span>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-4 py-6 space-y-2">
        {navigation.map((item) => {
          const isActive = pathname === item.href
          return (
            <Link
              key={item.name}
              href={item.href}
              className={cn(
                'flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors',
                isActive
                  ? 'bg-blue-50 text-blue-700 border-r-2 border-blue-700'
                  : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
              )}
            >
              <item.icon className="w-5 h-5 mr-3" />
              {item.name}
            </Link>
          )
        })}
      </nav>

      {/* User Info */}
      <div className="p-4 border-t">
        <div className="flex items-center space-x-3">
          <div className="w-8 h-8 bg-gray-300 rounded-full"></div>
          <div>
            <p className="text-sm font-medium text-gray-900">Admin User</p>
            <p className="text-xs text-gray-500">admin@saan.com</p>
          </div>
        </div>
      </div>
    </div>
  )
}
