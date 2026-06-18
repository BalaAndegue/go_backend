package controllers

import (
	"math"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"shopcart-api/models"
)

// getenvDefault returns the env var value or def when it is unset.
func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// envFloat reads a float env var, falling back to def when unset/invalid.
func envFloat(key string, def float64) float64 {
	if v, err := strconv.ParseFloat(os.Getenv(key), 64); err == nil {
		return v
	}
	return def
}

// round2 rounds a monetary amount to two decimals.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// computeOrderCharges derives shipping, tax and grand total from a subtotal.
// All rates default to 0 (no change in behaviour) and are configurable via env:
//   - TAX_RATE: fraction applied to subtotal (e.g. 0.20 for 20%)
//   - SHIPPING_FEE: flat shipping fee
//   - FREE_SHIPPING_THRESHOLD: subtotal at/above which shipping is free (0 = off)
func computeOrderCharges(subtotal float64) (shipping, tax, total float64) {
	taxRate := envFloat("TAX_RATE", 0)
	fee := envFloat("SHIPPING_FEE", 0)
	threshold := envFloat("FREE_SHIPPING_THRESHOLD", 0)

	shipping = fee
	if threshold > 0 && subtotal >= threshold {
		shipping = 0
	}
	tax = round2(subtotal * taxRate)
	total = round2(subtotal + shipping + tax)
	return shipping, tax, total
}

// canActOnOrder reports whether the authenticated caller may act on the given
// order: either they own it, or they hold a management role.
func canActOnOrder(c *gin.Context, order *models.Order) bool {
	user, ok := currentUser(c)
	if !ok {
		return false
	}
	return order.UserID == user.ID || user.IsManagement()
}

// paginationParams parses page / per_page query parameters with sane defaults
// and bounds, returning the page, per-page size, and the SQL OFFSET.
func paginationParams(c *gin.Context) (page, perPage, offset int) {
	page = 1
	perPage = 20
	if v, err := strconv.Atoi(c.Query("page")); err == nil && v > 0 {
		page = v
	}
	if v, err := strconv.Atoi(c.Query("per_page")); err == nil && v > 0 {
		perPage = v
	}
	if perPage > 100 {
		perPage = 100
	}
	offset = (page - 1) * perPage
	return page, perPage, offset
}

// paginationMeta builds the standard pagination metadata block.
func paginationMeta(page, perPage int, total int64) gin.H {
	lastPage := int((total + int64(perPage) - 1) / int64(perPage))
	if lastPage < 1 {
		lastPage = 1
	}
	return gin.H{
		"current_page": page,
		"per_page":     perPage,
		"total":        total,
		"last_page":    lastPage,
	}
}
