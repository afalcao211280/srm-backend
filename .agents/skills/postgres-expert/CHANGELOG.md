# Changelog — postgres-expert

## [2.1.0] - 2026-06-30

### Changed
- Consolidação em `skills-optimizadas` (cruzamento skills/ ⨯ skills-compare/postgresql-database-engineering ⨯ ADR-009).
- Versão canônica alinhada à última estável: **PostgreSQL 17+** (era 16+ no SKILL.md, 15/16 no stack.md).
- SKILL.md reescrito no formato Framework 2.0 do precedente (golang-expert): Princípios, Workflow Agentic, Quando Perguntar, Cross-references, seção de Segurança específica do domínio.
- Path do security baseline corrigido para `reference/security-baseline.md`.
- `description` enriquecida e `keywords` expandidas (1 → 8 reais) para trigger matching.

### Added
- `reference/operations.md` — operações de produção condensadas do compare (alto valor genérico): tipos de índice + matriz de decisão (B-tree/GIN/GiST/BRIN/SP-GiST), MVCC/VACUUM/wraparound, isolation levels, partition automation/lifecycle, PgBouncer, replicação streaming + lógica, backup/PITR, zero-downtime migrations, troubleshooting.
- Alinhamentos do ADR-009 ao SKILL.md: índices `CONCURRENTLY` em prod, FK com `ON DELETE` explícito + índice na coluna, tabela `_migrations`, caveat de prepared statements em PgBouncer transaction mode, extensões unidas (skill ∪ ADR: pgvector, pgcrypto, ltree, btree_gin, btree_gist, uuid-ossp, pg_trgm, pg_stat_statements, pgaudit, pg_repack, pg_cron).

### Removed
- `reference/patterns.md` — duplicata byte-a-byte de `reference/core.md` (sem valor adicional).

### Decisões de divergência ADR-009
- Nenhuma contradição dura. ADR-009 mais rico em pontos operacionais (CONCURRENTLY, FK ON DELETE, prepared statements vs pooling) → skill alinhada ao ADR.

## [2.0.0] - 2026-05-28
### Changed
- SKILL.md padronizado no formato Framework 2.0 (caveman style)
- Adicionado versionamento semantico
- Adicionado keywords para trigger matching
- Seguranca referenciada via _shared/security/baseline.md
- Referencias reorganizadas

## [1.0.0] - 2024-01-10
### Added
- Versao inicial
