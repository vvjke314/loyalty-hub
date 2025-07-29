package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/dto"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/model"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/repository"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/utils/passmanager"
	tokenmanager "github.com/vvjke314/itk-courses/loyalityhub/internal/utils/tokenManager"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type UserService struct {
	repo   *repository.Repositories
	tm     *tokenmanager.TokenManager
	logger *zap.Logger
}

func NewUserService(logger *zap.Logger,
	repos *repository.Repositories) *UserService {
	return &UserService{
		logger: logger,
		repo:   repos,
		tm:     tokenmanager.NewTokenManager(tokenmanager.NewTokenManagerConfig()),
	}
}

func (u *UserService) createTokens(userID uuid.UUID) (string, string, error) {
	// генерация рефреш токена
	refreshToken, err := u.tm.GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("can not generate refresh token %w", err)
	}

	// генерация ацесс токена
	accessToken, err := u.tm.GenerateAccessToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("can not generate access token %w", err)
	}
	return accessToken, refreshToken, nil
}

func (u *UserService) Register(ctx context.Context,
	req dto.RegisterRequest) (dto.AuthResponse, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "UserService.Register")
	defer span.End()

	tx, err := u.repo.BeginTx(ctx, pgx.ReadUncommitted)
	if err != nil {
		return dto.AuthResponse{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	// получаем методы для работы с пользователями
	userRepo := u.repo.NewUserRepo(tx)
	sessionRepo := u.repo.NewSessionRepo(tx)

	// маппим dto->model
	user, err := req.ToModel()
	if err != nil {
		u.logger.Error("can't parse req to model", zap.Error(err))
		return dto.AuthResponse{}, err
	}

	// проверка на отстутсвие пользователя
	tempUser, err := userRepo.GetByLogin(ctx, user.Login)
	if !errors.Is(err, repository.ErrNoUser) {
		if tempUser != nil {
			u.logger.Error("login already taken", zap.Error(err))
			return dto.AuthResponse{}, model.ErrAlreadyExits
		}
		u.logger.Error("error while gettins user login", zap.Error(err))
		return dto.AuthResponse{}, err
	}

	// создание пользователя
	err = userRepo.Create(ctx, user)
	if err != nil {
		u.logger.Error("error while creating user", zap.Error(err))
		return dto.AuthResponse{}, err
	}

	accessToken, refreshToken, err := u.createTokens(user.ID)
	if err != nil {
		u.logger.Error("error while creating tokens", zap.Error(err))
		return dto.AuthResponse{}, err
	}

	session := &model.Session{
		ID:                 uuid.New(),
		UserID:             user.ID,
		HashedRefreshToken: u.tm.HashToken(refreshToken),
		AuthAt:             time.Now(),
		ExpireAt:           time.Now().Add(u.tm.GetRefreshTokenTTL()),
	}
	// создаем сессию
	err = sessionRepo.Create(ctx, session)
	if err != nil {
		u.logger.Error("error while creating session", zap.Error(err))
		return dto.AuthResponse{}, fmt.Errorf("session.Repo %w", err)
	}

	// генерируем трейсы в спанах
	span.SetAttributes(attribute.String("user.id", user.ID.String()))
	u.logger.Info("user registered successfully", zap.String("user.id", user.ID.String()))
	// формируем ответ
	return dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *UserService) Auth(ctx context.Context,
	credentials dto.AuthRequest) (dto.AuthResponse, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "UserService.Auth")
	defer span.End()

	tx, err := u.repo.BeginTx(ctx, pgx.ReadUncommitted)
	if err != nil {
		u.logger.Error("failed to begin tx", zap.Error(err))
		return dto.AuthResponse{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	userRepo := u.repo.NewUserRepo(tx)
	sessionRepo := u.repo.NewSessionRepo(tx)

	user, err := userRepo.GetByLogin(ctx, credentials.Login)
	if err != nil {
		u.logger.Error("user not found", zap.Error(err))
		return dto.AuthResponse{}, errors.New("invalid login or password")
	}

	if !passmanager.CheckPass(credentials.Password, user.Password) {
		u.logger.Error("invalid password", zap.String("login", credentials.Login))
		return dto.AuthResponse{}, errors.New("invalid login or password")
	}

	accessToken, refreshToken, err := u.createTokens(user.ID)
	if err != nil {
		u.logger.Error("error while creating tokens", zap.Error(err))
		return dto.AuthResponse{}, err
	}

	session := &model.Session{
		ID:                 uuid.New(),
		UserID:             user.ID,
		HashedRefreshToken: u.tm.HashToken(refreshToken),
		AuthAt:             time.Now(),
		ExpireAt:           time.Now().Add(u.tm.GetRefreshTokenTTL()),
	}
	// создаем сессию
	err = sessionRepo.Create(ctx, session)
	if err != nil {
		u.logger.Error("error while creating session", zap.Error(err))
		return dto.AuthResponse{}, fmt.Errorf("session.Repo %w", err)
	}

	span.SetAttributes(attribute.String("user.id", user.ID.String()))
	u.logger.Info("user authenticated successfully", zap.String("user.id", user.ID.String()))
	return dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *UserService) GetNewAccessToken(ctx context.Context,
	refreshToken string) (dto.RefreshResponse, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "UserService.GetNewAccessToken")
	defer span.End()

	tx, err := u.repo.BeginTx(ctx, pgx.ReadUncommitted)
	if err != nil {
		u.logger.Error("failed to begin tx", zap.Error(err))
		return dto.RefreshResponse{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	sessionRepo := u.repo.NewSessionRepo(tx)
	hashedRefresh := u.tm.HashToken(refreshToken)
	session, err := sessionRepo.GetByRefreshToken(ctx, hashedRefresh)
	if err != nil || session == nil {
		u.logger.Error("session not found or error", zap.Error(err))
		return dto.RefreshResponse{}, errors.New("invalid refresh token")
	}

	if session.ExpireAt.Before(time.Now()) {
		u.logger.Error("refresh token expired", zap.String("session_id", session.ID.String()))
		return dto.RefreshResponse{}, errors.New("refresh token expired")
	}

	accessToken, err := u.tm.GenerateAccessToken(session.UserID)
	if err != nil {
		u.logger.Error("error while creating access token", zap.Error(err))
		return dto.RefreshResponse{}, err
	}

	span.SetAttributes(attribute.String("user.id", session.UserID.String()))
	u.logger.Info("access token refreshed successfully", zap.String("user.id", session.UserID.String()))
	return dto.RefreshResponse{
		AccessToken: accessToken,
	}, nil
}
