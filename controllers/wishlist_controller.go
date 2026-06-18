package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"shopcart-api/config"
	"shopcart-api/models"
)

// GetWishlist returns the caller's saved products.
func GetWishlist(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var items []models.WishlistItem
	config.DB.Preload("Product.Variants").Preload("Product.Category").
		Where("user_id = ?", userID.(uint)).Order("created_at DESC").Find(&items)
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": items, "code": 200})
}

type WishlistInput struct {
	ProductID uint `json:"product_id" binding:"required"`
}

// AddWishlist adds a product to the caller's wishlist (idempotent).
func AddWishlist(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uint)

	var input WishlistInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	var product models.Product
	if err := config.DB.First(&product, input.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Product not found"})
		return
	}

	var item models.WishlistItem
	err := config.DB.Where("user_id = ? AND product_id = ?", uid, input.ProductID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		item = models.WishlistItem{UserID: uid, ProductID: input.ProductID}
		config.DB.Create(&item)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Added to wishlist", "data": item, "code": 200})
}

// RemoveWishlist removes a product from the caller's wishlist.
func RemoveWishlist(c *gin.Context) {
	userID, _ := c.Get("user_id")
	productID := c.Param("product")
	config.DB.Where("user_id = ? AND product_id = ?", userID.(uint), productID).
		Delete(&models.WishlistItem{})
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Removed from wishlist", "code": 200})
}
