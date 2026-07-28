# sqlc — Queries SQL Type-Safe

## Filosofia

- Voce escreve SQL puro; sqlc gera codigo Go tipado
- Banco e a **fonte de verdade** do schema

## sqlc.yaml

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
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_interface: true
        emit_empty_slices: true
        emit_pointers_for_null_types: true
        omit_unused_structs: true
        query_parameter_limit: 3
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

**Pontos criticos:**
- `sql_package: "pgx/v5"` — usar pgx, nao `database/sql`
- `emit_interface: true` — gera `Querier` interface (mockavel)
- `schema` aponta para migrations

## Query sqlc

```sql
-- name: CreateUser :one
INSERT INTO users (email, name, status, metadata)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: UpdateUserStatus :exec
UPDATE users SET status = $2 WHERE id = $1;
```

## Anotacoes

- `:one` — retorna uma struct
- `:many` — retorna `[]struct`
- `:exec` — sem retorno
- `:execrows` — retorna `int64` (rows affected)

## Integracao Repository

```go
type UserRepository struct { q sqlc.Querier }

func NewUserRepository(q sqlc.Querier) *UserRepository {
    return &UserRepository{q: q}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
    row, err := r.q.CreateUser(ctx, sqlc.CreateUserParams{...})
    if err != nil { return nil, fmt.Errorf("criar usuario: %w", err) }
    return toDomainUser(row), nil
}
```

**Padroes:**
- `sqlc.Querier` (interface) no constructor
- Sempre converter `sqlc.X` → `domain.X`
- `pgx.ErrNoRows` → `apperrors.ErrNotFound`

## Anti-padroes

```sql
-- ❌ Sem RETURNING em INSERT que precisa do ID
INSERT INTO users (email, name) VALUES ($1, $2);
-- ✅ Sempre RETURNING quando precisa do registro
INSERT INTO users (email, name) VALUES ($1, $2) RETURNING *;

-- ❌ Down vazio
---- create above / drop below ----
-- TODO
```

```go
// ❌ Usar *Queries (concreto)
func NewUserRepository(q *sqlc.Queries) *UserRepository
// ✅ Usar Querier (interface)
func NewUserRepository(q sqlc.Querier) *UserRepository
```

## Makefile Targets

```makefile
sqlc-gen:
	sqlc generate
```
