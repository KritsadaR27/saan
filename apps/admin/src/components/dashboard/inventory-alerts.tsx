'use client'

import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { AlertTriangle, Package } from 'lucide-react'

interface Product {
  id: string
  name: string
  stock: number
  min_stock: number
}

export function InventoryAlerts() {
  const { data: lowStockProducts, isLoading } = useQuery<Product[]>({
    queryKey: ['low-stock-products'],
    queryFn: async () => {
      // Mock data - replace with actual API call
      return [
        { id: '1', name: 'เสื้อยืดสีขาว', stock: 5, min_stock: 10 },
        { id: '2', name: 'กางเกงยีนส์', stock: 2, min_stock: 5 },
        { id: '3', name: 'รองเท้าผ้าใบ', stock: 1, min_stock: 3 }
      ]
    }
  })

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <Package className="w-5 h-5" />
          <span>แจ้งเตือนสินค้า</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="space-y-3">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="animate-pulse">
                <div className="h-16 bg-gray-200 rounded"></div>
              </div>
            ))}
          </div>
        ) : lowStockProducts && lowStockProducts.length > 0 ? (
          <div className="space-y-3">
            {lowStockProducts.map((product) => (
              <Alert key={product.id} className="border-orange-200 bg-orange-50">
                <AlertTriangle className="h-4 w-4 text-orange-600" />
                <AlertDescription>
                  <div className="flex justify-between items-center">
                    <div>
                      <p className="font-medium text-orange-800">{product.name}</p>
                      <p className="text-sm text-orange-600">
                        เหลือ {product.stock} ชิ้น (ขั้นต่ำ {product.min_stock} ชิ้น)
                      </p>
                    </div>
                    <div className="text-right">
                      <span className="text-sm font-medium text-orange-800">
                        ต้องเติมสต็อก
                      </span>
                    </div>
                  </div>
                </AlertDescription>
              </Alert>
            ))}
          </div>
        ) : (
          <div className="text-center py-6">
            <Package className="w-12 h-12 text-gray-300 mx-auto mb-4" />
            <p className="text-gray-500">ไม่มีสินค้าที่ต้องแจ้งเตือน</p>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
