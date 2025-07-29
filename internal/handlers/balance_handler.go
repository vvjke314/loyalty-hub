package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/dto"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/model"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/services/interfaces"
	"go.opentelemetry.io/otel"
)

type BalanceHandler struct {
	hostname string
	serv     interfaces.BalanceServiceInterface
}

// фабрика
func NewBalanceHandler(hostname string,
	orderService interfaces.BalanceServiceInterface) *BalanceHandler {
	return &BalanceHandler{
		hostname: hostname,
		serv:     orderService,
	}
}

// GetBalance godoc
// @Summary      Текущий баланс пользователя
// @Description  Возвращает текущий баланс и сумму выведенных средств
// @Security     BearerAuth
// @Tags         balance
// @Produce      json
// @Success      200  {object}  dto.GetBalanceResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /api/v1/user/balance [get]
func (h *BalanceHandler) GetBalance(c *gin.Context) {
	ctx, span := otel.Tracer("handler").Start(c.Request.Context(), "BalanceHandler.GetBalance")
	defer span.End()

	res, err := h.serv.GetBalance(ctx)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("failed to get balance"))
		return
	}

	c.JSON(http.StatusOK, res)
}

// Withdraw godoc
// @Summary      Списание средств с баланса
// @Description  Позволяет списать средства на заказ
// @Security     BearerAuth
// @Tags         balance
// @Accept       json
// @Produce      json
// @Param        input  body      dto.NewWithdrawnRequest  true  "Данные списания"
// @Success      200    {string}  string  "успешное списание"
// @Failure      400    {object}  dto.ErrorResponse
// @Failure      401    {object}  dto.ErrorResponse
// @Failure      402    {object}  dto.ErrorResponse
// @Failure      500    {object}  dto.ErrorResponse
// @Router       /api/v1/user/balance/withdraw [post]
func (h *BalanceHandler) Withdraw(c *gin.Context) {
	ctx, span := otel.Tracer("handler").Start(c.Request.Context(), "BalanceHandler.Withdraw")
	defer span.End()

	var req dto.NewWithdrawnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("invalid request body"))
		return
	}

	err := h.serv.Withdraw(ctx, req)
	if err != nil {
		span.RecordError(err)
		switch {
		case errors.Is(err, model.ErrInsufficientFunds):
			c.JSON(http.StatusPaymentRequired, dto.NewErrorResponse("not enough funds"))
		default:
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("withdraw failed"))
		}
		return
	}

	c.JSON(http.StatusOK, "withdraw completed")
}

// GetWithdrawals godoc
// @Summary      История выводов средств
// @Description  Получение всех транзакций списания пользователя
// @Security     BearerAuth
// @Tags         balance
// @Produce      json
// @Success      200  {object}  dto.GetAllWithdrawalsResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /api/v1/user/withdrawals [get]
func (h *BalanceHandler) GetWithdrawals(c *gin.Context) {
	ctx, span := otel.Tracer("handler").Start(c.Request.Context(), "BalanceHandler.GetWithdrawals")
	defer span.End()

	res, err := h.serv.GetWithdrawals(ctx)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("failed to get withdrawals"))
		return
	}

	c.JSON(http.StatusOK, res)
}
