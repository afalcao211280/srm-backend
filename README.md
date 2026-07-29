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

Detalhes em `docs/guia-de-testes.md`.

## Acesso à API

A documentação OpenAPI 3.1 é gerada em runtime pelo Huma v2 — nenhum arquivo estático versionado. Interface Swagger em `http://localhost:8080/docs`, JSON bruto em `http://localhost:8080/openapi.json`.

Exemplo executável de simulação com `curl` (cenário "duplicata 45 dias" da spec de precificação). A simulação não persiste nada e não exige `cedente_id` existente, mas depende de uma taxa base vigente para a moeda na data da operação — cadastre uma antes:

```bash
curl -X POST http://localhost:8080/api/v1/taxas-base \
  -H "Content-Type: application/json" \
  -d '{"moeda":"BRL","taxa_mensal":"0.01","vigencia_inicio":"2026-01-01T00:00:00Z"}'

curl -X POST http://localhost:8080/api/v1/simulacoes \
  -H "Content-Type: application/json" \
  -d '{
    "cedente_id": 1,
    "tipo_recebivel": "DUPLICATA_MERCANTIL",
    "valor_face": "10000.00",
    "moeda_titulo": "BRL",
    "moeda_pagamento": "BRL",
    "data_operacao": "2026-07-01",
    "data_vencimento": "2026-08-15"
  }'
```

Resposta esperada (valores conferem com o cenário "duplicata 45 dias" da spec de precificação):

```json
{
  "valor_presente": "9636.38630878",
  "valor_liquido": "9636.39",
  "desagio": "363.61",
  "moeda_titulo": "BRL",
  "moeda_pagamento": "BRL",
  "data_operacao": "2026-07-01",
  "data_vencimento": "2026-08-15"
}
```

Se a massa de demonstração já foi carregada (`docker compose --profile carga run --rm carga`), a taxa base acima já existe e o segundo `curl` funciona sozinho.

## Observabilidade

- `GET /healthz` — liveness, não consulta o banco.
- `GET /readyz` — readiness, faz `PING` no banco. Retorna `503` se indisponível.
- `GET /metrics` — métricas Prometheus. Inclui:
  - `srm_http_requests_total{rota, metodo, status}` — contagem por rota, método e status.
  - `srm_http_request_duration_seconds` — histograma de latência.
  - `srm_precificacoes_total` — total de precificações.
  - `srm_liquidacoes_sucesso_total` / `srm_liquidacoes_conflito_total`.
  - `srm_circuito_cambio_estado` — estado do circuit breaker.

Logs estruturados em JSON com identificador de correlação. Para investigar uma requisição específica, filtre os logs pelo `correlation_id` retornado no cabeçalho `X-Correlation-ID` da resposta.

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
- `docs/diagramas/` — C4 (contexto e container), ER
- `docs/ddl.sql` — DDL consolidado
- `docs/alta-escala.md` — design para 1M transações/minuto
- `docs/eda.md` — proposta de arquitetura orientada a eventos
- `docs/criterios-de-aceite.md` — usabilidade, segurança, desempenho, escalabilidade
- `docs/guia-de-testes.md` — como rodar cada nível de teste
- `docs/runbook.md` — ciclo operacional da stack
- `docs/mapa-de-entrega.md` — cobertura das exigências do enunciado
- `docs/incidente.md` — simulação de gestão de crise
- `AI_USAGE.md` — registro do uso de IA
