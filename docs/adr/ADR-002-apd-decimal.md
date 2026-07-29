# ADR-002 — Aritmética decimal com apd/v3

**Decisão:** `cockroachdb/apd/v3` v3.2.3 com `Context` de 34 dígitos, `RoundHalfUp` e `Traps` ativos.

**Alternativas descartadas:**

- `shopspring/decimal` — `Pow` retorna `0` silenciosamente em casos de borda (documentação lista `0**y` com `y<0` e base negativa com expoente não-inteiro). Em motor financeiro, zero silencioso é deságio de 100%.
- `big.Rat` — sem exponenciação fracionária.
- `big.Float` — ponto flutuante binário.

**Motivo:** General Decimal Arithmetic (IEEE 754-2008), mesma semântica do `NUMERIC` do PostgreSQL. `Pow` retorna `Condition` sinalizando `Inexact`/`Rounded`. `Pow(0, -1)` devolve `Infinity` com `err == nil` — exigimos guard explícito (`mustFinite`).
