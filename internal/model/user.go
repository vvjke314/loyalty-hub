package model

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type User struct {
	ID        uuid.UUID
	Login     string
	Password  string
	Balance   decimal.Decimal
	Withdrawn decimal.Decimal
	Salt      string
}
