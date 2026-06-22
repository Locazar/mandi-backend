package http

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	_ "github.com/rohit221990/mandi-backend/cmd/api/docs"
	"github.com/rohit221990/mandi-backend/pkg/api/handler"
	handlerInterface "github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	mw "github.com/rohit221990/mandi-backend/pkg/api/middleware"
	"github.com/rohit221990/mandi-backend/pkg/api/routes"
	"github.com/rohit221990/mandi-backend/pkg/utils"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type ServerHTTP struct {
	Engine *gin.Engine
}

// @title						E-commerce Application Backend API
// @description				Backend API built with Golang using Clean Code architecture. \nGithub: [https://github.com/rohit221990/mandi-backend].
//
// @contact.name				For API Support
// @contact.email				rohit.jangid.social@gmail.com
//
// @license.name				MIT
// @license.url				https://opensource.org/licenses/MIT
//
// @BasePath					/api
// @SecurityDefinitions.apikey	BearerAuth
// @Name						Authorization
// @In							headerNewServerHTTP
// @Description				Add prefix of Bearer before  token Ex: "Bearer token"
// @Query.collection.format	multi
func NewServerHTTP(authHandler handlerInterface.AuthHandler, middleware mw.Middleware,
	adminHandler handlerInterface.AdminHandler, userHandler handlerInterface.UserHandler,
	cartHandler handlerInterface.CartHandler, paymentHandler handlerInterface.PaymentHandler,
	productHandler handlerInterface.ProductHandler, orderHandler handlerInterface.OrderHandler,
	couponHandler handlerInterface.CouponHandler, offerHandler handlerInterface.OfferHandler,
	stockHandler handlerInterface.StockHandler, branHandler handlerInterface.BrandHandler,
	notificationHandler handlerInterface.NotificationHandler, promotionHandler handlerInterface.PromotionHandler,
	fcmTokenHandler handlerInterface.FcmTokenHandler,
	searchHandler handlerInterface.SearchHandler,
	alertHandler handlerInterface.AlertHandler,
	uiHandler handlerInterface.UIHandler,
	alertTemplateHandler handlerInterface.AlertTemplateHandler,
	bannerUserHandler handlerInterface.BannerUserHandler,
	subscriptionPaymentHandler handlerInterface.SubscriptionPaymentHandler,
	subscriptionHandler handlerInterface.SubscriptionHandler,
	sellerGuideHandler handlerInterface.SellerGuideHandler,
	jobHandler *handler.JobHandler,
	jobCategoryHandler *handler.JobCategoryHandler,
	platformUserHandler handlerInterface.PlatformUserHandler,
	mobileAuthHandler handlerInterface.OTPAuthRequestHandler,
) *ServerHTTP {

	engine := gin.New()

	engine.RedirectTrailingSlash = false

	engine.LoadHTMLGlob("views/*.html")

	engine.Use(gin.Logger())
	engine.Use(utils.RecoveryMiddleware())
	engine.Use(mw.CORSMiddleware())

	// swagger docs
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Handle icon requests with fallback
	engine.GET("/uploads/icon/*filepath", func(c *gin.Context) {
		fileParam := c.Param("filepath")
		fullPath, err := filepath.Abs("./uploads/icon" + fileParam)
		if err != nil {
			c.Status(404)
			return
		}

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			defaultIconPath, err := filepath.Abs("./uploads/icon/default.svg")
			if err != nil {
				c.Status(404)
				return
			}
			if _, err := os.Stat(defaultIconPath); os.IsNotExist(err) {
				c.Status(404)
				return
			}
			c.File(defaultIconPath)
			return
		}

		c.File(fullPath)
	})

	file, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}

	// Send logs to file
	log.SetOutput(file)

	// Offers still write to local disk; keep until that namespace is migrated.
	engine.StaticFS("/uploads/offers", http.Dir("./uploads/offers"))
	// Bundled assets — not user uploads.
	engine.StaticFS("/uploads/promotions", http.Dir("./uploads/promotions"))

	// set up routes
	routes.UserRoutes(engine.Group("/api"), authHandler, middleware, userHandler, cartHandler,
		productHandler, paymentHandler, orderHandler, couponHandler, offerHandler, stockHandler, branHandler, notificationHandler, promotionHandler, subscriptionPaymentHandler, subscriptionHandler, searchHandler)
	routes.UserBannerRoutes(engine.Group("/api"), bannerUserHandler)
	routes.SellerGuideRoutes(engine.Group("/api"), sellerGuideHandler)
	routes.AdminRoutes(engine.Group("/api/admin"), authHandler, middleware, adminHandler,
		productHandler, paymentHandler, orderHandler, couponHandler, offerHandler, stockHandler, branHandler, promotionHandler, fcmTokenHandler, notificationHandler, alertHandler, uiHandler, alertTemplateHandler,
		jobHandler, jobCategoryHandler, platformUserHandler, mobileAuthHandler)
	routes.UIRoutes(engine.Group("/api/web"), middleware, uiHandler)

	// log registered routes for debug
	for _, route := range engine.Routes() {
		log.Printf("GIN route registered: %s %s -> %s", route.Method, route.Path, route.Handler)
	}

	// no handler
	engine.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "invalid url go to /swagger/index.html for api documentation",
		})
	})

	return &ServerHTTP{Engine: engine}
}

func (s *ServerHTTP) Start() error {
	return s.Engine.Run(":3000")
}
