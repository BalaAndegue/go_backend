package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health is a simple liveness probe.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// APIDocs returns a lightweight, machine-readable listing of the API surface.
// (A full Swagger/OpenAPI UI requires the swag toolchain; this dependency-free
// endpoint documents the available routes grouped by area.)
func APIDocs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":    "ShopCart API",
		"version": "1.0.0",
		"groups": gin.H{
			"auth": []string{
				"POST /api/register", "POST /api/registerAdmin", "POST /api/login",
				"POST /api/refresh", "POST /api/logout", "GET /api/user",
				"POST /api/password/forgot", "POST /api/password/reset",
				"POST /api/email/verify", "POST /api/email/resend",
			},
			"products": []string{
				"GET /api/products", "GET /api/products/featured",
				"GET /api/products/id/:id", "GET /api/products/:slug",
				"GET /api/products/id/:id/variants", "GET /api/products/id/:id/reviews",
				"POST /api/products", "PUT /api/products/:id", "DELETE /api/products/:id",
			},
			"cart": []string{
				"GET /api/cart", "POST /api/cart/add", "PUT /api/cart/items/:cartItem",
				"DELETE /api/cart/items/:cartItem", "DELETE /api/cart/clear",
			},
			"orders": []string{
				"GET /api/orders", "POST /api/orders", "GET /api/orders/:order",
				"GET /api/orders/my", "POST /api/orders/:order/cancel",
				"PUT /api/orders/:order/status",
			},
			"payments": []string{
				"POST /api/payments/intent", "POST /api/payments", "POST /api/payments/webhook",
			},
			"reviews_wishlist": []string{
				"POST /api/products/id/:id/reviews", "DELETE /api/reviews/:id",
				"GET /api/wishlist", "POST /api/wishlist", "DELETE /api/wishlist/:product",
			},
			"coupons": []string{
				"POST /api/coupons/validate", "GET /api/coupons",
				"POST /api/coupons", "DELETE /api/coupons/:id",
			},
			"delivery": []string{
				"GET /api/deliveries/my", "GET /api/deliveries/history",
				"PUT /api/deliveries/:order/status", "POST /api/deliveries/location",
				"POST /api/deliveries/:order/proof", "GET /api/deliveries/:order/proof",
				"GET /api/deliveries/pending", "POST /api/deliveries/:order/assign",
				"GET /api/deliveries/live/map",
			},
			"dashboard": []string{
				"GET /api/dashboard/kpis", "GET /api/dashboard/sales-over-time",
				"GET /api/dashboard/top-products", "GET /api/dashboard/order-status-distribution",
			},
		},
	})
}
