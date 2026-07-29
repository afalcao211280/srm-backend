# ADR-010 — pgx direto, query builder para relatórios

**Decisão:** `jackc/pgx/v5` para acesso, `Masterminds/squirrel` para queries filtráveis do extrato, `golang-migrate` para migrations SQL.

**Motivo:** §3.5 pede query builder ou SQL nativo. `squirrel` monta `WHERE` dinâmico sempre com placeholders.
