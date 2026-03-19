package models

import (
	"time"
)

// Role constants matching Laravel's User model
const (
	RoleCustomer   = "CUSTOMER"
	RoleAdmin      = "ADMIN"
	RoleVendor     = "VENDOR"
	RoleDelivery   = "DELIVERY"
	RoleManager    = "MANAGER"
	RoleSupervisor = "SUPERVISOR"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Email     string    `gorm:"size:255;unique;not null" json:"email"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Role      string    `gorm:"size:50;default:'CUSTOMER'" json:"role"`
	Phone     *string   `gorm:"size:100" json:"phone"`
	Address   *string   `gorm:"type:text" json:"address"`
	FcmToken  *string   `gorm:"size:500" json:"fcm_token,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) IsAdmin() bool      { return u.Role == RoleAdmin }
func (u *User) IsVendor() bool     { return u.Role == RoleVendor }
func (u *User) IsDelivery() bool   { return u.Role == RoleDelivery }
func (u *User) IsManager() bool    { return u.Role == RoleManager }
func (u *User) IsSupervisor() bool { return u.Role == RoleSupervisor }
func (u *User) IsManagement() bool {
	return u.Role == RoleAdmin || u.Role == RoleManager || u.Role == RoleSupervisor
}
