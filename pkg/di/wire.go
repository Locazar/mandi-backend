//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	http "github.com/rohit221990/mandi-backend/pkg/api"
	"github.com/rohit221990/mandi-backend/pkg/api/handler"
	"github.com/rohit221990/mandi-backend/pkg/api/middleware"
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/repository"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	elasticsearch "github.com/rohit221990/mandi-backend/pkg/service/elasticsearch"
	"github.com/rohit221990/mandi-backend/pkg/service/graphics"
	"github.com/rohit221990/mandi-backend/pkg/service/otp"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
)

func provideElasticURL(cfg config.Config) string {
	return cfg.ElasticsearchURL
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

		// graphics
		graphics.NewGraphicsService,

		// repository

		middleware.NewMiddleware,
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

		//usecase
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
		// handler
		handler.NewAuthHandler,
		handler.NewAdminHandler,
		handler.NewUserHandler,
		handler.NewCartHandler,
		handler.NewPaymentHandler,
		handler.NewProductHandler,
		handler.NewOrderHandler,
		handler.NewCouponHandler,
		handler.NewOfferHandler,
		handler.NewStockHandler,
		handler.NewBrandHandler,
		handler.NewNotificationHandler,
		handler.NewPromotionHandler,

		http.NewServerHTTP,
	)

	return &http.ServerHTTP{}, nil
}
