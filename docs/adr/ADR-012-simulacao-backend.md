# ADR-012 — Simulação em tempo real chama o backend

**Decisão:** o painel chama `POST /api/v1/simulacoes` com debounce. Fórmula **não** é reimplementada em TypeScript.

**Motivo:** JavaScript não tem decimal nativo. Uma única implementação da fórmula. A simulação é idempotente e sem efeito colateral.
