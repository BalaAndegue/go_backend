package models

import (
	"time"
)

type Payment struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	OrderID       uint      `gorm:"not null" json:"order_id"`
	Order         *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Amount        float64   `gorm:"type:numeric(12,2);not null" json:"amount"`
	Method        string    `gorm:"size:100" json:"method"`
	Status        string    `gorm:"size:50;default:'PENDING'" json:"status"`
	TransactionID *string   `gorm:"size:255" json:"transaction_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
