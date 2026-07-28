# Multi-Service — Multiplos Mains com DI Manual

## Quando Usar

- Monolito que precisa escalar para multiplos processos
- Worker/Consumer separado da API
- Scheduler/Cron jobs independentes
- Cada servico compartilha domain/service/repository mas tem entrypoint proprio

## Layout Multi-Service

```
cmd/
  server/main.go         # HTTP API (Gin + Huma)
  worker/main.go         # RabbitMQ/Kafka consumer
  scheduler/main.go      # Cron jobs

internal/
  server/
    server.go            # DI da API (sempre presente)
    worker.go            # DI do worker (MULTI_SERVICE=1)
    scheduler.go         # DI do scheduler (MULTI_SERVICE=1)
  domain/                # compartilhado
  service/               # compartilhado
  repository/            # compartilhado
  database/              # compartilhado
```

## Regra de Ouro

Cada `server/*.go` e um `New(ctx, cfg, log) (*Server, error)` independente.
Nenhum importa outro. Todos compartilham `domain/`, `service/`, `repository/`.

## Exemplo: server.go (HTTP API)

```go
package server

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) (*Server, error) {
    pool, err := database.NewPool(ctx, cfg.DB)
    if err != nil { return nil, fmt.Errorf("pool: %w", err) }

    queries := sqlc.New(pool)
    userRepo := repository.NewUserRepository(queries)
    userSvc := service.NewUserService(userRepo, log)

    ginEngine := gin.New()
    ginEngine.Use(middleware.CorrelationID(), middleware.Logger(log))
    handler.NewUserHandler(userSvc).RegisterRoutes(ginEngine)

    return &Server{
        httpServer: &http.Server{Addr: ":" + cfg.Port, Handler: ginEngine},
        closers:    []func() error{pool.Close},
    }, nil
}
```

## Exemplo: worker.go (RabbitMQ Consumer)

```go
package server

func NewWorker(ctx context.Context, cfg *config.Config, log *slog.Logger) (*Worker, error) {
    pool, err := database.NewPool(ctx, cfg.DB)
    if err != nil { return nil, fmt.Errorf("pool: %w", err) }

    queries := sqlc.New(pool)
    orderRepo := repository.NewOrderRepository(queries)
    orderSvc := service.NewOrderService(orderRepo, log)

    consumer, err := rabbitmq.NewConsumer(cfg.RabbitURL)
    if err != nil { return nil, fmt.Errorf("consumer: %w", err) }

    return &Worker{
        consumer: consumer,
        handler:  handler.NewOrderConsumer(orderSvc, log),
        closers:  []func() error{pool.Close, consumer.Close},
    }, nil
}

func (w *Worker) Run(ctx context.Context) error {
    return w.consumer.Run(ctx, "orders.queue", w.handler.Process)
}
```

## Exemplo: scheduler.go (Cron Jobs)

```go
package server

func NewScheduler(ctx context.Context, cfg *config.Config, log *slog.Logger) (*Scheduler, error) {
    pool, err := database.NewPool(ctx, cfg.DB)
    if err != nil { return nil, fmt.Errorf("pool: %w", err) }

    queries := sqlc.New(pool)
    reportRepo := repository.NewReportRepository(queries)
    reportSvc := service.NewReportService(reportRepo, log)

    return &Scheduler{
        cron:    cron.New(cron.WithSeconds()),
        handler: handler.NewReportScheduler(reportSvc, log),
        closers: []func() error{pool.Close},
    }, nil
}

func (s *Scheduler) Run(ctx context.Context) error {
    s.cron.AddFunc("0 0 * * * *", func() { s.handler.GenerateDailyReport(ctx) })
    s.cron.Start()
    <-ctx.Done()
    s.cron.Stop()
    return nil
}
```

## Exemplo: cmd/worker/main.go

```go
package main

func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    cfg, err := config.Load()
    if err != nil { slog.Default().Error("config", "error", err); os.Exit(1) }

    log := logger.New(cfg.LogLevel)

    worker, err := server.NewWorker(ctx, cfg, log)
    if err != nil { log.Error("worker init", "error", err); os.Exit(1) }

    go func() {
        if err := worker.Run(ctx); err != nil {
            log.Error("worker run", "error", err)
            cancel()
        }
    }()

    <-ctx.Done()
    log.Info("shutting down worker")
    worker.Shutdown(context.Background())
}
```

## Padroes

- **DI manual em cada constructor** — nao compartilhar servidor entre entrypoints
- **Closers explicitos** — cada servico gerencia seus proprios recursos
- **Signal handling identico** — todos usam `signal.NotifyContext`
- **Graceful shutdown** — todos chamam `Shutdown()` no sinal
- **Zero acoplamento entre entrypoints** — worker nao importa scheduler

## Anti-padroes

- ❌ Compartilhar instancia de servidor entre cmd/ (acoplamento)
- ❌ Worker importar pacote do scheduler (imports ciclicos)
- ❌ Inicializar recursos em `init()` (impossivel testar)
- ❌ Esquecer `defer tx.Rollback()` no worker

## Makefile

```makefile
.PHONY: run-server run-worker run-scheduler

run-server:
	go run ./cmd/server

run-worker:
	go run ./cmd/worker

run-scheduler:
	go run ./cmd/scheduler

build-all:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server
	CGO_ENABLED=0 go build -o bin/worker ./cmd/worker
	CGO_ENABLED=0 go build -o bin/scheduler ./cmd/scheduler
```

## Dockerfile Multi-Stage (Multi-Service)

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Builda todos os binarios
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/server ./cmd/server
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/worker ./cmd/worker
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/scheduler ./cmd/scheduler

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /bin/server /bin/worker /bin/scheduler /bin/
# Entrypoint definido no docker-compose ou k8s
```
