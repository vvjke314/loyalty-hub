package app

import (
	"github.com/vvjke314/itk-courses/loyalityhub/internal/router"
	"go.uber.org/zap"
)

type App struct {
	router *router.Router
	logger *zap.Logger
}

func NewApp(router *router.Router, logger *zap.Logger) *App {
	return &App{
		router: router,
		logger: logger,
	}
}

func Run()
