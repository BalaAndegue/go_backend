package controllers

import (
	"crypto/rand"
	"fmt"
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
	CouponCode      string `json:"coupon_code"`
}

func generateOrderNumber() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		panic(err) // crypto/rand should never fail; fail loudly if it does
	}
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return fmt.Sprintf("ORD-%s-%s", time.Now().Format("20060102"), string(b))
}

func GetMyOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")
	status := c.Query("status")
	applyFilters := func(db *gorm.DB) *gorm.DB {
		db = db.Where("user_id = ?", userID.(uint))
		if status != "" {
			db = db.Where("status = ?", status)
		}
		return db
	}

	page, perPage, offset := paginationParams(c)
	var total int64
	applyFilters(config.DB.Model(&models.Order{})).Count(&total)

	var orders []models.Order
	applyFilters(config.DB.Preload("Items")).Order("created_at DESC").
		Limit(perPage).Offset(offset).Find(&orders)

	c.JSON(http.StatusOK, gin.H{
		"message": "Orders retrieved", "data": orders,
		"meta": paginationMeta(page, perPage, total),
	})
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

	subtotal := cart.Total

	// Apply an optional coupon to the subtotal before computing charges.
	var coupon *models.Coupon
	var discount float64
	if input.CouponCode != "" {
		var cp models.Coupon
		if err := config.DB.Where("code = ?", models.NormalizeCode(input.CouponCode)).First(&cp).Error; err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Invalid coupon code"})
			return
		}
		if ok, reason := cp.Validate(subtotal); !ok {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Coupon invalid: " + reason})
			return
		}
		coupon = &cp
		discount = round2(cp.DiscountFor(subtotal))
	}

	shipping, tax, total := computeOrderCharges(subtotal - discount)

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
		Subtotal:        subtotal,
		Discount:        discount,
		Shipping:        shipping,
		Tax:             tax,
		Total:           total,
	}
	if coupon != nil {
		code := coupon.Code
		order.CouponCode = &code
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
			// Conditional decrement: only succeeds if enough stock remains.
			// This guards against overselling under concurrent checkouts, since
			// the WHERE clause is evaluated atomically by the database.
			res := tx.Model(&models.ProductVariant{}).
				Where("id = ? AND stock >= ?", cartItem.ProductVariantID, cartItem.Quantity).
				UpdateColumn("stock", gorm.Expr("stock - ?", cartItem.Quantity))
			if res.Error != nil || res.RowsAffected == 0 {
				tx.Rollback()
				c.JSON(http.StatusUnprocessableEntity, gin.H{
					"message": fmt.Sprintf("Stock insuffisant pour la variante ID %d.", cartItem.ProductVariantID),
				})
				return
			}
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

	// Record coupon usage.
	if coupon != nil {
		tx.Model(&models.Coupon{}).Where("id = ?", coupon.ID).
			UpdateColumn("used_count", gorm.Expr("used_count + 1"))
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

// CancelOrder lets a customer cancel their own order (or management cancel any)
// while it is still cancellable, restocking variant inventory in a transaction.
func CancelOrder(c *gin.Context) {
	orderID := c.Param("order")
	var order models.Order
	if err := config.DB.Preload("Items").First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Order not found"})
		return
	}

	if !canActOnOrder(c, &order) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Unauthorized"})
		return
	}

	switch order.Status {
	case models.StatusPending, models.StatusPendingPayment, models.StatusPaid, models.StatusProcessing:
		// still cancellable
	default:
		c.JSON(http.StatusConflict, gin.H{"message": "Order can no longer be cancelled"})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range order.Items {
			if item.ProductVariantID != nil {
				if err := tx.Model(&models.ProductVariant{}).
					Where("id = ?", *item.ProductVariantID).
					UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
					return err
				}
			}
		}
		return tx.Model(&order).Update("status", models.StatusCancelled).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not cancel order"})
		return
	}

	notifyOrderStatus(order.ID, models.StatusCancelled)
	config.DB.Preload("Items").First(&order, order.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully", "data": order})
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
	if !models.IsValidOrderStatus(input.Status) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid order status"})
		return
	}

	config.DB.Model(&order).Update("status", input.Status)
	notifyOrderStatus(order.ID, input.Status)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Order status updated to %s", input.Status), "data": order})
}
