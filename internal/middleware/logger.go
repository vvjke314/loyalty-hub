package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Выполнить следующий хендлер
		c.Next()

		// Получить статус ответа
		status := c.Writer.Status()

		// Получить userID из контекста (если установлен AuthMiddleware)
		userID, _ := c.Get("userID")

		// Получить Trace ID из контекста OpenTelemetry (если есть)
		var traceID string
		if span := trace.SpanFromContext(c.Request.Context()); span != nil && span.SpanContext().IsValid() {
			traceID = span.SpanContext().TraceID().String()
		}

		logger.Info("incoming request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()), // шаблон маршрута (если есть), иначе c.Request.URL.Path
			zap.Int("status", status),
			zap.Duration("duration", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
			zap.Any("user_id", userID),
			zap.String("trace_id", traceID),
		)
	}
}
