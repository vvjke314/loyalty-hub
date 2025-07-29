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

type SessionRepoPostgres struct {
	db     DBExecutor
	logger *zap.Logger
}

func NewSessionRepoPostgres(db DBExecutor, logger *zap.Logger) *SessionRepoPostgres {
	return &SessionRepoPostgres{
		db:     db,
		logger: logger.With(zap.String("repo", "session")),
	}
}

func (repo *SessionRepoPostgres) Create(ctx context.Context, session *model.Session) error {
	ctx, span := otel.Tracer("repository").Start(ctx, "SessionRepo.Create")
	defer span.End()

	query := `
		INSERT INTO sessions (id, user_id, hashed_refresh_token, auth_at, expire_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := repo.db.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.HashedRefreshToken,
		session.AuthAt,
		session.ExpireAt,
	)

	if err != nil {
		repo.logger.Error("failed to create session", zap.Error(err))
		span.RecordError(err)
		return fmt.Errorf("session create: %w", err)
	}

	span.SetAttributes(attribute.String("session.id", session.ID.String()))
	repo.logger.Info("session created", zap.String("session_id", session.ID.String()))
	return nil
}

func (repo *SessionRepoPostgres) GetByRefreshToken(ctx context.Context, token string) (*model.Session, error) {
	ctx, span := otel.Tracer("repository").Start(ctx, "SessionRepo.GetByRefreshToken")
	defer span.End()

	var s model.Session
	query := `
		SELECT id, user_id, hashed_refresh_token, auth_at, expire_at
		FROM sessions
		WHERE hashed_refresh_token = $1
	`

	err := repo.db.QueryRow(ctx, query, token).Scan(
		&s.ID,
		&s.UserID,
		&s.HashedRefreshToken,
		&s.AuthAt,
		&s.ExpireAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			repo.logger.Warn("session not found", zap.String("refresh_token", token))
			return nil, nil
		}
		repo.logger.Error("failed to get session", zap.Error(err))
		span.RecordError(err)
		return nil, fmt.Errorf("get session: %w", err)
	}

	span.SetAttributes(attribute.String("session.id", s.ID.String()))
	return &s, nil
}

func (repo *SessionRepoPostgres) Delete(ctx context.Context, sessionID string) error {
	ctx, span := otel.Tracer("repository").Start(ctx, "SessionRepo.Delete")
	defer span.End()

	query := `DELETE FROM sessions WHERE id = $1`

	_, err := repo.db.Exec(ctx, query, sessionID)
	if err != nil {
		repo.logger.Error("failed to delete session", zap.String("session_id", sessionID), zap.Error(err))
		span.RecordError(err)
		return fmt.Errorf("delete session: %w", err)
	}

	repo.logger.Info("session deleted", zap.String("session_id", sessionID))
	span.SetAttributes(attribute.String("session.id", sessionID))
	return nil
}
