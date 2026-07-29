package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/srm-asset/srm-backend/internal/platform/metricas"
)

func Metricas() gin.HandlerFunc {
	return func(c *gin.Context) {
		inicio := time.Now()
		c.Next()
		rota := c.FullPath()
		if rota == "" {
			rota = "sem_rota"
		}
		status := strconv.Itoa(c.Writer.Status())
		metricas.Requisicoes.WithLabelValues(rota, c.Request.Method, status).Inc()
		metricas.Latencia.WithLabelValues(rota, c.Request.Method, status).Observe(time.Since(inicio).Seconds())
	}
}
