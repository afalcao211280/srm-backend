# Testing Expert — Padrões de Código

Snippets prontos por stack. Princípios em `core.md`; BDD em `bdd.md`.

## Índice

1. Vitest — unit (Frontend/Node)
2. pytest — fixtures e parametrize (Python)
3. Go — table-driven + testify
4. Java — JUnit 5 + Mockito
5. Testcontainers — integração (Node/TS)
6. MSW — mock HTTP (Frontend)
7. Playwright — E2E
8. k6 — load test
9. Anti-patterns (com solução)

## 1. Vitest — Unit Test com Setup

```typescript
// src/services/user.service.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { UserService } from './user.service';
import type { UserRepository } from '@/repositories/user.repository';

// Mock factory — evita vazamento de estado entre testes
function createMockRepo(): UserRepository {
return {
findById: vi.fn(),
findByEmail: vi.fn(),
existsByEmail: vi.fn(),
create: vi.fn(),
findPaginated: vi.fn(),
updateById: vi.fn(),
deleteById: vi.fn(),
} as unknown as UserRepository;
}

describe('UserService', () => {
let service: UserService;
let mockRepo: ReturnType<typeof createMockRepo>;

beforeEach(() => {
mockRepo = createMockRepo();
service = new UserService(mockRepo);
});

describe('create', () => {
it('cria usuário quando email não existe', async () => {
vi.mocked(mockRepo.existsByEmail).mockResolvedValue(false);
vi.mocked(mockRepo.create).mockResolvedValue({
id: 'uuid-1',
email: 'joao@example.com',
name: 'João Silva',
role: 'viewer',
active: true,
createdAt: new Date(),
updatedAt: new Date(),
});

const result = await service.create({
email: 'joao@example.com',
name: 'João Silva',
password: 'senha123',
});

expect(result.email).toBe('joao@example.com');
expect(mockRepo.create).toHaveBeenCalledOnce();
});

it('lança ConflictError quando email já existe', async () => {
vi.mocked(mockRepo.existsByEmail).mockResolvedValue(true);

await expect(
service.create({ email: 'dup@example.com', name: 'Dup', password: 'senha123' }),
).rejects.toThrow(ConflictError);
});

it('não armazena senha em texto plano', async () => {
vi.mocked(mockRepo.existsByEmail).mockResolvedValue(false);
vi.mocked(mockRepo.create).mockResolvedValue({ id: '1' } as any);

await service.create({ email: 'a@b.com', name: 'A', password: 'mypassword' });

const callArgs = vi.mocked(mockRepo.create).mock.calls[0]?.[0];
expect(callArgs).not.toHaveProperty('password', 'mypassword');
expect(callArgs).toHaveProperty('passwordHash');
});
});
});
```

---

## 2. pytest — Fixtures e Parametrize

```python
# tests/conftest.py
import pytest
import pytest_asyncio
from httpx import AsyncClient, ASGITransport
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession

from src.main import app
from src.db import Base

@pytest_asyncio.fixture(scope="session")
async def engine():
"""Usar testcontainers — ver setup em conftest do projeto."""
engine = create_async_engine(
"postgresql+asyncpg://test:test@localhost/test_db",
echo=False,
)
async with engine.begin() as conn:
await conn.run_sync(Base.metadata.create_all)
yield engine
async with engine.begin() as conn:
await conn.run_sync(Base.metadata.drop_all)
await engine.dispose()

@pytest_asyncio.fixture
async def db_session(engine):
async with AsyncSession(engine) as session:
yield session
await session.rollback()

@pytest_asyncio.fixture
async def client():
async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as c:
yield c

# tests/test_user_service.py
import pytest
from src.services.user_service import UserService, EmailAlreadyExistsError

@pytest.mark.parametrize("email,name,should_raise", [
("valid@example.com", "Valid User", False),
("UPPER@EXAMPLE.COM", "Upper User", False),
("", "Empty Email", True),
("not-an-email", "Invalid Email", True),
])
async def test_create_user_validation(email, name, should_raise, db_session):
service = UserService(db_session)
if should_raise:
with pytest.raises(ValueError):
await service.create(email=email, name=name, password="senha123")
else:
user = await service.create(email=email, name=name, password="senha123")
assert user.email == email.lower()
```

---

## 3. Go — Table-Driven + testify

**Coverage para Sonar/CI (≥80%):**

```bash
go test./... -race -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out | tail -1 # total local
# Pipeline: sonar.go.coverage.reportPaths=coverage.out (cicd-expert)
```

```go
// internal/service/user_service_test.go
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mock do repository (interface declarada no service)
type mockRepo struct{ mock.Mock }

func (m *mockRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args:= m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}
func (m *mockRepo) Create(ctx context.Context, u User) (User, error) {
	args:= m.Called(ctx, u)
	return args.Get(0).(User), args.Error(1)
}

func TestUserService_Create(t *testing.T) {
	t.Parallel()
	tests:= []struct {
		name string
		email string
		existsRet bool
		existsErr error
		createErr error
		wantErr error
	}{
		{name: "cria quando email não existe", email: "joao@example.com"},
		{name: "rejeita email duplicado", email: "dup@example.com", existsRet: true, wantErr: ErrEmailExists},
		{name: "propaga erro do repo", email: "x@example.com", existsErr: errors.New("db down"), wantErr: ErrEmailExists},
	}
	for _, tt:= range tests {
		tt:= tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo:= new(mockRepo)
			repo.On("ExistsByEmail", mock.Anything, tt.email).Return(tt.existsRet, tt.existsErr)
			repo.On("Create", mock.Anything, mock.Anything).Return(User{ID: "1", Email: tt.email}, tt.createErr)

			svc:= NewUserService(repo)
			got, err:= svc.Create(context.Background(), tt.email, "senha123")

			if tt.wantErr!= nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.email, got.Email)
		})
	}
}
```

---

## 4. Java — JUnit 5 + Mockito

```java
// src/test/java/com/example/user/UserServiceTest.java
package com.example.user;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.ValueSource;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class UserServiceTest {

@Mock UserRepository repo;
@InjectMocks UserService service;

@Test
void createsUser_whenEmailDoesNotExist() {
when(repo.existsByEmail("joao@example.com")).thenReturn(false);
when(repo.save(any())).thenAnswer(inv -> inv.getArgument(0));

var user = service.create("joao@example.com", "senha123");

assertThat(user.email()).isEqualTo("joao@example.com");
verify(repo).save(any());
}

@Test
void throwsConflict_whenEmailExists() {
when(repo.existsByEmail("dup@example.com")).thenReturn(true);

assertThatThrownBy(() -> service.create("dup@example.com", "senha123"))
.isInstanceOf(EmailAlreadyExistsException.class);
}

@ParameterizedTest
@ValueSource(strings = {"", " ", "not-an-email"})
void rejectsInvalidEmail(String email) {
assertThatThrownBy(() -> service.create(email, "senha123"))
.isInstanceOf(IllegalArgumentException.class);
}
}
```

---

## 5. Testcontainers — Integration Test (Node/TS)

```typescript
// tests/integration/user.repository.test.ts
import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest';
import { PostgreSqlContainer, type StartedPostgreSqlContainer } from '@testcontainers/postgresql';
import { PrismaClient } from '@prisma/client';
import { execSync } from 'child_process';
import { UserRepository } from '@/repositories/user.repository';

let container: StartedPostgreSqlContainer;
let prisma: PrismaClient;
let repo: UserRepository;

beforeAll(async () => {
container = await new PostgreSqlContainer('postgres:16-alpine').start();
const connectionUri = container.getConnectionString();

execSync('npx prisma migrate deploy', {
env: {...process.env, DATABASE_URL: connectionUri },
});

prisma = new PrismaClient({ datasources: { db: { url: connectionUri } } });
repo = new UserRepository(prisma);
}, 60_000);

afterAll(async () => {
await prisma.$disconnect();
await container?.stop();
});

afterEach(async () => {
await prisma.user.deleteMany();
});

describe('UserRepository', () => {
it('cria e retorna usuário', async () => {
const user = await repo.create({ email: 'test@example.com', name: 'Test User', passwordHash: 'hash' });
expect(user.id).toBeDefined();
expect(user.email).toBe('test@example.com');
});

it('findByEmail é case-insensitive', async () => {
await repo.create({ email: 'test@example.com', name: 'Test', passwordHash: 'hash' });
const found = await repo.findByEmail('TEST@.COM');
expect(found).not.toBeNull();
});
});
```

---

## 6. MSW — Mock Service Worker (Frontend)

```typescript
// src/mocks/handlers.ts
import { http, HttpResponse } from 'msw';

export const handlers = [
http.get('/api/v1/users', () => {
return HttpResponse.json([
{ id: '1', name: 'João Silva', email: 'joao@example.com', active: true },
{ id: '2', name: 'Maria Santos', email: 'maria@example.com', active: true },
]);
}),

http.post('/api/v1/users', async ({ request }) => {
const body = await request.json() as { name: string; email: string };
return HttpResponse.json({ id: '3',...body }, { status: 201 });
}),

http.post('/api/v1/users', () => {
return HttpResponse.json({ error: 'Email already exists' }, { status: 409 });
}, { once: true }), // override apenas na primeira chamada
];

// src/mocks/server.ts (Node)
import { setupServer } from 'msw/node';
export const server = setupServer(...handlers);

// vitest.setup.ts
import { beforeAll, afterAll, afterEach } from 'vitest';
import { server } from './src/mocks/server';

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());
```

---

## 7. Playwright — E2E Tests

```typescript
// e2e/users.spec.ts
import { test, expect } from '@playwright/test';

test.describe('User management', () => {
test.beforeEach(async ({ page }) => {
// Setup: fazer login via API diretamente (evitar testar login em todo teste)
const response = await page.request.post('/api/auth/login', {
data: { email: 'admin@example.com', password: process.env['E2E_PASSWORD'] },
});
const { token } = await response.json();
await page.addInitScript((t) => {
localStorage.setItem('auth-token', t);
}, token);
});

test('cria novo usuário e aparece na lista', async ({ page }) => {
await page.goto('/dashboard/users');
await page.getByRole('button', { name: 'Novo Usuário' }).click();

await page.getByLabel('Nome').fill('Novo Usuário Teste');
await page.getByLabel('Email').fill('novo@example.com');
await page.getByLabel('Senha').fill('senha-segura-123');
await page.getByRole('button', { name: 'Criar Usuário' }).click();

await expect(page.getByText('Usuário criado com sucesso')).toBeVisible();
await expect(page.getByText('Novo Usuário Teste')).toBeVisible();
});

test('exibe erro quando email já existe', async ({ page }) => {
await page.goto('/dashboard/users/new');
await page.getByLabel('Email').fill('admin@example.com');
await page.getByRole('button', { name: 'Criar Usuário' }).click();

await expect(page.getByText('Email já cadastrado')).toBeVisible();
});
});
```

---

## 8. k6 — Load Test

```javascript
// tests/load/users-api.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const listDuration = new Trend('list_users_duration', true);

export const options = {
stages: [
{ duration: '1m', target: 10 }, // Ramp up
{ duration: '3m', target: 50 }, // Sustain
{ duration: '1m', target: 0 }, // Ramp down
],
thresholds: {
'http_req_duration': ['p(95)<500'], // 95% das requests < 500ms
'errors': ['rate<0.01'], // < 1% de erros
'http_req_failed': ['rate<0.01'],
},
};

export default function () {
const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';
const TOKEN = __ENV.API_TOKEN;

const headers = { Authorization: `Bearer ${TOKEN}` };

const listRes = http.get(`${BASE_URL}/api/v1/users`, { headers });
listDuration.add(listRes.timings.duration);
errorRate.add(listRes.status!== 200);

check(listRes, {
'list users status 200': (r) => r.status === 200,
'list users response time < 500ms': (r) => r.timings.duration < 500,
'list users has items': (r) => JSON.parse(r.body).length > 0,
});

sleep(1);
}
```

---

## 9. Anti-patterns

### ❌ Mockar o banco de dados em testes de integração
**Problema:** Substituir o banco real por um mock (objetos em memória) em testes de repositório/integração.
**Por quê evitar:** Testes passam mas produção falha — índices, constraints, tipos de dado, comportamento de transações e N+1 queries só aparecem com banco real.
**Solução:**
```typescript
// Use testcontainers — banco real, isolado, descartável
const pg = await new PostgreSqlContainer('postgres:16-alpine').start();
const dataSource = await AppDataSource.initialize({ url: pg.getConnectionUri() });
afterAll(() => pg.stop());
```

### ❌ Testar implementação em vez de comportamento
**Problema:** Assertar que um método privado foi chamado, ou verificar detalhes internos de implementação no teste.
**Por quê evitar:** Tests quebram em refactors que não mudam comportamento externo — cria fricção falsa ao melhorar código.
**Solução:**
```typescript
// ❌ Implementação: verifica que saveUser foi chamado
expect(mockRepo.saveUser).toHaveBeenCalledWith(expect.any(Object));

// ✅ Comportamento: verifica o resultado observável
const user = await userService.create({ email: 'x@y.com', name: 'X' });
expect(user.id).toBeDefined();
expect(user.email).toBe('x@y.com');
```

### ❌ Estado compartilhado entre testes
**Problema:** Modificar variáveis globais ou dados de banco sem limpar entre testes.
**Por quê evitar:** A ordem de execução dos testes determina o resultado — testes que passam sozinhos falham em conjunto (test order dependency).
**Solução:**
```typescript
beforeEach(async () => {
await db.user.deleteMany({}); // estado limpo para cada teste
});
// Ou use transactions com rollback automático (mais rápido)
beforeEach(() => { trx = await db.transaction(); });
afterEach(() => trx.rollback());
```

### ❌ Snapshot tests para objetos grandes e voláteis
**Problema:** Usar `expect(response).toMatchSnapshot()` para objetos com muitos campos ou timestamps.
**Por quê evitar:** Qualquer adição de campo quebra o snapshot — desenvolvedores atualizam snapshots sem ler o diff, snapshots viram rubber stamps sem valor real.
**Solução:**
```typescript
// Ao invés de snapshot completo, assertar campos que importam
expect(response.status).toBe(201);
expect(response.data.id).toMatch(/^[0-9a-f-]{36}$/);
expect(response.data.email).toBe('user@example.com');
// Reserve snapshots para estruturas estáveis (ASTs, HTML renderizado)
```

### ❌ Não testar caminhos de erro (só happy path)
**Problema:** Escrever testes apenas para o fluxo de sucesso.
**Por quê evitar:** Metade do comportamento real — tratamento de erros, edge cases, inputs inválidos — fica sem cobertura; bugs em produção vêm exatamente dessas situações.
**Solução:**
```typescript
describe('UserService.create', () => {
it('cria usuário com dados válidos',...) // happy path
it('rejeita email duplicado',...) // unicidade
it('rejeita email sem formato válido',...) // validação
it('propaga erro do banco para o chamador',...) // falha de infra
it('rejeita nome vazio ou só espaços',...) // edge case
});
```

### ❌ setTimeout em testes em vez de fake timers
**Problema:** Usar `await new Promise(r => setTimeout(r, 500))` em testes para esperar operações assíncronas.
**Por quê evitar:** Torna testes lentos e flaky em CI (timeouts variam com carga do servidor) — 500ms × N testes = minutos de espera desnecessária.
**Solução:**
```typescript
// Vitest / Jest: use fake timers
beforeEach(() => { vi.useFakeTimers(); });
afterEach(() => { vi.useRealTimers(); });

it('dispara callback após debounce', async () => {
const fn = vi.fn();
debouncedFn(fn);
vi.advanceTimersByTime(300); // instantâneo
expect(fn).toHaveBeenCalledOnce();
});
```

### ❌ E2E para lógica testável em unit/integration
**Problema:** Escrever testes Playwright/Cypress para validar lógica de negócio que poderia ser testada unitariamente.
**Por quê evitar:** E2E testes são 10-100x mais lentos, mais frágeis (dependem de browser, rede, dados) e mais difíceis de depurar — use na pirâmide: 70% unit, 20% integration, 10% E2E.
**Solução:**
```
Unit: cálculo de desconto, validação de CPF, formatação
Integration: API endpoint + banco, service + repositório
E2E: fluxo crítico de negócio (login → checkout → confirmação)
```
