package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"shopcart-api/config"
	"shopcart-api/models"
)

// GetProductReviews lists a product's reviews and its average rating (public).
func GetProductReviews(c *gin.Context) {
	productID := c.Param("id")

	page, perPage, offset := paginationParams(c)
	var total int64
	config.DB.Model(&models.Review{}).Where("product_id = ?", productID).Count(&total)

	var reviews []models.Review
	config.DB.Preload("User").Where("product_id = ?", productID).
		Order("created_at DESC").Limit(perPage).Offset(offset).Find(&reviews)

	var avg float64
	config.DB.Model(&models.Review{}).Where("product_id = ?", productID).
		Select("COALESCE(AVG(rating), 0)").Scan(&avg)

	meta := paginationMeta(page, perPage, total)
	meta["average_rating"] = round2(avg)

	c.JSON(http.StatusOK, gin.H{
		"status": "success", "code": 200, "data": reviews, "meta": meta,
	})
}

type ReviewInput struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

// CreateReview creates or updates the caller's review for a product. A user has
// at most one review per product.
func CreateReview(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uint)

	var product models.Product
	if err := config.DB.First(&product, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Product not found", "code": 404})
		return
	}

	var input ReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	var review models.Review
	err := config.DB.Where("product_id = ? AND user_id = ?", product.ID, uid).First(&review).Error
	if err == nil {
		updates := map[string]interface{}{"rating": input.Rating, "comment": nil}
		if input.Comment != "" {
			updates["comment"] = input.Comment
		}
		config.DB.Model(&review).Updates(updates)
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Review updated", "data": review, "code": 200})
		return
	}

	review = models.Review{ProductID: product.ID, UserID: uid, Rating: input.Rating}
	if input.Comment != "" {
		review.Comment = &input.Comment
	}
	if err := config.DB.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create review"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "success", "message": "Review created", "data": review, "code": 201})
}

// DeleteReview removes a review; only its author or management may do so.
func DeleteReview(c *gin.Context) {
	var review models.Review
	if err := config.DB.First(&review, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Review not found"})
		return
	}

	user, ok := currentUser(c)
	if !ok || (review.UserID != user.ID && !user.IsManagement()) {
		c.JSON(http.StatusForbidden, gin.H{"message": "Unauthorized"})
		return
	}

	config.DB.Delete(&review)
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Review deleted", "code": 200})
}
