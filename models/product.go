package models

import (
	"time"
)

type Product struct {
	ID           uint             `gorm:"primaryKey" json:"id"`
	Name         string           `gorm:"size:255;not null" json:"name"`
	Slug         string           `gorm:"size:255;unique;not null" json:"slug"`
	Description  *string          `gorm:"type:text" json:"description"`
	Price        float64          `gorm:"type:numeric(12,2);not null" json:"price"`
	ComparePrice *float64         `gorm:"type:numeric(12,2)" json:"compare_price,omitempty"`
	Stock        int              `gorm:"default:0" json:"stock"`
	SKU          *string          `gorm:"size:100;unique" json:"sku,omitempty"`
	Image        *string          `gorm:"size:500" json:"image,omitempty"`
	IsVisible    bool             `gorm:"default:false" json:"is_visible"`
	IsFeatured   bool             `gorm:"default:false" json:"is_featured"`
	CategoryID   uint             `gorm:"not null" json:"category_id"`
	Category     Category         `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Variants     []ProductVariant `gorm:"foreignKey:ProductID" json:"variants,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}
