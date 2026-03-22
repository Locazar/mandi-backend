package routes

import (
	"github.com/gin-gonic/gin"
	handlerInterface "github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/middleware"
)

func AdminRoutes(api *gin.RouterGroup, authHandler handlerInterface.AuthHandler, middleware middleware.Middleware,
	adminHandler handlerInterface.AdminHandler, productHandler handlerInterface.ProductHandler,
	paymentHandler handlerInterface.PaymentHandler, orderHandler handlerInterface.OrderHandler,
	couponHandler handlerInterface.CouponHandler, offerHandler handlerInterface.OfferHandler,
	stockHandler handlerInterface.StockHandler, branHandler handlerInterface.BrandHandler,
	promotionHandler handlerInterface.PromotionHandler, fcmTokenHandler handlerInterface.FcmTokenHandler,

) {

	auth := api.Group("/auth")
	{
		login := auth.Group("/signin")
		{
			login.POST("/", authHandler.AdminLogin)
		}

		signup := auth.Group("/signup")
		{
			signup.POST("/", adminHandler.AdminSignUp)
			signup.POST("/verify", adminHandler.AdminSignUpVerify)
		}

		auth.POST("/renew-access-token", authHandler.AdminRenewAccessToken())

		// Admin logout route
		auth.POST("/logout", authHandler.AdminLogout)
	}

	// Profile endpoint (accessible without full authentication)
	api.GET("/profile", adminHandler.GetAdminWithShopVerificationByPhone)

	// Product details endpoint (public access)
	productDetails := api.Group("/product-details")
	{
		productDetails.GET("/details", adminHandler.GetAllProductDetails)
	}

	// Shop time management
	shop := api.Group("/shop")
	{
		shop.POST("/time/:shop_id", adminHandler.SetShopTime)
		shop.GET("/time/:shop_id", adminHandler.GetShopTime)
	}

	// Web views for admin interface
	web := api.Group("/web")
	{
		web.GET("/login", func(c *gin.Context) {
			c.HTML(200, "goauth.html", nil)
		})
		web.GET("/payment", func(c *gin.Context) {
			c.HTML(200, "paymentForm.html", nil)
		})
		web.GET("/dashboard", func(c *gin.Context) {
			c.HTML(200, "dashboard.html", nil)
		})
		web.GET("/ui", func(c *gin.Context) {
			c.HTML(200, "dashboard.html", nil)
		})
	}

	api.Use(middleware.AuthenticateAdmin())
	{
		// Common routes
		api.GET("/banner", offerHandler.GetBanners)
		// user side
		user := api.Group("/users")
		{
			user.GET("/", adminHandler.GetAllUsers)
			user.PATCH("/block", adminHandler.BlockUser)
		}

		//department
		department := api.Group("/departments")
		{
			department.GET("/:department_id", middleware.TrimSpaces(), productHandler.GetDepartmentByID)
			department.POST("/", middleware.TrimSpaces(), productHandler.SaveDepartment)
			department.GET("/", middleware.TrimSpaces(), productHandler.GetAllDepartments)

			// category
			category := department.Group("/:department_id/categories")
			{
				category.GET("/", productHandler.GetAllCategoriesByDepartmentID)
				category.POST("/", middleware.TrimSpaces(), productHandler.SaveCategory)

				// Category images for a specific category
				categoryImages := category.Group("/:category_id/images")
				{
					categoryImages.POST("/", middleware.TrimSpaces(), productHandler.SaveCategoryImage)
					categoryImages.GET("/", productHandler.GetAllCategoryImages)
					categoryImages.GET("/:image_id", productHandler.GetCategoryImageByID)
					categoryImages.PUT("/:image_id", middleware.TrimSpaces(), productHandler.UpdateCategoryImage)
					categoryImages.DELETE("/:image_id", productHandler.DeleteCategoryImage)
				}

				// Sub-categories for a specific category
				subCategory := category.Group("/:category_id/sub-categories")
				{
					subCategory.POST("/", middleware.TrimSpaces(), productHandler.SaveSubCategory)
					subCategory.GET("/", productHandler.GetAllSubCategoriesByCategoryID)

					// Sub type attributes for a specific subcategory
					subTypeAttr := subCategory.Group("/:sub_category_id/attributes")
					{
						subTypeAttr.POST("/", middleware.TrimSpaces(), productHandler.SaveSubTypeAttribute)
						subTypeAttr.GET("/", productHandler.GetAllSubTypeAttributes)
						subTypeAttr.GET("/:attribute_id", productHandler.GetSubTypeAttributeByID)

						// Sub type attribute options for a specific attribute
						attrOption := subTypeAttr.Group("/:attribute_id/options")
						{
							attrOption.POST("/", middleware.TrimSpaces(), productHandler.SaveSubTypeAttributeOption)
							attrOption.GET("/", productHandler.GetAllSubTypeAttributeOptions)
							attrOption.GET("/:option_id", productHandler.GetSubTypeAttributeOptionByID)
						}
					}
				}

				// Variations for a specific category (sibling to subCategory, not child)
				variation := category.Group("/:category_id/variations")
				{
					variation.POST("/", middleware.TrimSpaces(), productHandler.SaveVariation)
					variation.GET("/", productHandler.GetAllVariations)

					variationOption := variation.Group("/:variation_id/options")
					{
						variationOption.POST("/", middleware.TrimSpaces(), productHandler.SaveVariationOption)
					}
				}
			}
		}
		// brand
		brand := api.Group("/brands")
		{
			brand.POST("", branHandler.Save)
			brand.GET("", branHandler.FindAll)
			brand.GET("/:brand_id", branHandler.FindOne)
			brand.PUT("/:brand_id", branHandler.Update)
			brand.DELETE("/:brand_id", branHandler.Delete)
		}
		// product
		product := api.Group("/products")
		{
			// product.GET("/", productHandler.GetAllProductsAdmin)
			product.GET("/filters/:shop_id", productHandler.FindProductItemFilters)
			product.GET("/:product_id", productHandler.GetProductByID)
			product.POST("/", middleware.TrimSpaces(), productHandler.SaveProduct)
			product.PUT("/", middleware.TrimSpaces(), productHandler.UpdateProduct)

		}

		productItem := api.Group("/items")
		{
			productItem.GET("", productHandler.GetAllProductItemsAdmin())
			productItem.GET("/lowViewproductitems", productHandler.FindLowViewProductItems)
			productItem.GET("/shop/:shop_id", productHandler.GetProductItemsByShopID())
			productItem.POST("", productHandler.SaveProductItem)
			productItem.GET("/:product_item_id", productHandler.GetProductItemByID)
			productItem.DELETE("/:product_item_id", productHandler.DeleteProductItem)
			productItem.PUT("/:product_item_id", productHandler.UpdateProductItem)
			// productItem.GET("/lowViewproductitems", productHandler.FindLowViewProductItems)

			productView := productItem.Group("/:product_item_id/view")
			{
				productView.GET("/", productHandler.GetProductItemViewCount)
			}

		}
		// 	// order
		order := api.Group("/orders")
		{
			order.GET("/all", orderHandler.GetAllShopOrders)
			order.GET("/:shop_order_id/items", orderHandler.GetAllOrderItemsAdmin())
			order.PUT("/", orderHandler.UpdateOrderStatus)

			status := order.Group("/statuses")
			{
				status.GET("/", orderHandler.GetAllOrderStatuses)
			}

			//return requests
			order.GET("/returns", orderHandler.GetAllOrderReturns)
			order.GET("/returns/pending", orderHandler.GetAllPendingReturns)
			order.PUT("/returns/pending", orderHandler.UpdateReturnRequest)
		}

		// payment_method
		paymentMethod := api.Group("/payment-methods")
		{
			paymentMethod.GET("/", paymentHandler.GetAllPaymentMethodsAdmin())
			// paymentMethod.POST("/", paymentHandler.AddPaymentMethod)
			paymentMethod.PUT("/", paymentHandler.UpdatePaymentMethod)
		}

		// offer
		offer := api.Group("/offers")
		{

			offer.POST("/", middleware.TrimSpaces(), offerHandler.SaveOffer) // add a new offer
			offer.GET("/", offerHandler.GetAllOffers)                        // get all offers
			offer.DELETE("/:product_item_id", offerHandler.RemoveOffer)
			offer.POST("/shop/:shop_id", middleware.TrimSpaces(), offerHandler.ApplyOfferToShop)
			offer.GET("/shop/:shop_id", offerHandler.GetShopOffersByShopID)
			offer.GET("/category", offerHandler.GetAllCategoryOffers)                        // to get all offers of categories
			offer.POST("/category", middleware.TrimSpaces(), offerHandler.SaveCategoryOffer) // add offer for categories
			offer.PATCH("/category", offerHandler.ChangeCategoryOffer)
			offer.DELETE("/category/:offer_category_id", offerHandler.RemoveCategoryOffer)

			offer.GET("/products", offerHandler.GetAllProductsOffers)                                // to get all offers of products
			offer.POST("/products_item", middleware.TrimSpaces(), offerHandler.SaveProductItemOffer) // add offer for products
			offer.PATCH("/products", offerHandler.ChangeProductOffer)
			offer.DELETE("/products/:offer_product_id", offerHandler.RemoveProductOffer)
			offer.GET("/active", offerHandler.GetActiveOffers)
			offer.GET("/post-login-offer", offerHandler.PostLoginOffer)
		}

		// coupons
		coupons := api.Group("/coupons")
		{
			coupons.POST("/", middleware.TrimSpaces(), couponHandler.SaveCoupon)
			coupons.GET("/", couponHandler.GetAllCouponsAdmin)
			coupons.PUT("/", middleware.TrimSpaces(), couponHandler.UpdateCoupon)
		}

		// sales report
		sales := api.Group("/sales")
		{
			sales.GET("/", adminHandler.GetFullSalesReport)
		}

		// stock
		stock := api.Group("/stocks")
		{
			stock.GET("/", stockHandler.GetAllStocks)

			stock.PATCH("/", stockHandler.UpdateStock)
		}

		// advertisement
		advertisement := api.Group("/advertisements")
		{
			advertisement.POST("/", adminHandler.CreateAdvertisement)
			advertisement.GET("/", adminHandler.GetAllAdvertisements)
			advertisement.PUT("/", adminHandler.UpdateAdvertisement)
			advertisement.DELETE("/:advertisement_id", adminHandler.DeleteAdvertisement)
		}

		// Shop details
		shop := api.Group("/shops")
		{
			shop.POST("/", adminHandler.CreateShop)
			shop.GET("/", adminHandler.GetAllShops)
			shop.GET("/:shop_id", adminHandler.GetShopByID)
			shop.PUT("/", adminHandler.UpdateShop)
			shop.PUT("/:shop_id", adminHandler.UploadShopById)
			shop.GET("/shop_details", adminHandler.GetShopByOwnerID)
			shop.POST("/verify", adminHandler.VerifyShop)
			shop.GET("/verify-status", adminHandler.GetVerificationStatus)

			shop.POST("/upload-profile-image", middleware.AuthenticateAdmin(), adminHandler.UploadAdminProfileImage)
			shop.PUT("/upload-profile-image/:shop_id", middleware.AuthenticateAdmin(), adminHandler.UploadAdminProfileImage)
			shop.GET("/shop-profile-image/:shop_id", middleware.AuthenticateAdmin(), adminHandler.GetShopProfileImageById)
			document := shop.Group("/business-document")
			{
				document.POST("/send-otp", adminHandler.UploadShopDocument)
				document.POST("/verify-otp", adminHandler.VerifyShopDocument)
			}

			address := shop.Group("/address-details")
			{
				address.POST("/save", adminHandler.UploadAddress)
			}
			// social := shop.Group("/social")
			// {
			// 	social.GET(":shop_id", adminHandler.GetShopSocialDetails)
			// }
			shop.GET("/:shop_id/social", adminHandler.GetShopSocialDetails)
		}

		// Notification
		notification := api.Group("/notifications")
		{
			notification.GET("/sendToUsersInRadius", adminHandler.SendNotificationToUsersInRadius)
		}

		// Promotion Categories and Types
		promotion := api.Group("/promotions")
		{
			// Promotion Categories
			categories := promotion.Group("/categories")
			{
				categories.GET("/", promotionHandler.GetAllPromotionCategories)
				categories.GET("/:category_id", promotionHandler.GetPromotionCategoryByID)
			}

			// Promotion Types
			types := promotion.Group("/types")
			{
				types.GET("/", promotionHandler.GetAllPromotionTypes)
				types.GET("/category/:category_id", promotionHandler.GetPromotionTypesByCategoryID)
				types.GET("/:type_id", promotionHandler.GetPromotionTypeByID)
			}

			// Promotions (instances)
			promotion.POST("/", promotionHandler.CreatePromotion)
			promotion.GET("/", promotionHandler.GetAllPromotions)
			promotion.GET("/:promotion_id", promotionHandler.GetPromotionByID)
			promotion.DELETE("/:promotion_id", promotionHandler.DeletePromotion)
		}

		identity := api.Group("/identity-document")
		{
			identity.POST("/send-otp", adminHandler.AdminDocumentOtpSend)
			identity.POST("/verify-otp", adminHandler.AdminDocumentOtpVerify)
		}

		verification := api.Group("/verification")
		{
			verification.GET("/shop/:shop_id", adminHandler.GetVerificationStatus)
		}

		fcm := api.Group("/fcm")
		{
			fcm.POST("/token", fcmTokenHandler.SaveFcmToken)
		}
	}
}
