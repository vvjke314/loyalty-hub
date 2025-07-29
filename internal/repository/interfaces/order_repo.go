package interfaces

import (
	"context"

	"github.com/vvjke314/itk-courses/loyalityhub/internal/model"
)

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByNumber(ctx context.Context, orderNumber string) (*model.Order, error)
	GetAll(ctx context.Context, userID string) ([]model.Order, error)
	Delete(ctx context.Context, orderNumber string) error
	GetAllPending(ctx context.Context) ([]string, error)
	Update(ctx context.Context, order model.Order) error
}
