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
	applogger "github.com/rohit221990/mandi-backend/pkg/logger"
	"github.com/rohit221990/mandi-backend/pkg/utils"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type ServerHTTP struct {
	Engine *gin.Engine
}

// @title						Mandi-backend API
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
	aiHandler *handler.AIHandler,
) *ServerHTTP {

	engine := gin.New()

	engine.RedirectTrailingSlash = true

	engine.LoadHTMLGlob("views/*.html")

	engine.Use(applogger.RequestLogger())
	engine.Use(utils.RecoveryMiddleware())
	engine.Use(mw.CORSMiddleware())
	engine.MaxMultipartMemory = 500 << 20 // 500 MB max for video uploads

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
	// Locazar brand icon used as the default image in push notifications.
	engine.StaticFile("/uploads/locazar_icon.png", "./uploads/locazar_icon.png")

	// Admin portal pages (no auth on the HTML itself; JS sends Bearer token)
	engine.GET("/admin/videos", func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin-videos.html", nil)
	})

	// set up routes
	routes.UserRoutes(engine.Group("/api"), authHandler, middleware, userHandler, cartHandler,
		productHandler, paymentHandler, orderHandler, couponHandler, offerHandler, stockHandler, branHandler, notificationHandler, promotionHandler, subscriptionPaymentHandler, subscriptionHandler, searchHandler)
	routes.UserBannerRoutes(engine.Group("/api"), bannerUserHandler)
	routes.SellerGuideRoutes(engine.Group("/api"), sellerGuideHandler)
	routes.AdminRoutes(engine.Group("/api/admin"), authHandler, middleware, adminHandler,
		productHandler, paymentHandler, orderHandler, couponHandler, offerHandler, stockHandler, branHandler, promotionHandler, fcmTokenHandler, notificationHandler, alertHandler, uiHandler, alertTemplateHandler,
		jobHandler, jobCategoryHandler, platformUserHandler, mobileAuthHandler, sellerGuideHandler)
	routes.UIRoutes(engine.Group("/api/web"), middleware, uiHandler)
	routes.AIRoutes(engine.Group("/api"), aiHandler)

	// Public advertisement endpoints — no auth required (used by mobile app)
	engine.GET("/api/advertisements/active", adminHandler.GetActiveAdvertisements)
	engine.GET("/api/advertisements/active/filter", adminHandler.GetActiveAdvertisementsFiltered)

	// Feature flags object — read by mobile apps on launch to gate features
	engine.GET("/api/feature-flags", adminHandler.GetFeatureFlagsObject)
	// App configs — read by clients without admin auth
	engine.GET("/api/app-configs", adminHandler.ListAppConfigs)
	engine.GET("/api/app-configs/:config_key", adminHandler.GetAppConfigByKey)
	// Structured global config (CDN, feature flags, HTTP, image upload) — read by
	// mobile apps on launch (RemoteConfigService). Same data as /api/admin/config;
	// this route just skips the admin-auth requirement for client reads.
	engine.GET("/api/app-config", adminHandler.GetGlobalConfig)
	// Help center (support contact details + FAQs) — read by clients without admin auth
	engine.GET("/api/help", adminHandler.GetHelpCenter)

	// PIN code lookup (city/district/state) — used by admin-portal's shop
	// address verification screen; no admin dependencies, so constructed
	// directly here rather than threaded through wire.
	pincodeHandler := handler.NewPincodeHandler()
	engine.GET("/api/pincode/:pincode", pincodeHandler.GetPincodeDetails)

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
	return http.ListenAndServe(":3000", alwaysCORS(s.Engine))
}

// alwaysCORS sets CORS headers on every response and short-circuits OPTIONS
// preflights outside of Gin's router entirely — before Gin (and its
// RedirectTrailingSlash) ever sees the request.
//
// Why this is needed: some registered routes end in a trailing slash (e.g.
// department.POST("/", ...) -> "/departments/") and some don't (e.g.
// attrOption.DELETE("/:option_id", ...) -> ".../options/:option_id"). Any
// request whose path trailing-slash doesn't match its route's registration
// makes Gin's RedirectTrailingSlash issue an internal 301/307 redirect
// *before* engine.Use() middleware — including CORSMiddleware — runs on
// that request, so the redirect response carries no CORS headers and the
// browser reports a CORS failure on the preflight instead of following a
// redirect. Setting headers here, before Gin's router runs at all, means
// they're present on every response Gin produces (including a redirect),
// and preflights never depend on routing/redirect behavior succeeding.
func alwaysCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
