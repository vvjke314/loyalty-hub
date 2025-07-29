package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Balance struct {
	Current   float64
	Withdrawn float64
}

type Withdrawal struct {
	ID          uuid.UUID
	OrderID     string
	UserID      uuid.UUID
	Amount      decimal.Decimal
	ProcessedAt time.Time
}
