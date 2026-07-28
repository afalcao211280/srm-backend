# Gin — Router e Middlewares

## Setup Basico

```go
r := gin.New()
r.Use(gin.Recovery())
r.Use(middleware.CorrelationID())
r.Use(middleware.Logger(log))
r.Use(middleware.Tracing("user-service"))
r.Use(metrics.Middleware())

r.GET("/health", handler.Health)
r.GET("/metrics", gin.WrapH(promhttp.Handler()))

v1 := r.Group("/api/v1")
handler.RegisterUsers(v1, userSvc, log)
```

## Middleware Correlation ID

```go
package middleware

import (
    "context"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type correlationKey struct{}
const headerCorrelationID = "X-Correlation-ID"

func CorrelationID() gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.GetHeader(headerCorrelationID)
        if id == "" { id = uuid.NewString() }
        ctx := context.WithValue(c.Request.Context(), correlationKey{}, id)
        c.Request = c.Request.WithContext(ctx)
        c.Header(headerCorrelationID, id)
        c.Next()
    }
}
```

## Middleware Logger

```go
func Logger(log *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        cid := CorrelationIDFrom(c.Request.Context())
        reqLog := log.With("correlation_id", cid, "method", c.Request.Method, "path", c.FullPath())
        ctx := logger.With(c.Request.Context(), reqLog)
        c.Request = c.Request.WithContext(ctx)
        c.Next()
        reqLog.Info("request concluido",
            "status", c.Writer.Status(),
            "duration_ms", time.Since(start).Milliseconds(),
            "bytes", c.Writer.Size(),
        )
    }
}
```

## Regras
- Sempre `c.Request.Context()` — nunca `context.Background()`
- Health e metrics sem auth
- Group `/api/v1` para versionamento
