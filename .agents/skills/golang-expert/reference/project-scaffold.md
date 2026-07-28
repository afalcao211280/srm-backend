# Project Scaffold — Novo Serviço Go

Template completo para criar um serviço Go novo do zero. Use quando o usuário pedir "scaffold", "novo microsserviço", "começar projeto Go" ou similar.

## Decisões Iniciais (Pergunte Primeiro)

Antes de gerar arquivos, confirme:

1. **Nome do serviço**: kebab-case (ex: `user-service`, `notification-worker`)
2. **Tipo**:
- HTTP API (Gin + Huma)
- Worker / consumer (RabbitMQ ou Kafka)
- CLI
- Job/CronJob
3. **Banco**:
- PostgreSQL com sqlc + tern
- MongoDB
- Nenhum
4. **Dependências externas** (Redis, Kafka, RabbitMQ, MinIO, APIs externas)
5. **Versão de Go**: padrão atual = `1.23` (ajuste conforme `golang.md` evoluir)
6. **Módulo Go**: `github.com/example-<service-name>` ou variação corporativa
7. **DI**: Manual (construtores explicitos em `internal/server/server.go`)

## Estrutura de Pastas a Criar

```
<service-name>/
├── cmd/
│ └── <service-name>/
│ ├── main.go
│ └── main.go
├── internal/
│ ├── config/
│ │ └── config.go
│ ├── handler/
│ │ ├── health.go
│ │ └── http.go # registro de rotas Gin+Huma
│ ├── middleware/
│ │ ├── correlation_id.go
│ │ ├── logger.go
│ │ └── tracing.go
│ ├── service/
│ │ └──.gitkeep
│ ├── repository/
│ │ └──.gitkeep
│ ├── domain/
│ │ └──.gitkeep
│ └── model/
│ └──.gitkeep
├── pkg/
│ ├── logger/
│ │ └── logger.go
│ ├── errors/
│ │ └── errors.go
│ ├── tracer/
│ │ └── tracer.go
│ └── metrics/
│ └── metrics.go
├── api/
│ └── openapi.yaml # gerado pelo Huma
├── migrations/ # tern
├── test/
│ └── integration/
│ └──.gitkeep
├── scripts/
│ ├── build.sh
│ └── generate-mocks.sh
├── docker/
│ ├── Dockerfile
│ └── docker-compose.dev.yml
├── k8s/
│ └──.gitkeep
├──.env.example
├──.gitignore
├──.golangci.yml
├──.mockery.yaml
├──.revive.toml
├── Makefile
├── README.md
├── go.mod
└── sqlc.yaml # se sqlc
```

## `go.mod`

```go
module github.com/example-<service-name>

go 1.23

require (
github.com/gin-gonic/gin v1.10.0
github.com/danielgtaylor/huma/v2 v2.20.0
github.com/go-resty/resty/v2 v2.14.0
github.com/sony/gobreaker/v2 v2.5.1
github.com/redis/go-redis/v9 v9.4.0
github.com/google/uuid v1.6.0
github.com/stretchr/testify v1.9.0
github.com/testcontainers/testcontainers-go v0.32.0
go.opentelemetry.io/otel v1.27.0
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.27.0
go.opentelemetry.io/otel/sdk v1.27.0
github.com/prometheus/client_golang v1.19.0
// se sqlc + tern:
github.com/jackc/pgx/v5 v5.6.0

)
```

## `cmd/<service-name>/main.go`

```go
package main

import (
"context"
"errors"
"log/slog"
"net/http"
"os"
"os/signal"
"syscall"
"time"

"github.com/example-<service-name>/internal/config"
"github.com/example-<service-name>/pkg/logger"
"github.com/example-<service-name>/pkg/tracer"
)

func main() {
ctx, stop:= signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

// Config
cfg, err:= config.Load()
if err!= nil {
slog.Error("config inválida", "error", err)
os.Exit(1)
}

// Logger
log:= logger.New(slog.Level(cfg.LogLevel))
slog.SetDefault(log)
log.Info("iniciando serviço", "service", cfg.ServiceName, "version", cfg.Version)

// Tracer
shutdownTracer, err:= tracer.Init(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
if err!= nil {
log.Error("falha ao iniciar tracer", "error", err)
// continua — tracer não é bloqueante
} else {
defer func() {
shutdownCtx, cancel:= context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
_ = shutdownTracer(shutdownCtx)
}()
}

// DI Manual — montagem das dependências
app, err:= server.New(ctx, cfg, log)
if err!= nil {
log.Error("falha ao iniciar app", "error", err)
os.Exit(1)
}
defer app.Close()

// HTTP server
srv:= &http.Server{
Addr: ":" + cfg.Port,
Handler: app.Router,
ReadHeaderTimeout: 10 * time.Second,
ReadTimeout: 30 * time.Second,
WriteTimeout: 30 * time.Second,
IdleTimeout: 120 * time.Second,
}

go func() {
log.Info("HTTP server escutando", "port", cfg.Port)
if err:= srv.ListenAndServe(); err!= nil &&!errors.Is(err, http.ErrServerClosed) {
log.Error("HTTP server falhou", "error", err)
stop()
}
}()

<-ctx.Done()
log.Info("sinal recebido, iniciando shutdown")

shutdownCtx, cancel:= context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err:= srv.Shutdown(shutdownCtx); err!= nil {
log.Error("erro no shutdown", "error", err)
}
log.Info("serviço encerrado")
}
```

## `internal/server/server.go` (DI Manual)

```go
package server

import (
"context"
"log/slog"
"net/http"

"github.com/jackc/pgx/v5/pgxpool"

"github.com/example-<service-name>/internal/config"
"github.com/example-<service-name>/internal/handler"
"github.com/example-<service-name>/internal/repository"
"github.com/example-<service-name>/internal/service"
"github.com/example-<service-name>/internal/sqlc"
)

type App struct {
Router http.Handler
closes []func() error
}

func (a *App) Close() {
for _, c:= range a.closes {
_ = c()
}
}

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) (*App, error) {
// Pool de conexões PostgreSQL
pool, err:= pgxpool.New(ctx, cfg.DB.DSN())
if err!= nil {
return nil, fmt.Errorf("criar pool: %w", err)
}

// sqlc queries
q:= sqlc.New(pool)

// Repositories
userRepo:= repository.NewUserRepository(q)

// Services
userSvc:= service.NewUserService(userRepo)

// Handlers + Router
userHandler:= handler.NewUserHandler(userSvc)
router:= handler.NewRouter(userHandler, log)

return &App{
Router: router,
closes: []func() error{pool.Close},
}, nil
}
```

## `internal/config/config.go`

```go
package config

import (
"fmt"
"os"
"strconv"
"time"
)

type Config struct {
ServiceName string
Version string
Port string
LogLevel int // -4 (Debug), 0 (Info), 4 (Warn), 8 (Error)
OTLPEndpoint string

DB DBConfig
Redis RedisConfig
}

type DBConfig struct {
Host string
Port int
User string
Password string
Name string
SSLMode string

MaxConns int32
MinConns int32
MaxConnLifetime time.Duration
}

type RedisConfig struct {
Addr string
Password string
DB int
}

func Load() (*Config, error) {
cfg:= &Config{
ServiceName: getEnv("SERVICE_NAME", "service"),
Version: getEnv("VERSION", "dev"),
Port: getEnv("PORT", "8080"),
LogLevel: getEnvInt("LOG_LEVEL", 0),
OTLPEndpoint: getEnv("OTLP_ENDPOINT", "localhost:4317"),

DB: DBConfig{
Host: mustEnv("DB_HOST"),
Port: getEnvInt("DB_PORT", 5432),
User: mustEnv("DB_USER"),
Password: mustEnv("DB_PASSWORD"),
Name: mustEnv("DB_NAME"),
SSLMode: getEnv("DB_SSLMODE", "require"),
MaxConns: int32(getEnvInt("DB_MAX_CONNS", 25)),
MinConns: int32(getEnvInt("DB_MIN_CONNS", 5)),
MaxConnLifetime: time.Duration(getEnvInt("DB_MAX_LIFETIME_MIN", 60)) * time.Minute,
},

Redis: RedisConfig{
Addr: getEnv("REDIS_ADDR", "localhost:6379"),
Password: os.Getenv("REDIS_PASSWORD"),
DB: getEnvInt("REDIS_DB", 0),
},
}
return cfg, nil
}

func getEnv(k, def string) string {
if v:= os.Getenv(k); v!= "" { return v }
return def
}

func getEnvInt(k string, def int) int {
if v:= os.Getenv(k); v!= "" {
if n, err:= strconv.Atoi(v); err == nil { return n }
}
return def
}

func mustEnv(k string) string {
v:= os.Getenv(k)
if v == "" { panic(fmt.Sprintf("env %s obrigatória", k)) }
return v
}
```

## `.env.example`

```bash
# Service
SERVICE_NAME=user-service
VERSION=dev
PORT=8080
LOG_LEVEL=0

# Observability
OTLP_ENDPOINT=otel-collector:4317

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=changeme
DB_NAME=userservice
DB_SSLMODE=disable
DB_MAX_CONNS=25
DB_MIN_CONNS=5

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

## `.golangci.yml`

```yaml
run:
timeout: 5m
modules-download-mode: readonly
go: "1.23"

linters:
disable-all: true
enable:
- errcheck
- errorlint
- funlen # espelha Sonar go:S138 (≤30 linhas)
- gocognit # espelha Sonar go:S3776 (cognitive ≤15)
- gocritic
- gofmt
- goimports
- gosec
- govet
- ineffassign
- misspell
- revive
- staticcheck
- stylecheck
- typecheck
- unconvert
- unused
- unparam

linters-settings:
funlen:
lines: 30
statements: -1 # só linhas (S138); statements desligado
gocognit:
min-complexity: 16 # flaga quando >15 (S3776)

revive:
rules:
- name: var-naming
- name: package-comments
- name: exported
disabled: false
- name: error-return
- name: error-strings
- name: error-naming
- name: increment-decrement
- name: range
- name: receiver-naming
- name: time-naming
- name: unexported-return
- name: indent-error-flow
- name: errorf
- name: empty-block
- name: superfluous-else
- name: unused-parameter
- name: argument-limit
arguments: [3] # alinha universal ≤3 params e Sonar go:S107 (antes: 6)

errorlint:
errorf: true
asserts: true
comparison: true

gocritic:
enabled-tags:
- diagnostic
- performance
- style
disabled-checks:
- dupImport
- ifElseChain
- octalLiteral
- whyNoLint

gosec:
excludes:
- G104 # tratado por errcheck

issues:
max-issues-per-linter: 0
max-same-issues: 0
exclude-rules:
- path: _test\.go
linters:
- gosec
- errcheck
- unparam
- path: cmd/
linters:
- gochecknoglobals
```

## `.mockery.yaml`

```yaml
quiet: false
with-expecter: true
inpackage: false
packages:
github.com/example-<service-name>/internal/service:
interfaces:
UserRepository:
OrderRepository:
config:
dir: "internal/service/mocks"
filename: "{{.InterfaceName | snakecase}}.go"
mockname: "Mock{{.InterfaceName}}"
outpkg: "mocks"
```

Gera com `go run github.com/vektra/mockery/v2`.

## `Makefile`

```makefile
SERVICE:= <service-name>
GO:= go
GOOS?= linux
GOARCH?= amd64

.PHONY: help
help: ## Lista targets disponíveis
	@grep -E '^[a-zA-Z_-]+:.*?##.*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf " \033[36m%-20s\033[0m %s\n", $$1, $$2}'

##@ Desenvolvimento

.PHONY: run
run: ## Roda o serviço localmente
	$(GO) run./cmd/$(SERVICE)

.PHONY: build
build: ## Build binário
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build \
		-ldflags="-s -w -X main.version=$$(git describe --tags --always)" \
		-o bin/$(SERVICE)./cmd/$(SERVICE)

.PHONY: clean
clean: ## Limpa artefatos
	rm -rf bin/ coverage.* dist/

##@ Qualidade

.PHONY: fmt
fmt: ## Formata o código
	$(GO) fmt./...
	$(GO) run golang.org/x/tools/cmd/goimports -w -local github.com/example-.

.PHONY: lint
lint: ## Roda linters
	$(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint run./...

.PHONY: lint-fix
lint-fix: ## Roda linters com auto-fix
	$(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix./...

.PHONY: vet
vet: ## go vet
	$(GO) vet./...

##@ Testes

.PHONY: test
test: ## Testes unitários
	$(GO) test -race -count=1 -short./...

.PHONY: test-integration
test-integration: ## Testes de integração (testcontainers-go)
	$(GO) test -race -count=1 -tags=integration -timeout=10m./test/integration/...

.PHONY: test-coverage
test-coverage: ## Coverage report HTML
	$(GO) test -race -coverprofile=coverage.out -covermode=atomic./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Abra coverage.html no navegador"

##@ Code generation

.PHONY: gen
gen: ## Gera mocks e sqlc
	$(GO) generate./...

.PHONY: mocks
mocks: ## Gera mocks (mockery)
	$(GO) run github.com/vektra/mockery/v2

##@ Database (sqlc + tern)

.PHONY: sqlc-gen
sqlc-gen: ## Gera código sqlc
	$(GO) run github.com/sqlc-dev/sqlc/cmd/sqlc generate

.PHONY: migrate-up
migrate-up: ## Aplica migrations
	$(GO) run github.com/jackc/tern/v2 migrate -m internal/database/migrations

.PHONY: migrate-down
migrate-down: ## Rollback de 1 migration
	$(GO) run github.com/jackc/tern/v2 migrate -m internal/database/migrations -d -1

.PHONY: migrate-new
migrate-new: ## Cria migration: make migrate-new NAME=descricao
	@if [ -z "$(NAME)" ]; then echo "Uso: make migrate-new NAME=descricao"; exit 1; fi
	$(GO) run github.com/jackc/tern/v2 new -m internal/database/migrations $(NAME)

##@ Docker

.PHONY: docker-build
docker-build: ## Build imagem Docker
	docker build -f docker/Dockerfile -t example/$(SERVICE):latest.

.PHONY: docker-run
docker-run: ## Roda container local
	docker run --rm -p 8080:8080 --env-file.env example/$(SERVICE):latest

.PHONY: docker-compose-dev
docker-compose-dev: ## Sobe ambiente de dev (db, redis, etc)
	docker compose -f docker/docker-compose.dev.yml up -d

.PHONY: docker-compose-down
docker-compose-down: ## Derruba ambiente de dev
	docker compose -f docker/docker-compose.dev.yml down

##@ Outros

.PHONY: deps
deps: ## Baixa dependências
	$(GO) mod download
	$(GO) mod tidy

.PHONY: tools
tools: ## Instala ferramentas de dev
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/vektra/mockery/v2@latest

	$(GO) install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	$(GO) install github.com/jackc/tern/v2@latest
```

## `docker/Dockerfile` (multi-stage, distroless)

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Cache de dependências
COPY go.mod go.sum./
RUN go mod download

# Build
COPY..
ARG SERVICE=<service-name>
ARG VERSION=dev

RUN CGO_ENABLED=0 GOOS=linux go build \
-ldflags="-s -w -X main.version=${VERSION}" \
-trimpath \
-o /out/app./cmd/${SERVICE}

# Runtime stage — distroless é menor e mais seguro
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/app /app

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/app"]
```

## `docker/docker-compose.dev.yml`

```yaml
services:
postgres:
image: postgres:16-alpine
environment:
POSTGRES_DB: ${DB_NAME:-userservice}
POSTGRES_USER: ${DB_USER:-postgres}
POSTGRES_PASSWORD: ${DB_PASSWORD:-changeme}
ports:
- "5432:5432"
volumes:
- postgres_data:/var/lib/postgresql/data
healthcheck:
test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-postgres}"]
interval: 5s
timeout: 3s
retries: 5

redis:
image: redis:7-alpine
ports:
- "6379:6379"
command: redis-server --save "" --appendonly no
healthcheck:
test: ["CMD", "redis-cli", "ping"]
interval: 5s
timeout: 3s
retries: 5

jaeger:
image: jaegertracing/all-in-one:latest
ports:
- "16686:16686" # UI
- "4317:4317" # OTLP gRPC

prometheus:
image: prom/prometheus:latest
volumes:
-./prometheus.yml:/etc/prometheus/prometheus.yml:ro
ports:
- "9090:9090"

volumes:
postgres_data:
```

## `.gitignore`

```gitignore
# Binários
bin/
dist/
*.exe

# Coverage
coverage.*
*.out

# IDE
.idea/
.vscode/
*.swp
*.swo

# Env
.env
.env.local

# OS
.DS_Store
Thumbs.db

# Gerados (se forem versionados, remova)
# Não ignore código sqlc — deve ser commitado!

# Build cache
.cache/
```

## `README.md` (esqueleto)

```markdown
# <service-name>

Breve descrição do serviço.

## Stack

- Go 1.23
- Gin + Huma v2
- PostgreSQL (sqlc + tern) | MongoDB
- Redis
- OpenTelemetry + Prometheus

## Desenvolvimento

### Pré-requisitos

- Go 1.23+
- Docker + Docker Compose
- Make

### Setup

\`\`\`bash
cp.env.example.env
make tools # instala ferramentas de dev
make docker-compose-dev # sobe DB, Redis, Jaeger
make migrate-up # aplica migrations
make run # roda o serviço
\`\`\`

### Comandos úteis

\`\`\`bash
make help # lista todos os targets
make test # testes unitários
make test-integration # testes com testcontainers
make lint # golangci-lint
make sqlc-gen # regenera código SQL
make docker-build # build imagem
\`\`\`

## Estrutura

Segue [golang-standards/project-layout](https://github.com/golang-standards/project-layout) e os padrões documentados em `golang.md`.

\`\`\`
cmd/<service>/ - entrypoint
internal/ - código privado (handler/service/repository/domain)
pkg/ - código público reutilizável
api/ - OpenAPI gerado
migrations/ - schema versionado
\`\`\`

## Observabilidade

- Logs: JSON estruturado via slog, com correlation_id em todos os requests
- Traces: OpenTelemetry → OTLP → Jaeger (local) / coletor central (prod)
- Métricas: `/metrics` Prometheus

## Convenções

- Commits semânticos em pt-BR (feat/fix/chore/docs/refactor/test)
- PRs com descrição clara, screenshots quando aplicável
- Cobertura mínima: 70% em service/domain, 60% em handler
```

## Workflow de Scaffold (para Claude executar)

Quando o usuário pedir "scaffold um novo serviço Go chamado X":

1. **Pergunte as decisões iniciais** (tipo, banco, DI, deps externas) — não chute.
2. **Crie a estrutura de pastas** via `mkdir -p`.
3. **Gere os arquivos na ordem**:
- `go.mod` primeiro
- `.gitignore`, `.env.example`
- `Makefile`, `.golangci.yml`, `.mockery.yaml`
- `pkg/` (logger, errors, tracer, metrics) — esses são quase invariantes
- `internal/config/config.go`
- `internal/middleware/` (correlation_id, logger, tracing)
- `internal/handler/health.go`
- `cmd/<service>/main.go`
- `internal/server/server.go`
- `docker/Dockerfile`, `docker/docker-compose.dev.yml`
- `README.md`
4. **Comente o que ficou para o usuário fazer**:
- Ajustar nome do módulo no `go.mod`
- Rodar `go mod tidy`
- `make tools` para instalar ferramentas
- Adicionar primeira entity de domínio
5. **Não gere domínio fictício**. Não invente `User` ou `Order`. Espere o usuário pedir.

## Checklist de Scaffold

- [ ] `go.mod` com nome de módulo correto
- [ ] `Makefile` com targets básicos funcionando
- [ ] `.golangci.yml` com lint passando em arquivo vazio (`make lint`)
- [ ] `main.go` compila (`make build`)
- [ ] `health.go` retorna 200 OK
- [ ] `/metrics` expõe métricas básicas
- [ ] Correlation ID propagado em logs
- [ ] Tracer inicializa sem erro (mesmo sem coletor disponível)
- [ ] Graceful shutdown via SIGTERM/SIGINT
- [ ] Docker build funciona
- [ ] `docker-compose.dev.yml` sobe DB/Redis/Jaeger
- [ ] README com instruções de setup
