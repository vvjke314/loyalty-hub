package interfaces

import (
	"context"

	"github.com/vvjke314/itk-courses/loyalityhub/internal/dto"
)

type UserServiceInterface interface {
	Register(ctx context.Context, credentials dto.RegisterRequest) (dto.AuthResponse, error)
	Auth(ctx context.Context, credentials dto.AuthRequest) (dto.AuthResponse, error)
	GetNewAccessToken(ctx context.Context, refresh string) (dto.RefreshResponse, error)
}
