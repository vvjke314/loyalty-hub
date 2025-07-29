package router

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/handlers"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/middleware"
	tokenmanager "github.com/vvjke314/itk-courses/loyalityhub/internal/utils/tokenManager"
	"go.uber.org/zap"
)

type Router struct {
	engine *gin.Engine
	logger *zap.Logger
}

func NewRouter(logger *zap.Logger, userHandler *handlers.UserHandler,
	orderHandler *handlers.OrderHandler, balanceHandler *handlers.BalanceHandler) *Router {
	// инициализация token manager
	tm := tokenmanager.NewTokenManager(tokenmanager.NewTokenManagerConfig())
	// Инициализация gin
	r := gin.Default()

	// включаем gzip
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(middleware.LoggerMiddleware(logger))
	api := r.Group("/api/v1")

	// Регистрация маршрутов для user
	api.POST("/register", userHandler.Register)
	api.POST("/auth", userHandler.Auth)
	api.GET("/refresh", userHandler.Refresh)

	auth := api.Group("/user")
	auth.Use(middleware.AuthMiddleware(tm))

	// Регистрация маршрутов по заказам (orders)
	auth.POST("/orders", orderHandler.LoadOrder)
	auth.GET("/orders", orderHandler.GetAllOrders)

	// Регистрация маршрутов по балансу
	auth.GET("/balance", balanceHandler.GetBalance)
	auth.POST("/balance/withdraw", balanceHandler.Withdraw)
	auth.GET("/withdrawals", balanceHandler.GetWithdrawals)

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return &Router{
		engine: r,
		logger: logger,
	}
}

func (r *Router) Run() {
	// Запуск сервера
	if err := r.engine.Run(); err != nil {
		r.logger.Fatal("failed to run server", zap.Error(err))
	}
}
