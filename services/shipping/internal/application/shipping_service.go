package application

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"saan/shipping/internal/domain"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type shippingService struct {
	db            *sql.DB
	redis         *redis.Client
	kafka         *kafka.Writer
	taskRepo      domain.DeliveryTaskRepository
	carrierRepo   domain.CarrierRepository
	routeRepo     domain.RouteRepository
}

func NewShippingService(db *sql.DB, redis *redis.Client, kafka *kafka.Writer) domain.ShippingService {
	return &shippingService{
		db:    db,
		redis: redis,
		kafka: kafka,
		// Initialize repositories here
	}
}

func (s *shippingService) CreateDeliveryTask(orderID uuid.UUID, customerAddressID uuid.UUID, codAmount float64) (*domain.DeliveryTask, error) {
	// Get customer address details
	_, err := s.getCustomerAddress(customerAddressID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer address: %w", err)
	}

	// Determine delivery method based on address
	deliveryOptions, err := s.GetDeliveryOptions(customerAddressID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery options: %w", err)
	}

	if len(deliveryOptions) == 0 {
		return nil, fmt.Errorf("no delivery options available for address")
	}

	// Use the recommended option
	selectedOption := deliveryOptions[0]
	for _, option := range deliveryOptions {
		if option.IsRecommended {
			selectedOption = option
			break
		}
	}

	// Create delivery task
	task := &domain.DeliveryTask{
		ID:                    uuid.New(),
		OrderID:               orderID,
		CustomerAddressID:     customerAddressID,
		DeliveryMethod:        selectedOption.Method,
		DeliveryRoute:         selectedOption.Route,
		PlannedDeliveryDate:   selectedOption.EstimatedDelivery,
		DeliveryFee:           selectedOption.DeliveryFee,
		CODAmount:             codAmount,
		Status:                domain.TaskPending,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Save to database
	if err := s.taskRepo.Create(task); err != nil {
		return nil, fmt.Errorf("failed to create delivery task: %w", err)
	}

	// Publish event to Kafka
	event := map[string]interface{}{
		"event_type":           "delivery_task_created",
		"order_id":             orderID.String(),
		"delivery_task_id":     task.ID.String(),
		"delivery_method":      string(task.DeliveryMethod),
		"planned_delivery_date": task.PlannedDeliveryDate,
		"delivery_fee":         task.DeliveryFee,
		"timestamp":            time.Now(),
	}

	s.publishEvent("shipping-events", event)

	return task, nil
}

func (s *shippingService) GetDeliveryOptions(customerAddressID uuid.UUID) ([]*domain.DeliveryOption, error) {
	// Get customer address
	address, err := s.getCustomerAddress(customerAddressID)
	if err != nil {
		return nil, err
	}

	var options []*domain.DeliveryOption

	// Check if province is in self-delivery area
	route, err := s.routeRepo.GetByProvince(address.Province)
	if err == nil && route != nil {
		// Self-delivery option available
		deliveryTime := time.Now().Add(24 * time.Hour) // Next day delivery
		option := &domain.DeliveryOption{
			Method:            domain.SelfDelivery,
			Route:             route.RouteCode,
			EstimatedHours:    route.EstimatedDeliveryHours,
			DeliveryFee:       route.BaseDeliveryFee,
			EstimatedDelivery: deliveryTime,
			IsRecommended:     true,
			Reason:           "Customer is in self-delivery area",
		}
		options = append(options, option)
	}

	// Get available carriers for this province
	carriers, err := s.carrierRepo.GetActiveCarriers()
	if err == nil {
		for _, carrier := range carriers {
			// Calculate delivery time based on carrier schedule
			deliveryTime := s.calculateCarrierDeliveryTime(carrier)
			fee := s.calculateCarrierFee(carrier, address.Province)

			option := &domain.DeliveryOption{
				Method:            domain.DeliveryMethod(carrier.CarrierName),
				CarrierName:       carrier.DisplayName,
				EstimatedHours:    48, // Default 2 days for third-party
				DeliveryFee:       fee,
				EstimatedDelivery: deliveryTime,
				IsRecommended:     false,
				Reason:           fmt.Sprintf("Third-party carrier: %s", carrier.DisplayName),
			}
			options = append(options, option)
		}
	}

	return options, nil
}

func (s *shippingService) UpdateTaskStatus(taskID uuid.UUID, status domain.TaskStatus) error {
	if err := s.taskRepo.UpdateStatus(taskID, status); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Publish status update event
	event := map[string]interface{}{
		"event_type":       "delivery_status_updated",
		"delivery_task_id": taskID.String(),
		"status":           string(status),
		"timestamp":        time.Now(),
	}

	s.publishEvent("shipping-events", event)

	return nil
}

func (s *shippingService) GetTaskByOrderID(orderID uuid.UUID) (*domain.DeliveryTask, error) {
	return s.taskRepo.GetByOrderID(orderID)
}

func (s *shippingService) PlanDailyRoutes(date time.Time) error {
	// Get all pending tasks for the date
	tasks, err := s.taskRepo.GetPendingTasks(date)
	if err != nil {
		return fmt.Errorf("failed to get pending tasks: %w", err)
	}

	// Group tasks by delivery route
	routeGroups := make(map[string][]*domain.DeliveryTask)
	for _, task := range tasks {
		if task.DeliveryMethod == domain.SelfDelivery {
			routeGroups[task.DeliveryRoute] = append(routeGroups[task.DeliveryRoute], task)
		}
	}

	// Plan routes for each group
	for route, routeTasks := range routeGroups {
		if err := s.planRoute(route, routeTasks); err != nil {
			return fmt.Errorf("failed to plan route %s: %w", route, err)
		}
	}

	return nil
}

// Helper methods
func (s *shippingService) getCustomerAddress(addressID uuid.UUID) (*domain.CustomerAddress, error) {
	// This would call Customer Service API
	// For now, return mock data
	return &domain.CustomerAddress{
		ID:            addressID,
		Province:      "กรุงเทพมหานคร",
		District:      "บางกะปิ",
		Subdistrict:   "หัวหมาก",
		DeliveryRoute: "route_a",
	}, nil
}

func (s *shippingService) calculateCarrierDeliveryTime(carrier *domain.DeliveryCarrier) time.Time {
	// Calculate based on carrier schedule
	return time.Now().Add(48 * time.Hour) // Default 2 days
}

func (s *shippingService) calculateCarrierFee(carrier *domain.DeliveryCarrier, province string) float64 {
	// Calculate based on pricing rules
	return 80.0 // Default fee
}

func (s *shippingService) planRoute(route string, tasks []*domain.DeliveryTask) error {
	// Optimize delivery sequence
	// Assign vehicle and driver
	// Update task status to planned
	for _, task := range tasks {
		task.Status = domain.TaskPlanned
		s.taskRepo.UpdateStatus(task.ID, domain.TaskPlanned)
	}
	return nil
}

func (s *shippingService) publishEvent(topic string, event map[string]interface{}) error {
	eventBytes, _ := json.Marshal(event)
	
	msg := kafka.Message{
		Topic: topic,
		Value: eventBytes,
	}

	return s.kafka.WriteMessages(nil, msg)
}
