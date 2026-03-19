package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"shopcart-api/config"
	"shopcart-api/models"
	"shopcart-api/utils"
)

// ===== AUTH CONTROLLER =====

type RegisterInput struct {
	Name                 string `json:"name" binding:"required"`
	Email                string `json:"email" binding:"required,email"`
	Password             string `json:"password" binding:"required,min=8"`
	PasswordConfirmation string `json:"password_confirmation" binding:"required,eqfield=Password"`
	Phone                string `json:"phone"`
	Address              string `json:"address"`
}

// @Summary Inscription utilisateur (CLIENT)
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body RegisterInput true "Données d'inscription"
// @Success 201 {object} map[string]interface{}
// @Failure 422 {object} map[string]interface{}
// @Router /register [post]
func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	createUser(c, input, models.RoleCustomer)
}

// @Summary Inscription administrateur
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body RegisterInput true "Données d'inscription admin"
// @Success 201 {object} map[string]interface{}
// @Failure 422 {object} map[string]interface{}
// @Router /registerAdmin [post]
func RegisterAdmin(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	createUser(c, input, models.RoleAdmin)
}

func createUser(c *gin.Context, input RegisterInput, role string) {
	var existing models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Email already exists"})
		return
	}

	user := models.User{
		Name:  input.Name,
		Email: input.Email,
		Role:  role,
	}
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}
	user.Password = hashedPassword

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

	token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registered",
		"token":   token,
		"user": gin.H{
			"id": user.ID, "name": user.Name, "email": user.Email,
			"role": user.Role, "phone": user.Phone, "address": user.Address,
		},
	})
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// @Summary Connexion utilisateur
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body LoginInput true "Identifiants"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /login [post]
func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !utils.CheckPassword(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"id": user.ID, "name": user.Name, "email": user.Email,
			"role": user.Role, "phone": user.Phone, "address": user.Address,
		},
	})
}

// @Summary Déconnexion
// @Tags Auth
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /logout [post]
func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// @Summary Profil utilisateur connecté
// @Tags Auth
// @Security BearerAuth
// @Success 200 {object} models.User
// @Router /user [get]
func GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// ===== PRODUCT CONTROLLER =====

func generateUniqueSlug(name string) string {
	slug := utils.GenerateSlug(name)
	original := slug
	count := 1
	for {
		var product models.Product
		if err := config.DB.Where("slug = ?", slug).First(&product).Error; err != nil {
			break
		}
		slug = fmt.Sprintf("%s-%d", original, count)
		count++
	}
	return slug
}

// @Summary Liste tous les produits visibles
// @Tags Products
// @Param search query string false "Recherche"
// @Param category query int false "Filtrer par catégorie ID"
// @Success 200 {object} map[string]interface{}
// @Router /products [get]
func GetProducts(c *gin.Context) {
	query := config.DB.Preload("Variants").Preload("Category").Where("is_visible = ?", true)

	if cat := c.Query("category"); cat != "" {
		query = query.Where("category_id = ?", cat)
	}
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var products []models.Product
	query.Order("is_featured DESC, updated_at DESC").Find(&products)
	c.JSON(http.StatusOK, gin.H{
		"status": "success", "message": "Products retrieved successfully",
		"code": 200, "data": products,
	})
}

// @Summary Produits mis en avant
// @Tags Products
// @Param limit query int false "Limite"
// @Success 200 {object} map[string]interface{}
// @Router /products/featured [get]
func GetFeaturedProducts(c *gin.Context) {
	limit := 8
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	var products []models.Product
	config.DB.Preload("Variants").Preload("Category").
		Where("is_visible = ? AND is_featured = ?", true, true).
		Order("updated_at DESC").Limit(limit).Find(&products)
	c.JSON(http.StatusOK, gin.H{
		"status": "success", "message": "Featured products retrieved successfully",
		"code": 200, "data": products,
	})
}

// @Summary Produit par slug
// @Tags Products
// @Param slug path string true "Slug"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /products/{slug} [get]
func GetProduct(c *gin.Context) {
	slug := c.Param("slug")
	var product models.Product
	if err := config.DB.Preload("Variants").Preload("Category").
		Where("slug = ? AND is_visible = ?", slug, true).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Product not found", "code": 404})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": product, "code": 200})
}

// @Summary Produit par ID
// @Tags Products
// @Param id path int true "ID"
// @Success 200 {object} map[string]interface{}
// @Router /products/id/{id} [get]
func GetProductByID(c *gin.Context) {
	id := c.Param("id")
	var product models.Product
	if err := config.DB.Preload("Variants").Preload("Category").
		Where("id = ? AND is_visible = ?", id, true).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Product not found", "code": 404})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": product, "code": 200})
}

type ProductInput struct {
	Name         string   `form:"name" binding:"required"`
	Description  string   `form:"description"`
	Price        float64  `form:"price" binding:"required"`
	ComparePrice *float64 `form:"compare_price"`
	Stock        int      `form:"stock" binding:"required"`
	SKU          string   `form:"sku"`
	CategoryID   uint     `form:"category_id" binding:"required"`
	IsVisible    bool     `form:"is_visible"`
	IsFeatured   bool     `form:"is_featured"`
}

// @Summary Créer un produit (Admin/Vendor)
// @Tags Products
// @Security BearerAuth
// @Accept multipart/form-data
// @Success 201 {object} map[string]interface{}
// @Failure 422 {object} map[string]interface{}
// @Router /products [post]
func CreateProduct(c *gin.Context) {
	role, _ := c.Get("role")
	roleStr := role.(string)
	if roleStr != models.RoleAdmin && roleStr != models.RoleVendor {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied"})
		return
	}

	var input ProductInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	product := models.Product{
		Name:         input.Name,
		Slug:         generateUniqueSlug(input.Name),
		Price:        input.Price,
		ComparePrice: input.ComparePrice,
		Stock:        input.Stock,
		CategoryID:   input.CategoryID,
		IsVisible:    input.IsVisible,
		IsFeatured:   input.IsFeatured,
	}
	if input.Description != "" {
		product.Description = &input.Description
	}
	if input.SKU != "" {
		product.SKU = &input.SKU
	}

	if file, err := c.FormFile("image"); err == nil {
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
		uploadPath := filepath.Join("uploads", filename)
		if err := c.SaveUploadedFile(file, uploadPath); err == nil {
			imageURL := "http://" + c.Request.Host + "/uploads/" + filename
			product.Image = &imageURL
		}
	}

	if err := config.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create product"})
		return
	}
	config.DB.Preload("Category").Preload("Variants").First(&product, product.ID)
	c.JSON(http.StatusCreated, gin.H{"status": "success", "code": 201, "message": "Product created successfully", "data": product})
}

// @Summary Mettre à jour un produit (Admin/Vendor)
// @Tags Products
// @Security BearerAuth
// @Accept multipart/form-data
// @Param id path int true "ID produit"
// @Success 200 {object} map[string]interface{}
// @Router /products/{id} [put]
func UpdateProduct(c *gin.Context) {
	role, _ := c.Get("role")
	roleStr := role.(string)
	if roleStr != models.RoleAdmin && roleStr != models.RoleVendor {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied"})
		return
	}

	id := c.Param("id")
	var product models.Product
	if err := config.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	updates := map[string]interface{}{}
	if name := c.PostForm("name"); name != "" {
		updates["name"] = name
		updates["slug"] = generateUniqueSlug(name)
	}
	if desc := c.PostForm("description"); desc != "" {
		updates["description"] = desc
	}
	if price := c.PostForm("price"); price != "" {
		if v, err := strconv.ParseFloat(price, 64); err == nil {
			updates["price"] = v
		}
	}
	if stock := c.PostForm("stock"); stock != "" {
		if v, err := strconv.Atoi(stock); err == nil {
			updates["stock"] = v
		}
	}
	if catID := c.PostForm("category_id"); catID != "" {
		if v, err := strconv.ParseUint(catID, 10, 64); err == nil {
			updates["category_id"] = uint(v)
		}
	}

	if file, err := c.FormFile("image"); err == nil {
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
		uploadPath := filepath.Join("uploads", filename)
		if err := c.SaveUploadedFile(file, uploadPath); err == nil {
			imageURL := "http://" + c.Request.Host + "/uploads/" + filename
			updates["image"] = imageURL
		}
	}

	config.DB.Model(&product).Updates(updates)
	config.DB.Preload("Category").Preload("Variants").First(&product, product.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully", "data": product})
}

// @Summary Supprimer un produit (Admin/Vendor)
// @Tags Products
// @Security BearerAuth
// @Param id path int true "ID produit"
// @Success 200 {object} map[string]interface{}
// @Router /products/{id} [delete]
func DeleteProduct(c *gin.Context) {
	role, _ := c.Get("role")
	roleStr := role.(string)
	if roleStr != models.RoleAdmin && roleStr != models.RoleVendor {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied"})
		return
	}

	id := c.Param("id")
	var product models.Product
	if err := config.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	config.DB.Delete(&product)
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Product deleted successfully", "code": 200})
}

// @Summary Mes produits (Admin/Vendor)
// @Tags Products
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /products/vendor/my-products [get]
func GetMyProducts(c *gin.Context) {
	role, _ := c.Get("role")
	roleStr := role.(string)
	if roleStr != models.RoleAdmin && roleStr != models.RoleVendor {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied"})
		return
	}
	var products []models.Product
	config.DB.Preload("Variants").Order("created_at DESC").Find(&products)
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": products, "message": "Products retrieved successfully", "code": 200})
}

// @Summary Statistiques produits (Admin/Vendor)
// @Tags Products
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /products/vendor/stats [get]
func GetVendorStats(c *gin.Context) {
	role, _ := c.Get("role")
	roleStr := role.(string)
	if roleStr != models.RoleAdmin && roleStr != models.RoleVendor {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied"})
		return
	}
	var total, visible, featured, outOfStock int64
	config.DB.Model(&models.Product{}).Count(&total)
	config.DB.Model(&models.Product{}).Where("is_visible = ?", true).Count(&visible)
	config.DB.Model(&models.Product{}).Where("is_featured = ?", true).Count(&featured)
	config.DB.Model(&models.Product{}).Where("stock = ?", 0).Count(&outOfStock)
	c.JSON(http.StatusOK, gin.H{
		"status": "success", "code": 200, "message": "Statistics retrieved successfully",
		"data": gin.H{
			"total_products": total, "visible_products": visible,
			"featured_products": featured, "out_of_stock": outOfStock,
		},
	})
}
