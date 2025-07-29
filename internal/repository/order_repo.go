package repository

import (
	"context"
	"fmt"

	"github.com/vvjke314/itk-courses/loyalityhub/internal/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type OrderRepoPostgres struct {
	db     DBExecutor
	logger *zap.Logger
}

func NewOrderRepoPostgres(db DBExecutor, logger *zap.Logger) *OrderRepoPostgres {
	return &OrderRepoPostgres{
		db:     db,
		logger: logger.With(zap.String("repo", "order")),
	}
}

func (repo *OrderRepoPostgres) Create(ctx context.Context, order *model.Order) error {
	ctx, span := otel.Tracer("repository").Start(ctx, "OrderRepo.Create")
	defer span.End()

	query := `
	INSERT INTO orders (number, user_id, status, accrual, uploaded_at) 
	VALUES ($1, $2, $3, $4, $5)
	`

	_, err := repo.db.Exec(ctx, query, order.Number, order.UserID.String(),
		order.Status, order.Accrual, order.UploadedAt)
	if err != nil {
		span.RecordError(err)
		repo.logger.Error("can't exec query", zap.Error(err))
		return fmt.Errorf("[db.Exec]: %w", err)
	}

	span.SetAttributes(attribute.String("order_number", order.Number))
	repo.logger.Info("order created", zap.String("order_number", order.Number))
	return nil
}

func (repo *OrderRepoPostgres) GetByNumber(ctx context.Context,
	orderNumber string) (*model.Order, error) {
	ctx, span := otel.Tracer("repository").Start(ctx, "OrderRepo.GetByNumber")
	defer span.End()

	query := `
	SELECT number, user_id, status, accrual, uploaded_at
	FROM orders
	WHERE number = $1
	`

	var order model.Order
	err := repo.db.QueryRow(ctx, query, orderNumber).Scan(
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)
	if err != nil {
		span.RecordError(err)
		repo.logger.Error("can't scan row", zap.Error(err), zap.String("query", query))
		return nil, fmt.Errorf("[db.QueryRow]: %w", err)
	}

	span.SetAttributes(attribute.String("order_number", order.Number))
	repo.logger.Info("got order by number", zap.String("order_number", order.Number))
	return &order, nil
}

func (repo *OrderRepoPostgres) GetAll(ctx context.Context,
	userID string) ([]model.Order, error) {
	ctx, span := otel.Tracer("repository").Start(ctx, "OrderRepo.GetAll")
	defer span.End()

	query := `
	SELECT number, user_id, status, accrual, uploaded_at
	FROM orders
	WHERE user_id = $1
	ORDER BY uploaded_at DESC
	`

	rows, err := repo.db.Query(ctx, query, userID)
	if err != nil {
		span.RecordError(err)
		repo.logger.Error("query exec error", zap.String("query", query), zap.Error(err))
		return nil, fmt.Errorf("[db.Query]: %w", err)
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			span.RecordError(err)
			repo.logger.Error("scan row error", zap.String("query", query), zap.Error(err))
			return nil, fmt.Errorf("[rows.Scan]: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		span.RecordError(err)
		repo.logger.Error("error occured while reading rows", zap.Error(err))
		return nil, fmt.Errorf("[rows.Err]: %w", err)
	}

	span.SetAttributes(attribute.Int("orders_count", len(orders)))
	repo.logger.Info("got all of user orders", zap.Int("order_number", len(orders)))
	return orders, nil
}

func (repo *OrderRepoPostgres) GetAllPending(ctx context.Context) ([]string, error) {
	ctx, span := otel.Tracer("repository").Start(ctx, "OrderRepo.GetAllPending")
	defer span.End()

	query := `
	SELECT number
	FROM orders
	WHERE status NOT IN ('INVALID', 'PROCESSED')
	ORDER BY uploaded_at
	`

	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		span.RecordError(err)
		repo.logger.Error("query exec error", zap.String("query", query), zap.Error(err))
		return nil, fmt.Errorf("[db.Query]: %w", err)
	}
	defer rows.Close()

	orders := make([]string, 0)
	for rows.Next() {
		var order string
		err := rows.Scan(&order)
		if err != nil {
			span.RecordError(err)
			repo.logger.Error("scan row error", zap.String("query", query), zap.Error(err))
			return nil, fmt.Errorf("[rows.Scan]: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		span.RecordError(err)
		repo.logger.Error("error occured while reading rows", zap.Error(err))
		return nil, fmt.Errorf("[rows.Err]: %w", err)
	}

	span.SetAttributes(attribute.Int("orders_count", len(orders)))
	return orders, nil
}

func (repo *OrderRepoPostgres) Update(ctx context.Context, order model.Order) error {
	ctx, span := otel.Tracer("repository").Start(ctx, "OrderRepo.Update")
	defer span.End()

	query := `
	UPDATE orders SET status = $1, accrual = $2
	WHERE number = $3
	`

	_, err := repo.db.Exec(ctx, query, order.Status, order.Accrual, order.Number)
	if err != nil {
		span.RecordError(err)
		repo.logger.Error("error while excuting query", zap.String("query", query), zap.Error(err))
		return fmt.Errorf("[db.Exec]: %w", err)
	}

	span.SetAttributes(attribute.String("order_number", order.Number))
	repo.logger.Info("order updated", zap.String("order_number", order.Number))
	return nil
}

func (repo *OrderRepoPostgres) Delete(ctx context.Context, orderNumber string) error {
	ctx, span := otel.Tracer("repository").Start(ctx, "OrderRepo.Delete")
	defer span.End()
	query := `
	DELETE FROM orders WHERE number = $1
	`

	_, err := repo.db.Exec(ctx, query, orderNumber)
	if err != nil {
		span.RecordError(err)
		repo.logger.Error("error while executing query", zap.String("query", query), zap.Error(err))
		return fmt.Errorf("[db.Exec]: %w", err)
	}

	span.SetAttributes(attribute.String("order_number", orderNumber))
	repo.logger.Info("order deleted", zap.String("order_number", orderNumber))
	return nil
}
