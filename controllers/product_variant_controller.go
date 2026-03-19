package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"shopcart-api/config"
	"shopcart-api/models"
	"shopcart-api/utils"
)

// ===== PRODUCT VARIANT CONTROLLER =====

func GetProductVariants(c *gin.Context) {
	productID := c.Param("product_id")
	var product models.Product
	if err := config.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Product not found", "code": 404})
		return
	}
	var variants []models.ProductVariant
	config.DB.Where("\"productId\" = ?", productID).Find(&variants)
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": variants, "message": "Variants retrieved", "code": 200})
}

func CreateProductVariant(c *gin.Context) {
	productID := c.Param("product_id")
	var product models.Product
	if err := config.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Product not found", "code": 404})
		return
	}

	name := c.PostForm("name")
	sku := c.PostForm("sku")
	color := c.PostForm("color")
	priceStr := c.PostForm("price")
	stockStr := c.PostForm("stock")
	attributesStr := c.PostForm("attributes")

	if name == "" || sku == "" || priceStr == "" || stockStr == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "name, sku, price, stock are required"})
		return
	}

	var price float64
	var stock int
	fmt.Sscanf(priceStr, "%f", &price)
	fmt.Sscanf(stockStr, "%d", &stock)

	variantIDUint, _ := fmt.Sscanf(productID, "%d", new(uint))
	_ = variantIDUint

	var productIDUint uint
	fmt.Sscanf(productID, "%d", &productIDUint)

	variant := models.ProductVariant{
		ProductID: productIDUint,
		Name:      name,
		SKU:       sku,
		Price:     price,
		Stock:     stock,
	}
	if color != "" {
		variant.Color = &color
	}
	if attributesStr != "" {
		var attrs models.JSONB
		if err := json.Unmarshal([]byte(attributesStr), &attrs); err == nil {
			variant.Attributes = attrs
		}
	}

	if file, err := c.FormFile("image"); err == nil {
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
		uploadPath := filepath.Join("uploads", filename)
		if err := c.SaveUploadedFile(file, uploadPath); err == nil {
			imageURL := "http://" + c.Request.Host + "/uploads/" + filename
			variant.Image = &imageURL
		}
	}

	if err := config.DB.Create(&variant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create variant"})
		return
	}

	// Update parent product stock total
	var totalStock int64
	config.DB.Model(&models.ProductVariant{}).Where("\"productId\" = ?", productIDUint).Select("COALESCE(SUM(stock), 0)").Scan(&totalStock)
	config.DB.Model(&product).Update("stock", totalStock)

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": variant, "message": "Variant created successfully", "code": 201})
}

func UpdateProductVariant(c *gin.Context) {
	variantID := c.Param("variant_id")
	var variant models.ProductVariant
	if err := config.DB.First(&variant, variantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Variant not found", "code": 404})
		return
	}

	updates := map[string]interface{}{}
	if name := c.PostForm("name"); name != "" {
		updates["name"] = name
	}
	if sku := c.PostForm("sku"); sku != "" {
		updates["sku"] = utils.GenerateSlug(sku)
	}
	if priceStr := c.PostForm("price"); priceStr != "" {
		var price float64
		fmt.Sscanf(priceStr, "%f", &price)
		updates["price"] = price
	}
	if stockStr := c.PostForm("stock"); stockStr != "" {
		var stock int
		fmt.Sscanf(stockStr, "%d", &stock)
		updates["stock"] = stock
	}
	if color := c.PostForm("color"); color != "" {
		updates["color"] = color
	}
	if file, err := c.FormFile("image"); err == nil {
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
		uploadPath := filepath.Join("uploads", filename)
		if err := c.SaveUploadedFile(file, uploadPath); err == nil {
			imageURL := "http://" + c.Request.Host + "/uploads/" + filename
			updates["image"] = imageURL
		}
	}

	config.DB.Model(&variant).Updates(updates)

	// Update parent product stock
	var totalStock int64
	config.DB.Model(&models.ProductVariant{}).Where("\"productId\" = ?", variant.ProductID).Select("COALESCE(SUM(stock), 0)").Scan(&totalStock)
	config.DB.Model(&models.Product{}).Where("id = ?", variant.ProductID).Update("stock", totalStock)

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": variant, "message": "Variant updated", "code": 200})
}

func DeleteProductVariant(c *gin.Context) {
	variantID := c.Param("variant_id")
	var variant models.ProductVariant
	if err := config.DB.First(&variant, variantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Variant not found", "code": 404})
		return
	}

	productID := variant.ProductID
	config.DB.Delete(&variant)

	var totalStock int64
	config.DB.Model(&models.ProductVariant{}).Where("\"productId\" = ?", productID).Select("COALESCE(SUM(stock), 0)").Scan(&totalStock)
	config.DB.Model(&models.Product{}).Where("id = ?", productID).Update("stock", totalStock)

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Variant deleted", "code": 200})
}
