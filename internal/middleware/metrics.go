package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vvjke314/itk-courses/loyalityhub/internal/metrics"
)

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// Получаем данные для лейблов
		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = "undefined"
		}

		// Записываем метрики
		metrics.HttpRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HttpRequestDuration.WithLabelValues(method, path, status).Observe(time.Since(start).Seconds())
	}
}
