package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/contextkeys"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/dto"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/model"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/repository"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type BalanceService struct {
	repo   *repository.Repositories
	logger *zap.Logger
}

func NewBalanceService(repo *repository.Repositories,
	logger *zap.Logger) *BalanceService {
	return &BalanceService{
		repo:   repo,
		logger: logger,
	}
}

// TO-DO: implement it
func (s *BalanceService) GetBalance(ctx context.Context) (dto.GetBalanceResponse, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "BalanceService.GetBalance")
	defer span.End()

	userIDStr, ok := ctx.Value(contextkeys.UserKeyID).(string)
	if !ok {
		err := fmt.Errorf("userID not found in context")
		span.RecordError(err)
		return dto.GetBalanceResponse{}, err
	}

	tx, err := s.repo.BeginTx(ctx, pgx.ReadUncommitted)
	if err != nil {
		span.RecordError(err)
		return dto.GetBalanceResponse{}, fmt.Errorf("error while starting transaction %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
		_ = tx.Commit(ctx)
	}()

	balanceRepo := s.repo.NewBalanceRepo(tx)
	balance, err := balanceRepo.Get(ctx, userIDStr)
	if err != nil {
		span.RecordError(err)
		return dto.GetBalanceResponse{}, fmt.Errorf("get balance error: %w", err)
	}

	return dto.GetBalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}, nil
}

func (s *BalanceService) Withdraw(ctx context.Context, req dto.NewWithdrawnRequest) error {
	ctx, span := otel.Tracer("service").Start(ctx, "BalanceService.Withdraw")
	defer span.End()

	userIDStr := ctx.Value(contextkeys.UserKeyID).(string)
	userID := uuid.MustParse(userIDStr)

	tx, err := s.repo.BeginTx(ctx, pgx.Serializable)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("start tx error: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
		_ = tx.Commit(ctx)
	}()

	balanceRepo := s.repo.NewBalanceRepo(tx)

	balance, err := balanceRepo.Get(ctx, userIDStr)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("get balance error: %w", err)
	}

	if balance.Current < req.Sum {
		err := model.ErrInsufficientFunds
		span.RecordError(err)
		return err
	}

	withdrawal := &model.Withdrawal{
		ID:          uuid.New(),
		OrderID:     req.Order,
		UserID:      userID,
		Amount:      decimal.NewFromFloat(req.Sum),
		ProcessedAt: time.Now(),
	}

	err = balanceRepo.AddWithdraw(ctx, withdrawal)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("add withdrawal error: %w", err)
	}

	return nil
}

func (s *BalanceService) GetWithdrawals(ctx context.Context) (dto.GetAllWithdrawalsResponse, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "BalanceService.GetWithdrawals")
	defer span.End()

	userIDStr, ok := ctx.Value(contextkeys.UserKeyID).(string)
	if !ok {
		err := fmt.Errorf("userID not found in context")
		span.RecordError(err)
		return dto.GetAllWithdrawalsResponse{}, err
	}

	tx, err := s.repo.BeginTx(ctx, pgx.Serializable)
	if err != nil {
		span.RecordError(err)
		return dto.GetAllWithdrawalsResponse{}, fmt.Errorf("start tx error: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
		_ = tx.Commit(ctx)
	}()

	balanceRepo := s.repo.NewBalanceRepo(tx)

	withdrawals, err := balanceRepo.GetAllWithdrawals(ctx, userIDStr)
	if err != nil {
		span.RecordError(err)
		return dto.GetAllWithdrawalsResponse{}, fmt.Errorf("get withdrawals error: %w", err)
	}

	var res []dto.Withdrawn
	for _, w := range withdrawals {
		res = append(res, dto.Withdrawn{
			Order:       w.OrderID,
			Sum:         w.Amount,
			ProcessedAt: w.ProcessedAt,
		})
	}

	return dto.GetAllWithdrawalsResponse{Withdrawals: res}, nil
}
