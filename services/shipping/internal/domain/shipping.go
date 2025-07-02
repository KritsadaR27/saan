package domain

import (
	"time"
	"github.com/google/uuid"
)

// DeliveryMethod represents the method of delivery
type DeliveryMethod string

const (
	SelfDelivery   DeliveryMethod = "self_delivery"
	InterExpress   DeliveryMethod = "inter_express"
	Lalamove       DeliveryMethod = "lalamove"
	Grab           DeliveryMethod = "grab"
	Flash          DeliveryMethod = "flash"
	NimExpress     DeliveryMethod = "nim_express"
)

// DeliveryTask represents a delivery task
type DeliveryTask struct {
	ID                     uuid.UUID      `json:"id" db:"id"`
	OrderID                uuid.UUID      `json:"order_id" db:"order_id"`
	CustomerAddressID      uuid.UUID      `json:"customer_address_id" db:"customer_address_id"`
	DeliveryMethod         DeliveryMethod `json:"delivery_method" db:"delivery_method"`
	DeliveryRoute          string         `json:"delivery_route" db:"delivery_route"`
	VehicleID              *uuid.UUID     `json:"vehicle_id,omitempty" db:"vehicle_id"`
	DriverID               *uuid.UUID     `json:"driver_id,omitempty" db:"driver_id"`
	CarrierID              *uuid.UUID     `json:"carrier_id,omitempty" db:"carrier_id"`
	CarrierTrackingNumber  *string        `json:"carrier_tracking_number,omitempty" db:"carrier_tracking_number"`
	PickupScheduledTime    *time.Time     `json:"pickup_scheduled_time,omitempty" db:"pickup_scheduled_time"`
	PlannedDeliveryDate    time.Time      `json:"planned_delivery_date" db:"planned_delivery_date"`
	EstimatedDeliveryTime  *time.Time     `json:"estimated_delivery_time,omitempty" db:"estimated_delivery_time"`
	ActualDeliveryTime     *time.Time     `json:"actual_delivery_time,omitempty" db:"actual_delivery_time"`
	DeliveryFee            float64        `json:"delivery_fee" db:"delivery_fee"`
	CODAmount              float64        `json:"cod_amount" db:"cod_amount"`
	Status                 TaskStatus     `json:"status" db:"status"`
	Notes                  string         `json:"notes" db:"notes"`
	CreatedAt              time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at" db:"updated_at"`
}

// TaskStatus represents delivery task status
type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskPlanned    TaskStatus = "planned"
	TaskDispatched TaskStatus = "dispatched"
	TaskInTransit  TaskStatus = "in_transit"
	TaskDelivered  TaskStatus = "delivered"
	TaskFailed     TaskStatus = "failed"
	TaskCancelled  TaskStatus = "cancelled"
)

// DeliveryCarrier represents a third-party delivery carrier
type DeliveryCarrier struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	CarrierName          string    `json:"carrier_name" db:"carrier_name"`
	DisplayName          string    `json:"display_name" db:"display_name"`
	CarrierType          string    `json:"carrier_type" db:"carrier_type"` // 'scheduled', 'on_demand'
	PickupAddress        string    `json:"pickup_address" db:"pickup_address"`
	ContactInfo          string    `json:"contact_info" db:"contact_info"` // JSON
	PricingRules         string    `json:"pricing_rules" db:"pricing_rules"` // JSON
	IsActive             bool      `json:"is_active" db:"is_active"`
	CutoffTime           string    `json:"cutoff_time" db:"cutoff_time"` // Time format "15:04"
	APIEndpoint          *string   `json:"api_endpoint,omitempty" db:"api_endpoint"`
	APIKey               *string   `json:"api_key,omitempty" db:"api_key"`
	TrackingURLTemplate  *string   `json:"tracking_url_template,omitempty" db:"tracking_url_template"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// DeliveryRoute represents a delivery route for self-delivery
type DeliveryRoute struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	RouteName           string    `json:"route_name" db:"route_name"`
	RouteCode           string    `json:"route_code" db:"route_code"`
	CoverageProvinces   string    `json:"coverage_provinces" db:"coverage_provinces"` // JSON array
	CoverageDistricts   string    `json:"coverage_districts" db:"coverage_districts"` // JSON array
	EstimatedDeliveryHours int    `json:"estimated_delivery_hours" db:"estimated_delivery_hours"`
	BaseDeliveryFee     float64   `json:"base_delivery_fee" db:"base_delivery_fee"`
	IsActive            bool      `json:"is_active" db:"is_active"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// CustomerAddress represents customer address (from Customer Service)
type CustomerAddress struct {
	ID             uuid.UUID `json:"id"`
	CustomerID     uuid.UUID `json:"customer_id"`
	LocationName   string    `json:"location_name"`
	HouseNumber    string    `json:"house_number"`
	AddressLine1   string    `json:"address_line1"`
	Subdistrict    string    `json:"subdistrict"`
	District       string    `json:"district"`
	Province       string    `json:"province"`
	PostalCode     string    `json:"postal_code"`
	DeliveryRoute  string    `json:"delivery_route"`
	IsDefault      bool      `json:"is_default"`
	IsActive       bool      `json:"is_active"`
}

// DeliveryOption represents available delivery options for an address
type DeliveryOption struct {
	Method             DeliveryMethod `json:"method"`
	CarrierName        string         `json:"carrier_name,omitempty"`
	Route              string         `json:"route,omitempty"`
	VehicleID          *uuid.UUID     `json:"vehicle_id,omitempty"`
	EstimatedHours     int            `json:"estimated_hours"`
	DeliveryFee        float64        `json:"delivery_fee"`
	EstimatedDelivery  time.Time      `json:"estimated_delivery"`
	IsRecommended      bool           `json:"is_recommended"`
	Reason             string         `json:"reason"`
}

// Repository interfaces
type DeliveryTaskRepository interface {
	Create(task *DeliveryTask) error
	GetByID(id uuid.UUID) (*DeliveryTask, error)
	GetByOrderID(orderID uuid.UUID) (*DeliveryTask, error)
	UpdateStatus(id uuid.UUID, status TaskStatus) error
	GetPendingTasks(date time.Time) ([]*DeliveryTask, error)
	GetTasksByRoute(route string, date time.Time) ([]*DeliveryTask, error)
}

type CarrierRepository interface {
	GetAll() ([]*DeliveryCarrier, error)
	GetByID(id uuid.UUID) (*DeliveryCarrier, error)
	GetActiveCarriers() ([]*DeliveryCarrier, error)
}

type RouteRepository interface {
	GetAll() ([]*DeliveryRoute, error)
	GetByCode(code string) (*DeliveryRoute, error)
	GetByProvince(province string) (*DeliveryRoute, error)
}

// Service interfaces
type ShippingService interface {
	CreateDeliveryTask(orderID uuid.UUID, customerAddressID uuid.UUID, codAmount float64) (*DeliveryTask, error)
	GetDeliveryOptions(customerAddressID uuid.UUID) ([]*DeliveryOption, error)
	UpdateTaskStatus(taskID uuid.UUID, status TaskStatus) error
	GetTaskByOrderID(orderID uuid.UUID) (*DeliveryTask, error)
	PlanDailyRoutes(date time.Time) error
}

type RouteService interface {
	OptimizeRoute(tasks []*DeliveryTask) ([]*DeliveryTask, error)
	GetRouteInfo(routeCode string) (*DeliveryRoute, error)
	CalculateDeliveryFee(route string, codAmount float64) (float64, error)
}

type CarrierService interface {
	GetAvailableCarriers(province string) ([]*DeliveryCarrier, error)
	SchedulePickup(carrierID uuid.UUID, tasks []*DeliveryTask) error
	GetTrackingInfo(carrierID uuid.UUID, trackingNumber string) (map[string]interface{}, error)
}
