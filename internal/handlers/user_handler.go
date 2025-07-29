package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/dto"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/model"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/services/interfaces"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type UserHandler struct {
	hostname    string
	UserService interfaces.UserServiceInterface
}

func NewUserHandler(hostname string,
	userService interfaces.UserServiceInterface) *UserHandler {
	return &UserHandler{
		hostname:    hostname,
		UserService: userService,
	}
}

// Register godoc
// @Summary      Регистрация пользователя
// @Description  Регистрирует нового пользователя и возвращает access/refresh токены
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        input  body      dto.RegisterRequest  true  "Данные для регистрации"
// @Success      200    {object}  dto.AuthResponse
// @Failure      400    {object}  map[string]string
// @Failure 	 409 	{object} map[string]string
// @Failure 	500 {object} map[string]string
// @Router       /api/v1/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	ctx, span := otel.Tracer("handler").Start(c.Request.Context(), "UserHandler.Register")
	defer span.End()

	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("invalid request"))
		span.RecordError(err)
		return
	}
	resp, err := h.UserService.Register(ctx, req)
	if err != nil {
		if errors.Is(err, model.ErrAlreadyExits) {
			c.JSON(http.StatusConflict, dto.NewErrorResponse(err.Error()))
			span.RecordError(err)
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err.Error()))
		span.RecordError(err)
		return
	}
	span.SetAttributes(attribute.String("user.login", req.Login))
	c.SetCookie("refresh_token", resp.RefreshToken, 2592000, "/", h.hostname, false, true) //только при разработке
	c.JSON(http.StatusOK, resp)
}

// Auth godoc
// @Summary      Аутентификация пользователя
// @Description  Аутентифицирует пользователя и возвращает access/refresh токены
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        input  body      dto.AuthRequest  true  "Данные для входа"
// @Success      200    {object}  dto.AuthResponse
// @Failure      400    {object}  map[string]string
// @Failure      401    {object}  map[string]string
// @Router       /api/v1/auth [post]
func (h *UserHandler) Auth(c *gin.Context) {
	ctx, span := otel.Tracer("handler").Start(c.Request.Context(), "UserHandler.Auth")
	defer span.End()

	var req dto.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("invalid request"))
		span.RecordError(err)
		return
	}
	resp, err := h.UserService.Auth(ctx, req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(err.Error()))
		span.RecordError(err)
		return
	}
	span.SetAttributes(attribute.String("user.login", req.Login))
	c.SetCookie("refresh_token", resp.RefreshToken, 2592000, "/", h.hostname, false, true) // только при разработке
	c.JSON(http.StatusOK, resp)
}

// Refresh godoc
// @Summary      Обновление access токена
// @Description  Обновляет access токен по refresh токену
// @Tags         user
// @Accept       json
// @Produce      json
// @Success      200    {object}  dto.RefreshResponse
// @Failure      400    {object}  dto.ErrorResponse
// @Failure      401    {object}  dto.ErrorResponse
// @Router       /api/v1/refresh [get]
func (h *UserHandler) Refresh(c *gin.Context) {
	ctx, span := otel.Tracer("handler").Start(c.Request.Context(), "UserHandler.Refresh")
	defer span.End()

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("no refresh_token cookie"))
		span.RecordError(err)
		return
	}
	req.RefreshToken = cookie
	resp, err := h.UserService.GetNewAccessToken(ctx, req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(err.Error()))
		span.RecordError(err)
		return
	}

	span.SetAttributes(attribute.String("refresh_token", req.RefreshToken))
	c.JSON(http.StatusOK, resp)
}
