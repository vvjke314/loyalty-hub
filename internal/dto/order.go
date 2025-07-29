package dto

import "github.com/vvjke314/itk-courses/loyalityhub/internal/model"

type AddOrderResponse struct {
	OrderNumber string `json:"orders"`
}

type GetAllOrdersResponse struct {
	Orders []model.Order `json:"orders"`
}
