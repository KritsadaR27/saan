package application

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"saan/payment/internal/domain"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type paymentService struct {
	db    *sql.DB
	redis *redis.Client
	kafka *kafka.Writer
}

func NewPaymentService(db *sql.DB, redis *redis.Client, kafka *kafka.Writer) domain.PaymentService {
	return &paymentService{
		db:    db,
		redis: redis,
		kafka: kafka,
	}
}

func (p *paymentService) CreatePayment(req *domain.PaymentRequest) (*domain.Payment, error) {
	payment := &domain.Payment{
		ID:            uuid.New(),
		OrderID:       req.OrderID,
		CustomerID:    req.CustomerID,
		PaymentMethod: req.PaymentMethod,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Status:        domain.PaymentPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if payment.Currency == "" {
		payment.Currency = "THB"
	}

	// Create snapshot
	p.createSnapshot(payment, "created")

	// Publish event
	event := map[string]interface{}{
		"event_type":     "payment_created",
		"payment_id":     payment.ID.String(),
		"order_id":       payment.OrderID.String(),
		"amount":         payment.Amount,
		"payment_method": string(payment.PaymentMethod),
		"timestamp":      time.Now(),
	}
	p.publishEvent("payment-events", event)

	return payment, nil
}

func (p *paymentService) ProcessPayment(paymentID uuid.UUID) (*domain.PaymentGatewayResponse, error) {
	// Mock payment processing
	response := &domain.PaymentGatewayResponse{
		Success:       true,
		TransactionID: fmt.Sprintf("txn_%d", time.Now().Unix()),
		PaymentID:     paymentID.String(),
		Status:        domain.PaymentProcessing,
		Amount:        100.0,
		Fee:           3.0,
		PaymentURL:    "https://payment.gateway.com/pay/12345",
	}

	return response, nil
}

func (p *paymentService) CompletePayment(paymentID uuid.UUID, externalTransactionID string) error {
	// Update payment status to completed
	// Create completed snapshot
	// Publish completion event
	
	event := map[string]interface{}{
		"event_type":              "payment_completed",
		"payment_id":              paymentID.String(),
		"external_transaction_id": externalTransactionID,
		"timestamp":               time.Now(),
	}
	p.publishEvent("payment-events", event)

	return nil
}

func (p *paymentService) FailPayment(paymentID uuid.UUID, reason string) error {
	event := map[string]interface{}{
		"event_type":     "payment_failed",
		"payment_id":     paymentID.String(),
		"failure_reason": reason,
		"timestamp":      time.Now(),
	}
	p.publishEvent("payment-events", event)

	return nil
}

func (p *paymentService) GetPaymentByID(id uuid.UUID) (*domain.Payment, error) {
	// Mock response
	return &domain.Payment{
		ID:     id,
		Status: domain.PaymentCompleted,
		Amount: 100.0,
	}, nil
}

func (p *paymentService) GetPaymentsByOrderID(orderID uuid.UUID) ([]*domain.Payment, error) {
	// Mock response
	return []*domain.Payment{
		{
			ID:      uuid.New(),
			OrderID: orderID,
			Status:  domain.PaymentCompleted,
			Amount:  100.0,
		},
	}, nil
}

func (p *paymentService) GetPaymentHistory(paymentID uuid.UUID) ([]*domain.PaymentSnapshot, error) {
	// Mock response
	return []*domain.PaymentSnapshot{}, nil
}

func (p *paymentService) createSnapshot(payment *domain.Payment, snapshotType string) error {
	_ = &domain.PaymentSnapshot{
		ID:           uuid.New(),
		PaymentID:    payment.ID,
		SnapshotType: snapshotType,
		Status:       payment.Status,
		Amount:       payment.Amount,
		CreatedAt:    time.Now(),
	}
	
	// Save snapshot to database (mock)
	return nil
}

func (p *paymentService) publishEvent(topic string, event map[string]interface{}) error {
	eventBytes, _ := json.Marshal(event)
	
	msg := kafka.Message{
		Topic: topic,
		Value: eventBytes,
	}

	return p.kafka.WriteMessages(nil, msg)
}

type refundService struct {
	db    *sql.DB
	redis *redis.Client
	kafka *kafka.Writer
}

func NewRefundService(db *sql.DB, redis *redis.Client, kafka *kafka.Writer) domain.RefundService {
	return &refundService{
		db:    db,
		redis: redis,
		kafka: kafka,
	}
}

func (r *refundService) CreateRefund(paymentID uuid.UUID, amount float64, reason string, createdBy uuid.UUID) (*domain.Refund, error) {
	refund := &domain.Refund{
		ID:           uuid.New(),
		PaymentID:    paymentID,
		RefundAmount: amount,
		RefundReason: reason,
		Status:       domain.PaymentPending,
		CreatedBy:    createdBy,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return refund, nil
}

func (r *refundService) ProcessRefund(refundID uuid.UUID) error {
	return nil
}

func (r *refundService) GetRefundByID(id uuid.UUID) (*domain.Refund, error) {
	return &domain.Refund{ID: id}, nil
}

func (r *refundService) GetRefundsByPaymentID(paymentID uuid.UUID) ([]*domain.Refund, error) {
	return []*domain.Refund{}, nil
}
