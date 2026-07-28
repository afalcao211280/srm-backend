# sqlc + tern — Workflow Completo

SQL puro type-safe com migrations versionadas. Use quando o projeto valoriza SQL versionado, controle fino de queries e migrations explícitas. Funciona apenas com **PostgreSQL** (tern é PG-only; sqlc também suporta MySQL/SQLite, mas tern não).

## Filosofia

- **sqlc**: você escreve SQL puro; sqlc gera código Go tipado a partir das queries.
- **tern**: migrations escritas à mão, versionadas, com `up/down` no mesmo arquivo.
- Banco é a **fonte de verdade do schema**. Código Go é gerado, não escrito à mão.

## Instalação

```bash
# sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# tern
go install github.com/jackc/tern/v2@latest

# Verificar
sqlc version
tern version
```

Em projetos, instale via `tools.go` para versionar:

```go
//go:build tools
// +build tools

package tools

import (
_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
_ "github.com/jackc/tern/v2"
)
```

E adicione `go.mod`:
```bash
go get -tool github.com/sqlc-dev/sqlc/cmd/sqlc github.com/jackc/tern/v2
```

## Estrutura de Pastas (padrão)

```
project/
├── internal/
│ └── database/
│ ├── migrations/
│ │ ├── 001_initial_schema.sql
│ │ ├── 002_add_users_table.sql
│ │ ├── 003_add_orders_table.sql
│ │ └── tern.conf
│ ├── queries/
│ │ ├── users.sql
│ │ └── orders.sql
│ └── sqlc/ # output do sqlc (gerado, não editar)
│ ├── db.go
│ ├── models.go
│ ├── users.sql.go
│ └── orders.sql.go
├── sqlc.yaml # raiz do projeto
└── Makefile
```

**Por que `internal/database/`?** Porque essas pastas são privadas do módulo. Migrations não devem ser importáveis de fora. O código gerado (`sqlc/`) também fica isolado.

## `sqlc.yaml` — Template

```yaml
version: "2"
sql:
- engine: "postgresql"
queries: "internal/database/queries"
schema: "internal/database/migrations"
gen:
go:
package: "sqlc"
out: "internal/database/sqlc"
sql_package: "pgx/v5" # pgx é o driver canônico
emit_json_tags: true
emit_interface: true # gera interface Querier (essencial para mocks)
emit_empty_slices: true # []T{} em vez de nil
emit_exported_queries: false
emit_prepared_queries: false
emit_pointers_for_null_types: true
omit_unused_structs: true
query_parameter_limit: 3 # acima disso, sqlc gera struct Params
overrides:
- db_type: "uuid"
go_type:
import: "github.com/google/uuid"
type: "UUID"
- db_type: "timestamptz"
go_type: "time.Time"
- db_type: "jsonb"
go_type:
import: "encoding/json"
type: "RawMessage"
```

**Pontos críticos:**
- `sql_package: "pgx/v5"` — usar pgx, não `database/sql`. pgx tem melhor performance e tipos nativos do Postgres.
- `emit_interface: true` — gera `Querier` interface; use ela no repository (mockável).
- `overrides` — mapeia tipos Postgres para tipos Go. UUID e JSONB são os mais comuns.
- `schema` aponta para migrations: sqlc lê o estado consolidado das migrations para gerar tipos.

## `tern.conf` — Template

`internal/database/migrations/tern.conf`:

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

Variáveis lidas do ambiente em runtime — nunca committar credenciais. Em dev, use `.env` carregado via `direnv` ou `godotenv`.

## Anatomia de uma Migration tern

`migrations/002_add_users_table.sql`:

```sql
-- Migration: criar tabela de usuários
-- Autor: rodrigo
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
CREATE INDEX idx_users_status_created_at ON users(status, created_at);

-- Trigger para updated_at automático
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

---- create above / drop below ----

DROP TRIGGER IF EXISTS users_set_updated_at ON users;
DROP FUNCTION IF EXISTS set_updated_at();
DROP INDEX IF EXISTS idx_users_status_created_at;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

**Convenções obrigatórias:**
- Nome do arquivo: `NNN_descricao_em_snake_case.sql` (NNN = número sequencial com zero-pad).
- Separador `---- create above / drop below ----` é **sintaxe do tern**. Acima é `up`, abaixo é `down`.
- O `down` deve **reverter completamente** o `up`. Teste rodando `tern migrate -d 1` e voltando.
- `IF EXISTS` no down para idempotência.
- Comentários no topo: o que faz, por que, quem, quando.

## Anatomia de uma Query sqlc

`queries/users.sql`:

```sql
-- name: CreateUser:one
INSERT INTO users (email, name, status, metadata)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByID:one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail:one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: ListUsers:many
SELECT * FROM users
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateUser:one
UPDATE users
SET name = $2, metadata = $3
WHERE id = $1
RETURNING *;

-- name: UpdateUserStatus:exec
UPDATE users
SET status = $2
WHERE id = $1;

-- name: DeleteUser:exec
DELETE FROM users WHERE id = $1;

-- name: CountUsersByStatus:one
SELECT COUNT(*) FROM users WHERE status = $1;
```

**Anotações sqlc:**
- `:one` — retorna uma struct (erro se não encontrar; usa `LIMIT 1` se quiser tratar como opcional)
- `:many` — retorna `[]struct`
- `:exec` — sem retorno (UPDATE/DELETE sem RETURNING)
- `:execrows` — retorna `int64` com rows affected
- `:batchone` / `:batchmany` / `:batchexec` — para `pgx` batch mode (alta performance)

## Código Gerado pelo sqlc

Depois de `sqlc generate`, `internal/database/sqlc/users.sql.go`:

```go
// Code generated by sqlc. DO NOT EDIT.
package sqlc

import (
"context"
"encoding/json"
"time"

"github.com/google/uuid"
)

type User struct {
ID uuid.UUID `json:"id"`
Email string `json:"email"`
Name string `json:"name"`
Status string `json:"status"`
Metadata json.RawMessage `json:"metadata"`
CreatedAt time.Time `json:"created_at"`
UpdatedAt time.Time `json:"updated_at"`
}

type CreateUserParams struct {
Email string `json:"email"`
Name string `json:"name"`
Status string `json:"status"`
Metadata json.RawMessage `json:"metadata"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
row:= q.db.QueryRow(ctx, createUser, arg.Email, arg.Name, arg.Status, arg.Metadata)
var i User
err:= row.Scan(&i.ID, &i.Email, &i.Name, &i.Status, &i.Metadata, &i.CreatedAt, &i.UpdatedAt)
return i, err
}

//... GetUserByID, GetUserByEmail, ListUsers, etc.
```

E `db.go` (interface `Querier` + `New`):

```go
type Querier interface {
CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
//... todas as queries
}

type Queries struct {
db DBTX
}

func New(db DBTX) *Queries {
return &Queries{db: db}
}
```

## Integração com Repository Pattern

O sqlc gera tipos "burros" de DB. O **repository** traduz para o domínio.

`internal/repository/user_repository.go`:

```go
package repository

import (
"context"
"encoding/json"
"errors"
"fmt"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"

"github.com/yourorg/project/internal/database/sqlc"
"github.com/yourorg/project/internal/domain"
apperrors "github.com/yourorg/project/pkg/errors"
)

type UserRepository struct {
q sqlc.Querier
}

func NewUserRepository(q sqlc.Querier) *UserRepository {
return &UserRepository{q: q}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
meta, err:= json.Marshal(u.Metadata)
if err!= nil {
return nil, fmt.Errorf("marshal metadata: %w", err)
}

row, err:= r.q.CreateUser(ctx, sqlc.CreateUserParams{
Email: u.Email,
Name: u.Name,
Status: string(u.Status),
Metadata: meta,
})
if err!= nil {
return nil, fmt.Errorf("criar usuário %q: %w", u.Email, err)
}
return toDomainUser(row), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
row, err:= r.q.GetUserByID(ctx, id)
if err!= nil {
if errors.Is(err, pgx.ErrNoRows) {
return nil, fmt.Errorf("usuário %s: %w", id, apperrors.ErrNotFound)
}
return nil, fmt.Errorf("buscar usuário %s: %w", id, err)
}
return toDomainUser(row), nil
}

func toDomainUser(r sqlc.User) *domain.User {
var meta map[string]any
_ = json.Unmarshal(r.Metadata, &meta)
return &domain.User{
ID: r.ID,
Email: r.Email,
Name: r.Name,
Status: domain.UserStatus(r.Status),
Metadata: meta,
CreatedAt: r.CreatedAt,
UpdatedAt: r.UpdatedAt,
}
}
```

**Padrões importantes:**
- `sqlc.Querier` (interface) no constructor — permite mock em testes
- Tradução `sqlc.User` ↔ `domain.User` sempre via função privada (`toDomainUser`)
- `pgx.ErrNoRows` virado em `apperrors.ErrNotFound` (sentinel da camada de domínio)
- Erros wrapped com contexto operacional

## Configuração da Conexão (pgx pool)

`internal/database/db.go`:

```go
package database

import (
"context"
"fmt"
"time"

"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
Host string
Port int
User string
Password string
Name string
SSLMode string

MaxConns int32 // padrão: 25
MinConns int32 // padrão: 5
MaxConnLifetime time.Duration // padrão: 1h
MaxConnIdleTime time.Duration // padrão: 30min
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
dsn:= fmt.Sprintf(
"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
)

poolCfg, err:= pgxpool.ParseConfig(dsn)
if err!= nil {
return nil, fmt.Errorf("parse dsn: %w", err)
}

poolCfg.MaxConns = cfg.MaxConns
poolCfg.MinConns = cfg.MinConns
poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime

pool, err:= pgxpool.NewWithConfig(ctx, poolCfg)
if err!= nil {
return nil, fmt.Errorf("criar pool pgx: %w", err)
}

if err:= pool.Ping(ctx); err!= nil {
return nil, fmt.Errorf("ping no banco: %w", err)
}
return pool, nil
}
```

E no `main.go`:

```go
pool, err:= database.NewPool(ctx, cfg.DB)
if err!= nil { log.Fatal("db: ", err) }
defer pool.Close()

queries:= sqlc.New(pool)
userRepo:= repository.NewUserRepository(queries)
```

## Targets do Makefile

```makefile
.PHONY: sqlc-gen migrate-up migrate-down migrate-status migrate-new db-up db-down

# Gera código sqlc a partir das queries
sqlc-gen:
	sqlc generate

# Aplica todas as migrations pendentes
migrate-up:
	tern migrate -m internal/database/migrations

# Rollback de N migrations (padrão: 1)
migrate-down:
	tern migrate -m internal/database/migrations -d -1

# Status atual das migrations
migrate-status:
	tern status -m internal/database/migrations

# Cria nova migration: make migrate-new NAME=add_orders_table
migrate-new:
	@if [ -z "$(NAME)" ]; then echo "Uso: make migrate-new NAME=descricao"; exit 1; fi
	tern new -m internal/database/migrations $(NAME)

# Sobe banco local de dev
db-up:
	docker compose up -d postgres
	@sleep 2
	$(MAKE) migrate-up

db-down:
	docker compose down postgres
```

## Workflow Diário

1. **Nova feature precisa de tabela nova**:
```bash
make migrate-new NAME=add_orders_table
# edita o arquivo gerado em migrations/
make migrate-up
```

2. **Nova query**:
- Edita `internal/database/queries/<entity>.sql` adicionando query nomeada
- `make sqlc-gen`
- Usa em repository

3. **Alteração de tabela existente**:
```bash
make migrate-new NAME=add_phone_to_users
# ALTER TABLE users ADD COLUMN phone TEXT;
# No down: ALTER TABLE users DROP COLUMN phone;
make migrate-up
make sqlc-gen # se a coluna afeta queries existentes
```

4. **Em CI**:
- `sqlc generate` é executado para garantir que o código gerado está atualizado (`git diff --exit-code internal/database/sqlc`)
- Migrations rodam em banco de teste antes dos testes de integração

## Anti-padrões sqlc + tern

```sql
-- ❌ Sem RETURNING em INSERT que precisa do ID
INSERT INTO users (email, name) VALUES ($1, $2);
-- ✅ Sempre RETURNING quando precisa do registro criado
INSERT INTO users (email, name) VALUES ($1, $2) RETURNING *;

-- ❌ Query genérica que sqlc não consegue tipar bem
SELECT * FROM users WHERE name LIKE '%' || $1 || '%';
-- ✅ Tipo explícito de parâmetro quando ambiguo
SELECT * FROM users WHERE name LIKE '%' || $1::text || '%';

-- ❌ Down vazio ou genérico
---- create above / drop below ----
-- TODO

-- ❌ Múltiplas migrations editando o mesmo objeto sem cuidado
-- 005_add_column.sql cria coluna X
-- 006_change_column.sql altera X
-- 007_change_column_again.sql altera X de novo
-- → consolide em PR review, ou use migrations descritivas

-- ❌ Dados de seed em migrations versionadas
INSERT INTO config VALUES ('feature_x_enabled', 'true');
-- ✅ Seeds ficam em scripts/seed.sql separado, idempotente
```

```go
// ❌ Usar *Queries (concreto) no repository
func NewUserRepository(q *sqlc.Queries) *UserRepository

// ✅ Usar Querier (interface) — permite mock
func NewUserRepository(q sqlc.Querier) *UserRepository

// ❌ Repassar sqlc.User no service / handler
// (vaza o tipo gerado para camadas externas)

// ✅ Sempre converter para domain.User no repository
```

## Transações com pgx

sqlc + pgx suporta transações elegantemente:

```go
func (r *UserRepository) CreateWithProfile(ctx context.Context, u *domain.User, p *domain.Profile) error {
tx, err:= r.pool.BeginTx(ctx, pgx.TxOptions{})
if err!= nil {
return fmt.Errorf("begin tx: %w", err)
}
defer tx.Rollback(ctx) // no-op se já commitado

qtx:= r.q.WithTx(tx) // sqlc gera WithTx automaticamente

user, err:= qtx.CreateUser(ctx, sqlc.CreateUserParams{...})
if err!= nil {
return fmt.Errorf("criar user: %w", err)
}

_, err = qtx.CreateProfile(ctx, sqlc.CreateProfileParams{
UserID: user.ID,
Bio: p.Bio,
})
if err!= nil {
return fmt.Errorf("criar profile: %w", err)
}

if err:= tx.Commit(ctx); err!= nil {
return fmt.Errorf("commit: %w", err)
}
return nil
}
```

Para isso o repository precisa do `*pgxpool.Pool` também (não só do `Querier`):

```go
type UserRepository struct {
q sqlc.Querier
pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
return &UserRepository{
q: sqlc.New(pool),
pool: pool,
}
}
```

## Testes com testcontainers-go

```go
package repository_test

import (
"context"
"testing"

"github.com/stretchr/testify/require"
"github.com/testcontainers/testcontainers-go/modules/postgres"
"github.com/jackc/pgx/v5/pgxpool"
"github.com/jackc/tern/v2/migrate"

"github.com/yourorg/project/internal/database/sqlc"
"github.com/yourorg/project/internal/domain"
"github.com/yourorg/project/internal/repository"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
t.Helper()
ctx:= context.Background()

container, err:= postgres.Run(ctx,
"postgres:16-alpine",
postgres.WithDatabase("testdb"),
postgres.WithUsername("test"),
postgres.WithPassword("test"),
)
require.NoError(t, err)
t.Cleanup(func() { _ = container.Terminate(ctx) })

dsn, err:= container.ConnectionString(ctx, "sslmode=disable")
require.NoError(t, err)

pool, err:= pgxpool.New(ctx, dsn)
require.NoError(t, err)
t.Cleanup(pool.Close)

// Roda migrations via tern
conn, _:= pool.Acquire(ctx)
defer conn.Release()
migrator, _:= migrate.NewMigrator(ctx, conn.Conn(), "schema_version")
_ = migrator.LoadMigrations(nil) // path para migrations
_ = migrator.Migrate(ctx)

return pool
}

func TestUserRepository_Create(t *testing.T) {
pool:= setupTestDB(t)
repo:= repository.NewUserRepository(pool)

user, err:= repo.Create(context.Background(), &domain.User{
Email: "test@example.com",
Name: "Test User",
})
require.NoError(t, err)
require.NotEmpty(t, user.ID)
require.Equal(t, "test@example.com", user.Email)
}
```

## Checklist sqlc + tern

Antes de fazer merge:

- [ ] Migration tem `up` e `down` que se cancelam mutuamente
- [ ] Migration testada em DB fresh: `tern migrate` + `tern migrate -d -1` + `tern migrate` (deve funcionar)
- [ ] `sqlc generate` rodado e arquivo `internal/database/sqlc/*` commitado
- [ ] `sqlc.yaml` não inclui novos overrides desnecessários
- [ ] Queries têm `-- name:` claro e `:one`/`:many`/`:exec` correto
- [ ] Repository converte `sqlc.X` para `domain.X` (não vaza tipo gerado)
- [ ] Constructor recebe `sqlc.Querier` (interface), não `*sqlc.Queries`
- [ ] Erros do pgx (especialmente `pgx.ErrNoRows`) traduzidos para sentinels de domínio
- [ ] Teste de integração com testcontainers-go cobre o caminho feliz da nova query
- [ ] CI roda `sqlc generate` e falha se `git diff` não vazio
