# ADR-009 — Optimistic locking por versão

**Decisão:** `transacoes.versao INTEGER NOT NULL DEFAULT 1`. Liquidação via `UPDATE` condicional em identificador, versão e status `PENDENTE`, incrementando a versão. Zero linhas afetadas → `409`.

**Motivo:** §6 🔴 cita optimistic locking. `SELECT FOR UPDATE` serializa sob carga; `SERIALIZABLE` espalha erros de serialização. A condição tripla torna a checagem atômica no próprio `UPDATE`.
