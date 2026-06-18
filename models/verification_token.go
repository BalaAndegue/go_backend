package models

import "time"

// Token purposes.
const (
	PurposePasswordReset = "password_reset"
	PurposeEmailVerify   = "email_verify"
)

// VerificationToken backs single-use, time-limited flows such as password reset
// and email verification. Only the hash of the token is stored.
type VerificationToken struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"index;not null" json:"user_id"`
	Purpose   string     `gorm:"size:50;index;not null" json:"purpose"`
	TokenHash string     `gorm:"size:64;index;not null" json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
