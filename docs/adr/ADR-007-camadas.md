# ADR-007 — 3 camadas, exceção verificada para relatórios

**Decisão:** handler → service → repository → domain. Exceção: `internal/report` tem apenas handler → query (2 camadas), e `report` **não importa** `internal/domain`. A regra é verificada por `depguard` no `golangci-lint` — uma violação quebra o CI.

**Motivo:** §3.6 do case exige a exceção explicitamente. A fronteira que não é verificada apodrece.
