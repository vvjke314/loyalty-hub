package model

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	HashedRefreshToken string
	AuthAt             time.Time
	ExpireAt           time.Time
}
