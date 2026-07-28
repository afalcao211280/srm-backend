# tern — Migrations SQL

## Filosofia

- Migrations escritas a mao, versionadas
- `up` + `down` no mesmo arquivo
- Down deve **reverter completamente** o up

## Estrutura

```
internal/database/migrations/
001_initial_schema.sql
002_add_users_table.sql
tern.conf
```

## tern.conf

```toml
[database]
host = {{env "DB_HOST"}}
port = {{env "DB_PORT"}}
database = {{env "DB_NAME"}}
user = {{env "DB_USER"}}
password = {{env "DB_PASSWORD"}}
sslmode = {{env "DB_SSLMODE"}}

[data]
prefix = app_
```

## Anatomia de Migration

```sql
-- Migration: criar tabela de usuarios
-- Autor: nome
-- Data: 2026-05-12

CREATE TABLE users (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
email TEXT NOT NULL UNIQUE,
name TEXT NOT NULL,
status TEXT NOT NULL DEFAULT 'active'
CHECK (status IN ('active', 'inactive', 'pending')),
metadata JSONB NOT NULL DEFAULT '{}',
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

CREATE TRIGGER users_set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

---- create above / drop below ----

DROP TRIGGER IF EXISTS users_set_updated_at ON users;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

## Convencoes

- Nome: `NNN_descricao_em_snake_case.sql`
- Separador: `---- create above / drop below ----`
- Down com `IF EXISTS` para idempotencia
- Comentarios no topo: o que faz, por que, quem, quando

## Makefile Targets

```makefile
migrate-up:
	tern migrate -m internal/database/migrations

migrate-down:
	tern migrate -m internal/database/migrations -d -1

migrate-status:
	tern status -m internal/database/migrations

migrate-new:
	tern new -m internal/database/migrations $(NAME)
```

## Workflow Diario

1. Nova tabela: `make migrate-new NAME=add_orders_table`
2. Edita migration
3. `make migrate-up`
4. Se afeta queries: `make sqlc-gen`

## Anti-padroes

- ❌ Down vazio ou genérico
- ❌ Múltiplas migrations editando mesmo objeto sem cuidado
- ❌ Dados de seed em migrations versionadas (usar scripts/seed.sql)
- ❌ Sem testar down: `migrate-up` + `migrate-down` + `migrate-up`
