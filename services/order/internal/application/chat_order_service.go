package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/saan/order-service/internal/application/dto"
	"github.com/saan/order-service/internal/application/template"
	"github.com/saan/order-service/internal/domain"
	"github.com/saan/order-service/internal/infrastructure/client"
	"github.com/saan/order-service/pkg/logger"
)

// ChatOrderItem represents an item from chat interaction
type ChatOrderItem struct {
	ProductName string  `json:"product_name"`
	ProductID   *uuid.UUID `json:"product_id,omitempty"`
	Quantity    int     `json:"quantity"`
	UnitPrice   *float64 `json:"unit_price,omitempty"`
}

// ChatOrderRequest represents a chat-based order creation request
type ChatOrderRequest struct {
	ChatID          string          `json:"chat_id"`
	CustomerID      string          `json:"customer_id"`
	Items           []ChatOrderItem `json:"items"`
	PaymentMethod   *string         `json:"payment_method,omitempty"`
	DeliveryMethod  *string         `json:"delivery_method,omitempty"`
	ShippingAddress *string         `json:"shipping_address,omitempty"`
	Notes           string          `json:"notes,omitempty"`
}

// ChatOrderService handles chat-based order operations
type ChatOrderService struct {
	orderService     *OrderService
	customerClient   client.CustomerClient
	inventoryClient  client.InventoryClient
	notificationClient client.NotificationClient
	templateSelector *template.TemplateSelector
	logger           logger.Logger
}

// NewChatOrderService creates a new chat order service
func NewChatOrderService(
	orderService *OrderService,
	customerClient client.CustomerClient,
	inventoryClient client.InventoryClient,
	notificationClient client.NotificationClient,
	templateSelector *template.TemplateSelector,
	logger logger.Logger,
) *ChatOrderService {
	return &ChatOrderService{
		orderService:     orderService,
		customerClient:   customerClient,
		inventoryClient:  inventoryClient,
		notificationClient: notificationClient,
		templateSelector: templateSelector,
		logger:           logger,
	}
}

// CreateOrderFromChat creates an order from chat interaction
func (s *ChatOrderService) CreateOrderFromChat(
	ctx context.Context,
	req *ChatOrderRequest,
) (*dto.OrderResponse, error) {
	s.logger.Info("Creating order from chat", "chat_id", req.ChatID, "customer_id", req.CustomerID)

	// 1. ดึงข้อมูลลูกค้า
	customerUUID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		s.logger.Error("Invalid customer ID format", "customer_id", req.CustomerID, "error", err)
		return nil, fmt.Errorf("invalid customer ID format: %w", err)
	}

	customer, err := s.customerClient.GetCustomer(ctx, customerUUID)
	if err != nil {
		s.logger.Error("Failed to get customer", "customer_id", req.CustomerID, "error", err)
		return nil, fmt.Errorf("failed to get customer information: %w", err)
	}

	// 2. ตรวจสอบสต็อกและดึงข้อมูลสินค้า
	orderItems := make([]dto.CreateOrderItemRequest, 0, len(req.Items))
	stockIssues := make([]string, 0)

	for _, chatItem := range req.Items {
		// ถ้าไม่มี ProductID ให้ค้นหาจากชื่อ
		var productID uuid.UUID
		var unitPrice float64

		if chatItem.ProductID != nil {
			productID = *chatItem.ProductID
		} else {
			// ในกรณีจริงควรมี ProductClient สำหรับค้นหาสินค้าจากชื่อ
			// สำหรับตอนนี้ให้ return error
			s.logger.Error("Product ID is required", "product_name", chatItem.ProductName)
			return nil, fmt.Errorf("cannot find product: %s", chatItem.ProductName)
		}

		// ดึงข้อมูลสินค้า
		product, err := s.inventoryClient.GetProduct(ctx, productID)
		if err != nil {
			s.logger.Error("Failed to get product", "product_id", productID, "error", err)
			return nil, fmt.Errorf("failed to get product %s: %w", chatItem.ProductName, err)
		}

		// ใช้ราคาจาก chat หรือราคาจากสินค้า
		if chatItem.UnitPrice != nil {
			unitPrice = *chatItem.UnitPrice
		} else {
			unitPrice = product.Price
		}

		// ตรวจสอบสต็อก
		stockCheck, err := s.inventoryClient.CheckStock(ctx, productID, chatItem.Quantity)
		if err != nil {
			s.logger.Warn("Failed to check stock", "product_id", productID, "error", err)
			// Continue แต่บันทึก warning
		} else if !stockCheck.CanFulfill {
			stockIssues = append(stockIssues, fmt.Sprintf("%s: มีเพียง %d ชิ้น (ต้องการ %d ชิ้น)", 
				product.Name, stockCheck.Available, chatItem.Quantity))
		}

		orderItems = append(orderItems, dto.CreateOrderItemRequest{
			ProductID: productID,
			Quantity:  chatItem.Quantity,
			UnitPrice: unitPrice,
		})
	}

	// 3. สร้าง order draft
	// สำหรับตอนนี้ใช้ชื่อลูกค้าเป็นที่อยู่ชั่วคราว ในกรณีจริงควรมี address service
	defaultAddress := fmt.Sprintf("%s %s", customer.FirstName, customer.LastName)
	shippingAddress := defaultAddress
	if req.ShippingAddress != nil && *req.ShippingAddress != "" {
		shippingAddress = *req.ShippingAddress
	}

	createOrderReq := &dto.CreateOrderRequest{
		CustomerID:      customerUUID,
		ShippingAddress: shippingAddress,
		BillingAddress:  defaultAddress,
		Notes:          req.Notes,
		Items:          orderItems,
	}

	// กำหนด payment method ถ้ามี
	if req.PaymentMethod != nil {
		paymentMethod := domain.PaymentMethod(*req.PaymentMethod)
		createOrderReq.PaymentMethod = &paymentMethod
	}

	order, err := s.orderService.CreateOrder(ctx, createOrderReq)
	if err != nil {
		s.logger.Error("Failed to create order", "chat_id", req.ChatID, "error", err)
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 4. สร้างข้อความสรุป
	orderSummary := s.GenerateOrderSummary(order, stockIssues)

	// 5. ส่ง message สรุป
	err = s.sendOrderSummaryMessage(ctx, req.ChatID, orderSummary)
	if err != nil {
		s.logger.Error("Failed to send order summary", "chat_id", req.ChatID, "order_id", order.ID, "error", err)
		// ไม่ให้ fail การสร้าง order เพราะ notification failure
	}

	s.logger.Info("Order created from chat successfully", 
		"chat_id", req.ChatID, 
		"order_id", order.ID, 
		"customer_id", req.CustomerID,
		"total_amount", order.TotalAmount)

	return order, nil
}

// GenerateOrderSummary สร้างข้อความสรุปออร์เดอร์
func (s *ChatOrderService) GenerateOrderSummary(order *dto.OrderResponse, stockIssues []string) string {
	// เลือก template ตามเงื่อนไข
	selectedTemplate := s.templateSelector.SelectTemplate(order)
	
	// สร้างข้อมูลสำหรับ template
	data := &template.OrderSummaryData{
		Order:       order,
		StockIssues: stockIssues,
	}

	return selectedTemplate.Generate(data)
}

// ConfirmChatOrder ยืนยันออร์เดอร์จาก chat
func (s *ChatOrderService) ConfirmChatOrder(ctx context.Context, chatID string, orderID uuid.UUID) (*dto.OrderResponse, error) {
	s.logger.Info("Confirming chat order", "chat_id", chatID, "order_id", orderID)

	// อัพเดทสถานะเป็น confirmed
	statusReq := &dto.UpdateOrderStatusRequest{
		Status: domain.OrderStatusConfirmed,
	}

	order, err := s.orderService.UpdateOrderStatus(ctx, orderID, statusReq)
	if err != nil {
		s.logger.Error("Failed to confirm chat order", "chat_id", chatID, "order_id", orderID, "error", err)
		return nil, fmt.Errorf("failed to confirm order: %w", err)
	}

	// ส่ง confirmation message
	confirmationMsg := s.generateConfirmationMessage(order)
	err = s.sendConfirmationMessage(ctx, chatID, confirmationMsg)
	if err != nil {
		s.logger.Error("Failed to send confirmation message", "chat_id", chatID, "order_id", orderID, "error", err)
		// ไม่ให้ fail การ confirm เพราะ notification failure
	}

	s.logger.Info("Chat order confirmed successfully", "chat_id", chatID, "order_id", orderID)
	return order, nil
}

// CancelChatOrder ยกเลิกออร์เดอร์จาก chat
func (s *ChatOrderService) CancelChatOrder(ctx context.Context, chatID string, orderID uuid.UUID, reason string) error {
	s.logger.Info("Cancelling chat order", "chat_id", chatID, "order_id", orderID, "reason", reason)

	// อัพเดทสถานะเป็น cancelled
	statusReq := &dto.UpdateOrderStatusRequest{
		Status: domain.OrderStatusCancelled,
	}

	_, err := s.orderService.UpdateOrderStatus(ctx, orderID, statusReq)
	if err != nil {
		s.logger.Error("Failed to cancel chat order", "chat_id", chatID, "order_id", orderID, "error", err)
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// ส่ง cancellation message
	cancellationMsg := s.generateCancellationMessage(orderID, reason)
	err = s.sendCancellationMessage(ctx, chatID, cancellationMsg)
	if err != nil {
		s.logger.Error("Failed to send cancellation message", "chat_id", chatID, "order_id", orderID, "error", err)
		// ไม่ให้ fail การ cancel เพราะ notification failure
	}

	s.logger.Info("Chat order cancelled successfully", "chat_id", chatID, "order_id", orderID)
	return nil
}

// sendOrderSummaryMessage ส่งข้อความสรุปออร์เดอร์
func (s *ChatOrderService) sendOrderSummaryMessage(ctx context.Context, chatID string, message string) error {
	// ส่งผ่าน notification service
	return s.notificationClient.SendChatMessage(ctx, chatID, message)
}

// sendConfirmationMessage ส่งข้อความยืนยัน
func (s *ChatOrderService) sendConfirmationMessage(ctx context.Context, chatID string, message string) error {
	return s.notificationClient.SendChatMessage(ctx, chatID, message)
}

// sendCancellationMessage ส่งข้อความยกเลิก
func (s *ChatOrderService) sendCancellationMessage(ctx context.Context, chatID string, message string) error {
	return s.notificationClient.SendChatMessage(ctx, chatID, message)
}

// generateConfirmationMessage สร้างข้อความยืนยัน
func (s *ChatOrderService) generateConfirmationMessage(order *dto.OrderResponse) string {
	return fmt.Sprintf(`✅ ยืนยันออร์เดอร์เรียบร้อยแล้ว!

🆔 หมายเลขออร์เดอร์: %s
💰 ยอดรวม: ฿%.2f
📦 สถานะ: %s

เราจะประมวลผลออร์เดอร์ของคุณและแจ้งข้อมูลการจัดส่งให้อีกครั้ง
ขอบคุณที่ใช้บริการครับ! 🙏`, 
		order.ID.String(), 
		order.TotalAmount, 
		string(order.Status))
}

// generateCancellationMessage สร้างข้อความยกเลิก
func (s *ChatOrderService) generateCancellationMessage(orderID uuid.UUID, reason string) string {
	msg := fmt.Sprintf(`❌ ยกเลิกออร์เดอร์แล้ว

🆔 หมายเลขออร์เดอร์: %s`, orderID.String())

	if reason != "" {
		msg += fmt.Sprintf(`
📝 เหตุผล: %s`, reason)
	}

	msg += `

หากมีข้อสอบถาม กรุณาติดต่อทีมงานครับ 🙏`

	return msg
}
