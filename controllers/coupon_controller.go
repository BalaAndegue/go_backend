package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"shopcart-api/config"
	"shopcart-api/models"
)

// ListCoupons returns all coupons (management).
func ListCoupons(c *gin.Context) {
	var coupons []models.Coupon
	config.DB.Order("created_at DESC").Find(&coupons)
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": coupons, "code": 200})
}

type CouponInput struct {
	Code        string     `json:"code" binding:"required"`
	Type        string     `json:"type" binding:"required"`
	Value       float64    `json:"value" binding:"required"`
	MinSubtotal float64    `json:"min_subtotal"`
	MaxUses     int        `json:"max_uses"`
	Active      *bool      `json:"active"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// CreateCoupon creates a discount code (management).
func CreateCoupon(c *gin.Context) {
	var input CouponInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	if input.Type != models.CouponPercent && input.Type != models.CouponFixed {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "type must be PERCENT or FIXED"})
		return
	}

	coupon := models.Coupon{
		Code:        models.NormalizeCode(input.Code),
		Type:        input.Type,
		Value:       input.Value,
		MinSubtotal: input.MinSubtotal,
		MaxUses:     input.MaxUses,
		Active:      true,
		ExpiresAt:   input.ExpiresAt,
	}
	if input.Active != nil {
		coupon.Active = *input.Active
	}
	if err := config.DB.Create(&coupon).Error; err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Could not create coupon (code may already exist)"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": coupon, "code": 201})
}

// DeleteCoupon removes a coupon (management).
func DeleteCoupon(c *gin.Context) {
	var coupon models.Coupon
	if err := config.DB.First(&coupon, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Coupon not found"})
		return
	}
	config.DB.Delete(&coupon)
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Coupon deleted", "code": 200})
}

type ValidateCouponInput struct {
	Code     string  `json:"code" binding:"required"`
	Subtotal float64 `json:"subtotal"`
}

// ValidateCoupon checks a coupon against a subtotal and returns the discount.
func ValidateCoupon(c *gin.Context) {
	var input ValidateCouponInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	var coupon models.Coupon
	if err := config.DB.Where("code = ?", models.NormalizeCode(input.Code)).First(&coupon).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"valid": false, "message": "Coupon not found"})
		return
	}
	if ok, reason := coupon.Validate(input.Subtotal); !ok {
		c.JSON(http.StatusOK, gin.H{"valid": false, "message": reason})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"valid": true, "discount": round2(coupon.DiscountFor(input.Subtotal)), "coupon": coupon,
	})
}
