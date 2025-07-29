package interfaces

import (
	"context"

	"github.com/vvjke314/itk-courses/loyalityhub/internal/model"
)

type BalanceRepository interface {
	Get(ctx context.Context, userID string) (*model.Balance, error)
	AddWithdraw(ctx context.Context, withdrawal *model.Withdrawal) error
	GetAllWithdrawals(ctx context.Context, userID string) ([]model.Withdrawal, error)
}
