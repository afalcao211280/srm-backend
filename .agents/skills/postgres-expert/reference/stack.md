# PostgreSQL Expert — Stack Canônica

## PostgreSQL 17+ — Versão Canônica

| Categoria | Lib/Ferramenta | Versão | Link | Notas |
|-----------|---------------|--------|------|-------|
| **Core** | | | | |
| Banco | PostgreSQL | 17.x | https://postgresql.org | 17 em produção nova. 16 ainda aceitável em legado |
| **ORM / Query Builder** | | | | |
| ORM (Go) | ent | 0.13+ | https://entgo.io | Code-gen. Padrão Go |
| ORM (Go) alt | sqlc | 2.x | https://sqlc.dev | SQL-first, zero reflection |
| ORM (Node/TS) | Prisma | 5.x | https://www.prisma.io | Schema-first. Padrão Node |
| ORM (Node/TS) alt | Drizzle | 0.30+ | https://orm.drizzle.team | SQL-like, type-safe, sem migrations gerenciadas |
| ORM (Node/TS) alt | TypeORM | 0.3.x | https://typeorm.io | Decorator-based. NestJS padrão |
| ORM (Python) | SQLAlchemy | 2.0+ | https://sqlalchemy.org | Async + Core + ORM. Padrão Python |
| ORM (Java) | Spring Data JPA | 3.3+ | https://spring.io/projects/spring-data | Hibernate under the hood |
| **Driver / Pool** | | | | |
| Driver (Node) | pg (node-postgres) | 8.x | https://node-postgres.com | Subjacente ao Prisma/Drizzle |
| Pool (Node) | pgpool / Prisma pooling | — | — | Prisma Accelerate para serverless |
| Driver (Python) | asyncpg | 0.29+ | https://magicstack.github.io/asyncpg | Async nativo. Mais rápido que psycopg2 |
| Driver (Python) alt | psycopg3 | 3.x | https://www.psycopg.org/psycopg3/ | Sync + Async. Mais maduro |
| Driver (Java) | postgresql JDBC | 42.x | https://jdbc.postgresql.org | |
| **Migrations** | | | | |
| Migrations (Go) | goose | 3.x | https://github.com/pressly/goose | SQL ou Go. Padrão Go |
| Migrations (Node) | Prisma Migrate | built-in | — | `prisma migrate dev` |
| Migrations (Node) alt | Drizzle Kit | latest | https://orm.drizzle.team/kit-docs | |
| Migrations (Python) | Alembic | 1.13+ | https://alembic.sqlalchemy.org | |
| Migrations (Java) | Flyway | 10.x | https://flywaydb.org | SQL versionado. Padrão Java |
| **Testes** | | | | |
| Containers | testcontainers | multi-lang | https://testcontainers.com | PostgreSQL real em testes. Obrigatório |
| Fixtures | Factory Boy (Python) / FactoryBot (Ruby) | — | — | Geração de dados de teste |
| **Monitoramento** | | | | |
| Query stats | pg_stat_statements | built-in extension | — | Habilitar sempre em produção |
| Slow query log | `log_min_duration_statement` | config | — | `500ms` em produção |
| Connection pool | PgBouncer | 1.22+ | https://pgbouncer.org | Transaction mode para alta concorrência |
| **Backup** | | | | |
| Backup lógico | pg_dump / pg_restore | built-in | — | `--format=custom` para restore seletivo |
| Backup físico | pgBackRest | 2.50+ | https://pgbackrest.org | WAL archiving + incremental |

## postgresql.conf — Parâmetros de Performance Essenciais

```ini
# Memória
shared_buffers = 256MB # 25% da RAM. Em prod: 1-4GB
effective_cache_size = 1GB # Estimativa de cache do OS. 50-75% RAM
work_mem = 16MB # Por operação de sort/hash. Cuidado com parallelism
maintenance_work_mem = 256MB # VACUUM, CREATE INDEX

# WAL / Durabilidade
wal_level = replica # Replicação lógica e física
synchronous_commit = on # Não alterar sem análise de risco
checkpoint_completion_target = 0.9

# Query planner
random_page_cost = 1.1 # Para SSD. Padrão é 4.0 (HDD)
effective_io_concurrency = 200 # Para SSD

# Logging
log_min_duration_statement = 500 # Log queries > 500ms
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
log_checkpoints = on
log_lock_waits = on
```
