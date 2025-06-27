export function Footer() {
  return (
    <footer className="bg-gray-50 border-t">
      <div className="container mx-auto px-4 py-8">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          {/* Company Info */}
          <div className="col-span-1 md:col-span-2">
            <div className="flex items-center space-x-2 mb-4">
              <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
                <span className="text-white font-bold">สาน</span>
              </div>
              <span className="text-xl font-bold text-gray-900">Saan System</span>
            </div>
            <p className="text-gray-600 mb-4">
              เชื่อมโยงการสื่อสาร กับการสั่งซื้อ และการจัดส่ง
            </p>
            <p className="text-gray-600">
              แพลตฟอร์มจัดการธุรกิจที่ทันสมัย ด้วยระบบแชทสู่การสั่งซื้อ
            </p>
          </div>

          {/* Services */}
          <div>
            <h3 className="font-semibold text-gray-900 mb-4">บริการ</h3>
            <ul className="space-y-2 text-gray-600">
              <li>ระบบแชท</li>
              <li>จัดการคำสั่งซื้อ</li>
              <li>คลังสินค้า</li>
              <li>จัดส่งสินค้า</li>
              <li>การเงิน</li>
            </ul>
          </div>

          {/* Contact */}
          <div>
            <h3 className="font-semibold text-gray-900 mb-4">ติดต่อ</h3>
            <ul className="space-y-2 text-gray-600">
              <li>อีเมล: info@saan.com</li>
              <li>โทร: 02-xxx-xxxx</li>
              <li>LINE: @saan</li>
              <li>Facebook: Saan System</li>
            </ul>
          </div>
        </div>

        <div className="border-t mt-8 pt-8 text-center text-gray-600">
          <p>&copy; 2025 Saan System. สงวนลิขสิทธิ์.</p>
        </div>
      </div>
    </footer>
  )
}
