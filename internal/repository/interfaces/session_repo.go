package interfaces

import (
	"context"

	"github.com/vvjke314/itk-courses/loyalityhub/internal/model"
)

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	GetByRefreshToken(ctx context.Context, refreshToken string) (*model.Session, error)
	Delete(ctx context.Context, sessionID string) error
}
