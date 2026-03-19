package models

import (
	"time"
)

type Cart struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     *uint      `gorm:"index" json:"user_id,omitempty"`
	User       *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	SessionID  string     `gorm:"size:100" json:"session_id,omitempty"`
	ItemsCount int        `gorm:"default:0" json:"items_count"`
	Total      float64    `gorm:"type:numeric(12,2);default:0" json:"total"`
	Items      []CartItem `gorm:"foreignKey:CartID" json:"items,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
