package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"shopcart-api/config"
	"shopcart-api/models"
)

// ===== DELIVERY CONTROLLER =====

func GetPendingDeliveries(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User
	config.DB.First(&user, userID.(uint))
	if !user.IsManagement() {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied. Management role required."})
		return
	}

	var orders []models.Order
	config.DB.Preload("Items").
		Where("delivery_user_id IS NULL AND status IN ?", []string{models.StatusPaid, models.StatusPendingPayment}).
		Order("created_at ASC").Find(&orders)

	c.JSON(http.StatusOK, gin.H{"message": "Pending orders retrieved successfully", "data": orders, "status": "success", "code": 200})
}

func AssignDelivery(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User
	config.DB.First(&user, userID.(uint))
	if !user.IsManagement() {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied. Management role required."})
		return
	}

	orderID := c.Param("order")
	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Order not found"})
		return
	}

	var body struct {
		DeliveryUserID uint `json:"delivery_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	var deliveryUser models.User
	if err := config.DB.Where("id = ? AND role = ?", body.DeliveryUserID, models.RoleDelivery).First(&deliveryUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Invalid delivery user ID or role."})
		return
	}

	config.DB.Model(&order).Updates(map[string]interface{}{
		"delivery_user_id": deliveryUser.ID,
		"status":           models.StatusAssigned,
	})
	config.DB.Preload("DeliveryUser").First(&order, order.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Order %d assigned to delivery user %s successfully", order.ID, deliveryUser.Name),
		"data": order, "status": "success", "code": 200,
	})
}

func GetMyDeliveries(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User
	config.DB.First(&user, userID.(uint))
	if !user.IsDelivery() {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied. Delivery role required."})
		return
	}

	var orders []models.Order
	config.DB.Preload("Items").
		Where("delivery_user_id = ? AND status NOT IN ?", userID.(uint), []string{models.StatusDelivered, models.StatusFailed}).
		Order("created_at ASC").Find(&orders)

	c.JSON(http.StatusOK, gin.H{"message": "Assigned deliveries retrieved successfully", "data": orders, "status": "success", "code": 200})
}

func GetDeliveryHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User
	config.DB.First(&user, userID.(uint))
	if !user.IsDelivery() {
		c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden. Only delivery personnel can access this resource."})
		return
	}

	var orders []models.Order
	config.DB.Preload("Items").
		Where("delivery_user_id = ? AND status IN ?", userID.(uint), []string{models.StatusDelivered, models.StatusFailed}).
		Order("updated_at DESC").Find(&orders)

	c.JSON(http.StatusOK, gin.H{"message": "Delivery history retrieved", "data": orders})
}

func UpdateDeliveryStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")
	orderID := c.Param("order")
	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Order not found"})
		return
	}

	var user models.User
	config.DB.First(&user, userID.(uint))
	if !user.IsDelivery() || order.DeliveryUserID == nil || *order.DeliveryUserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied. Not the assigned delivery person."})
		return
	}

	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	config.DB.Model(&order).Update("status", body.Status)
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Delivery status updated to %s for order %d", body.Status, order.ID),
		"data": order, "status": "success", "code": 200,
	})
}

func UpdateDeliveryLocation(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User
	config.DB.First(&user, userID.(uint))
	if !user.IsDelivery() {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied. Delivery role required."})
		return
	}

	var body struct {
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	uid := userID.(uint)
	var geo models.DeliveryGeolocation
	result := config.DB.Where("user_id = ?", uid).First(&geo)
	if result.Error != nil {
		geo = models.DeliveryGeolocation{UserID: uid, Latitude: body.Latitude, Longitude: body.Longitude}
		config.DB.Create(&geo)
	} else {
		config.DB.Model(&geo).Updates(map[string]interface{}{"latitude": body.Latitude, "longitude": body.Longitude})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Location updated successfully", "data": geo, "status": "success", "code": 200})
}

func GetLiveLocations(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User
	config.DB.First(&user, userID.(uint))
	if !user.IsManagement() {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied. Management role required."})
		return
	}

	var locations []models.DeliveryGeolocation
	config.DB.Preload("User").
		Joins("JOIN users ON users.id = delivery_geolocations.user_id AND users.role = ?", models.RoleDelivery).
		Find(&locations)

	c.JSON(http.StatusOK, gin.H{"message": "Live locations retrieved successfully", "data": locations, "status": "success", "code": 200})
}

func UploadProof(c *gin.Context) {
	userID, _ := c.Get("user_id")
	orderID := c.Param("order")
	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Order not found"})
		return
	}

	var user models.User
	config.DB.First(&user, userID.(uint))
	if !user.IsDelivery() || order.DeliveryUserID == nil || *order.DeliveryUserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied. Not the assigned delivery person."})
		return
	}

	if order.Status == models.StatusDelivered {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Order is already marked as delivered."})
		return
	}

	proofType := c.PostForm("proof_type")
	file, err := c.FormFile("proof_image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Image upload failed."})
		return
	}

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
	uploadPath := filepath.Join("uploads", "proofs", filename)
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Image upload failed."})
		return
	}
	proofURL := "http://" + c.Request.Host + "/uploads/proofs/" + filename
	config.DB.Model(&order).Updates(map[string]interface{}{
		"proof_path": proofURL, "proof_type": proofType, "status": models.StatusDelivered,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Proof of delivery uploaded successfully.", "data": order})
}

func GetProof(c *gin.Context) {
	userID, _ := c.Get("user_id")
	orderID := c.Param("order")
	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Order not found"})
		return
	}

	var user models.User
	config.DB.First(&user, userID.(uint))
	isManagement := user.IsManagement()
	isClient := order.UserID == userID.(uint)
	isAssigned := order.DeliveryUserID != nil && *order.DeliveryUserID == userID.(uint)

	if !isManagement && !isClient && !isAssigned {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied."})
		return
	}

	if order.ProofPath == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Delivery proof not yet available for this order."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Delivery proof path retrieved successfully.",
		"data": gin.H{"proof_url": *order.ProofPath, "proof_type": order.ProofType},
		"status": "success", "code": 200,
	})
}
