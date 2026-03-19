package models

import (
	"time"
)

// Order status constants
const (
	StatusPending        = "PENDING"
	StatusPendingPayment = "PENDING_PAYMENT"
	StatusPaid           = "PAID"
	StatusProcessing     = "PROCESSING"
	StatusAssigned       = "ASSIGNED"
	StatusInDelivery     = "IN_DELIVERY"
	StatusEnRoute        = "EN_ROUTE"
	StatusDelivered      = "DELIVERED"
	StatusFailed         = "FAILED"
	StatusCancelled      = "CANCELLED"
)

type Order struct {
	ID               uint        `gorm:"primaryKey" json:"id"`
	OrderNumber      string      `gorm:"size:100;unique;not null" json:"order_number"`
	Status           string      `gorm:"size:50;default:'PENDING'" json:"status"`
	UserID           uint        `gorm:"not null" json:"user_id"`
	User             *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	DeliveryUserID   *uint       `gorm:"index" json:"delivery_user_id,omitempty"`
	DeliveryUser     *User       `gorm:"foreignKey:DeliveryUserID" json:"delivery_user,omitempty"`
	CustomerName     string      `gorm:"size:255" json:"customer_name"`
	CustomerEmail    string      `gorm:"size:255" json:"customer_email"`
	CustomerPhone    string      `gorm:"size:50" json:"customer_phone"`
	ShippingAddress  string      `gorm:"type:text" json:"shipping_address"`
	ShippingCity     string      `gorm:"size:100" json:"shipping_city"`
	ShippingZipcode  string      `gorm:"size:20" json:"shipping_zipcode"`
	ShippingCountry  string      `gorm:"size:100" json:"shipping_country"`
	BillingAddress   *string     `gorm:"type:text" json:"billing_address,omitempty"`
	BillingCity      *string     `gorm:"size:100" json:"billing_city,omitempty"`
	BillingZipcode   *string     `gorm:"size:20" json:"billing_zipcode,omitempty"`
	BillingCountry   *string     `gorm:"size:100" json:"billing_country,omitempty"`
	PaymentMethod    string      `gorm:"size:100" json:"payment_method"`
	Notes            *string     `gorm:"type:text" json:"notes,omitempty"`
	Subtotal         float64     `gorm:"type:numeric(12,2);default:0" json:"subtotal"`
	Shipping         float64     `gorm:"type:numeric(12,2);default:0" json:"shipping"`
	Tax              float64     `gorm:"type:numeric(12,2);default:0" json:"tax"`
	Total            float64     `gorm:"type:numeric(12,2);default:0" json:"total"`
	ProofPath        *string     `gorm:"size:500" json:"proof_path,omitempty"`
	ProofType        *string     `gorm:"size:50" json:"proof_type,omitempty"`
	Items            []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}
