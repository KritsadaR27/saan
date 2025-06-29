package template

import (
	"github.com/saan/order-service/internal/application/dto"
	"github.com/saan/order-service/internal/domain"
)

// TemplateSelector selects appropriate message template based on order conditions
type TemplateSelector struct {
	templates map[string]MessageTemplate
}

// NewTemplateSelector creates a new template selector
func NewTemplateSelector() *TemplateSelector {
	templates := make(map[string]MessageTemplate)
	templates["cod_delivery"] = NewOrderSummaryTemplate("cod_delivery")
	templates["transfer_pickup"] = NewOrderSummaryTemplate("transfer_pickup")
	templates["credit_shipping"] = NewOrderSummaryTemplate("credit_shipping")
	templates["fallback"] = NewOrderSummaryTemplate("fallback")

	return &TemplateSelector{
		templates: templates,
	}
}

// SelectTemplate เลือก template ตามเงื่อนไข payment method และ delivery method
func (s *TemplateSelector) SelectTemplate(order *dto.OrderResponse) MessageTemplate {
	templateKey := s.determineTemplateKey(order)
	
	if template, exists := s.templates[templateKey]; exists {
		return template
	}
	
	// Fallback to default template
	return s.templates["fallback"]
}

// determineTemplateKey กำหนด template key ตามเงื่อนไข
func (s *TemplateSelector) determineTemplateKey(order *dto.OrderResponse) string {
	// ถ้าไม่มี payment method ให้ใช้ fallback
	if order.PaymentMethod == nil {
		return "fallback"
	}

	paymentMethod := *order.PaymentMethod
	
	// เงื่อนไขการเลือก template:
	// 1. COD + มีที่อยู่จัดส่ง = cod_delivery
	// 2. Bank Transfer + ไม่มีค่าจัดส่ง = transfer_pickup  
	// 3. Credit Card + มีค่าจัดส่ง = credit_shipping
	// 4. อื่นๆ = fallback

	switch paymentMethod {
	case domain.PaymentMethodCash:
		// COD delivery - ถ้ามีที่อยู่จัดส่ง
		if order.ShippingAddress != "" {
			return "cod_delivery"
		}
		return "fallback"
		
	case domain.PaymentMethodBankTransfer:
		// Transfer pickup - ถ้าไม่มีค่าจัดส่งหรือค่าจัดส่ง = 0
		if order.ShippingFee == 0 {
			return "transfer_pickup"
		}
		return "fallback"
		
	case domain.PaymentMethodCreditCard:
		// Credit card shipping - ถ้ามีค่าจัดส่ง
		if order.ShippingFee > 0 {
			return "credit_shipping"
		}
		return "fallback"
		
	case domain.PaymentMethodQRCode:
		// QR Code สามารถใช้กับทั้ง pickup และ delivery
		if order.ShippingFee == 0 {
			return "transfer_pickup" // ใช้ template เดียวกับ transfer
		}
		return "credit_shipping" // ใช้ template เดียวกับ credit card
		
	default:
		return "fallback"
	}
}

// GetAvailableTemplates returns list of available template keys
func (s *TemplateSelector) GetAvailableTemplates() []string {
	keys := make([]string, 0, len(s.templates))
	for key := range s.templates {
		keys = append(keys, key)
	}
	return keys
}

// AddCustomTemplate allows adding custom templates
func (s *TemplateSelector) AddCustomTemplate(key string, template MessageTemplate) {
	s.templates[key] = template
}
