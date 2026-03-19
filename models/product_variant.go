package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONB type for attributes
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONB")
	}
	return json.Unmarshal(bytes, j)
}

type ProductVariant struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ProductID  uint      `gorm:"column:productId;not null" json:"productId"`
	Product    *Product  `gorm:"foreignKey:ProductID;references:ID" json:"product,omitempty"`
	Name       string    `gorm:"size:255;not null" json:"name"`
	SKU        string    `gorm:"size:100;not null" json:"sku"`
	Price      float64   `gorm:"type:numeric(12,2);not null" json:"price"`
	Stock      int       `gorm:"default:0" json:"stock"`
	Color      *string   `gorm:"size:50" json:"color,omitempty"`
	Attributes JSONB     `gorm:"type:jsonb" json:"attributes,omitempty"`
	Image      *string   `gorm:"size:500" json:"image,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
