# PostgreSQL Expert — Operações de Produção

Referência para HA, escala, backup e incidentes. Genérico (genérico), condensado. Para padrões de código SQL do dia a dia, ver `core.md`.

## Índice

1. [Tipos de Índice e Matriz de Decisão](#1-tipos-de-índice-e-matriz-de-decisão)
2. [MVCC, VACUUM e Wraparound](#2-mvcc-vacuum-e-wraparound)
3. [Isolation Levels](#3-isolation-levels)
4. [Particionamento — Automação e Lifecycle](#4-particionamento--automação-e-lifecycle)
5. [Connection Pooling (PgBouncer)](#5-connection-pooling-pgbouncer)
6. [Replicação — Streaming e Lógica](#6-replicação--streaming-e-lógica)
7. [Backup e PITR](#7-backup-e-pitr)
8. [Zero-Downtime Migrations](#8-zero-downtime-migrations)
9. [Troubleshooting](#9-troubleshooting)

---

## 1. Tipos de Índice e Matriz de Decisão

| Padrão de query | Índice | Por quê |
|---|---|---|
| `WHERE id = x`, range, `ORDER BY`, prefix `LIKE 'x%'` | **B-tree** | Default. Igualdade, range, ordenação |
| `WHERE tags @> ARRAY[...]`, `data @> '{...}'`, full-text | **GIN** | Multi-valor (array/JSONB/tsvector). Query rápida, write mais lento |
| JSONB só com operador `@>` | **GIN `jsonb_path_ops`** | ~38% menor que GIN padrão, mais rápido para containment |
| Geometria, ranges, nearest-neighbor (`<->`) | **GiST** | PostGIS, `tstzrange`, exclusion constraints |
| Tabela enorme naturalmente ordenada (time-series append-only) | **BRIN** | Min/max por bloco. 100-1000x menor que B-tree; menos preciso |
| IP, quadtree, k-d tree (dados não-balanceados) | **SP-GiST** | Estruturas particionadas no espaço |

- **Composto**: ordem das colunas = igualdade → range/sort (regra ESR). Casar com o padrão real da query.
- **Parcial**: `... WHERE deleted_at IS NULL` — menor, write mais barato em linhas não-indexadas.
- **Expressão**: `((data->>'status'))`, `(lower(email))` — para função/cast inevitável na query.
- **Covering** (`INCLUDE`): index-only scan, dispensa heap fetch. Requer visibility map atualizado (VACUUM).
- **Exclusion constraint** (GiST + `btree_gist`): impede períodos sobrepostos — `EXCLUDE USING gist (room WITH =, period WITH &&)`.

```sql
-- Detectar índices nunca usados (candidatos a DROP)
SELECT schemaname, relname, indexrelname, idx_scan,
pg_size_pretty(pg_relation_size(indexrelid)) AS size
FROM pg_stat_user_indexes WHERE idx_scan = 0
ORDER BY pg_relation_size(indexrelid) DESC;

-- FK sem índice (causa seq scan em JOIN/cascade)
SELECT conrelid::regclass AS tbl, conname
FROM pg_constraint c WHERE contype='f' AND NOT EXISTS (
SELECT 1 FROM pg_index i WHERE i.indrelid=c.conrelid
AND (c.conkey[1] = ANY(i.indkey)));
```

---

## 2. MVCC, VACUUM e Wraparound

MVCC: cada `UPDATE`/`DELETE` cria nova versão da linha (`xmin`/`xmax`); leitores nunca bloqueiam escritores. Versões mortas (dead tuples) acumulam até o VACUUM recuperar. Sem VACUUM → table/index bloat → scans maiores → degradação.

- **Autovacuum sempre ligado**. Para tabelas hot, tunar por tabela (não global):
```sql
ALTER TABLE orders SET (
autovacuum_vacuum_scale_factor = 0.01, -- default 0.20
autovacuum_analyze_scale_factor = 0.005);
```
- **VACUUM regular** não bloqueia leitura/escrita. **`VACUUM FULL` bloqueia** (reescreve) — preferir `pg_repack` para rebuild online.
- **Long-running transaction bloqueia VACUUM** de coletar dead tuples globalmente — manter transações curtas; setar `idle_in_transaction_session_timeout`.
- **Transaction ID wraparound**: monitorar idade; autovacuum faz freeze. Risco real em bases write-heavy negligenciadas.
```sql
SELECT relname, n_dead_tup, n_live_tup,
round(n_dead_tup*100.0/NULLIF(n_live_tup+n_dead_tup,0),2) AS dead_pct,
last_autovacuum
FROM pg_stat_user_tables WHERE n_dead_tup > 10000 ORDER BY n_dead_tup DESC;

-- Idade do XID (alerta de wraparound)
SELECT datname, age(datfrozenxid) FROM pg_database ORDER BY 2 DESC;
```

---

## 3. Isolation Levels

| Nível | Vê | Uso |
|---|---|---|
| Read Committed (default) | Dados commitados no início de cada statement | Maioria das apps; melhor performance |
| Repeatable Read | Snapshot do início da transação | Relatórios/analytics que exigem consistência |
| Serializable | Serializável real (SSI) | Transações financeiras/críticas |

- Read Uncommitted é tratado como Read Committed no PostgreSQL.
- Repeatable Read e Serializable podem falhar com `could not serialize access` → a app deve ter **retry** (ver `core.md`, helper de transação).

---

## 4. Particionamento — Automação e Lifecycle

Para > 10M rows. Tipos: **range** (data — mais comum), **list** (tenant/região), **hash** (distribuição uniforme). A PK deve incluir a partition key.

```sql
CREATE TABLE events (
id BIGSERIAL, created_at TIMESTAMPTZ NOT NULL,...,
PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE events_2026_06 PARTITION OF events
FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE events_default PARTITION OF events DEFAULT;
```

- **Partition pruning**: a query DEVE filtrar pela partition key, senão varre todas as partições. Validar no `EXPLAIN`.
- **Criação automática**: função `CREATE TABLE... PARTITION OF` via `format()` + `pg_cron` agendando o próximo mês.
- **Lifecycle/retenção**: `ALTER TABLE... DETACH PARTITION` (rápido, não-bloqueante em PG14+) → `DROP TABLE` ou mover para schema `archive`. Muito mais barato que `DELETE` em massa.
- **Partition-wise joins/aggregation**: planner opera por partição → paralelismo.
- ⚠️ Particionar mal pode ser pior que tabela simples — medir antes (ADR-009).

---

## 5. Connection Pooling (PgBouncer)

Resolve connection exhaustion. **Nunca 1 conn por request** em produção.

```ini
[pgbouncer]
pool_mode = transaction
max_client_conn = 2000
default_pool_size = 25
reserve_pool_size = 10
auth_type = scram-sha-256
server_idle_timeout = 600
```

- **transaction mode** (recomendado, alta concorrência): conexão devolvida ao fim de cada transação. **Limita**: prepared statements server-side, `LISTEN/NOTIFY`, advisory locks, `SET` de sessão.
- **session mode**: conexão por cliente inteira. Use se precisar das features acima.
- Monitorar saturação: `SHOW POOLS;` — `cl_waiting > 0` ou `maxwait` alto = pool pequeno.
- App conecta na porta do PgBouncer (6432), não na do Postgres (5432).

---

## 6. Replicação — Streaming e Lógica

**Streaming (física)** — HA + read scaling. WAL binário para standby.
```conf
# primary
wal_level = replica
max_wal_senders = 10
max_replication_slots = 10
hot_standby = on
synchronous_standby_names = 'standby1' # sync = zero data loss, mais latência
```
```bash
pg_basebackup -h primary -D $PGDATA -U replicator -R -X stream -C -S standby1
# -R cria standby.signal + postgresql.auto.conf; -C cria o slot
```
- **sync** (zero perda, latência maior) vs **async** (rápido, perda possível) vs **quorum** (equilíbrio).
- Failover: `pg_promote()` ou `pg_ctl promote`. Switchback: `pg_rewind`. Automação: Patroni/repmgr.
- Monitorar lag:
```sql
-- primary
SELECT client_addr, state, sync_state, replay_lag,
pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn) AS lag_bytes
FROM pg_stat_replication;
-- standby
SELECT now() - pg_last_xact_replay_timestamp() AS lag, pg_is_in_recovery();
```

**Lógica** — replicação seletiva por tabela, cross-version, upgrades.
```sql
-- publisher
CREATE PUBLICATION pub FOR TABLE users, orders;
-- subscriber
CREATE SUBSCRIPTION sub CONNECTION 'host=... dbname=...' PUBLICATION pub;
SELECT * FROM pg_stat_subscription;
```
- Não replica DDL; tem overhead. Bom para multi-region selectivo e migração entre versões.

---

## 7. Backup e PITR

Regra **3-2-1**: 3 cópias, 2 mídias, 1 offsite. Testar restore regularmente (backup não-testado não é backup).

**Físico + WAL archiving** (base para PITR):
```conf
archive_mode = on
archive_command = 'test! -f /archive/%f && cp %p /archive/%f'
archive_timeout = 300
```
```bash
pg_basebackup -h localhost -U postgres -D /backup/base -F tar -z -X stream -P
```

**Lógico** (restore seletivo, cross-version):
```bash
pg_dump -Fc -f db.dump mydb # custom format
pg_restore -d mydb -j 4 db.dump # restore paralelo
pg_dump -Fc -t users -t orders... # tabelas específicas
```

**PITR** (recuperar a um ponto no tempo):
```conf
restore_command = 'cp /archive/%f %p'
recovery_target_time = '2026-06-29 14:30:00'
recovery_target_action = 'promote'
```
Cria `recovery.signal`, restaura a base, configura o target, inicia → recupera até o ponto. Schedule típico: WAL contínuo, base semanal, retenção mensal.

---

## 8. Zero-Downtime Migrations

- **Índice sem lock**: `CREATE INDEX CONCURRENTLY`. Se falhar, deixa índice `INVALID` — `DROP INDEX CONCURRENTLY` e refazer. Progresso em `pg_stat_progress_create_index`.
- **Adicionar NOT NULL sem table scan bloqueante**:
```sql
ALTER TABLE users ADD CONSTRAINT c_email_nn CHECK (email IS NOT NULL) NOT VALID; -- instantâneo
UPDATE...; -- backfill em lotes
ALTER TABLE users VALIDATE CONSTRAINT c_email_nn; -- ShareUpdateExclusiveLock (não bloqueia write)
ALTER TABLE users ALTER COLUMN email SET NOT NULL; -- agora sem scan
ALTER TABLE users DROP CONSTRAINT c_email_nn;
```
- **Backfill em lotes** (evita lock longo e bloat): loop com `LIMIT` + `pg_sleep` entre lotes.
- **Trocar tipo de coluna**: adicionar coluna nova → backfill → índices/FK na nova → swap em transação curta. Evita reescrita bloqueante.
- **Rebuild de tabela com bloat**: `pg_repack` (online, sem lock prolongado).

---

## 9. Troubleshooting

```sql
-- Cache hit ratio (alvo > 99% OLTP)
SELECT sum(heap_blks_hit)/NULLIF(sum(heap_blks_hit)+sum(heap_blks_read),0) AS ratio
FROM pg_statio_user_tables;

-- Queries lentas (pg_stat_statements)
SELECT substring(query,1,80) q, calls, round(total_exec_time) total_ms,
round(mean_exec_time) mean_ms, rows
FROM pg_stat_statements ORDER BY total_exec_time DESC LIMIT 20;

-- Bloqueios (quem bloqueia quem)
SELECT a.pid blocked, a.query blocked_q, b.pid blocking, b.query blocking_q
FROM pg_stat_activity a
JOIN pg_stat_activity b ON b.pid = ANY(pg_blocking_pids(a.pid));

-- Conexões por estado
SELECT state, count(*) FROM pg_stat_activity GROUP BY state;
```

| Sintoma | Causas prováveis |
|---|---|
| Query lenta | Falta índice (Seq Scan), estatísticas velhas (`ANALYZE`), função em coluna indexada, FK sem índice |
| CPU alto | Query cara (pg_stat_statements), seq scans, join/agg ineficiente |
| Connection exhaustion | Falta PgBouncer, leak na app, `max_connections` baixo |
| Autovacuum não acompanha | `autovacuum_max_workers` baixo, naptime alto, long-running transaction bloqueando |
| Replication lag | Banda, hardware do standby, query longa no standby, taxa de WAL |
| Wraparound | Idade do XID alta, autovacuum agressivo insuficiente, `VACUUM FREEZE` necessário |

> No `EXPLAIN ANALYZE`: `Seq Scan` em tabela grande → falta índice. `Nested Loop` com estimativas grandes → estatísticas desatualizadas (`ANALYZE`). `Hash Join` com `Batches > 1` → aumentar `work_mem`. `loops=N` (N>1) → nó executado N vezes (N+1 potencial).
