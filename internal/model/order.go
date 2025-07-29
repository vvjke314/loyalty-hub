package model

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderStatus string

func (s *OrderStatus) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan OrderStatus from %T", value)
	}
	*s = OrderStatus(str)
	return nil
}

func (s OrderStatus) Value() (driver.Value, error) {
	return string(s), nil
}

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	Number     string
	UserID     uuid.UUID
	Status     OrderStatus
	Accrual    decimal.Decimal
	UploadedAt time.Time
}
