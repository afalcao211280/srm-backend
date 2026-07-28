# Workflow — Scaffold e Geracao de Codigo

## Quando Usar

- Ao criar **projeto novo** do zero
- Ao gerar **feature nova** em projeto existente
- Ao criar **novo microsservico**

## Decisoes Iniciais (Pergunte Primeiro)

1. **Nome do servico**: kebab-case (ex: `user-service`)
2. **Tipo**: HTTP API / Worker / CLI / CronJob
3. **Banco**: PostgreSQL (sqlc+tern) / MongoDB / Nenhum
4. **Deps externas**: Redis, Kafka, RabbitMQ, MinIO, APIs externas
5. **Modulo Go**: `github.com/example-<service-name>`

## Scaffold de Projeto Novo

### Ordem de Criacao

1. `go.mod`
2. `.gitignore`, `.env.example`
3. `Makefile`, `.golangci.yml`
4. `pkg/` (logger, errors, tracer, metrics)
5. `internal/config/config.go`
6. `internal/middleware/` (correlation_id, logger, tracing)
7. `internal/handler/health.go`
8. `cmd/<service>/main.go`
9. `internal/server/server.go`
10. `docker/Dockerfile`, `docker/docker-compose.dev.yml`
11. `README.md`

### Estrutura de Pastas

```
<service-name>/
cmd/<service>/main.go
internal/
server/server.go
config/config.go
handler/health.go
middleware/
service/.gitkeep
repository/.gitkeep
domain/.gitkeep
model/.gitkeep
pkg/
logger/
errors/
tracer/
metrics/
docker/
Dockerfile
docker-compose.dev.yml
.env.example
.golangci.yml
Makefile
README.md
```

### Arquivos Invariantes (Gerar Sempre)

**`cmd/<service>/main.go`**:
- signal.NotifyContext para graceful shutdown
- Config load
- Logger init
- Tracer init (nao bloqueante)
- server.New() + DI manual
- HTTP server com timeouts

**`internal/server/server.go`**:
- Constructors `New*` explicitos
- Montagem de dependencias na ordem: infra → repo → service → handler
- Graceful shutdown com closers

**`pkg/logger/logger.go`**:
- slog JSONHandler
- With/From para contexto

**`pkg/tracer/tracer.go`**:
- OpenTelemetry OTLP gRPC
- Span helper

**`pkg/metrics/metrics.go`**:
- Prometheus histogram + counter
- Gin middleware

### Dockerfile Canonico

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum./
RUN go mod download
COPY..
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/server./cmd/<service>

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
```

## Workflow de Feature Nova

### Ordem de Dependencia

```
domain → migration → sqlc query → repository → service → handler → test
```

### Passos

1. **Domain**: struct pura, zero deps
2. **Migration**: `NNN_<nome>.sql` com up/down
3. **Query sqlc**: `queries/<entity>.sql` com anotacoes `:one`/`:many`/`:exec`
4. **Repository**: implementa interface do service; converte sqlc → domain
5. **Service**: declara interface do repository; logica de negocio; span OTel
6. **Handler**: Gin/Huma; validacao; chama service
7. **Testes**: unit (mock repo) + integracao (testcontainers)
8. **DI**: atualizar `internal/server/server.go`

### Checklist de Feature

- [ ] Domain puro (zero deps externas)
- [ ] Migration testada (up/down/up)
- [ ] Query sqlc gerada e compilando
- [ ] Repository converte tipos sqlc → domain
- [ ] Service tem span OTel
- [ ] Handler valida input
- [ ] Testes unit + integracao
- [ ] server.go atualizado com DI

## Checklist de Scaffold

- [ ] `go.mod` com nome correto
- [ ] `make build` compila
- [ ] `make lint` passa
- [ ] `health.go` retorna 200
- [ ] `/metrics` expoe metricas
- [ ] Correlation ID em logs
- [ ] Tracer inicializa (mesmo sem coletor)
- [ ] Graceful shutdown (SIGTERM/SIGINT)
- [ ] Docker build funciona
- [ ] docker-compose sobe DB/Redis
