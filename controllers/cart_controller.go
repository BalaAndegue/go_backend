package controllers

import (
	"math/rand"
	"net/http"
	"time"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"shopcart-api/config"
	"shopcart-api/models"
)

func getOrCreateCart(c *gin.Context) *models.Cart {
	userID, exists := c.Get("user_id")
	var cart models.Cart
	var err error
	if exists {
		uid := userID.(uint)
		err = config.DB.Where("user_id = ?", uid).First(&cart).Error
		if err == gorm.ErrRecordNotFound {
			sessionID := generateSessionID()
			cart = models.Cart{UserID: &uid, SessionID: sessionID}
			config.DB.Create(&cart)
		}
	}
	return &cart
}

func generateSessionID() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func updateCartTotals(cartID uint) {
	var items []models.CartItem
	config.DB.Where("cart_id = ?", cartID).Find(&items)

	var total float64
	var count int
	for _, item := range items {
		total += item.UnitPrice * float64(item.Quantity)
		count += item.Quantity
	}
	config.DB.Model(&models.Cart{}).Where("id = ?", cartID).Updates(map[string]interface{}{
		"total": total, "items_count": count,
	})
}

// @Summary Récupérer le panier
// @Tags Cart
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /cart [get]
func ShowCart(c *gin.Context) {
	cart := getOrCreateCart(c)
	config.DB.Preload("Items.ProductVariant.Product").First(cart, cart.ID)
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": cart, "message": "Panier récupéré avec succès", "code": 200})
}

// @Summary Créer/récupérer panier
// @Tags Cart
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /cart [post]
func StoreCart(c *gin.Context) {
	cart := getOrCreateCart(c)
	config.DB.Preload("Items.ProductVariant.Product").First(cart, cart.ID)
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": cart, "message": "Panier récupéré", "code": 200})
}

type AddItemInput struct {
	ProductVariantID uint `json:"product_variant_id" binding:"required"`
	Quantity         int  `json:"quantity" binding:"required,min=1"`
}

// @Summary Ajouter un article au panier
// @Tags Cart
// @Security BearerAuth
// @Accept json
// @Param body body AddItemInput true "Article à ajouter"
// @Success 200 {object} map[string]interface{}
// @Router /cart/add [post]
func AddCartItem(c *gin.Context) {
	var input AddItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "error", "message": err.Error(), "code": 422})
		return
	}

	var variant models.ProductVariant
	if err := config.DB.First(&variant, input.ProductVariantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Variant not found", "code": 404})
		return
	}

	if variant.Stock < input.Quantity {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "error", "message": fmt.Sprintf("Stock insuffisant. Disponible: %d", variant.Stock), "code": 422,
		})
		return
	}

	cart := getOrCreateCart(c)

	var existingItem models.CartItem
	err := config.DB.Where("cart_id = ? AND product_variant_id = ?", cart.ID, input.ProductVariantID).First(&existingItem).Error
	if err == nil {
		newQty := existingItem.Quantity + input.Quantity
		if variant.Stock < newQty {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "error", "message": "Stock insuffisant pour la quantité demandée", "code": 422})
			return
		}
		config.DB.Model(&existingItem).Updates(map[string]interface{}{
			"quantity": newQty, "total": float64(newQty) * existingItem.UnitPrice,
		})
	} else {
		total := float64(input.Quantity) * variant.Price
		item := models.CartItem{
			CartID:           cart.ID,
			ProductVariantID: &input.ProductVariantID,
			Quantity:         input.Quantity,
			UnitPrice:        variant.Price,
			Total:            total,
		}
		config.DB.Create(&item)
	}

	updateCartTotals(cart.ID)
	config.DB.Preload("Items.ProductVariant.Product").First(cart, cart.ID)
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Article ajouté au panier", "data": cart, "code": 200})
}

type UpdateItemInput struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

// @Summary Mettre à jour un article du panier
// @Tags Cart
// @Security BearerAuth
// @Param cartItem path int true "ID article"
// @Success 200 {object} map[string]interface{}
// @Router /cart/items/{cartItem} [put]
func UpdateCartItem(c *gin.Context) {
	itemID := c.Param("cartItem")
	var item models.CartItem
	if err := config.DB.First(&item, itemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Article non trouvé", "code": 404})
		return
	}

	var input UpdateItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	config.DB.Model(&item).Updates(map[string]interface{}{
		"quantity": input.Quantity, "total": float64(input.Quantity) * item.UnitPrice,
	})

	updateCartTotals(item.CartID)
	var cart models.Cart
	config.DB.Preload("Items.ProductVariant.Product").First(&cart, item.CartID)
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Quantité mise à jour", "data": cart, "code": 200})
}

// @Summary Retirer un article du panier
// @Tags Cart
// @Security BearerAuth
// @Param cartItem path int true "ID article"
// @Success 200 {object} map[string]interface{}
// @Router /cart/items/{cartItem} [delete]
func RemoveCartItem(c *gin.Context) {
	itemID := c.Param("cartItem")
	var item models.CartItem
	if err := config.DB.First(&item, itemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Article non trouvé", "code": 404})
		return
	}

	cartID := item.CartID
	config.DB.Delete(&item)
	updateCartTotals(cartID)
	var cart models.Cart
	config.DB.Preload("Items.ProductVariant.Product").First(&cart, cartID)
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Article retiré du panier", "data": cart, "code": 200})
}

// @Summary Vider le panier
// @Tags Cart
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /cart/clear [delete]
func ClearCart(c *gin.Context) {
	cart := getOrCreateCart(c)
	config.DB.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{})
	updateCartTotals(cart.ID)
	config.DB.First(cart, cart.ID)
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Panier vidé avec succès", "data": cart, "code": 200})
}

// @Summary Vider le panier d'un utilisateur
// @Tags Cart
// @Param userId path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /cart/user/{userId}/empty [get]
func EmptyUserCart(c *gin.Context) {
	userID, _ := c.Get("user_id")
	targetUserID := c.Param("userId")

	user := models.User{}
	config.DB.First(&user, userID.(uint))

	var uidStr string
	uidStr = fmt.Sprintf("%d", userID.(uint))

	if uidStr != targetUserID && !user.IsManagement() {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Non autorisé", "code": 403})
		return
	}

	var cart models.Cart
	if err := config.DB.Where("user_id = ?", targetUserID).First(&cart).Error; err == nil {
		config.DB.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{})
		updateCartTotals(cart.ID)
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Panier vidé avec succès", "code": 200})
}
