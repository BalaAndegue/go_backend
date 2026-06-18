package models

import (
	"strings"
	"time"
)

// Coupon discount types.
const (
	CouponPercent = "PERCENT"
	CouponFixed   = "FIXED"
)

// Coupon is a discount code applied at checkout.
type Coupon struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Code        string     `gorm:"size:50;unique;not null" json:"code"`
	Type        string     `gorm:"size:20;not null" json:"type"` // PERCENT | FIXED
	Value       float64    `gorm:"type:numeric(12,2);not null" json:"value"`
	MinSubtotal float64    `gorm:"type:numeric(12,2);default:0" json:"min_subtotal"`
	MaxUses     int        `gorm:"default:0" json:"max_uses"` // 0 = unlimited
	UsedCount   int        `gorm:"default:0" json:"used_count"`
	Active      bool       `gorm:"default:true" json:"active"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// NormalizeCode upper-cases and trims a coupon code for consistent lookup.
func NormalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// Validate reports whether the coupon can be applied to the given subtotal,
// returning a human-readable reason when it cannot.
func (c *Coupon) Validate(subtotal float64) (bool, string) {
	if !c.Active {
		return false, "Coupon is not active"
	}
	if c.ExpiresAt != nil && c.ExpiresAt.Before(time.Now()) {
		return false, "Coupon has expired"
	}
	if c.MaxUses > 0 && c.UsedCount >= c.MaxUses {
		return false, "Coupon usage limit reached"
	}
	if subtotal < c.MinSubtotal {
		return false, "Order subtotal is below the coupon minimum"
	}
	return true, ""
}

// DiscountFor computes the discount amount for a subtotal, never exceeding it.
func (c *Coupon) DiscountFor(subtotal float64) float64 {
	var d float64
	switch c.Type {
	case CouponPercent:
		d = subtotal * c.Value / 100
	case CouponFixed:
		d = c.Value
	}
	if d > subtotal {
		d = subtotal
	}
	if d < 0 {
		d = 0
	}
	return d
}
