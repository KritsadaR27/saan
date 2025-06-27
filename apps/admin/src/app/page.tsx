import { DashboardStats } from '@/components/dashboard/dashboard-stats'
import { RecentOrders } from '@/components/dashboard/recent-orders'
import { RecentChats } from '@/components/dashboard/recent-chats'
import { InventoryAlerts } from '@/components/dashboard/inventory-alerts'

export default function AdminDashboard() {
  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-gray-900">แดชบอร์ด</h1>
        <p className="text-gray-600 mt-2">ภาพรวมระบบ Saan</p>
      </div>

      {/* Stats Cards */}
      <DashboardStats />

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Recent Orders */}
        <div className="space-y-6">
          <RecentOrders />
          <InventoryAlerts />
        </div>

        {/* Recent Chats */}
        <div className="space-y-6">
          <RecentChats />
        </div>
      </div>
    </div>
  )
}
