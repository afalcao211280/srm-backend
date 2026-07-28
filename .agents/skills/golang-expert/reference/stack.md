# Stack — Golang Expert 2026

## Versão canônica

**Go 1.26.3** — versão de produção.

## Dependências principais

| Categoria | Biblioteca | Versão | Uso |
|-----------|-----------|--------|-----|
| **Web** | `github.com/gin-gonic/gin` | `v1.10+` | Router HTTP, middlewares |
| **OpenAPI** | `github.com/danielgtaylor/huma/v2` | `v2.x` | Geração automática de OpenAPI 3.1 |
| **SQL query** | `github.com/sqlc-dev/sqlc` | `v1.27+` | SQL type-safe code-gen |
| **Migrations** | `github.com/jackc/tern` | `v2.x` | Migrations SQL (par com sqlc) |
| **PostgreSQL** | `github.com/jackc/pgx/v5` | `v5.x` | Driver Postgres nativo |
| **MongoDB** | `go.mongodb.org/mongo-driver` | `v2.x` | Driver oficial Mongo |
| **Redis** | `github.com/redis/go-redis/v9` | `v9.x` | Cache + pub/sub + locks |
| **RabbitMQ** | `github.com/rabbitmq/amqp091-go` | `v1.x` | Consumer/publisher AMQP |
| **Kafka** | `github.com/twmb/franz-go/pkg/kgo` | `v1.x` | Consumer/producer Kafka |
| **HTTP client** | `github.com/go-resty/resty/v2` | `v2.x` | REST client com retry |
| **Circuit breaker** | `github.com/sony/gobreaker/v2` | `v2.x` | Resiliência para chamadas externas |
| **Object storage** | `github.com/minio/minio-go/v7` | `v7.x` | MinIO / S3-compatible |
| **PDF** | `github.com/starwalkn/gotenberg-go-client/v8` | `v8.x` | Geração de PDF via Gotenberg |
| **Tracing** | `go.opentelemetry.io/otel` | `v1.x` | OpenTelemetry SDK |
| **Metrics** | `github.com/prometheus/client_golang` | `v1.x` | Métricas Prometheus |
| **Logging** | `log/slog` | stdlib | Logger JSON estruturado (Go 1.21+) |
| **Testes** | `github.com/stretchr/testify` | `v1.x` | Assert / require / mock |
| **Testcontainers** | `github.com/testcontainers/testcontainers-go` | `v0.x` | Integração real com DB em testes |
| **DI** | — | — | Manual (construtores explicitos) |
| **UUID** | `github.com/google/uuid` | `v1.x` | Geração de UUIDs v4/v7 |
| **Linting** | `golangci-lint` + `revive` | latest | Qualidade de código em CI |
| **Load test** | `k6` | `v0.51+` | Teste de performance (externo) |

## Toolchain

```bash
# Versões fixas em.tool-versions ou Dockerfile
go 1.26.3
golangci-lint 1.64.x
```

## Dockerfile canônico (produção)

```dockerfile
FROM golang:1.26-alpine AS builder
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

## golangci-lint config mínimo

> Config completo (timeout, excludes de `_test.go`, etc.) em `project-scaffold.md`. Abaixo: núcleo + linters que espelham Sonar.

```yaml
#.golangci.yml
linters:
enable:
- revive
- govet
- errcheck
- staticcheck
- gosec
- misspell
- goimports
- funlen # go:S138 — funções ≤30 linhas
- gocognit # go:S3776 — cognitive complexity ≤15

linters-settings:
funlen:
lines: 30
statements: -1
gocognit:
min-complexity: 16
revive:
rules:
- name: exported
- name: var-naming
- name: unused-parameter
- name: argument-limit
arguments: [3] # go:S107 — ≤3 params; acima → Options/Params struct
```

Coverage para Sonar: `go test./... -coverprofile=coverage.out -covermode=atomic` e `sonar.go.coverage.reportPaths=coverage.out`.

## ADRs relacionados

- `docs/adrs/Backend/ADR-001-golang.md` — decisão de adoção de Go
- `docs/adrs/Database-Relational/ADR-009-postgresql.md` — stack Postgres
