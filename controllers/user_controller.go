package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"shopcart-api/config"
	"shopcart-api/models"
)

// ===== USER MANAGEMENT CONTROLLER =====

type CreateUserInput struct {
	Name                 string `json:"name" binding:"required"`
	Email                string `json:"email" binding:"required,email"`
	Password             string `json:"password" binding:"required,min=8"`
	PasswordConfirmation string `json:"password_confirmation" binding:"required,eqfield=Password"`
	Role                 string `json:"role" binding:"required"`
	Phone                string `json:"phone"`
	Address              string `json:"address"`
}

var allRoles = []string{
	models.RoleCustomer, models.RoleAdmin, models.RoleVendor,
	models.RoleDelivery, models.RoleManager, models.RoleSupervisor,
}

func isValidRole(role string) bool {
	for _, r := range allRoles {
		if r == role {
			return true
		}
	}
	return false
}

func ListUsers(c *gin.Context) {
	var users []models.User
	query := config.DB.Order("created_at DESC")
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}
	query.Find(&users)
	c.JSON(http.StatusOK, gin.H{"message": "Users retrieved successfully", "data": users})
}

func CreateUser(c *gin.Context) {
	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	if !isValidRole(input.Role) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid role"})
		return
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPwd),
		Role:     input.Role,
	}
	if input.Phone != "" {
		user.Phone = &input.Phone
	}
	if input.Address != "" {
		user.Address = &input.Address
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "data": user})
}

func GetUser(c *gin.Context) {
	var user models.User
	if err := config.DB.First(&user, c.Param("user")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User retrieved successfully", "data": user})
}

func GetUserMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User
	config.DB.First(&user, userID.(uint))
	c.JSON(http.StatusOK, gin.H{"message": "Profile retrieved successfully", "data": user})
}

func UpdateUser(c *gin.Context) {
	var user models.User
	if err := config.DB.First(&user, c.Param("user")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	updates := map[string]interface{}{}
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Role     string `json:"role"`
		Password string `json:"password"`
		Phone    string `json:"phone"`
		Address  string `json:"address"`
	}
	c.ShouldBindJSON(&body)
	if body.Name != "" {
		updates["name"] = body.Name
	}
	if body.Email != "" {
		updates["email"] = body.Email
	}
	if body.Role != "" && isValidRole(body.Role) {
		updates["role"] = body.Role
	}
	if body.Password != "" {
		pwd, _ := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		updates["password"] = string(pwd)
	}
	if body.Phone != "" {
		updates["phone"] = body.Phone
	}
	if body.Address != "" {
		updates["address"] = body.Address
	}

	config.DB.Model(&user).Updates(updates)
	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully", "data": user})
}

func DeleteUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User
	if err := config.DB.First(&user, c.Param("user")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}
	if user.ID == userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"message": "You cannot delete your own account"})
		return
	}
	config.DB.Delete(&user)
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func GetUserStats(c *gin.Context) {
	stats := gin.H{}
	for _, role := range allRoles {
		var count int64
		config.DB.Model(&models.User{}).Where("role = ?", role).Count(&count)
		stats[role] = count
	}
	var total int64
	config.DB.Model(&models.User{}).Count(&total)
	stats["total"] = total
	c.JSON(http.StatusOK, gin.H{"message": "User statistics retrieved successfully", "data": stats})
}

func UpdateFcmToken(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var body struct {
		FcmToken string `json:"fcm_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	config.DB.Model(&models.User{}).Where("id = ?", userID.(uint)).Update("fcm_token", body.FcmToken)
	c.JSON(http.StatusOK, gin.H{"message": "FCM token updated successfully."})
}
