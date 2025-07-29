package main

import (
	"context"
	"log"
	"os"
	"time"

	_ "github.com/vvjke314/itk-courses/loyalityhub/docs"
	"go.uber.org/zap"

	"github.com/joho/godotenv"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/client/accrual"
	_ "github.com/vvjke314/itk-courses/loyalityhub/internal/dto"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/handlers"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/logx"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/repository"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/router"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/services"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/tracing"
	accrualWorker "github.com/vvjke314/itk-courses/loyalityhub/internal/worker/accrual"
)

// @title           Loyaltyhub API
// @version         1.0
// @description     This is a sample server celler server.
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8080
// @BasePath  /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// подгружаем переменные окружения
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("can't parse .env config")
		panic(err)
	}
	// получаем логгер
	logger, err := logx.Get(os.Getenv("LOG_FILE"))
	if err != nil {
		log.Println("can't init logger")
		panic(err)
	}
	logger.Debug("logger successfully configurated and started")

	// инициализируем трейсер
	tp, err := tracing.StartTracing(os.Getenv("JAEGER_LISTEN_HOST_TEST") + ":" + os.Getenv("JAEGER_LISTEN_PORT"))
	if err != nil {
		logger.Fatal("can't init logger")
		panic(err)
	}
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// инициализируем репозиторий
	repos := repository.NewRepositories(logger)
	err = repos.Init(ctx, os.Getenv("ORDERS_DB_DSN_TEST"))
	if err != nil {
		logger.Fatal("can't init repo")
		panic(err)
	}
	defer repos.Close()
	logger.Debug("repository successfully configurated and started")

	// инициализация клиента
	accrualClient := accrual.NewAccrualClient(100, os.Getenv("ACCRUAL_SERVICE"))

	// инициализация сервисов
	userService := services.NewUserService(logger, repos)
	orderService := services.NewOrderService(repos, logger)
	balanceService := services.NewBalanceService(repos, logger)
	accrualWorkerService := services.NewAccrualWorkerService(repos, logger, accrualClient)

	// инициализация хендлеров
	userHandler := handlers.NewUserHandler(os.Getenv("APP_HOST"), userService)
	orderHandler := handlers.NewOrderHandler(os.Getenv("APP_HOST"), orderService)
	balanceHandler := handlers.NewBalanceHandler(os.Getenv("APP_HOST"), balanceService)

	// настройка роутера
	router := router.NewRouter(logger, userHandler, orderHandler, balanceHandler)

	// настройка фонового воркера
	worker := accrualWorker.NewAccrualWorker(1*time.Second, accrualWorkerService)
	errChan := make(chan error)

	go func() {
		router.Run()
	}()

	go func() {
		worker.Work(ctx, errChan)
	}()

	go func() {
		logger := logger.With(zap.String("service name", "accrual_worker"))
		for err := range errChan {
			logger.Error("error while working", zap.Error(err))
		}
	}()

	<-ctx.Done()
}
