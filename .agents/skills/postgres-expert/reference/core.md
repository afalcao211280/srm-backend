# PostgreSQL Expert — Padrões de Código

## 1. Schema Design — Convenções

```sql
-- Sempre usar UUID v7 (ordenável por tempo) ou BIGSERIAL para IDs
-- Prefira UUID quando IDs são expostos externamente

CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT NOT NULL,
    name        TEXT NOT NULL CHECK (char_length(name) BETWEEN 2 AND 100),
    password_hash TEXT NOT NULL,
    role        TEXT NOT NULL DEFAULT 'viewer' CHECK (role IN ('admin', 'editor', 'viewer')),
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Índice único em email (case-insensitive)
CREATE UNIQUE INDEX users_email_lower_idx ON users (lower(email));

-- Índice parcial — somente usuários ativos
CREATE INDEX users_active_created_idx ON users (created_at DESC)
    WHERE active = true;

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
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

---

## 2. Índices — Estratégia

```sql
-- B-tree padrão para igualdade e range
CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_orders_created_at ON orders (created_at DESC);

-- Índice composto — column order importa: coluna de maior seletividade primeiro
CREATE INDEX idx_orders_status_user ON orders (status, user_id)
    WHERE status != 'completed'; -- Partial index reduz tamanho

-- GIN para arrays e JSONB
CREATE INDEX idx_products_tags ON products USING GIN (tags);
CREATE INDEX idx_events_payload ON events USING GIN (payload jsonb_path_ops);

-- Text search
CREATE INDEX idx_products_name_tsv ON products
    USING GIN (to_tsvector('portuguese', name));

-- Verificar uso de índices
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
WHERE idx_scan = 0  -- índices nunca usados
ORDER BY pg_relation_size(indexrelid) DESC;
```

---

## 3. Query Patterns

```sql
-- Paginação com cursor (mais eficiente que OFFSET para grandes datasets)
-- Primeira página
SELECT id, name, email, created_at
FROM users
WHERE active = true
ORDER BY created_at DESC, id DESC
LIMIT 20;

-- Página seguinte (cursor = último registro)
SELECT id, name, email, created_at
FROM users
WHERE active = true
  AND (created_at, id) < ($last_created_at, $last_id)
ORDER BY created_at DESC, id DESC
LIMIT 20;

-- CTE para queries complexas (legibilidade > micro-otimização)
WITH active_users AS (
    SELECT id, name, email
    FROM users
    WHERE active = true AND created_at > NOW() - INTERVAL '30 days'
),
user_order_counts AS (
    SELECT user_id, COUNT(*) AS total_orders
    FROM orders
    WHERE status = 'completed'
    GROUP BY user_id
)
SELECT u.name, u.email, COALESCE(o.total_orders, 0) AS orders
FROM active_users u
LEFT JOIN user_order_counts o ON o.user_id = u.id
ORDER BY orders DESC;

-- UPSERT
INSERT INTO user_settings (user_id, key, value)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, key)
DO UPDATE SET value = EXCLUDED.value, updated_at = NOW();
```

---

## 4. EXPLAIN ANALYZE — Interpretação

```sql
-- Sempre usar EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT * FROM orders
WHERE user_id = '123e4567-e89b-12d3-a456-426614174000'
  AND status = 'pending'
ORDER BY created_at DESC
LIMIT 10;

-- Sinais de problema:
-- Seq Scan em tabelas grandes → falta índice
-- Nested Loop com grandes row estimates → estatísticas desatualizadas → ANALYZE
-- Hash Join spilling to disk (Batches > 1) → aumentar work_mem
-- cost=X..Y: X = startup cost, Y = total cost
-- actual time=X..Y rows=Z loops=W: W > 1 = chamado N vezes (nested)

-- Forçar atualização de estatísticas
ANALYZE users;
-- Reindex após muitos updates/deletes
REINDEX INDEX CONCURRENTLY users_email_lower_idx;
```

---

## 5. Prisma Schema (Node/TypeScript)

```prisma
// prisma/schema.prisma
generator client {
  provider = "prisma-client-js"
  previewFeatures = ["postgresqlExtensions"]
}

datasource db {
  provider   = "postgresql"
  url        = env("DATABASE_URL")
  extensions = [uuidOssp(map: "uuid-ossp"), pgcrypto]
}

model User {
  id           String    @id @default(dbgenerated("gen_random_uuid()")) @db.Uuid
  email        String    @unique
  name         String    @db.VarChar(100)
  passwordHash String    @map("password_hash")
  role         UserRole  @default(VIEWER)
  active       Boolean   @default(true)
  createdAt    DateTime  @default(now()) @map("created_at") @db.Timestamptz
  updatedAt    DateTime  @updatedAt @map("updated_at") @db.Timestamptz
  orders       Order[]

  @@index([createdAt(sort: Desc)], map: "users_created_at_idx")
  @@map("users")
}

enum UserRole {
  ADMIN
  EDITOR
  VIEWER

  @@map("user_role")
}
```

---

## 6. Transações com Retry (Node/TypeScript)

```typescript
// src/lib/db.ts — helper de transação com retry
import { db } from './prisma';
import type { Prisma } from '@prisma/client';

const MAX_RETRIES = 3;

export async function withTransaction<T>(
  fn: (tx: Prisma.TransactionClient) => Promise<T>,
): Promise<T> {
  let attempts = 0;
  while (attempts < MAX_RETRIES) {
    try {
      return await db.$transaction(fn, {
        isolationLevel: 'ReadCommitted',
        maxWait: 5000,
        timeout: 10000,
      });
    } catch (err) {
      // Retry apenas em serialization failures e deadlocks
      if (
        err instanceof Error &&
        (err.message.includes('could not serialize') || err.message.includes('deadlock detected'))
      ) {
        attempts++;
        if (attempts >= MAX_RETRIES) throw err;
        await new Promise((r) => setTimeout(r, attempts * 100));
        continue;
      }
      throw err;
    }
  }
  throw new Error('unreachable');
}
```

---

## 7. Testcontainers (TypeScript)

```typescript
// src/tests/setup.ts
import { PostgreSqlContainer, type StartedPostgreSqlContainer } from '@testcontainers/postgresql';
import { execSync } from 'child_process';

let container: StartedPostgreSqlContainer;

beforeAll(async () => {
  container = await new PostgreSqlContainer('postgres:17-alpine')
    .withDatabase('test_db')
    .withUsername('test')
    .withPassword('test')
    .start();

  process.env['DATABASE_URL'] = container.getConnectionUri();

  // Rodar migrations
  execSync('npx prisma migrate deploy', {
    env: { ...process.env, DATABASE_URL: container.getConnectionUri() },
  });
}, 60_000);

afterAll(async () => {
  await container?.stop();
});
```

---

## 8. Queries de Monitoramento

```sql
-- Top queries por tempo total
SELECT query, calls, total_exec_time::int AS total_ms,
       (total_exec_time / calls)::int AS avg_ms,
       rows
FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 20;

-- Locks ativos
SELECT pid, now() - pg_stat_activity.query_start AS duration,
       query, state, wait_event_type, wait_event
FROM pg_stat_activity
WHERE (now() - pg_stat_activity.query_start) > INTERVAL '5 seconds'
  AND state != 'idle';

-- Bloat de tabelas (tabelas com muitos dead tuples)
SELECT relname, n_dead_tup, n_live_tup,
       round(n_dead_tup * 100.0 / NULLIF(n_live_tup + n_dead_tup, 0), 2) AS dead_pct,
       last_autovacuum
FROM pg_stat_user_tables
WHERE n_dead_tup > 1000
ORDER BY n_dead_tup DESC;

-- Tamanho de tabelas e índices
SELECT tablename,
       pg_size_pretty(pg_total_relation_size(tablename::regclass)) AS total,
       pg_size_pretty(pg_relation_size(tablename::regclass)) AS table,
       pg_size_pretty(pg_indexes_size(tablename::regclass)) AS indexes
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(tablename::regclass) DESC;
```

---

## Anti-patterns

### ❌ LIKE com wildcard inicial
**Problema:** O desenvolvedor usa `LIKE '%termo'` ou `ILIKE '%termo'` para buscas textuais.
**Por quê evitar:** O B-tree index não pode ser usado quando o wildcard está no início do padrão, forçando um sequential scan completo na tabela independente do tamanho.
**Solução:**
```sql
-- Errado
SELECT * FROM products WHERE name LIKE '%laptop%';

-- Correto: full-text search com índice GIN
CREATE INDEX idx_products_name_tsv ON products
    USING GIN (to_tsvector('portuguese', name));

SELECT * FROM products
WHERE to_tsvector('portuguese', name) @@ plainto_tsquery('portuguese', 'laptop');

-- Ou para prefix search (wildcard só no final), B-tree funciona:
SELECT * FROM products WHERE name LIKE 'laptop%';
```

---

### ❌ SELECT * em queries de produção
**Problema:** O desenvolvedor usa `SELECT *` em vez de listar as colunas explicitamente.
**Por quê evitar:** Transfere dados desnecessários pela rede e para a aplicação, impede o uso de covering indexes, e quebra silenciosamente quando a schema muda (novas colunas ou remoções).
**Solução:**
```sql
-- Errado
SELECT * FROM users WHERE active = true;

-- Correto
SELECT id, email, name, role, created_at
FROM users
WHERE active = true;
```

---

### ❌ Foreign keys sem índice
**Problema:** O desenvolvedor cria foreign keys mas não cria índice na coluna filha.
**Por quê evitar:** JOINs e cascade deletes/updates precisam varrer a tabela filha inteira para encontrar registros relacionados, causando sequential scans mesmo com FK declarada.
**Solução:**
```sql
-- Errado: FK sem índice
ALTER TABLE orders ADD CONSTRAINT fk_orders_user
    FOREIGN KEY (user_id) REFERENCES users (id);

-- Correto: sempre criar índice na coluna de FK
ALTER TABLE orders ADD CONSTRAINT fk_orders_user
    FOREIGN KEY (user_id) REFERENCES users (id);

CREATE INDEX idx_orders_user_id ON orders (user_id);
```

---

### ❌ Transações longas sem timeout
**Problema:** O desenvolvedor abre transações de longa duração sem configurar `lock_timeout` ou `statement_timeout`.
**Por quê evitar:** Transações abertas retêm locks e impedem VACUUM de coletar dead tuples, podendo causar table bloat severo e bloquear outras queries por minutos ou horas.
**Solução:**
```sql
-- Configurar timeouts no nível de sessão ou conexão
SET lock_timeout = '5s';
SET statement_timeout = '30s';

-- Ou configurar no pool de conexões (ex: Prisma datasource url):
-- ?connect_timeout=10&statement_timeout=30000&lock_timeout=5000

-- No nível de role específico (persistente):
ALTER ROLE app_user SET statement_timeout = '30s';
ALTER ROLE app_user SET lock_timeout = '5s';
```

---

### ❌ Paginação com OFFSET em tabelas grandes
**Problema:** O desenvolvedor usa `LIMIT x OFFSET y` para paginar resultados em tabelas com muitos registros.
**Por quê evitar:** PostgreSQL precisa varrer e descartar todas as `y` linhas antes de retornar as `x` desejadas; com OFFSET alto (ex: 100.000), a query fica progressivamente mais lenta mesmo com índice.
**Solução:**
```sql
-- Errado: OFFSET cresce linearmente com o número de página
SELECT id, name, created_at FROM users
ORDER BY created_at DESC
LIMIT 20 OFFSET 100000;

-- Correto: keyset/cursor pagination
SELECT id, name, created_at FROM users
WHERE (created_at, id) < ($last_created_at, $last_id)
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

---

### ❌ NOT IN com subquery que pode retornar NULL
**Problema:** O desenvolvedor usa `NOT IN (SELECT ...)` em uma subquery cuja coluna pode conter valores NULL.
**Por quê evitar:** `NOT IN` retorna 0 linhas se qualquer valor da subquery for NULL, pois `x != NULL` é sempre `UNKNOWN` em SQL — bug silencioso difícil de diagnosticar.
**Solução:**
```sql
-- Errado: se a subquery retornar qualquer NULL, nenhuma linha é retornada
SELECT id FROM users
WHERE id NOT IN (SELECT banned_user_id FROM bans);

-- Correto: usar NOT EXISTS
SELECT u.id FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM bans b WHERE b.banned_user_id = u.id
);

-- Ou LEFT JOIN / IS NULL
SELECT u.id FROM users u
LEFT JOIN bans b ON b.banned_user_id = u.id
WHERE b.banned_user_id IS NULL;
```

---

### ❌ JSON onde colunas relacionais seriam adequadas
**Problema:** O desenvolvedor armazena dados estruturados e consultáveis em colunas JSONB quando esses dados poderiam ser colunas tipadas normais.
**Por quê evitar:** Perde-se type safety, validação de constraints, índices B-tree simples e legibilidade das queries; JOINs em campos JSON são significativamente mais lentos.
**Solução:**
```sql
-- Errado: atributos estáveis e conhecidos dentro de JSON
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    data JSONB -- { "user_id": "...", "total": 99.9, "status": "pending" }
);

-- Correto: colunas tipadas para atributos conhecidos; JSONB só para dados realmente flexíveis
CREATE TABLE orders (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users (id),
    total      NUMERIC(12, 2) NOT NULL,
    status     TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'cancelled')),
    metadata   JSONB  -- apenas dados variáveis/extensíveis
);
```

---

### ❌ TEXT para tudo em vez de tipos apropriados
**Problema:** O desenvolvedor usa `TEXT` para todas as colunas independente do domínio dos dados (e-mails, UUIDs, status, valores monetários).
**Por quê evitar:** Perde-se validação nativa do banco, otimizações de storage, e clareza semântica do schema; tipos errados também afetam performance de comparações e ordenações.
**Solução:**
```sql
-- Errado
CREATE TABLE payments (
    id         TEXT,
    user_id    TEXT,
    amount     TEXT,
    status     TEXT,
    created_at TEXT
);

-- Correto: usar o tipo mais específico disponível
CREATE TABLE payments (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users (id),
    amount     NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
    status     TEXT        NOT NULL CHECK (status IN ('pending', 'paid', 'refunded')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

### ❌ Cast implícito em WHERE impedindo uso de índice
**Problema:** O desenvolvedor compara uma coluna tipada com um literal de tipo diferente, causando cast implícito que invalida o índice.
**Por quê evitar:** PostgreSQL não pode usar o índice quando precisa converter o tipo da coluna a cada linha comparada; isso transforma um Index Scan em Seq Scan.
**Solução:**
```sql
-- Errado: coluna 'id' é INTEGER mas comparada com string
SELECT * FROM users WHERE id = '42';

-- Errado: coluna 'status' tem índice, mas função aplicada na coluna inibe índice
SELECT * FROM orders WHERE UPPER(status) = 'PENDING';

-- Correto: tipos compatíveis / índice funcional
SELECT * FROM users WHERE id = 42;

-- Ou criar índice funcional se a transformação for inevitável
CREATE INDEX idx_orders_status_upper ON orders (UPPER(status));
SELECT * FROM orders WHERE UPPER(status) = 'PENDING';
```

---

### ❌ Ignorar VACUUM e table bloat em tabelas com muitos UPDATEs
**Problema:** O desenvolvedor não monitora e não configura autovacuum em tabelas com alto volume de UPDATE ou DELETE.
**Por quê evitar:** PostgreSQL usa MVCC — cada UPDATE cria uma nova versão da linha; as versões antigas (dead tuples) acumulam e causam table bloat, aumentando o tamanho de scans e degradando performance ao longo do tempo.
**Solução:**
```sql
-- Monitorar bloat por tabela
SELECT relname,
       n_dead_tup,
       n_live_tup,
       round(n_dead_tup * 100.0 / NULLIF(n_live_tup + n_dead_tup, 0), 2) AS dead_pct,
       last_autovacuum,
       last_autoanalyze
FROM pg_stat_user_tables
WHERE n_dead_tup > 10000
ORDER BY n_dead_tup DESC;

-- Para tabelas muito ativas, ajustar autovacuum por tabela
ALTER TABLE orders SET (
    autovacuum_vacuum_scale_factor = 0.01,  -- vacuuma com 1% de dead tuples (default 20%)
    autovacuum_analyze_scale_factor = 0.005
);

-- Vacuum manual em janela de manutenção (não bloqueia leituras)
VACUUM (ANALYZE, VERBOSE) orders;
```
