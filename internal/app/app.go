package app

import (
	"context"
	"os"

	"github.com/vvjke314/itk-courses/loyalityhub/internal/handlers"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/repository"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/router"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/services"
	"go.uber.org/zap"
)

type App struct {
	router *router.Router
	logger *zap.Logger
}

func NewApp(logger *zap.Logger) *App {
	return &App{
		logger: logger,
	}
}

func (a *App) Init(ctx context.Context, repos *repository.Repositories) {
	// инициализация сервисов
	userService := services.NewUserService(a.logger, repos)
	orderService := services.NewOrderService(repos, a.logger)
	balanceService := services.NewBalanceService(repos, a.logger)

	// инициализация хендлеров
	userHandler := handlers.NewUserHandler(os.Getenv("APP_HOST"), userService)
	orderHandler := handlers.NewOrderHandler(os.Getenv("APP_HOST"), orderService)
	balanceHandler := handlers.NewBalanceHandler(os.Getenv("APP_HOST"), balanceService)

	// настройка роутера
	router := router.NewRouter(ctx, a.logger, userHandler, orderHandler, balanceHandler)
	a.router = router
}

func (a *App) Run() error {
	return a.router.Run()
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.router.Shutdown(ctx)
}
