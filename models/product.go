package models

import (
	"time"
)

type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Slug        string    `gorm:"size:255;unique;not null" json:"slug"`
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"type:numeric(10,2);not null" json:"price"`
	ImageURL    *string   `gorm:"size:255" json:"image_url"`
	CategoryID  uint      `gorm:"not null" json:"category_id"`
	Category    Category  `gorm:"foreignKey:CategoryID" json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
