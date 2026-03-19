package models

import (
	"time"
)

type DeliveryGeolocation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"unique;not null" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Latitude  float64   `gorm:"type:numeric(10,6);not null" json:"latitude"`
	Longitude float64   `gorm:"type:numeric(10,6);not null" json:"longitude"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
