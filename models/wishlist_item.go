package models

import "time"

// WishlistItem links a user to a product they saved. The pair is unique.
type WishlistItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_wishlist_user_product" json:"user_id"`
	ProductID uint      `gorm:"not null;uniqueIndex:idx_wishlist_user_product" json:"product_id"`
	Product   *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
