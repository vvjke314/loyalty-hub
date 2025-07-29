package interfaces

import (
	"context"

	"github.com/vvjke314/itk-courses/loyalityhub/internal/dto"
)

type BalanceServiceInterface interface {
	GetBalance(ctx context.Context) (dto.GetBalanceResponse, error)
	Withdraw(ctx context.Context, req dto.NewWithdrawnRequest) error
	GetWithdrawals(ctx context.Context) (dto.GetAllWithdrawalsResponse, error)
}
