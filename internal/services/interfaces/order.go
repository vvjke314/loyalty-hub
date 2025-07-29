package interfaces

import (
	"context"

	"github.com/vvjke314/itk-courses/loyalityhub/internal/dto"
)

type OrderServiceInterface interface {
	Load(ctx context.Context, orderNumber string) (dto.AddOrderResponse, error)
	GetAll(ctx context.Context, userID string) (dto.GetAllOrdersResponse, error)
}
