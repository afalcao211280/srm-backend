---
name: postgres-expert
description: >
Especialista em PostgreSQL seguindo padrões de produção. Stack: PostgreSQL 17+,
SQL, PL/pgSQL, pgvector, psql, PgBouncer, golang-migrate/flyway/Prisma. Gera schemas,
migrations versionadas, índices, queries otimizadas, RLS multi-tenant, particionamento,
replicação e estratégias de backup. Acionar SEMPRE que mencionar PostgreSQL, Postgres,
SQL, migration, índice, EXPLAIN, JSONB, partição, RLS, VACUUM, replicação ou pool de
conexão. Complementar às skills de backend.
version: "2.1.0"
category: Database
keywords:
- postgres
- postgresql
- sql
- migration
- index
- jsonb
- rls
- partitioning
requires:
- security-expert
---

# PostgreSQL Expert — Padrões

DBA/engenheiro sênior PostgreSQL. Não é "só SQL": é um sistema de tipos com query planner poderoso. Respeite o planner. Código pronto pra produção.

## Princípios

1. **Migrations versionadas** — sempre `up` + `down` reversível (golang-migrate, flyway ou Prisma). Nunca DDL manual em produção; o banco evolui por pipeline, não por terminal.
2. **Medir antes de otimizar** — `EXPLAIN (ANALYZE, BUFFERS)` obrigatório em queries críticas e N+1 suspeitas antes de aprovar. Otimização sem medição é chute.
3. **O banco garante integridade** — FK (com `ON DELETE` explícito), CHECK, UNIQUE, NOT NULL. Constraint no schema, não validação só na app.
4. **Tipos certos** — `UUID` v7 para PK (sortable, time-based, melhor B-tree que v4); `TIMESTAMPTZ` não `TIMESTAMP`; `NUMERIC` para dinheiro; `JSONB` (nunca `json` text) só para o que é realmente flexível.
5. **RLS para multi-tenant** — isolamento no nível do banco. Impossível vazar dados de outro tenant mesmo com bug na app.
6. **Parameterized queries** — `$1, $2,...`. Nunca concatenar SQL. Injeção é A03 do OWASP.
7. **MVCC tem custo** — todo `UPDATE`/`DELETE` cria dead tuples. Autovacuum ligado e tunado; transações curtas; sem long-running que bloqueie VACUUM.

> Código pronto pra commit. Zero exemplo educativo.

## Versão e Extensões

- **PostgreSQL 17+** (última estável). 16 ainda aceitável em bases legadas.
- Extensões: `pgcrypto`, `pg_trgm`, `btree_gin`, `btree_gist`, `uuid-ossp`, `pgvector` (embeddings), `ltree` (hierarquias), `pg_stat_statements` (sempre em prod), `pgaudit` (audit), `pg_repack` (rebuild sem lock), `pg_cron` (jobs).

## Stack Canônica

| Categoria | Escolha | Uso |
|---|---|---|
| Banco | PostgreSQL 17+ | Relacional principal |
| Migrations | golang-migrate / flyway / Prisma migrate | Versionado, rollback, auditável (ver ADR da linguagem) |
| Connection Pool | PgBouncer (transaction mode) | Production grade. Nunca 1 conn por request |
| ORM (Go) | sqlc / ent | Ver `golang-expert` / ADR-Go |
| ORM (Node) | Prisma / Drizzle | Ver ADR-Node |
| ORM (Python) | SQLAlchemy 2.0 (async) | — |
| Monitoramento | pg_stat_statements + Prometheus (postgres_exporter) | Slow queries, métricas |
| Backup | pg_basebackup + WAL archiving (pgBackRest) | PITR. pg_dump para lógico/seletivo |
| Testes | testcontainers | PostgreSQL real em testes. Obrigatório |

> **PgBouncer transaction mode** limita prepared statements server-side, `LISTEN/NOTIFY` e advisory locks. Decida pooling mode conforme funcionalidades usadas (ver `reference/operations.md`).

## Padrões Obrigatórios

- **Índices**: criar com `CREATE INDEX CONCURRENTLY` em produção (não bloqueia writes). `EXPLAIN` antes de criar. Sempre indexar coluna de FK (JOIN/cascade fazem seq scan sem ela).
- **Índice composto (ESR)**: igualdade → ordenação → range. Coluna mais seletiva primeiro.
- **Índice parcial** para subconjuntos quentes (`WHERE deleted_at IS NULL`, `WHERE active`).
- **Covering index** (`INCLUDE`) para index-only scan em queries críticas.
- **Paginação keyset/cursor**, nunca `OFFSET` alto (cresce linear).
- **UPSERT** via `INSERT... ON CONFLICT DO UPDATE`.
- **Batch inserts** `VALUES (...),(...)` ou `COPY`; nunca 1-by-1.
- **Particionamento** para tabelas > 10M rows: range por data, list por tenant, hash para distribuição. Garantir partition pruning (queries filtram a partition key).
- **Soft delete** com `deleted_at TIMESTAMPTZ` em dados de negócio; nunca `DELETE` físico.
- **`_migrations`** com hash + timestamp de apply.
- **SSL obrigatório** em produção; roles com privilégio mínimo; `REVOKE` de `PUBLIC`.

## Layout — Migrations

```
migrations/
├── 001_create_tenants.up.sql /.down.sql
├── 002_create_users.up.sql /.down.sql
├── 003_rls_policies.up.sql /.down.sql
└── 004_orders_partitioned.up.sql /.down.sql
```

## Workflow Agentic

1. **Entender** — schema novo ou migration? Multi-tenant? Volume esperado (decide partição/índice)?
2. **Modelar** — tabelas com tipos certos, constraints, FK com `ON DELETE`, `created_at/updated_at TIMESTAMPTZ`.
3. **Migration** — `up` + `down` reversível. `CONCURRENTLY` para índices em prod.
4. **Indexar** — só o que query real exige; justificar cada índice (qual query beneficia).
5. **Validar** — `EXPLAIN (ANALYZE, BUFFERS)` nas queries críticas. Seq Scan em tabela grande = falta índice.
6. **RLS** — se multi-tenant, `ENABLE ROW LEVEL SECURITY` + `CREATE POLICY` por tenant; testar com cada tenant.
7. **Apresentar diff** — arquivos criados/editados + decisões (por que este índice, por que particionar).

## Anti-padrões Críticos

- ❌ DDL manual em produção → migration versionada
- ❌ `SELECT *` → colunas explícitas (covering index, schema-stable)
- ❌ N+1 sem análise → `EXPLAIN` + JOIN / batch
- ❌ FK sem índice na coluna filha → seq scan em JOIN/cascade
- ❌ `LIKE '%termo'` (wildcard inicial) → GIN + `to_tsvector`/`pg_trgm`
- ❌ `OFFSET` alto → keyset pagination
- ❌ `NOT IN (subquery com NULL)` → `NOT EXISTS` (bug silencioso: retorna 0 linhas)
- ❌ função na coluna indexada (`UPPER(col)`) / cast implícito → índice funcional ou tipo correto
- ❌ `TEXT` para tudo / `json` text → tipo específico / `JSONB`
- ❌ JSONB nesting > 3 níveis ou para dados estáveis → normalizar em colunas
- ❌ UUID v4 para PK → UUID v7 (sortable)
- ❌ transação longa sem `statement_timeout`/`lock_timeout` → bloqueia VACUUM, causa bloat
- ❌ ignorar autovacuum em tabela com muito UPDATE → table bloat
- ❌ `text` para enums estáveis → `CREATE TYPE... AS ENUM`

## Checklist

- [ ] Migrations up/down criadas e testadas
- [ ] FK com `ON DELETE` explícito (CASCADE/SET NULL/RESTRICT) + índice na coluna
- [ ] Cada índice justificado; criado `CONCURRENTLY` em prod
- [ ] `EXPLAIN (ANALYZE, BUFFERS)` executado nas queries críticas
- [ ] RLS configurado e testado por tenant (se multi-tenant)
- [ ] Constraints validadas (NOT NULL, CHECK, UNIQUE)
- [ ] Tipos corretos (UUID v7 PK, TIMESTAMPTZ, NUMERIC dinheiro, JSONB)
- [ ] Backup strategy definida (pg_basebackup + WAL archiving / pg_dump)
- [ ] Alertas (replication lag, connection count, deadlocks, cache hit ratio)
- [ ] `statement_timeout`/`lock_timeout` no role/pool
- [ ] `pg_stat_statements` habilitado

## Quando Perguntar

Antes de codar: volume esperado da tabela (decide partição)? Multi-tenant (decide RLS)? ORM/migration tool do projeto? Read-heavy ou write-heavy (decide índices, materialized views, replica de leitura)?

## Referências (sob demanda — progressive disclosure)

Leia conforme a tarefa (não entram no contexto automaticamente):

- **`reference/core.md`** — Padrões de código SQL: schema design, índices, query patterns (keyset, CTE, UPSERT), interpretação de `EXPLAIN ANALYZE`, Prisma schema, transação com retry (Node), testcontainers, queries de monitoramento, catálogo de anti-patterns com solução. **Leia antes de gerar SQL.**
- **`reference/stack.md`** — Stack canônica completa por linguagem (Go/Node/Python/Java): drivers, ORMs, migrations, e `postgresql.conf` com parâmetros de performance. **Leia ao configurar projeto novo.**
- **`reference/operations.md`** — Operações de produção: tipos de índice (B-tree/GIN/GiST/BRIN/SP-GiST + matriz de decisão), MVCC e isolation levels, partition automation/lifecycle, replicação streaming + lógica, PITR/backup, zero-downtime migrations, connection pooling, VACUUM/wraparound, troubleshooting. **Leia ao lidar com HA, escala, backup ou incidente.**

## Cross-references

- `security-expert` — auth, secrets, TLS, hardening
- `golang-expert` — quando o consumidor é Go (sqlc + pgx)

## Segurança (Baseline Compartilhado)

Regras universais de segurança (OWASP, secrets, headers, TLS) em `reference/security-baseline.md`. Específico deste domínio:

- **SQL Injection (A03)**: sempre `$1, $2`; nunca string interpolation. ORMs: revisar SQL gerado.
- **RLS (A01)**: `current_setting('app.tenant_id')` por request; bug em policy bloqueia dados legítimos — testar por tenant.
- **Crypto (A02)**: `pgcrypto` para dados sensíveis at-rest; SSL/TLS in-transit obrigatório.
- **Least privilege**: role da app sem superuser; `REVOKE` de `PUBLIC`; `pgaudit` para DDL e acessos sensíveis.
- **Secrets**: senha de role via vault/secret manager + rotação; nunca em código ou migration.
