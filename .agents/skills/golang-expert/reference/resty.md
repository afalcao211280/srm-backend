# Resty + Circuit Breaker

## Resty HTTP Client

```go
package http

import (
    "context"
    "net/http"
    "time"
    "github.com/go-resty/resty/v2"
)

type Client struct { r *resty.Client }

type Config struct {
    BaseURL    string
    Timeout    time.Duration
    RetryCount int
    Token      string
}

func New(cfg Config) *Client {
    r := resty.New().
        SetBaseURL(cfg.BaseURL).
        SetTimeout(cfg.Timeout).
        SetRetryCount(cfg.RetryCount).
        SetRetryWaitTime(500 * time.Millisecond).
        SetRetryMaxWaitTime(5 * time.Second).
        AddRetryCondition(func(r *resty.Response, err error) bool {
            return err != nil || r.StatusCode() >= http.StatusInternalServerError
        })
    if cfg.Token != "" { r.SetAuthToken(cfg.Token) }
    return &Client{r: r}
}

func (c *Client) Get(ctx context.Context, path string, out any) error {
    resp, err := c.r.R().SetContext(ctx).SetResult(out).Get(path)
    if err != nil { return fmt.Errorf("GET %s: %w", path, err) }
    if resp.IsError() { return fmt.Errorf("GET %s: status=%d", path, resp.StatusCode()) }
    return nil
}

func (c *Client) Post(ctx context.Context, path string, in, out any) error {
    resp, err := c.r.R().SetContext(ctx).SetBody(in).SetResult(out).Post(path)
    if err != nil { return fmt.Errorf("POST %s: %w", path, err) }
    if resp.IsError() { return fmt.Errorf("POST %s: status=%d", path, resp.StatusCode()) }
    return nil
}
```

## Circuit Breaker (gobreaker)

```go
package circuitbreaker

import (
    "context"
    "log/slog"
    "time"
    "github.com/sony/gobreaker/v2"
)

type Config struct {
    Name         string
    MaxRequests  uint32
    Interval     time.Duration
    Timeout      time.Duration
    FailureRatio float64
    MinRequests  uint32
}

func New[T any](cfg Config) *gobreaker.CircuitBreaker[T] {
    return gobreaker.NewCircuitBreaker[T](gobreaker.Settings{
        Name:        cfg.Name,
        MaxRequests: cfg.MaxRequests,
        Interval:    cfg.Interval,
        Timeout:     cfg.Timeout,
        ReadyToTrip: func(c gobreaker.Counts) bool {
            if c.Requests < cfg.MinRequests { return false }
            ratio := float64(c.TotalFailures) / float64(c.Requests)
            return ratio >= cfg.FailureRatio
        },
    })
}
```

## Uso Combinado

```go
type ExternalAPIClient struct {
    http *http.Client
    cb   *gobreaker.CircuitBreaker[*UserDTO]
    log  *slog.Logger
}

func (c *ExternalAPIClient) GetUser(ctx context.Context, id string) (*UserDTO, error) {
    return c.cb.Execute(func() (*UserDTO, error) {
        var u UserDTO
        if err := c.http.Get(ctx, "/users/"+id, &u); err != nil {
            return nil, err
        }
        return &u, nil
    })
}
```

## Padroes
- Sempre `SetContext(ctx)` — propaga cancelamento
- Sempre tratar `resp.IsError()`
- Circuit breaker em servicos externos criticos
- Retry apenas em 5xx e erros de rede
