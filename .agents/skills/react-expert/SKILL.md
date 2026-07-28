---
name: react-expert
description: "Engenheiro frontend sênior especializado em React seguindo padrões de produção. Use SEMPRE que houver menção a React, arquivos.tsx/.jsx, componentes React, hooks, ou quando o contexto for um projeto React. Stack: React 18+, TypeScript, React Query/TanStack Query, Zustand ou Redux Toolkit, React Router, Vitest + RTL, Storybook, Tailwind CSS. Arquitetura: feature-based folders, componentes compositions, hooks customizados."
version: "2.1.2"
category: Frontend
keywords:
- react
- frontend
- typescript
- components
- hooks
- ui
requires:
- security-expert
- typescript-expert
- css-expert
---

# React Expert — Production Standard

## Princípio Central

Componentes production-grade: tipados, testados, composíveis. Código que escala sem refatoração.

## Versão

- React 18+
- TypeScript 5+

## Stack Canônica

| Camada | Lib |
|---|---|
| UI | React 18 |
| Tipagem | TypeScript 5 (strict) |
| Server State | TanStack Query v5 |
| Client State | Zustand ou Redux Toolkit |
| Roteamento | React Router v6 |
| Testes | Vitest + React Testing Library |
| Documentação | Storybook 8 |
| Estilos | Tailwind CSS |
| Componentes | Radix UI / shadcn/ui |

- Refira-se a `typescript-expert` para padrões de tipagem
- Refira-se a `css-expert` para padrões de estilização

## Contexto (ADR-005)

Em projetos fullstack, `react-expert` opera dentro do monorepo canônico:

- **React 19** como biblioteca base (o `18+` da stack é o piso mínimo).
- Monorepo `pnpm 10 + Turborepo 2`: `apps/web` (Next.js), `apps/api` (NestJS BFF), `packages/shared`.
- **Zod 4** como single source of truth: schemas e tipos vivem em `packages/shared` e são reusados por frontend e backend. Componentes/hooks derivam tipos de `z.infer`, nunca duplicam.
- **tRPC 11** é a ponte type-safe entre `apps/web` e `apps/api` — prefira tRPC a `fetch` REST manual quando ambos estão no mesmo monorepo.
- Refira-se a `nextjs-expert` (App Router/RSC) e `nestjs-expert` (BFF) para as camadas server.

## Project Layout

```
src/
├── features/<feature>/
│ ├── components/
│ ├── hooks/
│ ├── api/
│ └── types.ts
├── components/ui/ # primitivos (Button, Input, Dialog)
├── hooks/ # hooks globais
├── lib/ # utils, config, constants
├── styles/ # globals, tailwind
└── types/ # tipos compartilhados
```

## Padrões Obrigatórios

- Functional components only. Zero class components.
- Hooks composition: extract lógica em custom hooks, mantenha componentes limpos.
- TanStack Query para **server state**. Nunca `useState` para dados do servidor.
- Zustand para **client state** (UI, filtros, toggles). Never prop drilling.
- Error boundaries em cada route boundary.
- Suspense boundaries com fallbacks específicos (não spinners genéricos).
- `React.memo` apenas com profiler constatando re-renders. Never preventivo.
- `useCallback`/`useMemo` apenas quando profiler confirmar necessidade.
- Composition > props complexas: ` <Dialog><DialogTrigger/><DialogContent/></Dialog>`.

## Security

- XSS prevention: nunca renderizar user input como HTML.
- `DOMPurify.sanitize()` para qualquer `dangerouslySetInnerHTML`.
- CSP headers no servidor (`Content-Security-Policy`).
- Env vars para API keys. Nunca hardcoded no cliente.
- Tokens em HttpOnly cookies, nunca localStorage.
- Validação de input com Zod em boundaries (form submit, API calls).

## Anti-padrões

| Proibido | Alternativa |
|---|---|
| Prop drilling profundo (>2 níveis) | Context / Zustand / composition |
| `useEffect` para buscar dados | TanStack Query |
| `index` como key | ID estável |
| Inline styles dinâmicos | Tailwind classes / design tokens |
| `any` type | `unknown` + type guards |
| Barrel files que quebram tree-shaking | Import direto por path |
| Lógica de negócio em componentes | Custom hooks / services |
| `// @ts-ignore` | Fix the type |

## Checklist

- [ ] `tsconfig.json`: `strict: true`, `noUncheckedIndexedAccess`
- [ ] Testes RTL >80% coverage nos hooks e componentes críticos
- [ ] Storybook stories para componentes UI
- [ ] ESLint (eslint-config-prettier, react-hooks, react-exhaustive-deps)
- [ ] a11y: axe-core no CI
- [ ] Bundle size check (isolatedModules, tree-shaking)
- [ ] Zod schemas em API boundaries
- [ ] Error boundaries por route

## Quando Pedir Ajuda

- Escolha do state manager (Zustand vs Redux Toolkit vs Context)
- SSR strategy (Next.js vs custom)
- Composição de componentes complexos (compound components, render props)
- Performance: virtualização, code splitting, lazy loading

## Boas Práticas (referência)

- `reference/best-practices.md` — deep reference React 19: regras de hooks, RSC vs Client, `use()`, Actions/`useActionState`/`useOptimistic`, Suspense + Error Boundaries, efeitos (quando NÃO usar), performance + React Compiler, estado (colocation, server vs client), formulários + Zod. Leia ao gerar/revisar código.
- `reference/core.md` — código canônico (componente, hook com TanStack Query, form RHF+Zod, store Zustand, testes).
- `reference/stack.md` — versões e `tsconfig` canônico.

## Seguranca (Baseline Compartilhado)

Regras universais de segurança em `reference/security-baseline.md`.