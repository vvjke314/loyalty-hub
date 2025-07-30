package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/vvjke314/itk-courses/loyalityhub/docs"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/joho/godotenv"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/app"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/client/accrual"
	_ "github.com/vvjke314/itk-courses/loyalityhub/internal/dto"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/logx"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/metrics"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/repository"
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

	// инициализируем контекст для gracefull-shuttdown'a
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// получаем логгер
	logger, err := logx.Get(os.Getenv("LOG_FILE"))
	if err != nil {
		log.Println("can't init logger")
		panic(err)
	}
	logger.Debug("logger successfully configurated and started")

	// инициализируем метрики
	metrics.RegisterMetrics()

	// инициализируем трейсер
	tp, err := tracing.StartTracing(os.Getenv("JAEGER_LISTEN_HOST") + ":" + os.Getenv("JAEGER_LISTEN_PORT"))
	if err != nil {
		logger.Fatal("can't init logger")
		panic(err)
	}

	// инициализируем репозиторий
	repos := repository.NewRepositories(logger)
	err = repos.Init(ctx, os.Getenv("ORDERS_DB_DSN"))
	if err != nil {
		logger.Fatal("can't init repo")
		panic(err)
	}
	defer repos.Close()
	logger.Debug("repository successfully configurated and started")

	// группа ошибок
	errGrp, errCtx := errgroup.WithContext(ctx)

	// контекст для работы приложения
	appCtx, cancelAppCtx := context.WithCancel(ctx)
	defer cancelAppCtx()

	// инициализация клиента
	accrualClient := accrual.NewAccrualClient(100, os.Getenv("ACCRUAL_SERVICE"))

	// инициализация сервиса worker'a
	accrualWorkerService := services.NewAccrualWorkerService(repos, logger, accrualClient)

	// настройка фонового воркера
	worker := accrualWorker.NewAccrualWorker(1*time.Second, accrualWorkerService)
	errChan := make(chan error)

	// инициализация приложения
	app := app.NewApp(logger)
	app.Init(appCtx, repos)

	// запуск приложения
	errGrp.Go(func() error {
		if err := app.Run(); err != nil {
			cancelAppCtx()
			return err
		}
		return nil
	})

	// запуск воркера
	errGrp.Go(func() error {
		defer close(errChan)
		if err := worker.Work(ctx, errChan); err != nil {
			cancelAppCtx()
			return err
		}
		return nil
	})

	// запуск принта ошибок
	errGrp.Go(func() error {
		logger := logger.With(zap.String("service_name", "accrual_worker"))
		for err := range errChan {
			logger.Error("error while working", zap.Error(err))
		}
		return nil
	})

	// shutdown ctx
	shtDownCtx, cancelShtDown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShtDown()

	// shutdown
	errGrp.Go(func() error {
		<-errCtx.Done()

		if err := app.Shutdown(shtDownCtx); err != nil {
			logger.Error("error while shutdowning", zap.Error(err))
			return err
		}

		if err := tp.Shutdown(shtDownCtx); err != nil {
			logger.Error("error while shutdowning", zap.Error(err))
			return err
		}

		logger.Info("gracefully shutted down")
		return nil
	})

	if err := errGrp.Wait(); err != nil {
		return
	}
}
