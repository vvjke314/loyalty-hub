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
	"github.com/vvjke314/itk-courses/loyalityhub/internal/utils/lunavalidate"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type OrderService struct {
	repo   *repository.Repositories
	logger *zap.Logger
}

func NewOrderService(repo *repository.Repositories,
	logger *zap.Logger) *OrderService {
	return &OrderService{
		repo:   repo,
		logger: logger,
	}
}

func (os *OrderService) Load(ctx context.Context, orderNumber string) (dto.AddOrderResponse, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "OrderService.Load")
	defer span.End()

	// проверить заказ по алгоритму Луна
	if !lunavalidate.Validate(orderNumber) {
		return dto.AddOrderResponse{}, model.ErrBadOrderNumber
	}

	tx, err := os.repo.BeginTx(ctx, pgx.ReadUncommitted)
	if err != nil {
		span.RecordError(err)
		return dto.AddOrderResponse{}, fmt.Errorf("error while starting transaction %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
		_ = tx.Commit(ctx)
	}()

	orderRepo := os.repo.NewOrderRepo(tx)
	existOrder, err := orderRepo.GetByNumber(ctx, orderNumber)
	if existOrder != nil {
		span.RecordError(err)
		if existOrder.UserID.String() == ctx.Value(contextkeys.UserKeyID).(string) {
			return dto.AddOrderResponse{
				OrderNumber: existOrder.Number,
			}, model.ErrOrderAlreadyExists
		}
		return dto.AddOrderResponse{}, model.ErrOrderLoadedByAnotherPerson
	}

	order := model.Order{
		Number:     orderNumber,
		UserID:     uuid.MustParse(ctx.Value(contextkeys.UserKeyID).(string)),
		Status:     model.OrderStatusNew,
		Accrual:    decimal.Zero,
		UploadedAt: time.Now(),
	}

	err = orderRepo.Create(ctx, &order)
	if err != nil {
		span.RecordError(err)
		return dto.AddOrderResponse{}, fmt.Errorf("[orderRepo.Create] %w", err)
	}

	span.SetAttributes(attribute.String("order_number", orderNumber))
	return dto.AddOrderResponse{
		OrderNumber: orderNumber,
	}, nil
}

func (os *OrderService) GetAll(ctx context.Context,
	userID string) (dto.GetAllOrdersResponse, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "OrderService.GetAll")
	defer span.End()

	tx, err := os.repo.BeginTx(ctx, pgx.ReadUncommitted)
	if err != nil {
		span.RecordError(err)
		return dto.GetAllOrdersResponse{}, fmt.Errorf("error while starting transaction %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
		_ = tx.Commit(ctx)
	}()

	orderService := os.repo.NewOrderRepo(tx)

	orders, err := orderService.GetAll(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return dto.GetAllOrdersResponse{}, fmt.Errorf("[orderService.GetAll]: %w", err)
	}

	span.SetAttributes(attribute.String("user_id", userID), attribute.Int("orders_count", len(orders)))
	return dto.GetAllOrdersResponse{
		Orders: orders,
	}, nil
}
