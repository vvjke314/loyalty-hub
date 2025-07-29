package services

import (
	"context"

	"github.com/vvjke314/itk-courses/loyalityhub/internal/services/interfaces"
	"go.uber.org/zap"
)

type Services struct {
	logger *zap.Logger
	User   interfaces.UserServiceInterface
	
}

func NewServices(ctx context.Context, logger *zap.Logger,
	userService interfaces.UserServiceInterface) *Services {
	return &Services{
		logger: logger,
		User:   userService,
	}
}
