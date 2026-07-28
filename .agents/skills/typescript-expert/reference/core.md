# TypeScript Expert — Padrões de Código

TypeScript 5.x strict. Padrões prontos pra produção. Leia antes de gerar código.

## Índice

- [Type Design (interface vs type, generics, conditional/mapped/infer)](#0-type-design)
- [1. Configuração de Tipos Estritos](#1-configuração-de-tipos-estritos)
- [2. Zod — Schema, Type e Validation](#2-zod--schema-type-e-validation)
- [3. Discriminated Unions + Exhaustive Check](#3-discriminated-unions--exhaustive-check)
- [4. ts-pattern — Pattern Matching Complexo](#4-ts-pattern--pattern-matching-complexo)
- [5. Generics Avançados](#5-generics-avançados)
- [6. Type-safe Event Emitter](#6-type-safe-event-emitter)
- [7. Assertion Functions + Type Guards](#7-assertion-functions--type-guards)
- [8. Readonly e Immutability (+ type-fest)](#8-readonly-e-immutability)
- [9. Teste de Tipos com expect-type](#9-teste-de-tipos-com-expect-type)
- [Anti-patterns](#anti-patterns)

---

## 0. Type Design

Regras de escolha entre construtos de tipo:

- **`interface`** para object types extensíveis (contratos, classes, APIs que podem ser estendidas via `extends`).
- **`type`** para unions, intersections, mapped types, conditional types. Union só é expressável com `type`.
- **Generic constraints** com `extends`. Nunca generics sem restrição.
- **Conditional types**: `T extends U? X: Y`.
- **Mapped types**: `{ [K in keyof T]:... }`.
- **`infer`** em conditional types para extrair tipos.
- **Variadic tuples**: `[...T, U]`.
- **`Record`** sobre `{}` para dicionários.
- **Utility types** builtin: `Pick`, `Omit`, `Partial`, `Required`, `ReturnType`, `Parameters`.

> `interface` suporta declaration merging (perigoso para tipos de domínio — evite abrir tipos de domínio). Para domínio fechado, prefira `type`.

---

## 1. Configuração de Tipos Estritos

```typescript
// tsconfig strictness: noUncheckedIndexedAccess em ação
const items = ['a', 'b', 'c'];
const first = items[0]; // string | undefined — CORRETO com noUncheckedIndexedAccess

// Acesso seguro
if (first!== undefined) {
console.log(first.toUpperCase()); // OK
}

// Usando nullish coalescing
const name = items[0]?? 'default';
```

---

## 2. Zod — Schema, Type e Validation

```typescript
import { z } from 'zod';

// Schema como fonte única de verdade
const UserSchema = z.object({
id: z.string().uuid(),
name: z.string().min(2).max(100),
email: z.string().email(),
role: z.enum(['admin', 'editor', 'viewer']),
createdAt: z.coerce.date(),
metadata: z.record(z.string(), z.unknown()).optional(),
});

// Tipo derivado do schema — nunca definir tipo à mão
type User = z.infer<typeof UserSchema>;

// Parsing com erro tipado — use em trust boundaries (dado inválido = bug)
function parseUser(input: unknown): User {
return UserSchema.parse(input); // lança ZodError com detalhes
}

// Parsing seguro (sem throw) — use com input de usuário (falha esperada)
function tryParseUser(input: unknown): { success: true; data: User } | { success: false; error: z.ZodError } {
return UserSchema.safeParse(input);
}

// Composição DRY: partial/omit/pick/extend/merge
const UpdateUserSchema = UserSchema.partial().omit({ id: true, createdAt: true });
type UpdateUser = z.infer<typeof UpdateUserSchema>;

// Transform — normalização no parse (trim, parse de data)
const UserApiSchema = UserSchema.transform((u) => ({
...u,
displayName: u.name.trim(),
}));
```

**Regra `parse` vs `safeParse`:** `parse` em boundaries de confiança (contrato de API violado = bug); `safeParse` onde a falha é esperada (input de formulário do usuário).

---

## 3. Discriminated Unions + Exhaustive Check

```typescript
type ApiResult<T> =
| { status: 'success'; data: T; timestamp: Date }
| { status: 'error'; code: string; message: string }
| { status: 'loading' };

function handleResult<T>(result: ApiResult<T>): string {
switch (result.status) {
case 'success':
return `Data: ${JSON.stringify(result.data)}`;
case 'error':
return `Error ${result.code}: ${result.message}`;
case 'loading':
return 'Loading...';
default:
// TypeScript garante que este código é inacessível
result satisfies never;
return '';
}
}
```

Compare com o **anti-padrão** de flags booleanas — permite estados inválidos:
```typescript
// ❌ Ruim: permite { loading: true, error: Error } simultâneos
type BadState<T> = { loading: boolean; data?: T; error?: Error };
```

---

## 4. ts-pattern — Pattern Matching Complexo

```typescript
import { match, P } from 'ts-pattern';

type Command =
| { type: 'create'; payload: { name: string } }
| { type: 'update'; id: string; payload: Partial<{ name: string }> }
| { type: 'delete'; id: string };

function processCommand(cmd: Command): string {
return match(cmd)
.with({ type: 'create', payload: { name: P.string.minLength(2) } }, ({ payload }) =>
`Creating: ${payload.name}`,
)
.with({ type: 'update' }, ({ id, payload }) =>
`Updating ${id}: ${JSON.stringify(payload)}`,
)
.with({ type: 'delete' }, ({ id }) => `Deleting ${id}`)
.exhaustive(); // erro de compile se case não coberto
}
```

---

## 5. Generics Avançados

```typescript
// Generic com constraint e default
type Repository<T extends { id: string }, TCreate = Omit<T, 'id' | 'createdAt'>> = {
findById(id: string): Promise<T | null>;
findAll(): Promise<T[]>;
create(data: TCreate): Promise<T>;
update(id: string, data: Partial<TCreate>): Promise<T>;
delete(id: string): Promise<void>;
};

// Conditional types
type NonNullableFields<T> = {
[K in keyof T]-?: NonNullable<T[K]>;
};

// Mapped types com remapping
type Getters<T> = {
[K in keyof T as `get${Capitalize<string & K>}`]: () => T[K];
};

// Template literal types
type EventName<T extends string> = `on${Capitalize<T>}`;
type UserEvents = EventName<'created' | 'updated' | 'deleted'>;
// → 'onCreated' | 'onUpdated' | 'onDeleted'

// Infer em conditional types
type UnpackPromise<T> = T extends Promise<infer U>? U: T;
type UnpackArray<T> = T extends (infer U)[]? U: T;
```

---

## 6. Type-safe Event Emitter

```typescript
type EventMap = Record<string, unknown>;

class TypedEventEmitter<T extends EventMap> {
private listeners = new Map<keyof T, Set<(data: unknown) => void>>();

on<K extends keyof T>(event: K, listener: (data: T[K]) => void): this {
if (!this.listeners.has(event)) {
this.listeners.set(event, new Set());
}
this.listeners.get(event)!.add(listener as (data: unknown) => void);
return this;
}

emit<K extends keyof T>(event: K, data: T[K]): void {
this.listeners.get(event)?.forEach((l) => l(data));
}

off<K extends keyof T>(event: K, listener: (data: T[K]) => void): this {
this.listeners.get(event)?.delete(listener as (data: unknown) => void);
return this;
}
}

// Uso tipado
type AppEvents = {
userCreated: { id: string; name: string };
userDeleted: { id: string };
};

const emitter = new TypedEventEmitter<AppEvents>();
emitter.on('userCreated', ({ id, name }) => console.log(`Created: ${name} (${id})`));
emitter.emit('userCreated', { id: '1', name: 'João' });
```

---

## 7. Assertion Functions + Type Guards

```typescript
// Type guard
function isString(value: unknown): value is string {
return typeof value === 'string';
}

// Assertion function (lança se falso)
function assertString(value: unknown, label = 'value'): asserts value is string {
if (typeof value!== 'string') {
throw new TypeError(`Expected ${label} to be string, got ${typeof value}`);
}
}

// Narrowing com Array.isArray
function processItems(items: string | string[]): string[] {
return Array.isArray(items)? items: [items];
}

// Type predicate com Zod
import { z } from 'zod';
const UserSchema = z.object({ id: z.string(), name: z.string() });
type User = z.infer<typeof UserSchema>;

function isUser(value: unknown): value is User {
return UserSchema.safeParse(value).success;
}
```

---

## 8. Readonly e Immutability

```typescript
// Prefer readonly em todos os parâmetros
function sumAll(values: readonly number[]): number {
return values.reduce((a, b) => a + b, 0);
}

// ReadonlyDeep (via type-fest)
import type { ReadonlyDeep } from 'type-fest';

type Config = ReadonlyDeep<{
db: { host: string; port: number };
features: string[];
}>;

// Object.freeze para runtime immutability
const DEFAULT_CONFIG = Object.freeze({
timeout: 5000,
retries: 3,
baseUrl: 'https://api.example.com',
} as const);

// 'as const' para literal types
const ROLES = ['admin', 'editor', 'viewer'] as const;
type Role = (typeof ROLES)[number]; // 'admin' | 'editor' | 'viewer'
```

**type-fest** (opcional, quando builtin não basta): `Opaque<T, Token>` (branded types mais limpos que `& { __brand }`), `PartialDeep<T>`, `ReadonlyDeep<T>`, `SetRequired`/`SetOptional`, `Simplify<T>` (achata intersections no tooltip do IDE).

```typescript
import type { Opaque, PartialDeep } from 'type-fest';
type UserId = Opaque<string, 'UserId'>;
type UserPatch = PartialDeep<User>;
```

---

## 9. Teste de Tipos com expect-type

```typescript
import { expectTypeOf } from 'expect-type';
import { describe, it } from 'vitest';
import type { UnpackPromise } from './utils';

describe('Type tests', () => {
it('UnpackPromise unwraps promise type', () => {
expectTypeOf<UnpackPromise<Promise<string>>>().toEqualTypeOf<string>();
expectTypeOf<UnpackPromise<number>>().toEqualTypeOf<number>();
});

it('ApiResult success branch has data', () => {
type SuccessResult = Extract<ApiResult<string>, { status: 'success' }>;
expectTypeOf<SuccessResult['data']>().toEqualTypeOf<string>();
});
});
```

---

## Anti-patterns

### ❌ Usar `any` em vez de `unknown`
**Problema:** O desenvolvedor tipifica dados externos (JSON, respostas de API, erros de `catch`) como `any`.
**Por quê evitar:** `any` desativa completamente a checagem de tipos, propagando insegurança para todo o código que consome o valor — erros de runtime passam despercebidos em compile time.
**Solução:**
```typescript
// ❌ Errado
async function fetchUser(): Promise<any> {
const res = await fetch('/api/user');
return res.json();
}

// ✅ Correto — parse com Zod garante segurança em runtime
async function fetchUser(): Promise<User> {
const res = await fetch('/api/user');
return UserSchema.parse(await res.json());
}

// ✅ Em catch, use unknown
try {
doSomething();
} catch (err: unknown) {
if (err instanceof Error) console.error(err.message);
}
```

---

### ❌ Type assertion (`as SomeType`) sem validação runtime
**Problema:** O desenvolvedor usa `as SomeType` para "convencer" o compilador de que um valor tem determinado tipo sem verificar isso em runtime.
**Por quê evitar:** A assertion é apagada em JavaScript — se o valor real não corresponde ao tipo asserted, ocorrem crashes silenciosos ou comportamentos imprevisíveis em produção.
**Solução:**
```typescript
// ❌ Errado — o compilador confia cegamente
const user = response.data as User;

// ✅ Correto — Zod valida em runtime e o tipo é inferido
const user = UserSchema.parse(response.data);

// ✅ Alternativa com safeParse para não lançar exceção
const result = UserSchema.safeParse(response.data);
if (!result.success) {
logger.error('Invalid user payload', result.error.flatten());
return;
}
const user = result.data; // User — tipado e validado
```

---

### ❌ Non-null assertion (`!`) sem type guard
**Problema:** O desenvolvedor usa `valor!` para suprimir o aviso de possível `null` ou `undefined` sem verificar se o valor realmente existe.
**Por quê evitar:** Se o valor for `null`/`undefined` em runtime, a aplicação lança `TypeError`, e o stack trace não aponta para a assertion — difícil de depurar.
**Solução:**
```typescript
// ❌ Errado
const el = document.getElementById('root')!;

// ✅ Correto — verificação explícita
const el = document.getElementById('root');
if (!el) throw new Error('Root element #root not found in DOM');

// ✅ Assertion function reutilizável
function assertDefined<T>(value: T | null | undefined, label: string): asserts value is T {
if (value == null) throw new Error(`${label} must be defined`);
}
```

---

### ❌ Duplicar tipos em vez de usar `z.infer<typeof Schema>`
**Problema:** O desenvolvedor define manualmente uma `interface`/`type` que repete os campos de um schema Zod já existente.
**Por quê evitar:** Duplicação cria dessincronia inevitável — quando o schema muda, o tipo manual não é atualizado, gerando bugs silenciosos onde o tipo diz uma coisa e o parser valida outra.
**Solução:**
```typescript
// ❌ Errado — definição duplicada e propensa a divergir
const UserSchema = z.object({ id: z.string(), name: z.string() });
interface User { id: string; name: string; } // ← vai desatualizar

// ✅ Correto — única fonte de verdade
const UserSchema = z.object({ id: z.string(), name: z.string() });
type User = z.infer<typeof UserSchema>; // sempre sincronizado
```

---

### ❌ Usar `enum` em vez de `as const` object
**Problema:** O desenvolvedor usa `enum` do TypeScript para definir conjuntos de valores constantes.
**Por quê evitar:** `enum` gera código JS extra (objeto bidirecional), `const enum` tem pitfalls com isolatedModules/Babel/esbuild, e ambos impedem iteração simples. `as const` é zero-cost e universal.
**Solução:**
```typescript
// ❌ Errado
enum Role { Admin = 'admin', Editor = 'editor', Viewer = 'viewer' }

// ✅ Correto — as const + union type
const ROLE = { Admin: 'admin', Editor: 'editor', Viewer: 'viewer' } as const;
type Role = (typeof ROLE)[keyof typeof ROLE]; // 'admin' | 'editor' | 'viewer'

// ✅ Ou array para union direta
const ROLES = ['admin', 'editor', 'viewer'] as const;
type Role = (typeof ROLES)[number];
```

---

### ❌ Não habilitar `strict` e `noUncheckedIndexedAccess` no tsconfig
**Problema:** O projeto usa configuração padrão sem `"strict": true` ou `"noUncheckedIndexedAccess": true`.
**Por quê evitar:** Sem `strict`, `strictNullChecks` fica off e `null`/`undefined` são atribuíveis a qualquer tipo. Sem `noUncheckedIndexedAccess`, `arr[0]` retorna `T` em vez de `T | undefined`, escondendo bugs de acesso fora dos limites.
**Solução:**
```jsonc
{
"compilerOptions": {
"strict": true, // strictNullChecks, noImplicitAny, etc.
"noUncheckedIndexedAccess": true, // arr[0] → T | undefined
"noImplicitReturns": true, // toda branch deve retornar
"exactOptionalPropertyTypes": true // distingue undefined de ausente
}
}
```

---

### ❌ Funções exportadas sem tipo de retorno explícito
**Problema:** O desenvolvedor exporta funções sem anotar o tipo de retorno, confiando na inferência.
**Por quê evitar:** Inferência de retorno em funções exportadas é frágil — qualquer mudança interna acidental no tipo de retorno altera a API pública sem erro de compilação; o consumidor só descobre em runtime.
**Solução:**
```typescript
// ❌ Errado — retorno implícito, API pública não documentada no tipo
export function formatCurrency(value: number, locale = 'pt-BR') {
return new Intl.NumberFormat(locale, { style: 'currency', currency: 'BRL' }).format(value);
}

// ✅ Correto — retorno explícito fecha o contrato
export function formatCurrency(value: number, locale = 'pt-BR'): string {
return new Intl.NumberFormat(locale, { style: 'currency', currency: 'BRL' }).format(value);
}
```

---

### ❌ Generics com excesso de parâmetros não relacionados
**Problema:** Funções/tipos genéricos com 4+ parâmetros de tipo sem defaults ou constraints claros.
**Por quê evitar:** Assinaturas ilegíveis, inferência automática difícil, forçam o chamador a especificar tipos manualmente — na prática ninguém usa a API.
**Solução:**
```typescript
// ❌ Errado
function transform<TIn, TOut, TCtx, TErr, TOpts>(input: TIn, ctx: TCtx, opts: TOpts): Promise<TOut | TErr> {}

// ✅ Correto — agrupe em objeto de opções e use defaults
type TransformOptions<TCtx = Record<string, unknown>> = { ctx: TCtx; timeout?: number };
function transform<TIn, TOut>(input: TIn, transformer: (v: TIn) => TOut, options?: TransformOptions): Promise<TOut> {}
```

---

### ❌ Usar `interface` para union types
**Problema:** O desenvolvedor tenta descrever um union usando `interface` (erro do compilador) ou usa uma interface "aberta" demais.
**Por quê evitar:** `interface` só descreve objetos únicos e suporta declaration merging (perigoso para domínio). Union e discriminated union só são expressáveis com `type`.
**Solução:**
```typescript
// ❌ Errado — interface não suporta union
// interface Result = SuccessResult | ErrorResult; // SyntaxError

// ✅ Correto — type para unions
type Result<T> =
| { status: 'ok'; data: T }
| { status: 'error'; code: string; message: string };

// Use interface apenas para contratos de objetos extensíveis
interface Repository<T> {
findById(id: string): Promise<T | null>;
save(entity: T): Promise<T>;
}
```

---

### ❌ Narrowing ausente após `z.safeParse`
**Problema:** Chamar `safeParse` mas acessar `result.data` sem verificar `result.success`.
**Por quê evitar:** Quando `success` é `false`, `result.data` é `undefined` — acessá-lo causa crash, e o TS não protege se usar `as`.
**Solução:**
```typescript
// ❌ Errado
const result = UserSchema.safeParse(payload);
processUser(result.data as User); // crash se payload inválido

// ✅ Correto — narrowing via success flag
const result = UserSchema.safeParse(payload);
if (!result.success) {
return NextResponse.json({ errors: result.error.flatten().fieldErrors }, { status: 422 });
}
processUser(result.data); // User — garantido pelo narrowing
```
