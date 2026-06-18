package routes

import (
	"github.com/gin-gonic/gin"
	"shopcart-api/controllers"
	"shopcart-api/middlewares"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// =====================
		//  PUBLIC ROUTES
		// =====================
		api.POST("/register", controllers.Register)
		api.POST("/registerAdmin", controllers.RegisterAdmin)
		api.POST("/login", controllers.Login)
		api.POST("/refresh", controllers.RefreshToken)
		api.POST("/password/forgot", controllers.ForgotPassword)
		api.POST("/password/reset", controllers.ResetPassword)
		api.POST("/email/verify", controllers.VerifyEmail)
		api.GET("/email/verify", controllers.VerifyEmail)

		// Public Products
		api.GET("/products", controllers.GetProducts)
		api.GET("/products/featured", controllers.GetFeaturedProducts)
		api.GET("/products/id/:id", controllers.GetProductByID)
		api.GET("/products/:slug", controllers.GetProduct)

		// Use the existing /id/ prefix to avoid the collision.
		// Param name must match the others on this path segment (:id).
		api.GET("/products/id/:id/variants", controllers.GetProductVariants)

		// Public product reviews
		api.GET("/products/id/:id/reviews", controllers.GetProductReviews)

		// Public Categories
		api.GET("/categories", controllers.GetCategories)
		api.GET("/categories/:id", controllers.GetCategory)
		api.GET("/categories/:id/products", controllers.GetCategoryProducts)

		// =====================
		//  AUTHENTICATED ROUTES
		// =====================
		auth := api.Group("")
		auth.Use(middlewares.Auth())
		{
			auth.POST("/logout", controllers.Logout)
			auth.GET("/user", controllers.GetProfile)
			auth.POST("/email/resend", controllers.ResendVerification)

			// My profile
			auth.GET("/users/me", controllers.GetUserMe)
			auth.POST("/user/fcm-token", controllers.UpdateFcmToken)

			// Cart
			auth.GET("/cart", controllers.ShowCart)
			auth.POST("/cart", controllers.StoreCart)
			auth.POST("/cart/add", controllers.AddCartItem)
			auth.PUT("/cart/items/:cartItem", controllers.UpdateCartItem)
			auth.DELETE("/cart/items/:cartItem", controllers.RemoveCartItem)
			auth.DELETE("/cart/clear", controllers.ClearCart)
			auth.DELETE("/cart/user/:userId/empty", controllers.EmptyUserCart)

			// Orders (user's own)
			auth.GET("/orders", controllers.GetOrders)
			auth.POST("/orders", controllers.CreateOrder)
			auth.GET("/orders/:order", controllers.GetOrder)
			auth.GET("/orders/my", controllers.GetMyOrders)
			auth.POST("/orders/:order/cancel", controllers.CancelOrder)

			// Reviews
			auth.POST("/products/id/:id/reviews", controllers.CreateReview)
			auth.DELETE("/reviews/:id", controllers.DeleteReview)

			// Wishlist
			auth.GET("/wishlist", controllers.GetWishlist)
			auth.POST("/wishlist", controllers.AddWishlist)
			auth.DELETE("/wishlist/:product", controllers.RemoveWishlist)

			// Coupons
			auth.POST("/coupons/validate", controllers.ValidateCoupon)

			// Payment
			auth.POST("/payments/intent", controllers.CreatePaymentIntent)
			auth.POST("/payments", controllers.StorePayment)

			// Delivery user routes (DELIVERY role)
			auth.GET("/deliveries/my", controllers.GetMyDeliveries)
			auth.GET("/deliveries/history", controllers.GetDeliveryHistory)
			auth.PUT("/deliveries/:order/status", controllers.UpdateDeliveryStatus)
			auth.POST("/deliveries/location", controllers.UpdateDeliveryLocation)
			auth.POST("/deliveries/:order/proof", controllers.UploadProof)
			auth.GET("/deliveries/:order/proof", controllers.GetProof)

			// =====================
			//  PRODUCT MANAGEMENT ROUTES (ADMIN, SUPERADMIN, VENDOR)
			// =====================
			productMgmt := auth.Group("")
			productMgmt.Use(middlewares.ProductManagement())
			{
				productMgmt.POST("/products", controllers.CreateProduct)
				productMgmt.PUT("/products/:id", controllers.UpdateProduct)
				productMgmt.DELETE("/products/:id", controllers.DeleteProduct)
				productMgmt.GET("/products/vendor/stats", controllers.GetVendorStats)
				productMgmt.GET("/products/vendor/my-products", controllers.GetMyProducts)
			}

			// =====================
			//  MANAGEMENT ROUTES (ADMIN, SUPERADMIN, MANAGER, SUPERVISOR)
			// =====================
			mgmt := auth.Group("")
			mgmt.Use(middlewares.Management())
			{
				// Categories management
				mgmt.POST("/categories", controllers.CreateCategory)
				mgmt.PUT("/categories/:id", controllers.UpdateCategory)
				mgmt.DELETE("/categories/:id", controllers.DeleteCategory)

				// ProductVariant management
				mgmt.POST("/products/:product_id/variants", controllers.CreateProductVariant)
				mgmt.PUT("/variants/:variant_id", controllers.UpdateProductVariant)
				mgmt.DELETE("/variants/:variant_id", controllers.DeleteProductVariant)

				// Orders management
				mgmt.PUT("/orders/:order/status", controllers.UpdateOrderStatus)

				// Coupons management
				mgmt.GET("/coupons", controllers.ListCoupons)
				mgmt.POST("/coupons", controllers.CreateCoupon)
				mgmt.DELETE("/coupons/:id", controllers.DeleteCoupon)

				// User management
				mgmt.GET("/users", controllers.ListUsers)
				mgmt.POST("/users", controllers.CreateUser)
				mgmt.GET("/users/stats", controllers.GetUserStats)
				mgmt.GET("/users/:user", controllers.GetUser)
				mgmt.PUT("/users/:user", controllers.UpdateUser)
				mgmt.DELETE("/users/:user", controllers.DeleteUser)

				// Dashboard
				mgmt.GET("/dashboard/kpis", controllers.GetKpis)
				mgmt.GET("/dashboard/sales-over-time", controllers.GetSalesOverTime)
				mgmt.GET("/dashboard/top-products", controllers.GetTopProducts)
				mgmt.GET("/dashboard/order-status-distribution", controllers.GetOrderStatusDistribution)

				// Delivery Management
				mgmt.GET("/deliveries/pending", controllers.GetPendingDeliveries)
				mgmt.POST("/deliveries/:order/assign", controllers.AssignDelivery)
				mgmt.GET("/deliveries/live/map", controllers.GetLiveLocations)
			}
		}
	}
}
