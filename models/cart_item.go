package models

import (
	"time"
)

type CartItem struct {
	ID               uint            `gorm:"primaryKey" json:"id"`
	CartID           uint            `gorm:"not null" json:"cart_id"`
	Cart             *Cart           `gorm:"foreignKey:CartID" json:"cart,omitempty"`
	ProductVariantID *uint           `gorm:"index" json:"product_variant_id,omitempty"`
	ProductVariant   *ProductVariant `gorm:"foreignKey:ProductVariantID" json:"product_variant,omitempty"`
	Quantity         int             `gorm:"not null" json:"quantity"`
	UnitPrice        float64         `gorm:"type:numeric(12,2);not null" json:"unit_price"`
	Total            float64         `gorm:"type:numeric(12,2);not null" json:"total"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}
