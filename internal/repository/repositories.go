package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/repository/interfaces"
	"go.uber.org/zap"
)

type Repositories struct {
	pgxpool *pgxpool.Pool
	logger  *zap.Logger
}

func NewRepositories(logger *zap.Logger) *Repositories {
	return &Repositories{
		logger: logger.With(zap.String("layer", "repository")),
	}
}

// TO-DO: сделать настройку c pgx-zap adapter
func (repos *Repositories) Init(ctx context.Context, dsn string) error {
	// connConfig, err := pgxpool.ParseConfig(dsn)
	// if err != nil {
	// 	return fmt.Errorf("can't parse dsn: %w", err)
	// }

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		repos.logger.Error("can't create connection pool", zap.Error(err))
		return fmt.Errorf("can't parse db: %w", err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		repos.logger.Error("ping to DB failed", zap.Error(err))
		return fmt.Errorf("can't ping db: %w", err)
	}

	repos.pgxpool = pool
	return nil
}

// Возвращает транзакцию для работыт с БД
func (r *Repositories) BeginTx(ctx context.Context, isoLevel pgx.TxIsoLevel) (pgx.Tx, error) {
	return r.pgxpool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: isoLevel,
	})
}

func (repos *Repositories) Close() {
	repos.pgxpool.Close()
}

// фабрики репозиториев
func (repos *Repositories) NewUserRepo(exec DBExecutor) interfaces.UserRepositoryInterface {
	return NewUserRepoPostgres(exec, repos.logger)
}

func (repos *Repositories) NewSessionRepo(exec DBExecutor) interfaces.SessionRepository {
	return NewSessionRepoPostgres(exec, repos.logger)
}

func (repos *Repositories) NewOrderRepo(exec DBExecutor) interfaces.OrderRepository {
	return NewOrderRepoPostgres(exec, repos.logger)
}

func (repos *Repositories) NewBalanceRepo(exec DBExecutor) interfaces.BalanceRepository {
	return NewBalanceRepoPostgres(exec, repos.logger)
}
