//go:build wireinject
// +build wireinject

package di

import (
	"database/sql"

	"github.com/google/wire"
	"gorm.io/gorm"

	http "github.com/rohit221990/mandi-backend/pkg/api"
	"github.com/rohit221990/mandi-backend/pkg/api/handler"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/middleware"
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/repository"
	aiservice "github.com/rohit221990/mandi-backend/pkg/service/ai"
	"github.com/rohit221990/mandi-backend/pkg/service/alert_engine"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	"github.com/rohit221990/mandi-backend/pkg/service/crypto"
	elasticsearch "github.com/rohit221990/mandi-backend/pkg/service/elasticsearch"
	"github.com/rohit221990/mandi-backend/pkg/service/graphics"
	"github.com/rohit221990/mandi-backend/pkg/service/otp"
	"github.com/rohit221990/mandi-backend/pkg/service/sms"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
)

func provideElasticURL(cfg config.Config) string {
	return cfg.ElasticsearchURL
}

// provideQRCodeHandler wires the QR handler with the public origin from config.
// A dedicated provider (vs. a bare string) avoids ambiguity with other
// string-returning providers in the graph.
func provideQRCodeHandler(uc *usecase.QRCodeUseCase, cfg config.Config) *handler.QRCodeHandler {
	return handler.NewQRCodeHandler(uc, cfg.PublicBaseURL)
}

func provideSQLDB(gormDB *gorm.DB) (*sql.DB, error) {
	return gormDB.DB()
}

func provideTwoFactorSMSService(cfg config.Config) *sms.TwoFactorSMSService {
	return sms.NewTwoFactorSMSService(cfg.TwoFactorAPIKey)
}

func provideAIServiceClient(cfg config.Config) *aiservice.Client {
	return aiservice.NewClient(cfg.AIServiceURL)
}

func provideCryptoService(cfg config.Config) (*crypto.Service, error) {
	keys, err := crypto.ParseKeyring(cfg.PIIEncryptionKeys)
	if err != nil {
		return nil, err
	}
	return crypto.NewService(keys, cfg.PIIEncryptionActiveKey)
}

func provideAlertRuleRegistry() *alert_engine.RuleRegistry {
	registry := alert_engine.NewRuleRegistry()
	// Register default rules
	registry.RegisterMultiple(
		alert_engine.MissingShopPhotoRule{},
		alert_engine.NoProductsRule{},
		alert_engine.ShopNotVerifiedRule{},
	)
	return registry
}

func InitializeApi(cfg config.Config) (*http.ServerHTTP, error) {

	wire.Build(db.ConnectDatabase,
		//external
		token.NewTokenService,
		otp.NewOtpAuth,
		cloud.NewObjectStorageService,

		// elasticsearch
		elasticsearch.NewElasticService,
		provideElasticURL,

		// ai service
		provideAIServiceClient,

		// graphics
		graphics.NewGraphicsService,

		// PII field encryption
		provideCryptoService,

		// alert engine
		provideAlertRuleRegistry,

		// middleware
		middleware.NewMiddleware,

		// repository — all constructors return interface directly, no Bind needed
		repository.NewAuthRepository,
		repository.NewPaymentRepository,
		repository.NewAdminRepository,
		repository.NewUserRepository,
		repository.NewCartRepository,
		repository.NewProductRepository,
		repository.NewOrderRepository,
		repository.NewCouponRepository,
		repository.NewOfferRepository,
		repository.NewStockRepository,
		repository.NewBrandDatabaseRepository,
		repository.NewPromotionRepository,
		repository.NewShopTimeRepository,
		repository.NewBannerRepository,
		repository.NewFcmTokenRepository,
		repository.NewSubscriptionRepository,
		repository.NewSearchRepository,
		repository.NewNotificationRepository,
		repository.NewAlertRepository,
		repository.NewPlatformUserRepository,
		repository.NewShopUpdateRepository,
		repository.NewLanguageRepository,
		repository.NewQRCodeRepository,

		//usecase — constructors that return interface directly need no Bind;
		//          constructors that return *concrete need Bind
		usecase.NewAuthUseCase,
		usecase.NewAdminUseCase,
		usecase.NewUserUseCase,
		usecase.NewCartUseCase,
		usecase.NewPaymentUseCase,
		usecase.NewProductUseCase,
		usecase.NewOrderUseCase,
		usecase.NewCouponUseCase,
		usecase.NewOfferUseCase,
		usecase.NewStockUseCase,
		usecase.NewBrandUseCase,
		usecase.NewNotificationUseCase,
		usecase.NewPromotionUseCase,
		usecase.NewShopTimeUseCase,
		usecase.NewFcmTokenUseCase,
		usecase.NewSubscriptionPaymentUseCase,
		usecase.NewSubscriptionUseCase,
		usecase.NewSearchUseCase,
		usecase.NewAlertUseCase,
		usecase.NewAlertTemplateUseCase,
		usecase.NewBannerUseCase,
		usecase.NewPlatformUserUseCase,
		usecase.NewShopUpdateUseCase,
		usecase.NewLanguageUseCase,
		usecase.NewQRCodeUseCase,

		// handler
		handler.NewAuthHandler,
		handler.NewAdminHandler,
		handler.NewUserHandler,
		wire.Bind(new(interfaces.UserHandler), new(*handler.UserHandler)),
		handler.NewCartHandler,
		handler.NewPaymentHandler,
		handler.NewProductHandler,
		handler.NewOrderHandler,
		handler.NewCouponHandler,
		handler.NewOfferHandler,
		handler.NewStockHandler,
		handler.NewBrandHandler,
		handler.NewNotificationHandler,
		wire.Bind(new(interfaces.NotificationHandler), new(*handler.NotificationHandler)),
		handler.NewPromotionHandler,
		wire.Bind(new(interfaces.PromotionHandler), new(*handler.PromotionHandler)),
		handler.NewFcmTokenHandler,
		wire.Bind(new(interfaces.FcmTokenHandler), new(*handler.FcmTokenHandler)),
		handler.NewSubscriptionPaymentHandler,
		handler.NewSubscriptionHandler,
		handler.NewSearchHandler,
		handler.NewAlertHandler,
		wire.Bind(new(interfaces.AlertHandler), new(*handler.AlertHandler)),
		handler.NewAlertTemplateHandler,
		wire.Bind(new(interfaces.AlertTemplateHandler), new(*handler.AlertTemplateHandler)),
		handler.NewUIHandler,
		handler.NewBannerUserHandler,
		wire.Bind(new(interfaces.BannerUserHandler), new(*handler.BannerUserHandler)),
		handler.NewSellerGuideHandler,
		wire.Bind(new(interfaces.SellerGuideHandler), new(*handler.SellerGuideHandler)),
		handler.NewJobHandler,
		usecase.NewJobService,
		handler.NewJobCategoryHandler,
		usecase.NewJobCategoryService,
		wire.Bind(new(handler.JobCategoryService), new(*usecase.JobCategoryService)),
		handler.NewPlatformUserHandler,
		handler.NewShopUpdateHandler,
		handler.NewLanguageHandler,
		provideQRCodeHandler,

		// mobile OTP auth (seller signup)
		provideSQLDB,
		provideTwoFactorSMSService,
		otp.NewMobileOTPService,
		repository.NewMobileAuthRepository,
		usecase.NewMobileAuthUseCase,
		handler.NewHandler,
		wire.Bind(new(interfaces.OTPAuthRequestHandler), new(*handler.Handler)),
		// ai service proxy
		handler.NewAIHandler,

		http.NewServerHTTP,
	)

	return &http.ServerHTTP{}, nil
}
