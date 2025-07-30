package router

import (
	"context"
	"net"
	"net/http"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/handlers"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/middleware"
	tokenmanager "github.com/vvjke314/itk-courses/loyalityhub/internal/utils/tokenManager"
	"go.uber.org/zap"
)

type Router struct {
	server *http.Server
	logger *zap.Logger
}

func NewRouter(ctx context.Context, logger *zap.Logger, userHandler *handlers.UserHandler,
	orderHandler *handlers.OrderHandler, balanceHandler *handlers.BalanceHandler) *Router {
	// инициализация token manager
	tm := tokenmanager.NewTokenManager(tokenmanager.NewTokenManagerConfig())
	// Инициализация gin
	r := gin.Default()

	// включаем gzip
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.Metrics())
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

	// init prometheus metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return &Router{
		server: &http.Server{
			Addr:        ":8080",
			BaseContext: func(l net.Listener) context.Context { return ctx },
			Handler:     r,
		},
		logger: logger,
	}
}

func (r *Router) Run() error {
	// Запуск сервера
	return r.server.ListenAndServe()
}

// шатдаун сервера
func (r *Router) Shutdown(ctx context.Context) error {
	return r.server.Shutdown(ctx)
}
