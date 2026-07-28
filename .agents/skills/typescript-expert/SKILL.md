---
name: typescript-expert
description: >
Especialista em TypeScript seguindo padrões de produção (cross-cutting,
ADR-008). Stack: TypeScript 5.x strict, Zod 4, branded types, discriminated
unions, type guards, ts-pattern, type-fest, tsd/expect-type, ESLint
`@typescript-eslint`. Complementar às skills de framework (react/nextjs/vue/
angular). Gera tipos, schemas Zod, type guards e configs prontos pra produção.
Acionar SEMPRE que mencionar TypeScript,.ts/.tsx, tsconfig, tipagem, generics,
interface, type, branded type, Zod, type guard, discriminated union.
version: "2.1.0"
category: Frontend
keywords:
- typescript
- types
- strict
- generics
- zod
- type-guards
- branded-types
- discriminated-unions
- tsconfig
requires:
- security-expert
---

# TypeScript Expert — Padrões

Especialista TypeScript. `strict` sempre, zero `any`. Tipos servem o domínio, não o contrário. Código que entra em produção.

## Princípios

1. **`strict: true` + `noUncheckedIndexedAccess`** — pega bugs de null/undefined e acesso fora de limites em build, não em runtime.
2. **Zero `any`** — use `unknown` + type guards / Zod. `any` desliga a checagem e propaga insegurança pra todo consumidor. Alvo: type-coverage > 95%.
3. **Make illegal states unrepresentable** — o tipo impede combinações inválidas (discriminated union em vez de flags booleanas soltas).
4. **Schema é a fonte única de verdade** — Zod no boundary (API/form/env), tipo derivado com `z.infer<>`. Nunca duplicar tipo à mão.
5. **Narrowing explícito** — type guards, assertion functions, exhaustive `switch` com `never`. Nunca `as`/`!` sem verificação.
6. **Tipos servem o consumidor** — retorno explícito em funções exportadas (fecha o contrato da API pública); generics com `extends` e defaults, nunca 4+ parâmetros soltos.

> Esta skill é **complementar** e referenciada pelas skills de framework. Ela cobre idiomas de linguagem; padrões específicos (hooks, SFC, Signals) ficam na skill do framework.

## Stack Canônica

| Categoria | Lib | Uso |
|---|---|---|
| Linguagem | TypeScript 5.x | `strict: true` sempre |
| Validação runtime | Zod 4 | Schema → tipo (`z.infer`). Padrão (ADR-008) |
| Utility types | type-fest 4.x | `Opaque`, `Simplify`, `SetOptional`, `ReadonlyDeep` |
| Pattern matching | ts-pattern 5.x | `match().with().exhaustive()` tipado |
| Runtime script | tsx 4.x | `tsx script.ts` |
| Build lib | tsup | libs; `tsc --noEmit` para type-check |
| Teste de tipos | tsd / expect-type | `expectTypeOf<T>()` no CI |
| Lint | ESLint 9 + `@typescript-eslint` | `recommendedTypeChecked` |
| Format | Prettier 3 | — |

> **FIXO** (ADR-008): TypeScript strict, Zod 4 para validação runtime, zero `any`. **Alternativas** (perguntar): Valibot (bundle menor que Zod), io-ts (FP-style). type-fest é opcional (só quando utilitários builtin não bastam).

## Padrões Obrigatórios

- **Discriminated unions** para estados mutuamente exclusivos (state machines) — não flags booleanas que permitem `{ loading: true, error: Error }`.
- **Branded types** para domain IDs — `type UserId = string & { readonly __brand: 'UserId' }` impede passar `OrderId` onde `UserId` é esperado.
- **`as const`** para literal unions; deriva `type Role = (typeof ROLES)[number]` — array e tipo sempre em sincronia. Preferir a `enum`.
- **`satisfies`** para validar sem alargar o tipo (type narrowing sem widening).
- **Zod**: `parse` em trust boundaries (dado inválido é bug → lança); `safeParse` para input de usuário (falha é esperada → narrowing via `result.success`). Compor com `.extend()`/`.pick()`/`.omit()`/`.merge()`; normalizar com `.transform()`.
- **Type design**: `interface` para objetos extensíveis (contratos, classes); `type` para unions/intersections/mapped/conditional. Generics com constraint (`extends`), nunca sem restrição.
- **Runtime safety**: type guards (`value is T`), assertion functions (`asserts value is T`), exhaustive `switch` (default com `const _: never = x`).
- **Error handling**: `Result<T, E>` / discriminated error types para erros esperados; `catch (e: unknown)` + `instanceof`.
- **Config**: `strict`, `noUncheckedIndexedAccess`, `noImplicitReturns`, `forceConsistentCasingInFileNames`. Ver `reference/stack.md`.

## Workflow Agentic

1. **Modelar o domínio** — quais estados são válidos? Modele com discriminated union / branded types para tornar o inválido inexpressável.
2. **Schema no boundary** — Zod para todo dado externo (API/form/env). Tipo derivado com `z.infer<>`.
3. **Guards e narrowing** — type guards / assertion functions onde o tipo precisa estreitar.
4. **Verificar** — `tsc --noEmit` limpo; teste de tipos (`expect-type`/`tsd`) para utilitários genéricos; ESLint `no-explicit-any` sem violação.

## Anti-padrões

| Proibido | Alternativa |
|---|---|
| `any` | `unknown` + type guard / Zod |
| `as SomeType` sem validação runtime | `Schema.parse()` (Zod valida e infere) |
| Non-null `!` sem guard | narrowing explícito / assertion function |
| `// @ts-ignore` / `@ts-expect-error` sem justificativa | corrigir o tipo |
| `enum` | `as const` + union |
| Duplicar tipo do schema Zod | `z.infer<typeof Schema>` |
| `interface` para union | `type` |
| Função exportada sem tipo de retorno | retorno explícito (fecha contrato) |
| Generics com 4+ parâmetros soltos | objeto de opções + defaults |
| Acessar `result.data` sem checar `result.success` | narrowing via `if (!result.success) return` |
| `Object.keys().forEach` sem tipo | `for...of` / cast tipado de `keyof` |
| Implicit `any` (param sem anotação) | anotação explícita |

## Checklist

- [ ] `strict: true` + `noUncheckedIndexedAccess`
- [ ] Zero `any` (type-coverage > 95%)
- [ ] Zod schemas em todos API/form/env boundaries; tipos via `z.infer<>`
- [ ] Branded types para domain IDs
- [ ] Discriminated unions para state machines
- [ ] Exhaustive `switch` com `never` check
- [ ] Error handling com `Result`/`Either` ou discriminated error
- [ ] ESLint `@typescript-eslint` (`no-explicit-any`, `no-unsafe-assignment`)
- [ ] Imports diretos por path (não barrel que quebra tree-shaking)
- [ ] `tsc --noEmit` no CI

## Quando Perguntar

Antes de tipar: Zod vs Valibot (bundle) vs io-ts (FP)? Precisa de type-fest? Estratégia de migração JS → TS (incremental)?
Decisões que não se assume: tipos recursivos/condicionais complexos, generics avançados (variance, `infer` aninhado), design de branded types para o domínio.

## Referências (sob demanda — progressive disclosure)

- **`reference/core.md`** — Padrões de código completos: `strict` em ação, Zod (schema/parse/safeParse/transform), discriminated unions + exhaustive check, ts-pattern, generics avançados (mapped/conditional/template literal/`infer`), event emitter tipado, assertion functions/type guards, readonly/immutability, teste de tipos, + 12 anti-padrões detalhados com o porquê. Leia antes de gerar código.
- **`reference/stack.md`** — Stack canônica: versões, `tsconfig.json` (template "strictest"), ESLint flat config `@typescript-eslint`. Leia ao configurar projeto novo.

## Cross-references

- `react-expert`, `nextjs-expert`, `vue-expert`, `angular-expert` — esta skill fornece a base de tipagem
- `css-expert` — tipos de Tailwind config / theme
- `security-expert` — validação de input, Zod em boundaries

## Segurança (Baseline Compartilhado)

Regras universais em `reference/security-baseline.md`. Específico de tipagem:

- **Validação no boundary**: Zod `parse`/`safeParse` em TODO dado externo (API, form, env, `JSON.parse`). Tipos não existem em runtime — sem validação, `any`/`as` deixam dado malicioso entrar tipado.
- **Nunca `as` para dado não confiável**: assertion apaga em runtime; use Zod que valida de fato.
- **Env vars tipadas e validadas** com Zod no bootstrap — falhar cedo se faltar segredo.
