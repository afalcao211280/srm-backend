# Mapa de entrega — Backend

| Exigência do enunciado | Atendido em |
|---|---|
| §3.1 Gestão de Câmbio | `internal/infra/postgres/repos.go` (CotacaoRepo, TaxaBaseRepo) + `internal/server/operacoes.go` (rotas Huma) |
| §3.2 Motor de Precificação | `internal/domain/precificacao/motor.go` + `internal/domain/recebivel/strategy.go` |
| §3.3 Persistência e Integridade | `internal/infra/postgres/transacao.go` (optimistic locking) + `internal/platform/migracao` (schema aplicado no boot) |
| §3.4 API REST | `internal/server/` (Huma v2 sobre `*gin.Engine`) — OpenAPI real em `/openapi.json`, docs em `/docs` |
| §3.5 Consultas Analíticas | `internal/report/extrato.go` (2 camadas, squirrel, paginação — testado com EXPLAIN confirmando uso de índice) |
| §3.6 Exceção de 2 camadas para relatórios | `internal/report/` + ADR-007 + lint `depguard` (habilitado de verdade no schema v2) |
| §5.1 Tratamento de Exceções | `internal/app/middleware/middleware.go` + `internal/app/middleware/huma.go` (mesmo formato de erro em rotas Gin e Huma) |
| §6 🟡 Pleno — Docker | `docker-compose.yml` + `Dockerfile` — `docker compose up -d db api` verificado de ponta a ponta, healthcheck real |
| §6 🟡 Pleno — Testes Unitários | `internal/domain/*_test.go` + integração real via testcontainers em `internal/infra/postgres`, `internal/report`, `internal/app/handler`, `internal/infra/cambio` |
| §6 🔴 Sênior — Optimistic Locking | `internal/infra/postgres/transacao.go`, testado com 10 liquidações concorrentes reais |
| §6 🔴 Sênior — Observabilidade | `internal/platform/metricas/metricas.go` + middleware — `/metrics` verificado ao vivo |
| §6 🔴 Sênior — Resiliência | `internal/infra/cambio/cliente.go` (gobreaker) — as 4 transições de estado testadas de verdade |
| §6 🔴 Sênior — CI/CD | `.github/workflows/ci.yml` + `lefthook.yml` |
| §6 🟣 Staff — C4 | `docs/diagramas/` |
| §6 🟣 Staff — ADRs | `docs/adr/` |
| §6 🟣 Staff — Alta Escala | `docs/alta-escala.md` |
| §6 🟣 Staff — EDA | `docs/eda.md` |
| §6 🟣 Staff — Branching | ADR-014 + este README |
| §6 🟣 Staff — Gestão de Crise | grupo 13 do plano (simulação) |
| §7 Modelagem de Dados | `migrations/` + `docs/ddl.sql` + `docs/diagramas/er.mmd` |
| §9 Entrega | tags `v1.0.0`, `AI_USAGE.md`, runbook, mapa de entrega |

## Fora de escopo

- **IOF, tarifas, tributos** — não constam no enunciado.
- **IaC (Terraform/K8s)** — marcado como "(Opcional)" no case.
- **Broker de eventos** — proposta de EDA é documental.
- **Autenticação** — não consta no enunciado.
- **Provedor real de câmbio** — §3.1 pede integração mockada.

## Ambiguidades resolvidas

- **Taxa Base**: modelada como tabela versionada (ADR-004).
- **Prazo**: dias corridos ÷ 30, fracionário, capitalização composta (ADR-003).
