import Link from 'next/link'
import { Button } from '@/components/ui/button'

export function Header() {
  return (
    <header className="bg-white shadow-sm border-b">
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <Link href="/" className="flex items-center space-x-2">
            <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
              <span className="text-white font-bold">สาน</span>
            </div>
            <span className="text-xl font-bold text-gray-900">Saan System</span>
          </Link>

          {/* Navigation */}
          <nav className="hidden md:flex items-center space-x-6">
            <Link href="/" className="text-gray-600 hover:text-gray-900">
              หน้าแรก
            </Link>
            <Link href="/products" className="text-gray-600 hover:text-gray-900">
              สินค้า
            </Link>
            <Link href="/orders" className="text-gray-600 hover:text-gray-900">
              คำสั่งซื้อ
            </Link>
            <Link href="/tracking" className="text-gray-600 hover:text-gray-900">
              ติดตามสินค้า
            </Link>
          </nav>

          {/* Actions */}
          <div className="flex items-center space-x-4">
            <Button variant="outline" size="sm">
              เข้าสู่ระบบ
            </Button>
            <Button size="sm">
              สมัครสมาชิก
            </Button>
          </div>
        </div>
      </div>
    </header>
  )
}
