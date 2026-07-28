# SRM Credit Engine — Backend

API REST em Go 1.26.5 que precifica recebíveis com deságio e registra a liquidação de forma auditável. Stack fixa: Gin + slog + sqlc/tern + pgx + gobreaker, DI manual, OpenAPI via Huma.

## Pré-requisitos

- Go 1.26.5
- Docker + Docker Compose
- lefthook (opcional, para hooks de pré-commit)
- golangci-lint v2.12+ (opcional, lint local)

## Execução local

```bash
cp .env.example .env
docker compose up -d db
go run ./cmd/api
```

A API sobe em `http://localhost:8080`. Endpoints:

- `GET  /healthz` — liveness, não consulta o banco
- `GET  /readyz` — readiness, faz `PING` no banco
- `GET  /metrics` — métricas Prometheus
- `POST /api/v1/simulacoes` — simula sem persistir
- `POST /api/v1/transacoes` — cria transação
- `GET  /api/v1/transacoes` — lista paginada
- `GET  /api/v1/transacoes/{id}` — detalha
- `POST /api/v1/transacoes/{id}/liquidacao` — liquida
- `POST /api/v1/cotacoes` — registra cotação
- `POST /api/v1/taxas-base` — registra taxa base
- `GET  /api/v1/taxas-base/vigente` — consulta taxa vigente
- `GET  /api/v1/tipos-recebivel` — lista tipos
- `GET  /api/v1/relatorios/extrato-liquidacao` — extrato filtrado

## Execução com Docker

```bash
docker compose up
docker compose --profile carga run --rm carga
```

A massa de demonstração popula 8 cedentes, taxas base para BRL e USD, cotações USD→BRL e 50 transações calculadas pelo próprio motor.

## Comandos de teste

```bash
go test ./...                        # testes unitários
go test -tags=integration ./...     # testes de integração (precisa DATABASE_URL)
go test ./... -race -coverprofile=coverage.out -covermode=atomic
```

## Lint

```bash
golangci-lint run
```

## Hooks de pré-commit

```bash
lefthook install
```

## Variáveis de ambiente

Veja `.env.example`. As obrigatórias são `DATABASE_URL`. As demais têm padrão seguro.

## Documentação adicional

- `docs/adr/` — Architecture Decision Records
- `docs/diagramas/` — C4, ER
- `docs/ddl.sql` — DDL consolidado
- `AI_USAGE.md` — registro do uso de IA

## Runbook

Veja `docs/runbook.md`. Cobre subir/derrubar a stack, acompanhar logs, reiniciar descartando dados, problemas comuns.
