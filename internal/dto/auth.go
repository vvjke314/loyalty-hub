package dto

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/model"
	passmanager "github.com/vvjke314/itk-courses/loyalityhub/internal/utils/passmanager"
)

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (r *RegisterRequest) ToModel() (*model.User, error) {
	hashedPass, err := passmanager.HashPassword(r.Password)
	if err != nil {
		return nil, fmt.Errorf("[passmanager.HashPassword]: %w", err)
	}

	return &model.User{
		ID:        uuid.New(),
		Login:     r.Login,
		Password:  hashedPass,
		Balance:   decimal.NewFromInt(0),
		Withdrawn: decimal.NewFromInt(0),
	}, nil
}

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}
