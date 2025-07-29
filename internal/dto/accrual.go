package dto

type AccrualServiceResponse struct {
	OrderNumber string  `json:"order"`
	Status      string  `json:"status"`
	Accrual     float64 `json:"accrual"`
}
