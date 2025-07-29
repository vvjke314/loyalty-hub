package handlers

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/contextkeys"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/dto"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/model"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/services/interfaces"
	"go.opentelemetry.io/otel"
)

type OrderHandler struct {
	hostname string
	serv     interfaces.OrderServiceInterface
}

func NewOrderHandler(hostname string,
	orderService interfaces.OrderServiceInterface) *OrderHandler {
	return &OrderHandler{
		hostname: hostname,
		serv:     orderService,
	}
}

// LoadOrder godoc
// @Summary      Загрузка заказа
// @Description  Загрудает в систему новый товар
// @Security BearerAuth
// @Tags         order
// @Accept       plain
// @Param        input  body      string  true  "Номер заказа"
// @Produce      json
// @Success      200    {object}  dto.AddOrderResponse
// @Success 	 202    {object}  dto.ErrorResponse
// @Failure      400    {object}  dto.ErrorResponse
// @Failure      401    {object}  dto.ErrorResponse
// @Failure      409    {object}  dto.ErrorResponse
// @Failure 	 422    {object}  dto.ErrorResponse
// @Failure      500    {object}  dto.ErrorResponse
// @Router       /api/v1/user/orders [post]
func (h *OrderHandler) LoadOrder(c *gin.Context) {
	ctx, span := otel.Tracer("handler").Start(c.Request.Context(), "OrderHandler.LoadOrder")
	defer span.End()

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("failed to read request body"))
		return
	}

	orderNumber := strings.TrimSpace(string(bodyBytes))
	if orderNumber == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("order ID is required"))
		return
	}

	resp, err := h.serv.Load(ctx, orderNumber)
	if err != nil {
		var status int
		var message string
		switch {
		case errors.Is(err, model.ErrBadOrderNumber):
			status = http.StatusUnprocessableEntity
			message = "wrong order number format"
		case errors.Is(err, model.ErrOrderAlreadyExists):
			status = http.StatusOK
			message = "order already loaded"
		case errors.Is(err, model.ErrOrderLoadedByAnotherPerson):
			status = http.StatusConflict
			message = "this order loaded by another person"
		default:
			status = http.StatusInternalServerError
			message = "internal server error"
		}
		c.JSON(status, dto.NewErrorResponse(message))
		span.RecordError(err)
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// GetAllOrders godoc
// @Summary      Возвращает информацию по всем товарам пользователя
// @Description  Возвращает информацию по всем товарам пользователя
// @Security BearerAuth
// @Tags         order
// @Accept       json
// @Produce      json
// @Success      200    {object}  dto.GetAllOrdersResponse
// @Success 	 204    {object}  dto.ErrorResponse
// @Failure      401    {object}  dto.ErrorResponse
// @Failure      500    {object}  dto.ErrorResponse
// @Router       /api/v1/user/orders [get]
func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	ctx, span := otel.Tracer("handler").Start(c.Request.Context(), "OrderHandler.GetAllOrders")
	defer span.End()

	orders, err := h.serv.GetAll(ctx, ctx.Value(contextkeys.UserKeyID).(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("internal server error"))
		span.RecordError(err)
		return
	}

	if len(orders.Orders) == 0 {
		c.JSON(http.StatusNoContent, dto.NewErrorResponse("no data"))
		span.RecordError(err)
		return
	}

	c.JSON(http.StatusOK, orders)
}
