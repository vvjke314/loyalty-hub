package interfaces

import (
	"context"

	"github.com/vvjke314/itk-courses/loyalityhub/internal/model"
)

type UserRepositoryInterface interface {
	Create(ctx context.Context, user *model.User) error
	GetByLogin(ctx context.Context, login string) (*model.User, error)
}
