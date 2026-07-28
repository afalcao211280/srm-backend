# Guia de execução dos testes

## Backend

| Nível | Comando | Cobertura | Dependência |
|---|---|---|---|
| Unitário | `go test ./...` | strategies, fórmula, arredondamento, guards, exception handler | nenhuma |
| Integração | `go test -tags=integration ./...` | repositórios, optimistic locking, atomicidade, constraints, SQL injection | PostgreSQL real (`TEST_DATABASE_URL`) |
| Cobertura | `go test ./... -race -coverprofile=coverage.out -covermode=atomic` | ≥ 80% (gate Sonar) | nenhuma |

### Em container

```bash
docker compose up -d db
docker compose run --rm api go test ./...
docker compose run --rm -e TEST_DATABASE_URL=postgres://srm:srm@db:5432/srm_test api go test -tags=integration ./...
```

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
