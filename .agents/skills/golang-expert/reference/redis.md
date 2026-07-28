# Redis (go-redis/v9)

## Setup

```go
package cache

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "time"
    "github.com/redis/go-redis/v9"
)

type Cache struct { client *redis.Client }

func New(addr, password string, db int) *Cache {
    return &Cache{client: redis.NewClient(&redis.Options{
        Addr: addr, Password: password, DB: db,
        DialTimeout: 5 * time.Second,
        ReadTimeout: 3 * time.Second,
    })}
}
```

## Get/Set

```go
func (c *Cache) Get(ctx context.Context, key string, dest any) error {
    val, err := c.client.Get(ctx, key).Result()
    if err != nil {
        if errors.Is(err, redis.Nil) { return ErrCacheMiss }
        return fmt.Errorf("redis get %s: %w", key, err)
    }
    if err := json.Unmarshal([]byte(val), dest); err != nil {
        return fmt.Errorf("unmarshal %s: %w", key, err)
    }
    return nil
}

func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil { return fmt.Errorf("marshal %s: %w", key, err) }
    return c.client.Set(ctx, key, data, ttl).Err()
}
```

## Lock Distribuido

```go
func (c *Cache) Lock(ctx context.Context, key string, ttl time.Duration) (func() error, error) {
    ok, err := c.client.SetNX(ctx, key, "1", ttl).Result()
    if err != nil { return nil, fmt.Errorf("lock %s: %w", key, err) }
    if !ok { return nil, ErrLockHeld }
    return func() error {
        return c.client.Del(context.Background(), key).Err()
    }, nil
}

var ErrCacheMiss = errors.New("cache miss")
var ErrLockHeld  = errors.New("lock ja adquirido")
```

## Padroes
- Context em todas as operacoes
- JSON para serializacao de structs
- SetNX para locks distribuidos
- Erros especificos (nao genericos)
