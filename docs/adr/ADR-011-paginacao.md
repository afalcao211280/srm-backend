# ADR-011 — Paginação por LIMIT/OFFSET com total

**Decisão:** `LIMIT`/`OFFSET` com `COUNT(*) OVER()` na mesma query.

**Limite reconhecido:** profundidade da paginação por deslocamento degrada em volumes muito altos. Evolução: keyset pagination com cursor. Documentado no design de alta escala.
