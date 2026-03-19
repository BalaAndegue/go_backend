package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"shopcart-api/config"
	"shopcart-api/models"
)

// ===== DASHBOARD CONTROLLER =====

func checkDashboardAccess(c *gin.Context) (*models.User, bool) {
	userID, _ := c.Get("user_id")
	var user models.User
	config.DB.First(&user, userID.(uint))
	if !user.IsManagement() {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied. Management role required."})
		return nil, false
	}
	return &user, true
}

func GetKpis(c *gin.Context) {
	if _, ok := checkDashboardAccess(c); !ok {
		return
	}

	var (
		totalOrders, deliveredOrders, readyToShip, inDelivery, successDeliveries, failedDeliveries int64
		totalRevenue                                                                                float64
	)

	config.DB.Model(&models.Order{}).Count(&totalOrders)
	config.DB.Model(&models.Order{}).Where("status IN ?", []string{"PAID", "DELIVERED"}).Select("COALESCE(SUM(total), 0)").Scan(&totalRevenue)
	config.DB.Model(&models.Order{}).Where("status = ?", "DELIVERED").Count(&deliveredOrders)
	config.DB.Model(&models.Order{}).Where("status = ?", "PAID").Count(&readyToShip)
	config.DB.Model(&models.Order{}).Where("status = ?", "IN_DELIVERY").Count(&inDelivery)
	config.DB.Model(&models.Order{}).Where("status = ?", "DELIVERED").Count(&successDeliveries)
	config.DB.Model(&models.Order{}).Where("status = ?", "FAILED").Count(&failedDeliveries)

	var activeUsers int64
	config.DB.Model(&models.User{}).Where("role = ?", models.RoleCustomer).Count(&activeUsers)

	deliveryRate := 0.0
	if totalOrders > 0 {
		deliveryRate = float64(deliveredOrders) / float64(totalOrders)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "KPIs retrieved successfully",
		"data": gin.H{
			"total_revenue": totalRevenue, "total_orders": totalOrders,
			"active_users": activeUsers, "delivery_rate": deliveryRate,
			"orders_ready_to_ship": readyToShip, "deliveries_in_progress": inDelivery,
			"deliveries_successful": successDeliveries, "deliveries_failed": failedDeliveries,
		},
	})
}

func GetSalesOverTime(c *gin.Context) {
	if _, ok := checkDashboardAccess(c); !ok {
		return
	}

	days := 30
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil {
			days = v
		}
	}

	startDate := time.Now().AddDate(0, 0, -days)

	type SalesData struct {
		Date       string  `json:"date"`
		TotalSales float64 `json:"total_sales"`
	}
	var sales []SalesData
	config.DB.Model(&models.Order{}).
		Select("CAST(created_at AS DATE) as date, COALESCE(SUM(total), 0) as total_sales").
		Where("status IN ? AND created_at >= ?", []string{"DELIVERED", "PAID"}, startDate).
		Group("CAST(created_at AS DATE)").
		Order("date ASC").
		Scan(&sales)

	c.JSON(http.StatusOK, gin.H{"message": "Sales data retrieved successfully", "data": sales})
}

func GetTopProducts(c *gin.Context) {
	if _, ok := checkDashboardAccess(c); !ok {
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}

	type TopProduct struct {
		ProductName   string  `json:"product_name"`
		TotalQuantity int     `json:"total_quantity"`
		TotalRevenue  float64 `json:"total_revenue"`
	}
	var topProducts []TopProduct
	config.DB.Table("order_items").
		Select("order_items.product_name, COALESCE(SUM(order_items.quantity), 0) as total_quantity, COALESCE(SUM(order_items.total), 0) as total_revenue").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.status IN ?", []string{"PAID", "DELIVERED"}).
		Group("order_items.product_name").
		Order("total_quantity DESC").
		Limit(limit).
		Scan(&topProducts)

	c.JSON(http.StatusOK, gin.H{"message": "Top selling products retrieved successfully", "data": topProducts})
}

func GetOrderStatusDistribution(c *gin.Context) {
	if _, ok := checkDashboardAccess(c); !ok {
		return
	}

	type StatusCount struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	var distribution []StatusCount
	config.DB.Model(&models.Order{}).Select("status, COUNT(*) as count").Group("status").Scan(&distribution)

	result := gin.H{}
	for _, s := range distribution {
		result[s.Status] = s.Count
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status distribution retrieved successfully", "data": result})
}

// ===== PAYMENT CONTROLLER =====

type CreatePaymentInput struct {
	OrderID uint    `json:"order_id" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
	Method  string  `json:"method" binding:"required"`
}

func CreatePaymentIntent(c *gin.Context) {
	var input CreatePaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	var order models.Order
	if err := config.DB.First(&order, input.OrderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Payment intent stub (real integration with Stripe would go here)
	c.JSON(http.StatusOK, gin.H{
		"message": "Payment intent created",
		"data": gin.H{
			"client_secret": "pi_stub_" + strconv.FormatUint(uint64(input.OrderID), 10),
			"amount":        input.Amount,
			"currency":      "usd",
		},
	})
}

func StorePayment(c *gin.Context) {
	var input struct {
		OrderID       uint    `json:"order_id" binding:"required"`
		Amount        float64 `json:"amount" binding:"required"`
		Method        string  `json:"method" binding:"required"`
		TransactionID string  `json:"transaction_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	payment := models.Payment{
		OrderID: input.OrderID,
		Amount:  input.Amount,
		Method:  input.Method,
		Status:  "PAID",
	}
	if input.TransactionID != "" {
		payment.TransactionID = &input.TransactionID
	}

	if err := config.DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not store payment"})
		return
	}

	// Update order status to PAID
	config.DB.Model(&models.Order{}).Where("id = ?", input.OrderID).Update("status", models.StatusPaid)

	c.JSON(http.StatusCreated, gin.H{"message": "Payment registered", "data": payment})
}

// suppress unused import
var _ = gorm.ErrRecordNotFound
