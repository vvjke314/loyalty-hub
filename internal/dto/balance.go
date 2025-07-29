package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type GetBalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type NewWithdrawnRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type Withdrawn struct {
	Order       string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"processed_at"`
}

type GetAllWithdrawalsResponse struct {
	Withdrawals []Withdrawn
}
