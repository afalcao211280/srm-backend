# ADR-001 — Go 1.26 e Gin

**Decisão:** Go 1.26.5, router Gin v1.10, DTOs e middlewares com Gin nativo.

**Motivo:** tipagem forte, binário estático, concorrência de primeira classe, e o stack canônico do `golang-expert` da organização.

**Consequência:** uso de `gin.HandlerFunc` e `gin.Context` em toda a camada HTTP. Middlewares via `r.Use(...)`. Não usamos `chi`.
