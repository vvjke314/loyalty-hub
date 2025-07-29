package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// конкретная реализация postgres
type UserRepoPostgres struct {
	db     DBExecutor
	logger *zap.Logger
}

func NewUserRepoPostgres(db DBExecutor, logger *zap.Logger) *UserRepoPostgres {
	return &UserRepoPostgres{
		db:     db,
		logger: logger.With(zap.String("repo", "user")),
	}
}

func (repo *UserRepoPostgres) Create(ctx context.Context, user *model.User) error {
	ctx, span := otel.Tracer("repository").Start(ctx, "UserRepo.Create")
	defer span.End()

	query := "INSERT INTO users (id, login, password, balance, withdrawn) " +
		"VALUES ($1, $2, $3, $4, $5)"

	_, err := repo.db.Exec(ctx, query, user.ID, user.Login, user.Password,
		user.Balance, user.Withdrawn)

	if err != nil {
		repo.logger.Error("can't insert row", zap.Error(err))
		span.RecordError(err)
		return fmt.Errorf("[pgxpool.Conn.Exec]: %w", err)
	}

	span.SetAttributes(attribute.String("user.id", user.ID.String()))
	span.SetAttributes(attribute.String("db.query", query))
	repo.logger.Info("user registered successfully", zap.String("user_id", user.ID.String()))
	return nil
}

func (repo *UserRepoPostgres) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	ctx, span := otel.Tracer("repository").Start(ctx, "UserRepo.GetByID")
	defer span.End()

	var user model.User
	query := "SELECT id, login, password, balance, withdrawn " +
		"FROM users WHERE login=$1"

	err := repo.db.QueryRow(ctx, query, login).Scan(&user.ID, &user.Login, &user.Password,
		&user.Balance, &user.Withdrawn)

	if err != nil {
		if err == pgx.ErrNoRows {
			repo.logger.Error("no such user", zap.Error(err))
			span.RecordError(err)
			return nil, ErrNoUser
		}

		repo.logger.Error("[pgxpool.Conn.QueryRow]", zap.Error(err))
		span.RecordError(err)
		return nil, fmt.Errorf("[pgxpool.Conn.QueryRow]: %w", err)
	}

	span.SetAttributes(attribute.String("user.id", user.ID.String()))
	span.SetAttributes(attribute.String("db.query", query))
	repo.logger.Info("succcessfully get user by ID", zap.String("user_id", user.ID.String()))
	return &user, nil
}
