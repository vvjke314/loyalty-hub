package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/contextkeys"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/dto"
	tokenmanager "github.com/vvjke314/itk-courses/loyalityhub/internal/utils/tokenManager"
	"go.opentelemetry.io/otel"
)

func AuthMiddleware(tm *tokenmanager.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := otel.Tracer("middleware").Start(c.Request.Context(), "AuthMiddleware")
		defer span.End()
		bearerRawToken := c.GetHeader("Authorization")
		parts := strings.SplitN(bearerRawToken, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			span.RecordError(errors.New("missing or malformed token"))
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse("missing or malformed token"))
			return
		}

		token := parts[1]

		userID, err := tm.ParseToken(token)
		if err != nil {
			span.RecordError(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse("invalid token"))
			return
		}

		// добавляем в контекст id клиента
		ctx = context.WithValue(ctx, contextkeys.UserKeyID, userID.String())

		// прокидываем наш контекст дальше
		c.Request = c.Request.WithContext(ctx)
		c.Set(string(contextkeys.UserKeyID), userID.String())
		c.Next()
	}
}
