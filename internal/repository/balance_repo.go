package repository

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/model"
	"go.uber.org/zap"
)

type BalanceRepoPostgres struct {
	db     DBExecutor
	logger *zap.Logger
}

func NewBalanceRepoPostgres(db DBExecutor, logger *zap.Logger) *BalanceRepoPostgres {
	return &BalanceRepoPostgres{
		db:     db,
		logger: logger.With(zap.String("repo", "balance")),
	}
}

func (r *BalanceRepoPostgres) Get(ctx context.Context, userID string) (*model.Balance, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN o.accrual IS NOT NULL THEN o.accrual ELSE 0 END), 0) AS current,
			COALESCE(SUM(w.amount), 0) AS withdrawn
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id
		LEFT JOIN withdrawals w ON u.id = w.user_id
		WHERE u.id = $1;
	`

	var balance model.Balance
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&balance.Current,
		&balance.Withdrawn,
	)
	if err != nil {
		r.logger.Error("failed to get balance", zap.Error(err))
		return nil, fmt.Errorf("get balance: %w", err)
	}

	return &balance, nil
}

func (r *BalanceRepoPostgres) AddWithdraw(ctx context.Context, withdrawal *model.Withdrawal) error {
	query := `
		INSERT INTO withdrawals (id, order_id, user_id, amount, processed_at)
		VALUES ($1, $2, $3, $4, $5);
	`

	_, err := r.db.Exec(
		ctx,
		query,
		withdrawal.ID,
		withdrawal.OrderID,
		withdrawal.UserID,
		withdrawal.Amount,
		withdrawal.ProcessedAt,
	)

	if err != nil {
		r.logger.Error("failed to add withdrawal", zap.Error(err))
		return fmt.Errorf("add withdrawal: %w", err)
	}

	return nil
}

func (r *BalanceRepoPostgres) GetAllWithdrawals(ctx context.Context, userID string) ([]model.Withdrawal, error) {
	query := `
		SELECT id, order_id, user_id, amount, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC;
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		r.logger.Error("failed to get withdrawals", zap.Error(err))
		return nil, fmt.Errorf("get withdrawals: %w", err)
	}
	defer rows.Close()

	var withdrawals []model.Withdrawal

	for rows.Next() {
		var w model.Withdrawal
		var amount decimal.Decimal

		err := rows.Scan(&w.ID, &w.OrderID, &w.UserID, &amount, &w.ProcessedAt)
		if err != nil {
			r.logger.Error("failed to scan withdrawal", zap.Error(err))
			return nil, fmt.Errorf("scan withdrawal: %w", err)
		}

		w.Amount = amount
		withdrawals = append(withdrawals, w)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("rows error", zap.Error(err))
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return withdrawals, nil
}
