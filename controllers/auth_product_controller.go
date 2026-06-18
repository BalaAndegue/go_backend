package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

	// This public endpoint exists only to bootstrap the very first admin on a
	// fresh deployment. Once any admin exists, further admin accounts must be
	// created via the authenticated user-management endpoints, so we refuse to
	// avoid privilege escalation by anonymous callers.
	var adminCount int64
	if err := config.DB.Model(&models.User{}).
		Where("role IN ?", []string{models.RoleAdmin, models.RoleSuperAdmin}).
		Count(&adminCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify admin state"})
		return
	}
	if adminCount > 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin registration is disabled"})
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

	token, refresh, err := issueTokens(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	resp := gin.H{
		"message":       "Registered",
		"token":         token,
		"refresh_token": refresh,
		"user": gin.H{
			"id": user.ID, "name": user.Name, "email": user.Email,
			"role": user.Role, "phone": user.Phone, "address": user.Address,
		},
	}
	if raw, err := issueEmailVerifyToken(user.ID); err == nil && exposeTokens() {
		resp["verification_token"] = raw
	}
	c.JSON(http.StatusCreated, resp)
}

// issueTokens generates an access + refresh token pair for the given user.
func issueTokens(user *models.User) (access, refresh string, err error) {
	access, err = utils.GenerateToken(user.ID, user.Role, user.TokenVersion)
	if err != nil {
		return "", "", err
	}
	refresh, err = utils.GenerateRefreshToken(user.ID, user.TokenVersion)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
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

	token, refresh, err := issueTokens(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"token":         token,
		"refresh_token": refresh,
		"user": gin.H{
			"id": user.ID, "name": user.Name, "email": user.Email,
			"role": user.Role, "phone": user.Phone, "address": user.Address,
		},
	})
}

// RefreshToken exchanges a valid refresh token for a new access/refresh pair.
func RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	claims, err := utils.ValidateToken(input.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}
	if t, _ := claims["type"].(string); t != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not a refresh token"})
		return
	}

	userID := uint(claims["user_id"].(float64))
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	if ver, ok := claims["ver"].(float64); !ok || int(ver) != user.TokenVersion {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token revoked"})
		return
	}

	token, refresh, err := issueTokens(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "refresh_token": refresh})
}

// @Summary Déconnexion
// @Tags Auth
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /logout [post]
func Logout(c *gin.Context) {
	userID, _ := c.Get("user_id")
	// Bumping the token version invalidates every access/refresh token issued
	// before this logout.
	config.DB.Model(&models.User{}).Where("id = ?", userID.(uint)).
		UpdateColumn("token_version", gorm.Expr("token_version + 1"))
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
	cat := c.Query("category")
	search := c.Query("search")
	applyFilters := func(db *gorm.DB) *gorm.DB {
		db = db.Where("is_visible = ?", true)
		if cat != "" {
			db = db.Where("category_id = ?", cat)
		}
		if search != "" {
			db = db.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
		}
		return db
	}

	page, perPage, offset := paginationParams(c)
	var total int64
	applyFilters(config.DB.Model(&models.Product{})).Count(&total)

	var products []models.Product
	applyFilters(config.DB.Preload("Variants").Preload("Category")).
		Order("is_featured DESC, updated_at DESC").
		Limit(perPage).Offset(offset).Find(&products)

	c.JSON(http.StatusOK, gin.H{
		"status": "success", "message": "Products retrieved successfully",
		"code": 200, "data": products, "meta": paginationMeta(page, perPage, total),
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

	if imageURL, err := saveUploadedImage(c, "image", ""); err == nil {
		product.Image = &imageURL
	} else if err != errNoFile {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
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

	if imageURL, err := saveUploadedImage(c, "image", ""); err == nil {
		updates["image"] = imageURL
	} else if err != errNoFile {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
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
