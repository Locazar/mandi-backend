package routes

import (
	"github.com/gin-gonic/gin"
	handlerInterface "github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/middleware"
)

func UserRoutes(api *gin.RouterGroup, authHandler handlerInterface.AuthHandler, middleware middleware.Middleware,
	userHandler handlerInterface.UserHandler, cartHandler handlerInterface.CartHandler,
	productHandler handlerInterface.ProductHandler, paymentHandler handlerInterface.PaymentHandler,
	orderHandler handlerInterface.OrderHandler, couponHandler handlerInterface.CouponHandler,
	offerHandle handlerInterface.OfferHandler, stockHandler handlerInterface.StockHandler,
	branHandler handlerInterface.BrandHandler, notificationHandler handlerInterface.NotificationHandler,
) {

	auth := api.Group("/auth")
	{
		signup := auth.Group("/sign-up")
		{
			signup.POST("/", authHandler.UserSignUp)
			signup.POST("/verify", authHandler.UserSignUpVerify)
			signup.POST("/resend-otp", authHandler.UserLoginOtpSend)
			signup.POST("/email/otp/send", authHandler.UserLoginOtpSendEmail)
		}

		login := auth.Group("/sign-in")
		{
			login.POST("/", authHandler.UserLogin)
			login.POST("/otp/send", authHandler.UserLoginOtpSend)
			login.POST("/otp/verify", authHandler.UserLoginOtpVerify)
		}

		goath := auth.Group("/google-auth")
		{
			goath.GET("/", authHandler.UserGoogleAuthLoginPage)
			goath.GET("/initialize", authHandler.UserGoogleAuthInitialize)
			goath.GET("/callback", authHandler.UserGoogleAuthCallBack)
		}

		auth.POST("/renew-access-token", authHandler.UserRenewAccessToken())

		// api.POST("/logout")

	}

	api.Use(middleware.AuthenticateUser())
	{

		// api.POST("/logout", userHandler.UserLogout)

		product := api.Group("/products")
		{
			// product.GET("/", productHandler.GetAllProductsUser)

			productItem := product.Group("/items")
			{
				product.GET("/:product_item_id", productHandler.GetProductItemByID)
				productItem.GET("/", productHandler.GetAllProductItemsUser())
				productItem.GET("/:product_item_id", productHandler.GetProductItemByID)
				productItem.GET("/:product_item_id/filters", productHandler.FindProductItemFilters)
			}

			product.GET("/search", productHandler.SearchProducts)
			product.GET("/suggestions", productHandler.GetProductSearchSuggestions)
			product.GET("/filters", productHandler.GetProductSearchFilters)
			product.GET("/locations", productHandler.GetProductSearchLocations)
			product.GET("/radius", productHandler.GetProductsByRadius)
			product.GET("/nearby", productHandler.GetNearbyProductsByPincode)

			productViewed := product.Group("/viewed-products")
			{
				productViewed.GET("/:product_item_id/view-count", productHandler.GetProductItemViewCount)
			}
		}

		// 	// cart
		cart := api.Group("/carts")
		{
			cart.GET("/", cartHandler.GetCart)
			cart.POST("/:product_item_id", cartHandler.AddToCart)
			cart.PUT("/", cartHandler.UpdateCart)
			cart.DELETE("/:product_item_id", cartHandler.RemoveFromCart)

			cart.PATCH("/apply-coupon", couponHandler.ApplyCouponToCart)

			cart.GET("/checkout/payment-select-page", paymentHandler.CartOrderPaymentSelectPage)
			// 		cart.GET("/payment-methods", orderHandler.GetAllPaymentMethods)
			cart.POST("/place-order", orderHandler.SaveOrder)

			// 		//cart.GET("/checkout", userHandler.CheckOutCart, orderHandler.GetAllPaymentMethods)
			cart.POST("/place-order/cod", paymentHandler.PaymentCOD)

			// razorpay payment
			cart.POST("/place-order/razorpay-checkout", paymentHandler.RazorpayCheckout)
			cart.POST("/place-order/razorpay-verify", paymentHandler.RazorpayVerify)

			// 	stripe payment
			cart.POST("/place-order/stripe-checkout", paymentHandler.StripPaymentCheckout)
			cart.POST("/place-order/stripe-verify", paymentHandler.StripePaymentVeify)
		}

		// profile
		account := api.Group("/account")
		{
			account.GET("/", userHandler.GetProfile)
			account.PUT("/", userHandler.UpdateProfile)
			account.POST("upload-profile-image/:id", userHandler.UploadProfileImage)

			account.GET("/address", userHandler.GetAllAddresses) // to show all address and // show countries
			account.POST("/address", userHandler.SaveAddress)    // to add a new address
			account.PUT("/address", userHandler.UpdateAddress)   // to edit address
			// account.DELETE("/address", userHandler.DeleteAddress)

			//wishlist
			wishList := account.Group("/wishlist")
			{
				wishList.GET("/", userHandler.GetWishList)
				wishList.POST("/:product_item_id", userHandler.SaveToWishList)
				wishList.DELETE("/:product_item_id", userHandler.RemoveFromWishList)
			}

			wallet := account.Group("/wallet")
			{
				wallet.GET("/", orderHandler.GetUserWallet)
				wallet.GET("/transactions", orderHandler.GetUserWalletTransactions)
			}

			coupons := account.Group("/coupons")
			{
				coupons.GET("/", couponHandler.GetAllCouponsForUser)
			}
		}

		paymentMethod := api.Group("/payment-methods")
		{
			paymentMethod.GET("/", paymentHandler.GetAllPaymentMethodsUser())
		}

		// 	// order
		orders := api.Group("/orders")
		{
			orders.GET("/", orderHandler.GetUserOrder)                               // get all order list for user
			orders.GET("/:shop_order_id/items", orderHandler.GetAllOrderItemsUser()) //get order items for specific order

			orders.POST("/return", orderHandler.SubmitReturnRequest)
			orders.POST("/:shop_order_id/cancel", orderHandler.CancelOrder) // cancel an order
		}

		// Product Search
		search := api.Group("/search")
		{
			search.GET("/", productHandler.SearchProducts)
			search.GET("/suggestions", productHandler.GetProductSearchSuggestions)
			search.GET("/filters", productHandler.GetProductSearchFilters)
			search.GET("/locations", productHandler.GetProductSearchLocations)
		}

		// Shop Search - Unified endpoint supporting: name search, geolocation (lat+lng+radius), and pincode filtering
		shop := api.Group("/shop")
		{
			shop.GET("/search", userHandler.SearchShopList)
		}

		// Shop by Category
		category := api.Group("/categories")
		{
			category.GET("/", productHandler.GetAllCategories)
			category.GET("/:category_id/products", productHandler.GetProductsByCategory)
			category.GET("/:category_id/product-items", userHandler.GetProductItemsByCategory)
		}

		// Sub-categories product items
		subCategory := api.Group("/sub-categories")
		{
			subCategory.GET("/:sub_category_id/product-items", userHandler.GetProductItemsBySubCategory)
		}

		// Shops (by admin id) product items
		shops := api.Group("/shops")
		{
			shops.GET("/:admin_id/products", userHandler.GetProductItemsByShop)
		}

		// Shop by Brand
		brand := api.Group("/brands")
		{
			brand.GET("/", productHandler.GetAllBrands)
			brand.GET("/:brand_id/products", productHandler.GetProductsByBrand)
		}

		// Shop by Offers
		offer := api.Group("/offers")
		{
			offer.GET("/", offerHandle.GetAllOffers)                 // get all offers
			offer.GET("/category", offerHandle.GetAllCategoryOffers) // to get all offers of categories
			offer.GET("/active", offerHandle.GetActiveOffers)        // get active offers

		}

		// Shop Search Filters
		filters := api.Group("/filters")
		{
			filters.GET("/categories", productHandler.GetCategoryFilters)
			filters.GET("/brands", productHandler.GetBrandFilters)
			filters.GET("/location", productHandler.GetLocationFilter)
		}

		// Shop by Location
		location := api.Group("/locations")
		{
			location.GET("/", productHandler.GetProductsByLocation)
			location.GET("/areas", productHandler.GetAllAreas)
			location.GET("/cities", productHandler.GetAllCities)
			location.GET("/states", productHandler.GetAllStates)
			location.GET("/countries", productHandler.GetAllCountries)
			location.GET("/pincodes", productHandler.GetAllPincodes)
			location.GET("/states/:state_id/cities", productHandler.GetCitiesByState)
			location.GET("/cities/:city_id/areas", productHandler.GetAreasByCity)
			location.GET("/areas/:area_id/pincodes", productHandler.GetPincodesByArea)
			location.GET("/pincodes/:pincode_id/location", productHandler.GetLocationByPincode)
		}

		notification := api.Group("/notifications")
		{
			notification.POST("/", notificationHandler.SaveNotification)
			notification.GET("/", notificationHandler.GetNotificationsBy)
			notification.PUT("/:notification_id/read", notificationHandler.MarkNotificationAsRead)
			notification.POST("/generateFCMToken", notificationHandler.GenerateFCMToken)
		}

		feedback := api.Group("/feedback")
		{
			feedback.POST("/shop", orderHandler.SubmitShoppingFeedback)
		}

		viewedProducts := api.Group("/viewed-products")
		{
			viewedProducts.POST("/:product_item_id", productHandler.IncrementProductItemViewCount)
			viewedProducts.GET("/:product_item_id/view-count", productHandler.GetProductItemViewCount)
		}

		department := api.Group("/departments")
		{
			department.GET("/", productHandler.GetAllDepartments)

			category := department.Group("/:department_id/categories")
			{
				category.GET("/", productHandler.GetAllCategoriesByDepartmentID)
				subCategory := category.Group("/:category_id/sub-categories")
				{
					subCategory.GET("/", productHandler.GetAllSubCategoriesByCategoryID)
					subCategory.GET("/:category_id/subcategories", productHandler.GetAllSubCategoriesByCategoryID)
				}
			}
		}

		// Job search
		// jobs := api.Group("/jobs")
		// {
		// 	jobs.GET("/", userHandler.GetAllJobs)
		// 	jobs.POST("/apply/:job_id", userHandler.ApplyToJob)
		// 	jobs.GET("/applications", userHandler.GetUserJobApplications)
		// 	jobs.DELETE("/applications/:application_id", userHandler.DeleteJobApplication)
		// 	jobs.GET("/suggestions", userHandler.GetJobSearchSuggestions)
		// 	jobs.GET("/filters", userHandler.GetJobSearchFilters)
		// 	jobs.GET("/search", userHandler.SearchJobs)
		// 	jobs.GET("/locations", userHandler.GetJobSearchLocations)
		// }

		// //Job Categories
		// jobCategories := api.Group("/job-categories")
		// {
		// 	jobCategories.GET("/", userHandler.GetAllJobCategories)
		// 	jobCategories.GET("/:category_id/jobs", userHandler.GetJobsByCategory)
		// 	jobCategories.GET("/:category_id/subcategories", userHandler.GetJobSubCategories)
		// 	jobCategories.GET("/subcategories/:subcategory_id/jobs", userHandler.GetJobsBySubCategory)
		// 	jobCategories.GET("/filters", userHandler.GetJobCategoryFilters)
		// 	jobCategories.GET("/locations", userHandler.GetJobCategoryLocations)
		// 	jobCategories.GET("/search", userHandler.SearchJobsInCategory)
		// }

		// Shop offers
		api.GET("/shop-offers", offerHandle.GetShopOffers)
	}
}
