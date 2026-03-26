//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	http "github.com/rohit221990/mandi-backend/pkg/api"
	"github.com/rohit221990/mandi-backend/pkg/api/handler"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/middleware"
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/repository"
	repointerfaces "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	aiservice "github.com/rohit221990/mandi-backend/pkg/service/ai"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	elasticsearch "github.com/rohit221990/mandi-backend/pkg/service/elasticsearch"
	"github.com/rohit221990/mandi-backend/pkg/service/graphics"
	"github.com/rohit221990/mandi-backend/pkg/service/otp"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	usecaseinterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

func provideElasticURL(cfg config.Config) string {
	return cfg.ElasticsearchURL
}

func provideAIServiceClient(cfg config.Config) *aiservice.Client {
	return aiservice.NewClient(cfg.AIServiceURL)
}

func InitializeApi(cfg config.Config) (*http.ServerHTTP, error) {

	wire.Build(db.ConnectDatabase,
		//external
		token.NewTokenService,
		otp.NewOtpAuth,
		cloud.NewAWSCloudService,

		// elasticsearch
		elasticsearch.NewElasticService,
		provideElasticURL,

		// ai service
		provideAIServiceClient,

		// graphics
		graphics.NewGraphicsService,

		// middleware
		middleware.NewMiddleware,

		// repository
		repository.NewAuthRepository,
		wire.Bind(new(repointerfaces.AuthRepository), new(*repository.authRepository)),
		repository.NewPaymentRepository,
		wire.Bind(new(repointerfaces.PaymentRepository), new(*repository.paymentRepository)),
		repository.NewAdminRepository,
		wire.Bind(new(repointerfaces.AdminRepository), new(*repository.adminRepository)),
		repository.NewUserRepository,
		wire.Bind(new(repointerfaces.UserRepository), new(*repository.userRepository)),
		repository.NewCartRepository,
		wire.Bind(new(repointerfaces.CartRepository), new(*repository.cartRepository)),
		repository.NewProductRepository,
		wire.Bind(new(repointerfaces.ProductRepository), new(*repository.productRepository)),
		repository.NewOrderRepository,
		wire.Bind(new(repointerfaces.OrderRepository), new(*repository.orderRepository)),
		repository.NewCouponRepository,
		wire.Bind(new(repointerfaces.CouponRepository), new(*repository.couponRepository)),
		repository.NewOfferRepository,
		wire.Bind(new(repointerfaces.OfferRepository), new(*repository.offerRepository)),
		repository.NewStockRepository,
		wire.Bind(new(repointerfaces.StockRepository), new(*repository.stockRepository)),
		repository.NewBrandDatabaseRepository,
		wire.Bind(new(repointerfaces.BrandRepository), new(*repository.brandRepository)),
		repository.NewPromotionRepository,
		wire.Bind(new(repointerfaces.PromotionRepository), new(*repository.promotionRepository)),
		repository.NewShopTimeRepository,
		wire.Bind(new(repointerfaces.ShopTimeRepository), new(*repository.shopTimeRepository)),
		repository.NewBannerRepository,
		wire.Bind(new(repointerfaces.BannerRepository), new(*repository.bannerRepository)),
		repository.NewFcmTokenRepository,
		wire.Bind(new(repointerfaces.FcmTokenRepository), new(*repository.fcmTokenRepository)),

		//usecase
		usecase.NewAuthUseCase,
		wire.Bind(new(usecaseinterfaces.AuthUseCase), new(*usecase.authUseCase)),
		usecase.NewAdminUseCase,
		wire.Bind(new(usecaseinterfaces.AdminUseCase), new(*usecase.adminUseCase)),
		usecase.NewUserUseCase,
		wire.Bind(new(usecaseinterfaces.UserUseCase), new(*usecase.userUseCase)),
		usecase.NewCartUseCase,
		wire.Bind(new(usecaseinterfaces.CartUseCase), new(*usecase.cartUseCase)),
		usecase.NewPaymentUseCase,
		wire.Bind(new(usecaseinterfaces.PaymentUseCase), new(*usecase.paymentUseCase)),
		usecase.NewProductUseCase,
		wire.Bind(new(usecaseinterfaces.ProductUseCase), new(*usecase.productUseCase)),
		usecase.NewOrderUseCase,
		wire.Bind(new(usecaseinterfaces.OrderUseCase), new(*usecase.orderUseCase)),
		usecase.NewCouponUseCase,
		wire.Bind(new(usecaseinterfaces.CouponUseCase), new(*usecase.couponUseCase)),
		usecase.NewOfferUseCase,
		wire.Bind(new(usecaseinterfaces.OfferUseCase), new(*usecase.offerUseCase)),
		usecase.NewStockUseCase,
		wire.Bind(new(usecaseinterfaces.StockUseCase), new(*usecase.stockUseCase)),
		usecase.NewBrandUseCase,
		wire.Bind(new(usecaseinterfaces.BrandUseCase), new(*usecase.brandUseCase)),
		usecase.NewNotificationUseCase,
		wire.Bind(new(usecaseinterfaces.NotificationUseCase), new(*usecase.notificationUseCase)),
		usecase.NewPromotionUseCase,
		wire.Bind(new(usecaseinterfaces.PromotionUseCase), new(*usecase.promotionUseCase)),
		usecase.NewShopTimeUseCase,
		wire.Bind(new(usecaseinterfaces.ShopTimeUseCase), new(*usecase.shopTimeUseCase)),
		usecase.NewFcmTokenUseCase,
		wire.Bind(new(usecaseinterfaces.FcmTokenUseCase), new(*usecase.fcmTokenUseCase)),

		// handler
		handler.NewAuthHandler,
		wire.Bind(new(interfaces.AuthHandler), new(*handler.AuthHandler)),
		handler.NewAdminHandler,
		wire.Bind(new(interfaces.AdminHandler), new(*handler.AdminHandler)),
		handler.NewUserHandler,
		wire.Bind(new(interfaces.UserHandler), new(*handler.UserHandler)),
		handler.NewCartHandler,
		wire.Bind(new(interfaces.CartHandler), new(*handler.CartHandler)),
		handler.NewPaymentHandler,
		wire.Bind(new(interfaces.PaymentHandler), new(*handler.PaymentHandler)),
		handler.NewProductHandler,
		wire.Bind(new(interfaces.ProductHandler), new(*handler.ProductHandler)),
		handler.NewOrderHandler,
		wire.Bind(new(interfaces.OrderHandler), new(*handler.OrderHandler)),
		handler.NewCouponHandler,
		wire.Bind(new(interfaces.CouponHandler), new(*handler.CouponHandler)),
		handler.NewOfferHandler,
		wire.Bind(new(interfaces.OfferHandler), new(*handler.OfferHandler)),
		handler.NewStockHandler,
		wire.Bind(new(interfaces.StockHandler), new(*handler.StockHandler)),
		handler.NewBrandHandler,
		wire.Bind(new(interfaces.BrandHandler), new(*handler.BrandHandler)),
		handler.NewNotificationHandler,
		wire.Bind(new(interfaces.NotificationHandler), new(*handler.NotificationHandler)),
		handler.NewPromotionHandler,
		wire.Bind(new(interfaces.PromotionHandler), new(*handler.PromotionHandler)),
		handler.NewFcmTokenHandler,
		wire.Bind(new(interfaces.FcmTokenHandler), new(*handler.FcmTokenHandler)),

		http.NewServerHTTP,
	)

	return &http.ServerHTTP{}, nil
}
