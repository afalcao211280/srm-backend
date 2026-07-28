# pgx/v5 — PostgreSQL Driver

## Pool de Conexoes

```go
package database

import (
    "context"
    "fmt"
    "time"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
    Host, Password, Name, SSLMode string
    Port, MaxConns, MinConns      int
    MaxConnLifetime               time.Duration
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
    
    poolCfg, err := pgxpool.ParseConfig(dsn)
    if err != nil { return nil, fmt.Errorf("parse dsn: %w", err) }
    
    poolCfg.MaxConns = int32(cfg.MaxConns)
    poolCfg.MinConns = int32(cfg.MinConns)
    poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
    
    pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
    if err != nil { return nil, fmt.Errorf("criar pool: %w", err) }
    
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("ping: %w", err)
    }
    return pool, nil
}
```

## Transacoes

```go
func (r *UserRepository) CreateWithProfile(ctx context.Context, u *domain.User, p *domain.Profile) error {
    tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
    if err != nil { return fmt.Errorf("begin tx: %w", err) }
    defer tx.Rollback(ctx)
    
    qtx := r.q.WithTx(tx)
    
    user, err := qtx.CreateUser(ctx, sqlc.CreateUserParams{...})
    if err != nil { return fmt.Errorf("criar user: %w", err) }
    
    _, err = qtx.CreateProfile(ctx, sqlc.CreateProfileParams{UserID: user.ID, ...})
    if err != nil { return fmt.Errorf("criar profile: %w", err) }
    
    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit: %w", err)
    }
    return nil
}
```

**Requisito:** Repository precisa do `*pgxpool.Pool` para transacoes:
```go
type UserRepository struct {
    q    sqlc.Querier
    pool *pgxpool.Pool
}
```

## Padroes
- MaxConns: 25, MinConns: 5
- Sempre Ping apos criar pool
- Transacoes curtas e focadas
