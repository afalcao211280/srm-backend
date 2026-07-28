# Core — Principios, Patterns e Anti-padroes Go

Leia SEMPRE antes de gerar codigo Go.

## Principios Detalhados

1. **Stack canonica** — nao inventar libs. A stack e definida, siga.
2. **Clean Architecture** — handler → service → repository → domain. Interfaces pelo consumidor.
3. **`context.Context`** — primeiro parametro, sempre. Sem excecao.
4. **Erros wrapped** — `fmt.Errorf("contexto: %w", err)`. Nunca `_ = err`.
5. **slog** — JSON estruturado. correlation ID no contexto.
6. **Observabilidade nativa** — span no service, metricas no middleware, trace_id no log.
7. **Testes junto** — `*_test.go` no mesmo pacote. testcontainers pra integracao.

> Codigo pronto pra commit. Zero exemplo educativo.

## Regras de Importacao Inviolaveis

- `cmd/` nunca e importado. So importa `internal/` e `pkg/`.
- `internal/` importa `pkg/`, nunca o contrario.
- `pkg/` zero deps de `internal/`.
- `domain/` zero deps externas (so stdlib).

## Multi-Service: Multiplos Mains

Monolito com opcao de expandir via env var `MULTI_SERVICE=1`.

Cada entrypoint (`cmd/server/main.go`, `cmd/worker/main.go`, `cmd/scheduler/main.go`) tem seu proprio constructor `New*(ctx, cfg, log)` em `internal/server/`.

**Ver `reference/multi-service.md` para exemplos completos de DI manual em multiplos mains.**

## Workflow Agentic Detalhado

1. **Entender escopo** — servico novo ou existente? HTTP, worker, CLI?
2. **Planejar arquivos** (de dentro pra fora):
   - `internal/domain/<entity>.go`, `internal/model/dto.go`
   - `internal/database/migrations/NNN_<nome>.sql` (tern)
   - `internal/database/queries/<entity>.sql` (sqlc)
   - `internal/repository/<entity>_repository.go`
   - `internal/service/<entity>_service.go` (interface do repository aqui)
   - `internal/handler/<entity>_handler.go`
   - `internal/handler/<entity>_handler_test.go`
   - `internal/service/<entity>_service_test.go`
   - Atualizar `internal/server/server.go` (DI manual)
3. **Gerar completo** — imports, errors, logs, spans, validacao
4. **Testar** — unit (testify mock) + integracao (testcontainers)
5. **Apresentar diff** — "criei/editei: A, B, C; decisao: sqlc pq ja tem sqlc.yaml"

## Checklist Completo

- [ ] `gofmt` / imports stdlib → externos → internos
- [ ] `context.Context` primeiro parametro em I/O
- [ ] Erros wrapped com `%w`
- [ ] Logs com `slog` via contexto (nunca `fmt.Println`)
- [ ] Interfaces no consumidor (service declara, repository implementa)
- [ ] Testes: 1 feliz + 1 erro validacao + 1 erro repo
- [ ] Tags JSON/BSON em snake_case
- [ ] Span OTel no service
- [ ] Zero globais de estado
- [ ] Imports ciclicos: handler → service → repo → domain (nunca contrario)

## Quando Perguntar

Antes de codar: banco? fila? auth? JWT? OAuth?
Depois da 1a versao: quer refinar algum ponto?

## Anti-padroes Comuns

### ❌ Ignorar erros com `_ = err`
**Problema:** Falhas silenciosas, programa continua em estado invalido.
**Solucao:** Sempre tratar ou propagar.

### ❌ Nao passar `context.Context` como primeiro parametro
**Problema:** Sem propagacao de cancelamento/deadline/tracing.
**Solucao:** `func (r *Repo) GetByID(ctx context.Context, id uuid.UUID)`

### ❌ Usar `fmt.Errorf` sem `%w`
**Problema:** Quebra cadeia de erros; `errors.Is`/`errors.As` param.
**Solucao:** `fmt.Errorf("buscar: %w", err)`

### ❌ Concatenacao de strings para SQL
**Problema:** SQL injection; impede query plan cache.
**Solucao:** Usar sqlc ou parametros `$1, $2`.

### ❌ Goroutine leak por falta de cancelamento
**Problema:** Goroutines bloqueadas consomem memoria indefinidamente.
**Solucao:** Observar `ctx.Done()` em select.

### ❌ Estado global em `init()`
**Problema:** Testes dificeis; falhas ocultas.
**Solucao:** Constructors explicitos `New*()`.

### ❌ N+1 queries
**Problema:** Loop com query DB por iteracao.
**Solucao:** JOIN unico ou batch query.

### ❌ Bloquear em handler sem propagar contexto da request
**Problema:** `context.Background()` nao respeita cancelamento do cliente.
**Solucao:** `c.Request.Context()` no Gin.
