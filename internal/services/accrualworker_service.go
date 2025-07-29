package services

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/sony/gobreaker/v2"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/client/accrual"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/dto"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/model"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/repository"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type AccrualWorkerService struct {
	repo   *repository.Repositories
	client *accrual.AccrualClient
	logger *zap.Logger
	cb     *gobreaker.CircuitBreaker[dto.AccrualServiceResponse]
}

func NewAccrualWorkerService(repos *repository.Repositories,
	logger *zap.Logger, client *accrual.AccrualClient) *AccrualWorkerService {
	return &AccrualWorkerService{
		repo:   repos,
		client: client,
		logger: logger.With(zap.String("layer", "service")),
		cb: gobreaker.NewCircuitBreaker[dto.AccrualServiceResponse](gobreaker.Settings{
			Name: "accrual service breaker",
		}),
	}
}

func (s *AccrualWorkerService) UpdateOrders(ctx context.Context) error {
	ctx, span := otel.Tracer("worker").Start(ctx, "AccrualWorker.UpdateOrders")
	defer span.End()

	// начинаем транзакцию
	tx, err := s.repo.BeginTx(ctx, pgx.RepeatableRead)
	if err != nil {
		span.RecordError(err)
		s.logger.Error("can't start transaction", zap.Error(err))
		return fmt.Errorf("can't start transaction %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
		_ = tx.Commit(ctx)
	}()

	orderRepo := s.repo.NewOrderRepo(tx)
	pendingOrders, err := orderRepo.GetAllPending(ctx)
	if err != nil {
		span.RecordError(err)
		s.logger.Error("can't get all pending requests", zap.Error(err))
		return fmt.Errorf("can't get all pending requests %w", err)
	}

	for _, orderNumber := range pendingOrders {
		resp, err := s.cb.Execute(func() (dto.AccrualServiceResponse, error) {
			return s.client.GetData(orderNumber)
		})
		if err != nil {
			span.RecordError(err)
			s.logger.Error("can't get order data", zap.Error(err))
			return fmt.Errorf("can't get order data %w", err)
		}

		order := model.Order{
			Number:  resp.OrderNumber,
			Status:  model.OrderStatus(resp.Status),
			Accrual: decimal.NewFromFloat(resp.Accrual),
		}
		if err = orderRepo.Update(ctx, order); err != nil {
			span.RecordError(err)
			s.logger.Error("can't update order data", zap.Error(err))
			return fmt.Errorf("can't update order data %w", err)
		}
	}

	if len(pendingOrders) == 0 {
		return nil
	}

	return nil
}
