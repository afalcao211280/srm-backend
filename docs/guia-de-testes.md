# Guia de execução dos testes

## Backend

| Nível | Comando | Cobertura | Dependência |
|---|---|---|---|
| Unitário | `go test ./...` | strategies, fórmula, arredondamento, guards, exception handler, validação Huma | nenhuma |
| Integração | `go test -tags=integration ./...` | repositórios, optimistic locking, atomicidade, constraints, SQL injection, circuit breaker, round-trip HTTP completo, plano de execução do extrato | Docker (sobe o próprio Postgres via testcontainers-go — não precisa de um banco pré-existente) |
| Cobertura | `go test ./... -race -coverprofile=coverage.out -covermode=atomic` | ≥ 80% (gate Sonar) | nenhuma |

Os testes de integração usam [testcontainers-go](https://golang.testcontainers.org/): cada pacote sobe seu próprio container PostgreSQL descartável, aplica as migrations e roda os testes contra ele — não é preciso `docker compose up -d db` antes, nem configurar `TEST_DATABASE_URL`. A única dependência é o Docker do host estar acessível.

### Em container

Os alvos `test-unit`/`test-integration` do compose usam o estágio `builder` do `Dockerfile` (com toolchain Go) — a imagem final do serviço `api` é distroless e não tem `go` instalado.

```bash
docker compose --profile test run --rm test-unit
docker compose --profile test run --rm test-integration
```

`test-integration` monta o socket do Docker do host (`/var/run/docker.sock`) para o testcontainers-go, rodando *dentro* do container de teste, conseguir subir o Postgres descartável (padrão "Docker outside of Docker").

## Frontend

| Nível | Comando | Cobertura | Dependência |
|---|---|---|---|
| Componente | `bun run test` | comportamento do painel, debounce, validação | jsdom |
| E2E | `bun run test:e2e` | fluxo completo, navegação entre páginas | stack em execução |

### Em container

```bash
docker compose run --rm frontend bun run test
docker compose run --rm frontend bun run test:e2e
```

## Massa de demonstração

```bash
docker compose --profile carga run --rm carga
```

Popula: 8 cedentes, 2 moedas (BRL, USD), 2 tipos de recebível, taxas base em duas vigências para BRL e uma para USD, cotações USD→BRL em duas vigências, 50 transações calculadas pelo próprio motor distribuídas entre os dois tipos de recebível e com mistura de status (PENDENTE, LIQUIDADA, CANCELADA).
