package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"shopcart-api/config"
	"shopcart-api/models"
	"shopcart-api/utils"
)

// ===== CATEGORY CONTROLLER =====

type CategoryInput struct {
	Name string `json:"name" binding:"required"`
}

func GetCategories(c *gin.Context) {
	var categories []models.Category
	config.DB.Find(&categories)
	c.JSON(http.StatusOK, categories)
}

func GetCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category
	if err := config.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}
	c.JSON(http.StatusOK, category)
}

func GetCategoryProducts(c *gin.Context) {
	id := c.Param("id")
	var products []models.Product
	config.DB.Preload("Variants").Preload("Category").
		Where("category_id = ? AND is_visible = ?", id, true).Find(&products)
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": products})
}

func CreateCategory(c *gin.Context) {
	var input CategoryInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	category := models.Category{Name: input.Name, Slug: utils.GenerateSlug(input.Name)}

	// Handle image upload
	if imageURL, err := saveUploadedImage(c, "image", ""); err == nil {
		category.Image = &imageURL
	} else if err != errNoFile {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create category"})
		return
	}
	c.JSON(http.StatusCreated, category)
}

func UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category
	if err := config.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	updates := map[string]interface{}{}
	if name := c.PostForm("name"); name != "" {
		updates["name"] = name
		updates["slug"] = utils.GenerateSlug(name)
	}
	if imageURL, err := saveUploadedImage(c, "image", ""); err == nil {
		updates["image"] = imageURL
	} else if err != errNoFile {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	config.DB.Model(&category).Updates(updates)
	c.JSON(http.StatusOK, category)
}

func DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category
	if err := config.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}
	config.DB.Delete(&category)
	c.Status(http.StatusNoContent)
}
