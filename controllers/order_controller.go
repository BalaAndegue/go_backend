package controllers

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"shopcart-api/config"
	"shopcart-api/models"
)

type CreateOrderInput struct {
	CustomerName    string `json:"customer_name" binding:"required"`
	CustomerEmail   string `json:"customer_email" binding:"required,email"`
	CustomerPhone   string `json:"customer_phone" binding:"required"`
	ShippingAddress string `json:"shipping_address" binding:"required"`
	ShippingCity    string `json:"shipping_city" binding:"required"`
	ShippingZipcode string `json:"shipping_zipcode" binding:"required"`
	ShippingCountry string `json:"shipping_country" binding:"required"`
	BillingAddress  string `json:"billing_address"`
	BillingCity     string `json:"billing_city"`
	BillingZipcode  string `json:"billing_zipcode"`
	BillingCountry  string `json:"billing_country"`
	PaymentMethod   string `json:"payment_method" binding:"required"`
	Notes           string `json:"notes"`
}

func generateOrderNumber() string {
	rand.Seed(time.Now().UnixNano())
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("ORD-%s-%s", time.Now().Format("20060102"), string(b))
}

func GetMyOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var orders []models.Order
	query := config.DB.Preload("Items").Where("user_id = ?", userID.(uint)).Order("created_at DESC")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	query.Find(&orders)
	c.JSON(http.StatusOK, gin.H{"message": "Orders retrieved", "data": orders})
}

func GetOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var orders []models.Order
	config.DB.Preload("Items").Where("user_id = ?", userID.(uint)).Order("created_at DESC").Find(&orders)
	if len(orders) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No orders found for this user.", "data": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Orders retrieved successfully", "data": orders})
}

func GetOrder(c *gin.Context) {
	orderID := c.Param("order")
	userID, _ := c.Get("user_id")
	var order models.Order
	if err := config.DB.Preload("Items.Product").First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Order not found"})
		return
	}
	if order.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Unauthorized"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func CreateOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uint)

	var cart models.Cart
	if err := config.DB.Where("user_id = ?", uid).First(&cart).Error; err != nil || cart.ItemsCount == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Your cart is empty or could not be found."})
		return
	}

	var input CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	config.DB.Preload("Items.ProductVariant").First(&cart, cart.ID)

	// Stock verification
	for _, item := range cart.Items {
		if item.ProductVariant != nil && item.ProductVariant.Stock < item.Quantity {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"message": fmt.Sprintf("Stock insuffisant pour la variante ID %d. Disponible: %d", item.ProductVariantID, item.ProductVariant.Stock),
			})
			return
		}
	}

	initialStatus := models.StatusPending
	if input.PaymentMethod != "cash_on_delivery" {
		initialStatus = models.StatusPendingPayment
	}

	billingAddress := &input.BillingAddress
	billingCity := &input.BillingCity
	billingZipcode := &input.BillingZipcode
	billingCountry := &input.BillingCountry
	if input.BillingAddress == "" {
		billingAddress = &input.ShippingAddress
		billingCity = &input.ShippingCity
		billingZipcode = &input.ShippingZipcode
		billingCountry = &input.ShippingCountry
	}

	var notes *string
	if input.Notes != "" {
		notes = &input.Notes
	}

	order := models.Order{
		OrderNumber:     generateOrderNumber(),
		Status:          initialStatus,
		UserID:          uid,
		CustomerName:    input.CustomerName,
		CustomerEmail:   input.CustomerEmail,
		CustomerPhone:   input.CustomerPhone,
		ShippingAddress: input.ShippingAddress,
		ShippingCity:    input.ShippingCity,
		ShippingZipcode: input.ShippingZipcode,
		ShippingCountry: input.ShippingCountry,
		BillingAddress:  billingAddress,
		BillingCity:     billingCity,
		BillingZipcode:  billingZipcode,
		BillingCountry:  billingCountry,
		PaymentMethod:   input.PaymentMethod,
		Notes:           notes,
		Subtotal:        cart.Total,
		Total:           cart.Total,
	}

	tx := config.DB.Begin()
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create order"})
		return
	}

	for _, cartItem := range cart.Items {
		var productName, productSKU string
		var productID *uint

		if cartItem.ProductVariant != nil {
			productName = cartItem.ProductVariant.Name
			productSKU = cartItem.ProductVariant.SKU
			productID = &cartItem.ProductVariant.ProductID
			tx.Model(&models.ProductVariant{}).Where("id = ?", cartItem.ProductVariantID).
				UpdateColumn("stock", gorm.Expr("stock - ?", cartItem.Quantity))
		}

		total := cartItem.UnitPrice * float64(cartItem.Quantity)
		orderItem := models.OrderItem{
			OrderID:          order.ID,
			ProductID:        productID,
			ProductVariantID: cartItem.ProductVariantID,
			ProductName:      productName,
			Quantity:         cartItem.Quantity,
			UnitPrice:        cartItem.UnitPrice,
			Total:            total,
		}
		if productSKU != "" {
			orderItem.ProductSKU = &productSKU
		}
		tx.Create(&orderItem)
	}

	// Clear cart
	tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{})
	tx.Model(&cart).Updates(map[string]interface{}{"items_count": 0, "total": 0})
	tx.Commit()

	config.DB.Preload("Items").First(&order, order.ID)
	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Order created successfully. Status: %s", initialStatus),
		"data":    order,
	})
}

type UpdateOrderStatusInput struct {
	Status string `json:"status" binding:"required"`
}

func UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("order")
	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Order not found"})
		return
	}

	var input UpdateOrderStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	config.DB.Model(&order).Update("status", input.Status)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Order status updated to %s", input.Status), "data": order})
}
