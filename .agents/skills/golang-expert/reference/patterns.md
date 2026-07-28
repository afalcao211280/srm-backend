# Stack & Patterns — Snippets Canônicos

Referência rápida de patterns de código por biblioteca. Leia a seção da lib que vai usar **antes** de codar — esses snippets são o ponto de partida, não exemplos educativos.

## Índice

1. [Logger (slog)](#1-logger-slog)
2. [Gin + Middlewares](#2-gin--middlewares)
3. [Huma v2 (OpenAPI)](#3-huma-v2-openapi)
4. [Resty HTTP Client](#4-resty-http-client)
5. [gobreaker Circuit Breaker](#5-gobreaker-circuit-breaker)
6. [DI Manual + sqlc Repository](#6-di-manual--sqlc-repository)
7. [MongoDB](#7-mongodb)
8. [Redis (go-redis)](#8-redis-go-redis)
9. [RabbitMQ (amqp091-go)](#9-rabbitmq-amqp091-go)
10. [Kafka (franz-go)](#10-kafka-franz-go)
11. [MinIO](#11-minio)
12. [Gotenberg PDF](#12-gotenberg-pdf)
13. [OpenTelemetry](#13-opentelemetry)
14. [Prometheus](#14-prometheus)
15. [Testes (testify + testcontainers-go)](#15-testes-testify--testcontainers-go)
16. [Qualidade / SonarQube](#16-qualidade--sonarqube)

---

## 1. Logger (slog)

`pkg/logger/logger.go`:

```go
package logger

import (
"context"
"log/slog"
"os"
)

type ctxKey struct{}

type Level = slog.Level

const (
LevelDebug = slog.LevelDebug
LevelInfo = slog.LevelInfo
LevelWarn = slog.LevelWarn
LevelError = slog.LevelError
)

// New cria um logger JSON com nível configurável.
// Em dev pode usar slog.NewTextHandler para output legível.
func New(level Level) *slog.Logger {
return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
Level: level,
ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
if a.Key == slog.TimeKey {
a.Key = "timestamp"
}
return a
},
}))
}

// With injeta um logger no contexto.
func With(ctx context.Context, log *slog.Logger) context.Context {
return context.WithValue(ctx, ctxKey{}, log)
}

// From recupera o logger do contexto. Fallback para Default.
func From(ctx context.Context) *slog.Logger {
if log, ok:= ctx.Value(ctxKey{}).(*slog.Logger); ok {
return log
}
return slog.Default()
}
```

**Uso:**
```go
log:= logger.From(ctx)
log.Info("evento", "user_id", userID, "action", "login")
log.Error("falha", "error", err)
```

**Não faça:**
- `log.Printf("user %s...", id)` — use atributos estruturados
- Logar campos sensíveis (senha, token, CPF, cartão)

---

## 2. Gin + Middlewares

`cmd/<service>/main.go` (esqueleto):

```go
r:= gin.New()
r.Use(gin.Recovery())
r.Use(middleware.CorrelationID())
r.Use(middleware.Logger(log))
r.Use(middleware.Tracing("user-service"))
r.Use(metrics.Middleware())

r.GET("/health", handler.Health)
r.GET("/metrics", gin.WrapH(promhttp.Handler()))

v1:= r.Group("/api/v1")
handler.RegisterUsers(v1, userSvc, log)
```

`internal/middleware/correlation_id.go`:

```go
package middleware

import (
"context"

"github.com/gin-gonic/gin"
"github.com/google/uuid"
)

type correlationKey struct{}

const headerCorrelationID = "X-Correlation-ID"

func CorrelationID() gin.HandlerFunc {
return func(c *gin.Context) {
id:= c.GetHeader(headerCorrelationID)
if id == "" {
id = uuid.NewString()
}
ctx:= context.WithValue(c.Request.Context(), correlationKey{}, id)
c.Request = c.Request.WithContext(ctx)
c.Header(headerCorrelationID, id)
c.Next()
}
}

func CorrelationIDFrom(ctx context.Context) string {
if v, ok:= ctx.Value(correlationKey{}).(string); ok {
return v
}
return ""
}
```

`internal/middleware/logger.go`:

```go
func Logger(log *slog.Logger) gin.HandlerFunc {
return func(c *gin.Context) {
start:= time.Now()
cid:= CorrelationIDFrom(c.Request.Context())

reqLog:= log.With(
"correlation_id", cid,
"method", c.Request.Method,
"path", c.FullPath(),
)
ctx:= logger.With(c.Request.Context(), reqLog)
c.Request = c.Request.WithContext(ctx)

c.Next()

reqLog.Info("request concluído",
"status", c.Writer.Status(),
"duration_ms", time.Since(start).Milliseconds(),
"bytes", c.Writer.Size(),
)
}
}
```

---

## 3. Huma v2 (OpenAPI)

Huma vai **junto** com Gin (não substitui). Use Huma para rotas que precisam de OpenAPI gerado automaticamente; use Gin direto para health/metrics/probes simples.

```go
import (
"github.com/danielgtaylor/huma/v2"
"github.com/danielgtaylor/huma/v2/adapters/humagin"
)

// No main.go
ginEngine:= gin.New()
ginEngine.Use(...) // middlewares

api:= humagin.New(ginEngine, huma.DefaultConfig("Users API", "1.0.0"))

// Registra recursos
handler.RegisterUsersHuma(api, userSvc)
```

`internal/handler/users_huma.go`:

```go
package handler

import (
"context"
"net/http"

"github.com/danielgtaylor/huma/v2"
"github.com/google/uuid"

"github.com/yourorg/project/internal/service"
"github.com/yourorg/project/pkg/logger"
)

type CreateUserInput struct {
Body struct {
Name string `json:"name" minLength:"1" maxLength:"100" doc:"Nome completo"`
Email string `json:"email" format:"email" doc:"Email único"`
Age *int `json:"age,omitempty" minimum:"0" maximum:"150"`
}
}

type UserOutput struct {
Body struct {
ID uuid.UUID `json:"id"`
Name string `json:"name"`
Email string `json:"email"`
Status string `json:"status"`
CreatedAt string `json:"created_at"`
}
}

func RegisterUsersHuma(api huma.API, svc *service.UserService) {
huma.Register(api, huma.Operation{
OperationID: "create-user",
Method: http.MethodPost,
Path: "/api/v1/users",
Summary: "Cria um usuário",
Tags: []string{"Users"},
}, func(ctx context.Context, in *CreateUserInput) (*UserOutput, error) {
log:= logger.From(ctx)
log.Info("criando usuário", "email", in.Body.Email)

user, err:= svc.Create(ctx, service.CreateUserInput{
Name: in.Body.Name,
Email: in.Body.Email,
Age: in.Body.Age,
})
if err!= nil {
return nil, huma.Error400BadRequest("falha ao criar usuário", err)
}

out:= &UserOutput{}
out.Body.ID = user.ID
out.Body.Name = user.Name
out.Body.Email = user.Email
out.Body.Status = string(user.Status)
out.Body.CreatedAt = user.CreatedAt.Format(time.RFC3339)
return out, nil
})
}
```

**Convenções Huma:**
- Struct de input/output **sempre tem campo `Body`** envolvendo o payload real.
- Validação via tags (`minLength`, `maxLength`, `format`, `minimum`, `maximum`, `pattern`).
- Erros: `huma.Error400BadRequest`, `huma.Error404NotFound`, `huma.Error500InternalServerError`.
- `OperationID` é único no projeto — usado para nomes de SDK gerado.

---

## 4. Resty HTTP Client

`pkg/http/client.go`:

```go
package http

import (
"context"
"net/http"
"time"

"github.com/go-resty/resty/v2"
)

type Client struct {
r *resty.Client
}

type Config struct {
BaseURL string
Timeout time.Duration
RetryCount int
Token string // opcional, p/ Bearer
}

func New(cfg Config) *Client {
r:= resty.New().
SetBaseURL(cfg.BaseURL).
SetTimeout(cfg.Timeout).
SetRetryCount(cfg.RetryCount).
SetRetryWaitTime(500 * time.Millisecond).
SetRetryMaxWaitTime(5 * time.Second).
AddRetryCondition(func(r *resty.Response, err error) bool {
return err!= nil || r.StatusCode() >= http.StatusInternalServerError
})

if cfg.Token!= "" {
r.SetAuthToken(cfg.Token)
}

return &Client{r: r}
}

func (c *Client) Get(ctx context.Context, path string, out any) error {
resp, err:= c.r.R().
SetContext(ctx).
SetResult(out).
Get(path)
if err!= nil {
return fmt.Errorf("GET %s: %w", path, err)
}
if resp.IsError() {
return fmt.Errorf("GET %s: status=%d body=%s", path, resp.StatusCode(), resp.String())
}
return nil
}

func (c *Client) Post(ctx context.Context, path string, in, out any) error {
resp, err:= c.r.R().
SetContext(ctx).
SetBody(in).
SetResult(out).
Post(path)
if err!= nil {
return fmt.Errorf("POST %s: %w", path, err)
}
if resp.IsError() {
return fmt.Errorf("POST %s: status=%d body=%s", path, resp.StatusCode(), resp.String())
}
return nil
}
```

**Regras:**
- Sempre `SetContext(ctx)` — propaga cancelamento e deadline
- Sempre tratar `resp.IsError()` (Resty não erra em 4xx/5xx por padrão)
- Para serviços externos críticos, envolver em circuit breaker (próxima seção)

---

## 5. gobreaker Circuit Breaker

`pkg/circuitbreaker/breaker.go`:

```go
package circuitbreaker

import (
"context"
"log/slog"
"time"

"github.com/sony/gobreaker/v2"
)

type Config struct {
Name string
MaxRequests uint32 // requests permitidos em half-open
Interval time.Duration // janela de contagem
Timeout time.Duration // tempo em open antes de tentar half-open
FailureRatio float64 // ratio para abrir (ex: 0.6 = 60%)
MinRequests uint32 // mínimo de requests antes de calcular ratio
OnStateChange func(name string, from, to gobreaker.State)
}

func New[T any](cfg Config) *gobreaker.CircuitBreaker[T] {
return gobreaker.NewCircuitBreaker[T](gobreaker.Settings{
Name: cfg.Name,
MaxRequests: cfg.MaxRequests,
Interval: cfg.Interval,
Timeout: cfg.Timeout,
ReadyToTrip: func(c gobreaker.Counts) bool {
if c.Requests < cfg.MinRequests {
return false
}
ratio:= float64(c.TotalFailures) / float64(c.Requests)
return ratio >= cfg.FailureRatio
},
OnStateChange: cfg.OnStateChange,
})
}
```

**Uso envolvendo Resty:**

```go
type ExternalAPIClient struct {
http *http.Client
cb *gobreaker.CircuitBreaker[*UserDTO]
log *slog.Logger
}

func (c *ExternalAPIClient) GetUser(ctx context.Context, id string) (*UserDTO, error) {
return c.cb.Execute(func() (*UserDTO, error) {
var u UserDTO
if err:= c.http.Get(ctx, "/users/"+id, &u); err!= nil {
return nil, err
}
return &u, nil
})
}
```

---

## 6. DI Manual + sqlc Repository

### DI Manual — `internal/server/server.go`

`internal/server/server.go` faz a injeção de dependências manual:

```go
package server

import (
"context"
"fmt"
"log/slog"
"net/http"

"github.com/gin-gonic/gin"
"github.com/yourorg/project/internal/config"
"github.com/yourorg/project/internal/database"
"github.com/yourorg/project/internal/database/sqlc"
"github.com/yourorg/project/internal/handler"
"github.com/yourorg/project/internal/middleware"
"github.com/yourorg/project/internal/repository"
"github.com/yourorg/project/internal/service"
"github.com/yourorg/project/pkg/logger"
"github.com/yourorg/project/pkg/tracer"
)

type Server struct {
httpServer *http.Server
log *slog.Logger
closers []func() error
}

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) (*Server, error) {
// 1. Infra: PostgreSQL
pool, err:= database.NewPool(ctx, cfg.DB)
if err!= nil {
return nil, fmt.Errorf("new pool: %w", err)
}

// 2. Observability
shutdownTracer, err:= tracer.Init(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
if err!= nil {
return nil, fmt.Errorf("tracer init: %w", err)
}

// 3. Repository (sqlc gerado)
queries:= sqlc.New(pool)
userRepo:= repository.NewUserRepository(queries)

// 4. Service (interface repository declarada aqui)
userSvc:= service.NewUserService(userRepo, log)

// 5. Handler (Gin + Huma)
ginEngine:= gin.New()
ginEngine.Use(
middleware.CorrelationID(),
middleware.Logger(log),
middleware.Tracing(cfg.ServiceName),
middleware.Recovery(),
)
userHandler:= handler.NewUserHandler(userSvc)
userHandler.RegisterRoutes(ginEngine)

// 6. HTTP Server
srv:= &http.Server{
Addr: fmt.Sprintf(":%s", cfg.HTTPPort),
Handler: ginEngine,
}

return &Server{
httpServer: srv,
log: log,
closers: []func() error{pool.Close, shutdownTracer},
}, nil
}

func (s *Server) Run(ctx context.Context) error {
s.log.Info("server starting", "addr", s.httpServer.Addr)
if err:= s.httpServer.ListenAndServe(); err!= nil && err!= http.ErrServerClosed {
return fmt.Errorf("server: %w", err)
}
return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
for _, c:= range s.closers {
if err:= c(); err!= nil {
s.log.Error("shutdown error", "error", err)
}
}
return s.httpServer.Shutdown(ctx)
}
```

### cmd/server/main.go

```go
package main

import (
"context"
"log/slog"
"os"
"os/signal"
"syscall"

"github.com/yourorg/project/internal/config"
"github.com/yourorg/project/internal/server"
"github.com/yourorg/project/pkg/logger"
)

func main() {
ctx, cancel:= signal.NotifyContext(context.Background(),
syscall.SIGINT, syscall.SIGTERM)
defer cancel()

cfg, err:= config.Load()
if err!= nil {
slog.Default().Error("config", "error", err)
os.Exit(1)
}

log:= logger.New(cfg.Env)

srv, err:= server.New(ctx, cfg, log)
if err!= nil {
log.Error("server init", "error", err)
os.Exit(1)
}

go func() {
if err:= srv.Run(ctx); err!= nil {
log.Error("server run", "error", err)
cancel()
}
}()

<-ctx.Done()
log.Info("shutting down")
if err:= srv.Shutdown(context.Background()); err!= nil {
log.Error("shutdown", "error", err)
}
}
```

### sqlc Repository Pattern

Interface declarada no **consumidor** (`internal/service/user_service.go`):

```go
package service

import (
"context"
"github.com/google/uuid"
"github.com/yourorg/project/internal/domain"
)

type UserRepository interface {
Create(ctx context.Context, u *domain.User) (*domain.User, error)
GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
List(ctx context.Context, status string, limit, offset int32) ([]*domain.User, error)
Update(ctx context.Context, u *domain.User) (*domain.User, error)
Delete(ctx context.Context, id uuid.UUID) error
}
```

Implementação no `repository/` usando sqlc:

```go
package repository

import (
"context"
"fmt"
"github.com/google/uuid"
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
params:= sqlc.CreateUserParams{Name: u.Name, Email: u.Email, Status: u.Status}
result, err:= r.q.CreateUser(ctx, params)
if err!= nil {
return nil, fmt.Errorf("criar usuário: %w", err)
}
return toDomainUser(result), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
result, err:= r.q.GetUserByID(ctx, id)
if err!= nil {
return nil, fmt.Errorf("buscar usuário %s: %w", id, apperrors.ErrNotFound)
}
return toDomainUser(result), nil
}
```

---

## 7. MongoDB

`internal/repository/mongo/user_repository.go`:

```go
package mongo

import (
"context"
"errors"
"fmt"
"time"

"go.mongodb.org/mongo-driver/bson"
"go.mongodb.org/mongo-driver/bson/primitive"
"go.mongodb.org/mongo-driver/mongo"
"go.mongodb.org/mongo-driver/mongo/options"

"github.com/yourorg/project/internal/domain"
apperrors "github.com/yourorg/project/pkg/errors"
)

type userDoc struct {
ID primitive.ObjectID `bson:"_id,omitempty"`
Name string `bson:"name"`
Email string `bson:"email"`
Status string `bson:"status"`
CreatedAt time.Time `bson:"created_at"`
UpdatedAt time.Time `bson:"updated_at"`
}

type UserRepository struct {
coll *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
return &UserRepository{coll: db.Collection("users")}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
now:= time.Now().UTC()
doc:= userDoc{
Name: u.Name, Email: u.Email, Status: string(u.Status),
CreatedAt: now, UpdatedAt: now,
}
res, err:= r.coll.InsertOne(ctx, doc)
if err!= nil {
return nil, fmt.Errorf("inserir usuário: %w", err)
}
doc.ID = res.InsertedID.(primitive.ObjectID)
return toDomainUser(doc), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
var doc userDoc
err:= r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&doc)
if err!= nil {
if errors.Is(err, mongo.ErrNoDocuments) {
return nil, fmt.Errorf("email %s: %w", email, apperrors.ErrNotFound)
}
return nil, fmt.Errorf("buscar por email: %w", err)
}
return toDomainUser(doc), nil
}
```

**Setup do client:**

```go
clientOpts:= options.Client().
ApplyURI(cfg.URI).
SetMaxPoolSize(50).
SetMinPoolSize(5).
SetServerSelectionTimeout(5 * time.Second)

client, err:= mongo.Connect(ctx, clientOpts)
if err!= nil { return fmt.Errorf("conectar mongo: %w", err) }

if err:= client.Ping(ctx, nil); err!= nil {
return fmt.Errorf("ping mongo: %w", err)
}
```

---

## 8. Redis (go-redis)

`internal/cache/redis.go`:

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

type Cache struct {
client *redis.Client
}

func New(addr, password string, db int) *Cache {
return &Cache{
client: redis.NewClient(&redis.Options{
Addr: addr, Password: password, DB: db,
DialTimeout: 5 * time.Second,
ReadTimeout: 3 * time.Second,
}),
}
}

func (c *Cache) Get(ctx context.Context, key string, dest any) error {
val, err:= c.client.Get(ctx, key).Result()
if err!= nil {
if errors.Is(err, redis.Nil) {
return ErrCacheMiss
}
return fmt.Errorf("redis get %s: %w", key, err)
}
if err:= json.Unmarshal([]byte(val), dest); err!= nil {
return fmt.Errorf("unmarshal cache %s: %w", key, err)
}
return nil
}

func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
data, err:= json.Marshal(value)
if err!= nil {
return fmt.Errorf("marshal cache %s: %w", key, err)
}
return c.client.Set(ctx, key, data, ttl).Err()
}

// Lock com SetNX para distribuído. Retorna função de unlock.
func (c *Cache) Lock(ctx context.Context, key string, ttl time.Duration) (func() error, error) {
ok, err:= c.client.SetNX(ctx, key, "1", ttl).Result()
if err!= nil {
return nil, fmt.Errorf("lock %s: %w", key, err)
}
if!ok {
return nil, ErrLockHeld
}
return func() error {
return c.client.Del(context.Background(), key).Err()
}, nil
}

var (
ErrCacheMiss = errors.New("cache miss")
ErrLockHeld = errors.New("lock já adquirido")
)
```

---

## 9. RabbitMQ (amqp091-go)

**Publisher:**

```go
package rabbitmq

import (
"context"
"encoding/json"
"fmt"

amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
conn *amqp.Connection
ch *amqp.Channel
}

func NewPublisher(url string) (*Publisher, error) {
conn, err:= amqp.Dial(url)
if err!= nil { return nil, fmt.Errorf("dial: %w", err) }

ch, err:= conn.Channel()
if err!= nil {
_ = conn.Close()
return nil, fmt.Errorf("channel: %w", err)
}

if err:= ch.Confirm(false); err!= nil {
return nil, fmt.Errorf("confirm mode: %w", err)
}
return &Publisher{conn: conn, ch: ch}, nil
}

func (p *Publisher) Publish(ctx context.Context, exchange, routingKey string, body any) error {
data, err:= json.Marshal(body)
if err!= nil { return fmt.Errorf("marshal: %w", err) }

return p.ch.PublishWithContext(ctx,
exchange, routingKey,
true, // mandatory
false, // immediate
amqp.Publishing{
ContentType: "application/json",
Body: data,
DeliveryMode: amqp.Persistent,
Timestamp: time.Now(),
},
)
}

func (p *Publisher) Close() error {
if err:= p.ch.Close(); err!= nil { return err }
return p.conn.Close()
}
```

**Consumer (com graceful shutdown):**

```go
func (c *Consumer) Run(ctx context.Context, queue string, handler func(context.Context, []byte) error) error {
msgs, err:= c.ch.Consume(queue, "",
false, // auto-ack desabilitado: ack manual
false, false, false, nil)
if err!= nil { return fmt.Errorf("consume: %w", err) }

for {
select {
case <-ctx.Done():
return ctx.Err()
case msg, ok:= <-msgs:
if!ok { return nil }

if err:= handler(ctx, msg.Body); err!= nil {
logger.From(ctx).Error("processar mensagem", "error", err)
_ = msg.Nack(false, true) // requeue
continue
}
_ = msg.Ack(false)
}
}
}
```

---

## 10. Kafka (franz-go)

```go
package kafka

import (
"context"
"fmt"

"github.com/twmb/franz-go/pkg/kgo"
)

// Producer assíncrono com confirmação via callback.
func NewProducer(brokers []string) (*kgo.Client, error) {
return kgo.NewClient(
kgo.SeedBrokers(brokers...),
kgo.RequiredAcks(kgo.AllISRAcks()),
kgo.ProducerLinger(10*time.Millisecond),
)
}

func ProduceSync(ctx context.Context, client *kgo.Client, topic string, key, value []byte) error {
record:= &kgo.Record{Topic: topic, Key: key, Value: value}
res:= client.ProduceSync(ctx, record)
return res.FirstErr()
}

// Consumer com group e processamento.
func NewConsumer(brokers []string, group string, topics...string) (*kgo.Client, error) {
return kgo.NewClient(
kgo.SeedBrokers(brokers...),
kgo.ConsumerGroup(group),
kgo.ConsumeTopics(topics...),
kgo.DisableAutoCommit(),
)
}

func RunConsumer(ctx context.Context, client *kgo.Client, handler func(ctx context.Context, rec *kgo.Record) error) error {
for {
fetches:= client.PollFetches(ctx)
if errs:= fetches.Errors(); len(errs) > 0 {
return fmt.Errorf("poll: %v", errs)
}

iter:= fetches.RecordIter()
for!iter.Done() {
rec:= iter.Next()
if err:= handler(ctx, rec); err!= nil {
// estratégia: log + DLQ; não commitar offset
logger.From(ctx).Error("processar record", "error", err, "topic", rec.Topic)
continue
}
}
if err:= client.CommitUncommittedOffsets(ctx); err!= nil {
return fmt.Errorf("commit offsets: %w", err)
}
}
}
```

---

## 11. MinIO

```go
package storage

import (
"context"
"fmt"
"io"
"time"

"github.com/minio/minio-go/v7"
"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
client *minio.Client
bucket string
}

func New(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*Storage, error) {
cli, err:= minio.New(endpoint, &minio.Options{
Creds: credentials.NewStaticV4(accessKey, secretKey, ""),
Secure: useSSL,
})
if err!= nil { return nil, fmt.Errorf("minio new: %w", err) }

ctx:= context.Background()
ok, err:= cli.BucketExists(ctx, bucket)
if err!= nil { return nil, fmt.Errorf("bucket exists: %w", err) }
if!ok {
if err:= cli.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err!= nil {
return nil, fmt.Errorf("create bucket: %w", err)
}
}
return &Storage{client: cli, bucket: bucket}, nil
}

func (s *Storage) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
_, err:= s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{
ContentType: contentType,
})
return err
}

func (s *Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
return s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
}

func (s *Storage) PresignedGet(ctx context.Context, key string, expiry time.Duration) (string, error) {
u, err:= s.client.PresignedGetObject(ctx, s.bucket, key, expiry, nil)
if err!= nil { return "", err }
return u.String(), nil
}
```

---

## 12. Gotenberg PDF

```go
package pdf

import (
"bytes"
"context"
"fmt"
"html/template"
"io"

"github.com/starwalkn/gotenberg-go-client/v8"
)

type Service struct {
client *gotenberg.Client
tpls *template.Template
}

func New(gotenbergURL, templatesGlob string) (*Service, error) {
c, err:= gotenberg.NewClient(gotenbergURL, nil)
if err!= nil { return nil, fmt.Errorf("gotenberg client: %w", err) }

tpls, err:= template.ParseGlob(templatesGlob)
if err!= nil { return nil, fmt.Errorf("parse templates: %w", err) }

return &Service{client: c, tpls: tpls}, nil
}

func (s *Service) Render(ctx context.Context, templateName string, data any) ([]byte, error) {
var buf bytes.Buffer
if err:= s.tpls.ExecuteTemplate(&buf, templateName, data); err!= nil {
return nil, fmt.Errorf("execute template: %w", err)
}

req:= gotenberg.NewHTMLRequest(gotenberg.NewDocumentFromBytes("index.html", buf.Bytes()))
req.PaperSize(gotenberg.A4)
req.Margins(gotenberg.NormalMargins)

resp, err:= s.client.Send(ctx, req)
if err!= nil { return nil, fmt.Errorf("gotenberg send: %w", err) }
defer resp.Body.Close()

return io.ReadAll(resp.Body)
}
```

---

## 13. OpenTelemetry

`pkg/tracer/tracer.go`:

```go
package tracer

import (
"context"
"fmt"

"go.opentelemetry.io/otel"
"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
"go.opentelemetry.io/otel/propagation"
"go.opentelemetry.io/otel/sdk/resource"
sdktrace "go.opentelemetry.io/otel/sdk/trace"
semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
"go.opentelemetry.io/otel/trace"
)

func Init(ctx context.Context, serviceName, endpoint string) (func(context.Context) error, error) {
exp, err:= otlptracegrpc.New(ctx,
otlptracegrpc.WithEndpoint(endpoint),
otlptracegrpc.WithInsecure(),
)
if err!= nil { return nil, fmt.Errorf("otlp exporter: %w", err) }

res, err:= resource.New(ctx,
resource.WithAttributes(semconv.ServiceName(serviceName)),
)
if err!= nil { return nil, fmt.Errorf("resource: %w", err) }

tp:= sdktrace.NewTracerProvider(
sdktrace.WithBatcher(exp),
sdktrace.WithResource(res),
sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1))),
)
otel.SetTracerProvider(tp)
otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
propagation.TraceContext{}, propagation.Baggage{},
))
return tp.Shutdown, nil
}

func Span(ctx context.Context, name string) (context.Context, trace.Span) {
return otel.Tracer("example").Start(ctx, name)
}
```

**Uso em service:**

```go
func (s *UserService) Create(ctx context.Context, in CreateUserInput) (*domain.User, error) {
ctx, span:= tracer.Span(ctx, "UserService.Create")
defer span.End()

span.SetAttributes(attribute.String("user.email", in.Email))

user, err:= s.repo.Create(ctx, &domain.User{...})
if err!= nil {
span.RecordError(err)
span.SetStatus(codes.Error, err.Error())
return nil, err
}
return user, nil
}
```

---

## 14. Prometheus

`pkg/metrics/metrics.go`:

```go
package metrics

import (
"strconv"
"time"

"github.com/gin-gonic/gin"
"github.com/prometheus/client_golang/prometheus"
"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
HTTPDuration = promauto.NewHistogramVec(
prometheus.HistogramOpts{
Name: "http_request_duration_seconds",
Help: "Latência das requisições HTTP em segundos",
Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
},
[]string{"method", "path", "status"},
)

HTTPRequests = promauto.NewCounterVec(
prometheus.CounterOpts{
Name: "http_requests_total",
Help: "Total de requisições HTTP",
},
[]string{"method", "path", "status"},
)
)

func Middleware() gin.HandlerFunc {
return func(c *gin.Context) {
start:= time.Now()
c.Next()

path:= c.FullPath()
if path == "" { path = "unknown" }
status:= strconv.Itoa(c.Writer.Status())

HTTPDuration.WithLabelValues(c.Request.Method, path, status).
Observe(time.Since(start).Seconds())
HTTPRequests.WithLabelValues(c.Request.Method, path, status).Inc()
}
}
```

---

## 15. Testes (testify + testcontainers-go)

**Unit test com mock (testify/mock):**

```go
package service_test

import (
"context"
"errors"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/mock"
"github.com/stretchr/testify/require"

"github.com/yourorg/project/internal/domain"
"github.com/yourorg/project/internal/service"
)

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
args:= m.Called(ctx, u)
if args.Get(0) == nil { return nil, args.Error(1) }
return args.Get(0).(*domain.User), args.Error(1)
}

func TestUserService_Create_Sucesso(t *testing.T) {
repo:= new(mockUserRepo)
svc:= service.NewUserService(repo)

input:= service.CreateUserInput{Name: "Ana", Email: "ana@example.com.br"}
expected:= &domain.User{Name: "Ana", Email: "ana@example.com.br"}

repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
Return(expected, nil)

user, err:= svc.Create(context.Background(), input)

require.NoError(t, err)
assert.Equal(t, expected.Email, user.Email)
repo.AssertExpectations(t)
}

func TestUserService_Create_RepoFalha(t *testing.T) {
repo:= new(mockUserRepo)
svc:= service.NewUserService(repo)

repo.On("Create", mock.Anything, mock.Anything).
Return(nil, errors.New("db down"))

_, err:= svc.Create(context.Background(), service.CreateUserInput{
Name: "Ana", Email: "ana@example.com.br",
})

require.Error(t, err)
assert.Contains(t, err.Error(), "db down")
}
```

**Integration test com testcontainers-go (Postgres):**

```go
func setupPostgres(t *testing.T) string {
t.Helper()
ctx:= context.Background()

container, err:= postgres.Run(ctx,
"postgres:16-alpine",
postgres.WithDatabase("test"),
postgres.WithUsername("test"),
postgres.WithPassword("test"),
testcontainers.WithWaitStrategy(
wait.ForLog("database system is ready to accept connections").
WithOccurrence(2).
WithStartupTimeout(30*time.Second),
),
)
require.NoError(t, err)
t.Cleanup(func() { _ = container.Terminate(ctx) })

dsn, err:= container.ConnectionString(ctx, "sslmode=disable")
require.NoError(t, err)
return dsn
}
```

**Table-driven tests (estilo idiomático Go):**

```go
func TestValidateEmail(t *testing.T) {
tests:= []struct {
name string
input string
wantErr bool
}{
{"válido", "user@example.com", false},
{"sem @", "userexample.com", true},
{"sem domínio", "user@", true},
{"vazio", "", true},
}

for _, tt:= range tests {
t.Run(tt.name, func(t *testing.T) {
err:= domain.ValidateEmail(tt.input)
if tt.wantErr {
assert.Error(t, err)
} else {
assert.NoError(t, err)
}
})
}
}
```

**Convenções:**
- Subtests com `t.Run` sempre que houver múltiplos cenários
- `require` para precondições (interrompe se falhar), `assert` para verificações (continua)
- `t.Cleanup` em vez de `defer` quando setup é compartilhado
- Naming: `Test<Tipo>_<Método>_<Cenário>` (ex: `TestUserService_Create_Sucesso`)
- Coverage alvo: 70%+ em `internal/service` e `internal/domain`, 60%+ em `internal/handler`, integração obrigatória em `internal/repository`

---

## Anti-patterns

### ❌ Ignorar erros com `_ = err`
**Problema:** O desenvolvedor descarta erros silenciosamente usando `_ = err` ou simplesmente não captura o retorno de erro.
**Por quê evitar:** Falhas silenciosas são a principal causa de bugs difíceis de rastrear em produção; o programa continua em estado inválido sem nenhum sinal de alerta.
**Solução:**
```go
// Errado
_ = r.q.CreateUser(ctx, sqlc.CreateUserParams{Name: u.Name})

// Correto
result, err:= r.q.CreateUser(ctx, sqlc.CreateUserParams{Name: u.Name})
if err!= nil {
return nil, fmt.Errorf("criar usuário: %w", err)
}
```

### ❌ Não passar `context.Context` como primeiro parâmetro
**Problema:** O desenvolvedor cria funções que fazem I/O (banco, HTTP, cache) sem aceitar `context.Context` como primeiro argumento.
**Por quê evitar:** Sem contexto não é possível propagar cancelamento, deadline ou valores de tracing; a operação não pode ser interrompida quando o cliente desconecta ou o timeout estoura.
**Solução:**
```go
// Errado
func (r *UserRepository) GetByID(id uuid.UUID) (*domain.User, error) {... }

// Correto
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {... }
```

### ❌ Usar `fmt.Errorf` sem `%w` (quebrando a cadeia de erros)
**Problema:** O desenvolvedor usa `fmt.Errorf("mensagem: %s", err)` em vez de `fmt.Errorf("mensagem: %w", err)` ao encapsular erros.
**Por quê evitar:** Sem `%w` o erro original é convertido em string e a cadeia é quebrada; `errors.Is` e `errors.As` param de funcionar, impossibilitando verificações como `errors.Is(err, pgx.ErrNoRows)`.
**Solução:**
```go
// Errado
return nil, fmt.Errorf("buscar usuário: %s", err)

// Correto
return nil, fmt.Errorf("buscar usuário: %w", err)
// Depois, no handler:
if errors.Is(err, pgx.ErrNoRows) {... }
```

### ❌ Concatenação de strings para construir queries (SQL injection)
**Problema:** O desenvolvedor monta queries SQL concatenando strings com valores vindos do usuário em vez de usar parâmetros do ORM ou placeholders.
**Por quê evitar:** Abre vetor direto de SQL injection; também impede o banco de reutilizar planos de execução (query plan cache).
**Solução:**
```go
// Errado
query:= "SELECT * FROM users WHERE email = '" + email + "'"
rows, _:= db.QueryContext(ctx, query)

// Correto — usando sqlc
result, err:= r.q.GetUserByEmail(ctx, email)
if err!= nil {
return nil, fmt.Errorf("buscar usuário por email: %w", err)
}
```

### ❌ Goroutine leak por falta de cancelamento via contexto
**Problema:** O desenvolvedor lança goroutines que ficam bloqueadas em canais ou operações sem observar `ctx.Done()`.
**Por quê evitar:** Goroutines vazadas consomem memória e file descriptors indefinidamente; em serviços de longa duração isso leva a OOM gradual.
**Solução:**
```go
// Errado
go func() {
result:= <-someChan // bloqueia para sempre se someChan nunca receber
process(result)
}()

// Correto
go func() {
select {
case <-ctx.Done():
return
case result:= <-someChan:
process(result)
}
}()
```

### ❌ Estado global e efeitos colaterais em `init()`
**Problema:** O desenvolvedor inicializa conexões de banco, clientes HTTP ou configurações dentro de funções `init()` ou em variáveis globais de nível de pacote.
**Por quê evitar:** Efeitos em `init()` são executados na importação do pacote, tornando testes difíceis (não há como injetar dependências) e ocultando falhas de inicialização.
**Solução:**
```go
// Errado
var pool *pgxpool.Pool

func init() {
var err error
pool, err = pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
if err!= nil { log.Fatal(err) }
}

// Correto — inicialização explícita via construtor
func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
return pgxpool.New(ctx, cfg.DatabaseURL)
}
```

### ❌ Retornar tipos concretos em vez de interfaces em construtores
**Problema:** O desenvolvedor expõe o tipo concreto de repositório ou serviço (`*UserRepository`) no lugar de uma interface no retorno do construtor.
**Por quê evitar:** Acopla quem chama ao tipo concreto; impede substituição por mock em testes sem alterar assinaturas de funções dependentes.
**Solução:**
```go
// Errado
func NewUserRepository(c *pgxpool.Pool) *UserRepository {... }

// Correto
type UserRepositoryIface interface {
Create(ctx context.Context, u *domain.User) (*domain.User, error)
GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

func NewUserRepository(q sqlc.Querier) UserRepositoryIface {
return &UserRepository{q: q}
}
```

### ❌ Não usar table-driven tests
**Problema:** O desenvolvedor escreve uma função de teste separada para cada caso (`TestCreate_Valid`, `TestCreate_EmptyName`, `TestCreate_DuplicateEmail`...) em vez de um único table-driven test.
**Por quê evitar:** Duplicação de código de setup e teardown; adicionar novos casos exige criar novas funções; dificulta ver de relance quais cenários estão cobertos.
**Solução:**
```go
// Correto
func TestUserService_Create(t *testing.T) {
tests:= []struct {
name string
input service.CreateUserInput
wantErr bool
}{
{"sucesso", service.CreateUserInput{Name: "Ana", Email: "ana@example.com.br"}, false},
{"email vazio", service.CreateUserInput{Name: "Ana", Email: ""}, true},
{"nome vazio", service.CreateUserInput{Name: "", Email: "x@example.com.br"}, true},
}
for _, tt:= range tests {
t.Run(tt.name, func(t *testing.T) {
_, err:= svc.Create(context.Background(), tt.input)
if tt.wantErr {
require.Error(t, err)
} else {
require.NoError(t, err)
}
})
}
}
```

### ❌ Bloquear em handler Gin sem propagar o contexto da request
**Problema:** O desenvolvedor chama operações de I/O dentro do handler Gin usando `context.Background()` em vez de `c.Request.Context()`.
**Por quê evitar:** Quando o cliente cancela a requisição (ex: browser fecha a aba), o contexto da request é cancelado; usar `context.Background()` faz a operação continuar consumindo recursos desnecessariamente.
**Solução:**
```go
// Errado
func (h *UserHandler) Create(c *gin.Context) {
user, err:= h.svc.Create(context.Background(), input)
...
}

// Correto
func (h *UserHandler) Create(c *gin.Context) {
user, err:= h.svc.Create(c.Request.Context(), input)
...
}
```

### ❌ N+1 queries por falta de JOIN
**Problema:** O desenvolvedor faz query de lista e dentro do loop faz outra query para dados relacionados.
**Por quê evitar:** Em listas de 100 usuários com pedidos pendentes isso gera 101 queries ao banco (1 lista + 1 por usuário).
**Solução:**
```go
// Errado — N+1 queries
users, _:= r.q.ListActiveUsers(ctx)
for _, u:= range users {
orders, _:= r.q.ListOrdersByUser(ctx, u.ID) // 1 query por iteração
process(orders)
}

// Correto — JOIN único
usersWithOrders, err:= r.q.ListActiveUsersWithOrders(ctx)
if err!= nil {
return nil, fmt.Errorf("listar usuários com pedidos: %w", err)
}
for _, u:= range usersWithOrders {
process(u.Orders)
}
```

---

## 16. Qualidade / SonarQube

Limiares alinhados ao Quality Gate e ao `.golangci.yml` canônico (`project-scaffold.md`). Prevenir reincidência — não “apagar o sintoma” no Sonar.

| Regra | Limiar | Sonar | Linter local |
|---|---|---|---|
| Tamanho de função | ≤30 linhas | `go:S138` | `funlen` |
| Parâmetros | ≤3 (inclui `context.Context`) | `go:S107` | revive `argument-limit: 3` |
| Complexidade cognitiva | ≤15 | `go:S3776` | `gocognit` (min-complexity: 16) |
| TODO/FIXME | só com issue (`// TODO ABC-123:...`) | `go:S1135` | review / checklist |
| Coverage | ≥80% linhas | Quality Gate | `go test -coverprofile=coverage.out` |

### Options / Params struct (S107)

Mesmo padrão do sqlc (`query_parameter_limit: 3`): acima de 3 params → uma struct.

```go
// ❌ go:S107 — 5 parâmetros
func CreateUser(ctx context.Context, name, email, status string, roleID int64) error

// ✅ ctx + Options (2 parâmetros)
type CreateUserOptions struct {
Name string
Email string
Status string
RoleID int64
}

func CreateUser(ctx context.Context, opts CreateUserOptions) error
```

### Extrair funções (S138 / S3776)

- Função >30 linhas ou cognitive >15 → extrair helpers privados com 1 responsabilidade (early return, sem nesting profundo).
- Preferir tabelas/`switch` claros a cadeias longas de `if/else`.

### Coverage para Sonar (Go)

```bash
go test./... -race -coverprofile=coverage.out -covermode=atomic
# sonar-project.properties: sonar.go.coverage.reportPaths=coverage.out
```

Sem `*_test.go` relevantes → coverage 0% e o gate falha. Ver `testing-expert` + `cicd-expert`.
