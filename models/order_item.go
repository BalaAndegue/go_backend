package models

import (
	"time"
)

type OrderItem struct {
	ID               uint            `gorm:"primaryKey" json:"id"`
	OrderID          uint            `gorm:"not null" json:"order_id"`
	Order            *Order          `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	ProductID        *uint           `gorm:"index" json:"product_id,omitempty"`
	Product          *Product        `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	ProductVariantID *uint           `gorm:"index" json:"product_variant_id,omitempty"`
	ProductVariant   *ProductVariant `gorm:"foreignKey:ProductVariantID" json:"product_variant,omitempty"`
	ProductName      string          `gorm:"size:255" json:"product_name"`
	ProductSKU       *string         `gorm:"size:100" json:"product_sku,omitempty"`
	Quantity         int             `gorm:"not null" json:"quantity"`
	UnitPrice        float64         `gorm:"type:numeric(12,2);not null" json:"unit_price"`
	Total            float64         `gorm:"type:numeric(12,2);not null" json:"total"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}
