# Prometheus — Metrics

## Setup

```go
package metrics

import (
    "strconv"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    HTTPDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "Latencia HTTP",
            Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
        },
        []string{"method", "path", "status"},
    )
    
    HTTPRequests = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total de requests HTTP",
        },
        []string{"method", "path", "status"},
    )
)
```

## Middleware

```go
func Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        
        path := c.FullPath()
        if path == "" { path = "unknown" }
        status := strconv.Itoa(c.Writer.Status())
        
        HTTPDuration.WithLabelValues(c.Request.Method, path, status).
            Observe(time.Since(start).Seconds())
        HTTPRequests.WithLabelValues(c.Request.Method, path, status).Inc()
    }
}
```

## Padroes
- Histogram para latencia (buckets padrao)
- Counter para throughput
- Labels: method, path, status
- Path vazio → "unknown"
- Endpoint `/metrics` exposto via Gin
