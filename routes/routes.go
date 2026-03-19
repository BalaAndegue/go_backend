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

		// Public Products
		api.GET("/products", controllers.GetProducts)
		api.GET("/products/featured", controllers.GetFeaturedProducts)
		api.GET("/products/id/:id", controllers.GetProductByID)
		api.GET("/products/:slug", controllers.GetProduct)
		api.GET("/products/:product_id/variants", controllers.GetProductVariants)

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
			//  MANAGEMENT ROUTES (ADMIN, MANAGER, SUPERVISOR)
			// =====================
			mgmt := auth.Group("")
			mgmt.Use(middlewares.Management())
			{
				// Products management
				mgmt.POST("/products", controllers.CreateProduct)
				mgmt.PUT("/products/:id", controllers.UpdateProduct)
				mgmt.DELETE("/products/:id", controllers.DeleteProduct)
				mgmt.GET("/products/vendor/stats", controllers.GetVendorStats)
				mgmt.GET("/products/vendor/my-products", controllers.GetMyProducts)

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
