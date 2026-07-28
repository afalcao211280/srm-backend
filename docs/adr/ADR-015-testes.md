# ADR-015 — Pirâmide de testes em três níveis

**Decisão:**

| Nível | Ferramenta | Alvo |
|---|---|---|
| Unitário | testify | strategies, fórmula, arredondamento, guards |
| Integração | testcontainers-go | repositórios contra PostgreSQL real |
| E2E | Playwright | fluxo do painel e do grid contra a stack |

**Motivo:** mockar o banco em integração apaga o que se quer testar (`NUMERIC`, isolamento transacional, `UPDATE` condicional do optimistic locking).
