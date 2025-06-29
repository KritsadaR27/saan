package template

import (
	"fmt"

	"github.com/saan/order-service/internal/application/dto"
)

// MessageTemplate defines the interface for message templates
type MessageTemplate interface {
	Generate(data *OrderSummaryData) string
}

// OrderSummaryData contains data for generating order summary messages
type OrderSummaryData struct {
	Order       *dto.OrderResponse
	StockIssues []string
}

// OrderSummaryTemplate implements MessageTemplate for order summaries
type OrderSummaryTemplate struct {
	templateType string
}

// NewOrderSummaryTemplate creates a new order summary template
func NewOrderSummaryTemplate(templateType string) *OrderSummaryTemplate {
	return &OrderSummaryTemplate{
		templateType: templateType,
	}
}

// Generate creates order summary message based on template type
func (t *OrderSummaryTemplate) Generate(data *OrderSummaryData) string {
	switch t.templateType {
	case "cod_delivery":
		return t.generateCODDeliveryTemplate(data)
	case "transfer_pickup":
		return t.generateTransferPickupTemplate(data)
	case "credit_shipping":
		return t.generateCreditShippingTemplate(data)
	default:
		return t.generateFallbackTemplate(data)
	}
}

// generateCODDeliveryTemplate สำหรับ COD + ส่งรถสาย
func (t *OrderSummaryTemplate) generateCODDeliveryTemplate(data *OrderSummaryData) string {
	order := data.Order
	
	message := fmt.Sprintf(`🛍️ สรุปออร์เดอร์ (เก็บเงินปลายทาง + ส่งรถสาย)

🆔 หมายเลขออร์เดอร์: %s
👤 ลูกค้า: %s
📍 ที่อยู่จัดส่ง: %s

📦 รายการสินค้า:`, 
		order.ID.String(),
		order.CustomerID.String(), // ในกรณีจริงควรมีชื่อลูกค้า
		order.ShippingAddress)

	// เพิ่มรายการสินค้า
	for i, item := range order.Items {
		message += fmt.Sprintf(`
%d. สินค้า ID: %s
   จำนวน: %d ชิ้น @ ฿%.2f = ฿%.2f`,
			i+1, item.ProductID.String(), item.Quantity, item.UnitPrice, item.TotalPrice)
	}

	message += fmt.Sprintf(`

💰 ยอดรวม: ฿%.2f
💳 ชำระเงิน: เก็บเงินปลายทาง (COD)
🚛 จัดส่ง: รถส่งสาย`, order.TotalAmount)

	// เพิ่มข้อมูลปัญหาสต็อกถ้ามี
	if len(data.StockIssues) > 0 {
		message += "\n\n⚠️ แจ้งเตือนสต็อก:"
		for _, issue := range data.StockIssues {
			message += "\n• " + issue
		}
	}

	message += `

✅ กดยืนยันเพื่อดำเนินการต่อ
❌ หรือพิมพ์ "ยกเลิก" เพื่อยกเลิกออร์เดอร์`

	return message
}

// generateTransferPickupTemplate สำหรับ โอนเงิน + นัดรับ
func (t *OrderSummaryTemplate) generateTransferPickupTemplate(data *OrderSummaryData) string {
	order := data.Order
	
	message := fmt.Sprintf(`🛍️ สรุปออร์เดอร์ (โอนเงิน + นัดรับสินค้า)

🆔 หมายเลขออร์เดอร์: %s
👤 ลูกค้า: %s

📦 รายการสินค้า:`, 
		order.ID.String(),
		order.CustomerID.String())

	// เพิ่มรายการสินค้า
	for i, item := range order.Items {
		message += fmt.Sprintf(`
%d. สินค้า ID: %s
   จำนวน: %d ชิ้น @ ฿%.2f = ฿%.2f`,
			i+1, item.ProductID.String(), item.Quantity, item.UnitPrice, item.TotalPrice)
	}

	message += fmt.Sprintf(`

💰 ยอดรวม: ฿%.2f
💳 ชำระเงิน: โอนเงินผ่านธนาคาร
🏪 รับสินค้า: นัดรับที่หน้าร้าน

🏦 ข้อมูลการโอนเงิน:
ธนาคารกสิกรไทย
เลขที่บัญชี: 123-4-56789-0
ชื่อบัญชี: บริษัท สาอัน จำกัด`, order.TotalAmount)

	// เพิ่มข้อมูลปัญหาสต็อกถ้ามี
	if len(data.StockIssues) > 0 {
		message += "\n\n⚠️ แจ้งเตือนสต็อก:"
		for _, issue := range data.StockIssues {
			message += "\n• " + issue
		}
	}

	message += `

✅ กดยืนยันเพื่อดำเนินการต่อ
❌ หรือพิมพ์ "ยกเลิก" เพื่อยกเลิกออร์เดอร์

📝 หมายเหตุ: กรุณาโอนเงินภายใน 24 ชั่วโมง และส่งสลิปมาให้ตรวจสอบ`

	return message
}

// generateCreditShippingTemplate สำหรับ บัตรเครดิต + ขนส่ง
func (t *OrderSummaryTemplate) generateCreditShippingTemplate(data *OrderSummaryData) string {
	order := data.Order
	
	message := fmt.Sprintf(`🛍️ สรุปออร์เดอร์ (บัตรเครดิต + ขนส่งเอกชน)

🆔 หมายเลขออร์เดอร์: %s
👤 ลูกค้า: %s
📍 ที่อยู่จัดส่ง: %s

📦 รายการสินค้า:`, 
		order.ID.String(),
		order.CustomerID.String(),
		order.ShippingAddress)

	// เพิ่มรายการสินค้า
	for i, item := range order.Items {
		message += fmt.Sprintf(`
%d. สินค้า ID: %s
   จำนวน: %d ชิ้น @ ฿%.2f = ฿%.2f`,
			i+1, item.ProductID.String(), item.Quantity, item.UnitPrice, item.TotalPrice)
	}

	shippingFee := order.ShippingFee
	if shippingFee == 0 {
		shippingFee = 50.0 // ค่าจัดส่งเริ่มต้น
	}

	message += fmt.Sprintf(`

💰 ยอดสินค้า: ฿%.2f
🚛 ค่าจัดส่ง: ฿%.2f
💰 ยอดรวมทั้งสิ้น: ฿%.2f
💳 ชำระเงิน: บัตรเครดิต/เดบิต
📦 จัดส่ง: ขนส่งเอกชน (Kerry/Flash)`, 
		order.TotalAmount - shippingFee, shippingFee, order.TotalAmount)

	// เพิ่มข้อมูลปัญหาสต็อกถ้ามี
	if len(data.StockIssues) > 0 {
		message += "\n\n⚠️ แจ้งเตือนสต็อก:"
		for _, issue := range data.StockIssues {
			message += "\n• " + issue
		}
	}

	message += `

✅ กดยืนยันเพื่อดำเนินการชำระเงิน
❌ หรือพิมพ์ "ยกเลิก" เพื่อยกเลิกออร์เดอร์

📝 หมายเหตุ: ระบบจะประมวลผลการชำระเงินทันที หลังจากยืนยัน`

	return message
}

// generateFallbackTemplate template สำรอง
func (t *OrderSummaryTemplate) generateFallbackTemplate(data *OrderSummaryData) string {
	order := data.Order
	
	message := fmt.Sprintf(`🛍️ สรุปออร์เดอร์

🆔 หมายเลขออร์เดอร์: %s
👤 ลูกค้า: %s
📍 ที่อยู่: %s

📦 รายการสินค้า:`, 
		order.ID.String(),
		order.CustomerID.String(),
		order.ShippingAddress)

	// เพิ่มรายการสินค้า
	for i, item := range order.Items {
		message += fmt.Sprintf(`
%d. สินค้า ID: %s
   จำนวน: %d ชิ้น @ ฿%.2f = ฿%.2f`,
			i+1, item.ProductID.String(), item.Quantity, item.UnitPrice, item.TotalPrice)
	}

	message += fmt.Sprintf(`

💰 ยอดรวม: ฿%.2f`, order.TotalAmount)

	// เพิ่มข้อมูลการชำระเงินถ้ามี
	if order.PaymentMethod != nil {
		message += fmt.Sprintf(`
💳 วิธีชำระเงิน: %s`, string(*order.PaymentMethod))
	}

	// เพิ่มข้อมูลปัญหาสต็อกถ้ามี
	if len(data.StockIssues) > 0 {
		message += "\n\n⚠️ แจ้งเตือนสต็อก:"
		for _, issue := range data.StockIssues {
			message += "\n• " + issue
		}
	}

	message += `

✅ กดยืนยันเพื่อดำเนินการต่อ
❌ หรือพิมพ์ "ยกเลิก" เพื่อยกเลิกออร์เดอร์

📞 ติดต่อสอบถาม: 02-xxx-xxxx`

	return message
}
