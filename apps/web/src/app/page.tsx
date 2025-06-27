import { ChatInterface } from '@/components/chat/chat-interface'
import { Header } from '@/components/layout/header'
import { Footer } from '@/components/layout/footer'

export default function HomePage() {
  return (
    <div className="min-h-screen flex flex-col">
      <Header />
      <main className="flex-1 container mx-auto px-4 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Welcome Section */}
          <div className="space-y-6">
            <div className="text-center lg:text-left">
              <h1 className="text-4xl font-bold text-gray-900 mb-4">
                ยินดีต้อนรับสู่ <span className="text-blue-600">สาน</span>
              </h1>
              <p className="text-xl text-gray-600 mb-6">
                เชื่อมโยงการสื่อสาร กับการสั่งซื้อ และการจัดส่ง
              </p>
              <div className="space-y-4">
                <div className="flex items-center space-x-3">
                  <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                    <span className="text-blue-600 font-semibold">💬</span>
                  </div>
                  <span className="text-gray-700">แชทเพื่อสั่งซื้อง่าย ๆ</span>
                </div>
                <div className="flex items-center space-x-3">
                  <div className="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center">
                    <span className="text-green-600 font-semibold">📋</span>
                  </div>
                  <span className="text-gray-700">จัดการคำสั่งซื้ออัตโนมัติ</span>
                </div>
                <div className="flex items-center space-x-3">
                  <div className="w-8 h-8 bg-purple-100 rounded-full flex items-center justify-center">
                    <span className="text-purple-600 font-semibold">🚚</span>
                  </div>
                  <span className="text-gray-700">ติดตามการจัดส่งแบบเรียลไทม์</span>
                </div>
              </div>
            </div>
          </div>

          {/* Chat Interface */}
          <div className="lg:col-span-1">
            <div className="bg-white rounded-lg shadow-lg p-6">
              <h2 className="text-2xl font-semibold text-gray-900 mb-4">
                เริ่มแชทเพื่อสั่งซื้อ
              </h2>
              <ChatInterface />
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  )
}
