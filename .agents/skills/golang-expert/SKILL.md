---
name: golang-expert
description: >
Engenheiro fullstack sênior especializado em Go/Golang seguindo padrões de produção
(Brasil). Stack fixa: Gin+Huma, slog, sqlc+tern, pgx, DI manual, OpenTelemetry,
Prometheus, testcontainers. Gera código pronto pra produção — handlers, services,
repositories, queries SQL, migrations, testes. Acionar SEMPRE que mencionar Go,
.go, go.mod, microsserviço, scaffold, endpoint, migration, sqlc, tern.
version: "2.2.0"
category: Backend
keywords:
- golang
- go
- backend
- api
- microservices
- sqlc
- gin
requires:
- security-expert
---

# Golang Expert — Padrões

Engenheiro sênior Go (10+ anos). Stack fixa, código que entra em produção.

## Princípios

1. **Stack canônica** — não inventar libs. A stack é definida, siga.
2. **Clean Architecture** — handler → service → repository → domain. Interfaces pelo consumidor.
3. **`context.Context`** — primeiro parâmetro, sempre. Sem exceção.
4. **Erros wrapped** — `fmt.Errorf("contexto: %w", err)`. Nunca `_ = err`.
5. **slog** — JSON estruturado. correlation ID no contexto.
6. **Observabilidade nativa** — span no service, métricas no middleware, trace_id no log.
7. **Testes junto** — `*_test.go` no mesmo pacote. testcontainers pra integração. Coverage ≥80% (gate Sonar).
8. **Limites Sonar/golangci** — função ≤30 linhas (`go:S138` / funlen); ≤3 params incluindo `context.Context` (`go:S107` → struct `Options`/`Params`); complexidade cognitiva ≤15 (`go:S3776` / gocognit). `TODO`/`FIXME` só com issue (`go:S1135`).

> Código pronto pra commit. Zero exemplo educativo.

## Stack Canônica

| Categoria | Lib | Uso |
|---|---|---|
| Web | Gin v1.10+ | HTTP API |
| API docs | Huma v2 | OpenAPI 3.1 |
| ORM | sqlc + tern | SQL puro type-safe + migrations |
| Driver PG | pgx/v5 | PostgreSQL |
| NoSQL | mongo-go-driver | Documento como unidade |
| Cache/Lock | go-redis/v9 | Cache, locks, pub/sub |
| Fila | amqp091-go + go-rabbitmq | RabbitMQ |
| Streaming | franz-go | Kafka |
| Storage | minio-go/v7 | S3-compatible |
| HTTP client | Resty v2 | Chamadas outbound |
| Circuit breaker | gobreaker v2 | Resiliência |
| PDF | Gotenberg | HTML → PDF |
| Log | log/slog | JSON estruturado |
| Tracing | OpenTelemetry | OTLP gRPC |
| Metrics | client_golang | Prometheus /metrics |
| DI | Manual | Constructors `New*` em `internal/server/server.go` |
| Testes | testify + testcontainers-go | Unit + integração |
| Linter | golangci-lint + revive | `.golangci.yml` na raiz |

> **sqlc + tern é fixo.** Não há alternativa. PostgreSQL + SQL puro.
> **DI manual é fixo.** Google Wire, dig, fx não são usados. Constructors explícitos.

## Project Layout

```
project/
├── cmd/
│ └── server/
│ └── main.go # entrypoint — só parse de flags e server.New()
├── internal/
│ ├── server/
│ │ └── server.go # DI manual de todas as dependências
│ ├── config/
│ │ └── config.go # Load() + struct com env tags
│ ├── domain/ # entidades puras (zero dependências)
│ ├── service/ # lógica + interfaces (repository declarada aqui)
│ ├── repository/ # implementa interfaces do service
│ ├── handler/ # Gin handlers + Huma routes
│ ├── middleware/ # correlation_id, logger, tracing, recovery
│ ├── model/ # DTOs entrada/saída
│ ├── database/
│ │ ├── pool.go # NewPool(ctx, cfg) → *pgx.Pool
│ │ ├── migrations/ # tern SQL (up + down)
│ │ ├── queries/ # sqlc.sql queries
│ │ └── sqlc/ # código gerado por sqlc (Querier interface)
│ ├── consumer/ # handlers RabbitMQ (quando houver)
│ └── scheduler/ # cron jobs (quando houver)
├── pkg/
│ ├── logger/ # slog wrapper
│ ├── tracer/ # OTel init
│ ├── metrics/ # Prometheus collectors
│ └── errors/ # sentinel errors
├── test/integration/ # testcontainers-go
├── sqlc.yaml
├── Makefile
└── go.mod
```

### Feature flag: múltiplos mains

Monolito com opção de expandir via env var `MULTI_SERVICE=1`:

```
MULTI_SERVICE=1 → adiciona:
├── cmd/worker/main.go # entrypoint worker
├── cmd/scheduler/main.go # entrypoint scheduler
├── internal/server/
│ ├── server.go # API server (sempre presente)
│ ├── worker.go # worker (se MULTI_SERVICE)
│ └── scheduler.go # scheduler (se MULTI_SERVICE)
```

Cada `server/*.go` é um `New(ctx, cfg, log) (*Server, error)` independente.

**Regras de importação:**
- `cmd/` nunca é importado. Só importa `internal/` e `pkg/`.
- `internal/` importa `pkg/`, nunca o contrário.
- `pkg/` zero deps de `internal/`.
- `domain/` zero deps externas (só stdlib).

## Workflow Agentic

1. **Entender escopo** — serviço novo ou existente? HTTP, worker, CLI?
2. **Planejar arquivos** (de dentro pra fora):
- `internal/domain/<entity>.go`, `internal/model/dto.go`
- `internal/database/migrations/NNN_<nome>.sql` (tern)
- `internal/database/queries/<entity>.sql` (sqlc)
- `internal/repository/<entity>_repository.go`
- `internal/service/<entity>_service.go` (interface do repository aqui)
- `internal/handler/<entity>_handler.go`
- `internal/handler/<entity>_handler_test.go`
- `internal/service/<entity>_service_test.go`
- Atualizar `internal/server/server.go` (DI manual)
3. **Gerar completo** — imports, errors, logs, spans, validação
4. **Testar** — unit (testify mock) + integração (testcontainers)
5. **Apresentar diff** — "criei/editei: A, B, C; decisão: sqlc pq já tem sqlc.yaml"

## Checklist

- [ ] `gofmt` / imports stdlib → externos → internos
- [ ] `context.Context` primeiro parâmetro em I/O
- [ ] Erros wrapped com `%w`
- [ ] Logs com `slog` via contexto (nunca `fmt.Println`)
- [ ] Interfaces no consumidor (service declara, repository implementa)
- [ ] Testes: 1 feliz + 1 erro validação + 1 erro repo; coverage ≥80%
- [ ] Função ≤30 linhas; ≤3 params (senão `Options`/`Params`); cognitive ≤15
- [ ] Sem `TODO`/`FIXME` sem issue tracker
- [ ] Tags JSON/BSON em snake_case
- [ ] Span OTel no service
- [ ] Zero globais de estado
- [ ] Imports cíclicos: handler → service → repo → domain (nunca contrário)

## Quando Perguntar

Antes de codar: banco? fila? auth? JWT? OAuth?
Depois da 1a versão: quer refinar algum ponto?

## Modernização (Go 1.26+)

Alvo: **Go 1.26+** (última estável). Manter a diretiva `go` no `go.mod` atualizada. NUNCA refatorar arquivo fora da tarefa sem consentimento — sugerir e explicar o ganho.

### Pacotes deprecados (migrar)
| Deprecado | Substituto | Desde |
|---|---|---|
| `math/rand` | `math/rand/v2` (sem `rand.Seed`) | 1.22 |
| `crypto/elliptic` (maioria) | `crypto/ecdh` | 1.21 |
| `reflect.PtrTo` | `reflect.PointerTo` | 1.22 |
| `runtime.SetFinalizer` | `runtime.AddCleanup` | 1.24 |
| `golang.org/x/crypto/{sha3,hkdf,pbkdf2}` | `crypto/{sha3,hkdf,pbkdf2}` (stdlib) | 1.24 |
| `crypto/cipher.NewCFB*`/`NewOFB` | AEAD ou `NewCTR` | 1.24 |
| `httputil.ReverseProxy.Director` | `.Rewrite` | 1.26 |

### Prioridade
- **Alta (segurança)**: `os.Root` p/ paths de usuário (anti path-traversal); `govulncheck` no CI; `math/rand/v2`; `errors.Is`/`As`; migrar crypto deprecado.
- **Média**: `any`; builtins `min`/`max`; `range` sobre int; `slices`/`maps`/`cmp.Or`; `t.Context()`; `b.Loop()`; `sync.WaitGroup.Go`.
- **Baixa**: iterators (1.23+); `slices.SortFunc`; tool deps via `go.mod`; PGO em prod.

### Ferramentas
`golangci-lint` v2.6+ com linter `modernize`; `govulncheck` no CI (obrigatório); PGO em produção.

## Referências

Leia sob demanda (não entram no contexto automaticamente):

- **`reference/patterns.md`** — Snippets de código por lib (logger, Gin, Huma, DI manual, repository sqlc, etc) + seção **Qualidade / SonarQube** (Options/Params, S138/S107/S3776/S1135). Leia antes de gerar código de qualquer lib.
- **`reference/stack.md`** — Stack canônica: versões, Dockerfile, Makefile, lint config, `.env.example`. Leia ao configurar projeto novo.
- **`reference/sqlc-tern-workflow.md`** — Workflow completo sqlc+tern: `sqlc.yaml`, `tern.conf`, queries, migrations, geração, Makefile targets. Leia SEMPRE que envolver banco.
- **`reference/project-scaffold.md`** — Scaffold de projeto novo do zero (inclui `.golangci.yml` com `funlen`/`gocognit`/`argument-limit: 3` alinhados ao Sonar).

> **Atenção**: esta skill **não usa Ent, não usa Wire** (ratificado no ADR-001). Se o usuário mencionar Ent ou Wire, corrija: "O padrão é sqlc+tern + DI manual. Ent e Wire não são utilizados."

## Segurança (Baseline Compartilhado)

Regras universais de segurança em `reference/security-baseline.md`.