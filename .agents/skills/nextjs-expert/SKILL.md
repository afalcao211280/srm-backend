---
name: nextjs-expert
description: >
Engenheiro fullstack sênior especializado em Next.js seguindo padrões de
produção. Use SEMPRE que houver menção a Next.js, App Router, RSC, Server
Components, Server Actions, next.config, ou quando o contexto for um
projeto Next.js. Stack: Next.js 14+ (App Router), React Server Components,
Server Actions, Next Auth, Prisma/Drizzle, Tailwind CSS, Vercel/AWS
deploy.
version: "2.1.2"
category: Frontend
keywords:
- nextjs
- react
- app-router
- server-components
- server-actions
- frontend
requires: ['security-expert']
---

# Next.js Expert — Production Standard

## Princípio Central

Server-first. Maximizar Server Components. Client-side apenas quando interatividade exige.

## Versão

- Next.js 14+ (App Router)
- React 18+ (RSC)

## Stack Canônica

| Camada | Lib |
|---|---|
| Framework | Next.js 14+ (App Router) |
| Rendering | React Server Components |
| Mutations | Server Actions |
| Auth | next-auth v5 / Custom |
| ORM | Prisma **ou** Drizzle (perguntar) |
| Client State | TanStack Query v5 |
| Estilos | Tailwind CSS |
| Validação | Zod |
| Testes | Vitest + RTL + Playwright |
| Deploy | Vercel **ou** Docker |

- Refira-se a `react-expert` para padrões de componentes cliente
- Refira-se a `typescript-expert` para padrões de tipagem
- Refira-se a `css-expert` para padrões de estilização

## Contexto (ADR-005)

Em projetos fullstack, `nextjs-expert` é a app `apps/web` do monorepo canônico:

- **Next.js 15** (App Router, RSC, Server Actions) + **React 19** (o `14+`/`18+` da stack é o piso mínimo).
- Monorepo `pnpm 10 + Turborepo 2`: `apps/web` (Next.js), `apps/api` (NestJS BFF), `packages/shared`.
- **Zod 4** como single source of truth em `packages/shared` — Server Actions e o BFF derivam tipos do mesmo schema (`z.infer`), nunca duplicam.
- **tRPC 11** é a API type-safe interna entre `apps/web` e o BFF NestJS (`apps/api`). Route Handlers (`app/api/`) ficam apenas para webhooks/integrações HTTP externas.
- Refira-se a `nestjs-expert` para a camada BFF/API.

## Project Layout

```
app/
├── layout.tsx
├── page.tsx
├── loading.tsx
├── error.tsx
├── not-found.tsx
├── (auth)/
│ ├── login/page.tsx
│ └── register/page.tsx
└── dashboard/
├── page.tsx
└── [id]/page.tsx
components/
├── ui/ # primitivos
└── features/ # componentes de domínio
lib/
├── db.ts # Prisma/Drizzle client
├── auth.ts # auth config
└── utils.ts
server/
├── actions/ # Server Actions
└── queries/ # data fetchers
types/
public/
```

## Padrões Obrigatórios

- **Server Components por default**. `'use client'` apenas quando necessário (event handlers, hooks, browser APIs).
- Server Actions para mutations. Sempre com Zod validation.
- `loading.tsx` em cada route segment. Nunca tela em branco.
- `error.tsx` em cada route segment. Error boundary automático.
- Suspense boundaries para streaming de dados pesados.
- Metadata API para SEO. Nunca `<head>` manual.
- `next/image` para todas as imagens. Nunca `<img>`.
- Parallel routes + intercepting routes para modals.
- Route Handlers (`app/api/`) apenas para webhooks e integrações externas.

## Security

- Server Actions: **sempre** validação Zod no input.
- CSRF protection: Server Actions têm proteção nativa. Route Handlers precisam middleware.
- Auth verification em server actions: `const session = await auth()`.
- Rate limiting em actions sensíveis.
- CSP headers via `next.config.js` headers.
- `'use server'` boundary: nunca exponha server code ao client.
- Environment isolation: `NEXT_PUBLIC_` apenas para valores públicos. Secrets ficam no servidor.

## Anti-padrões

| Proibido | Alternativa |
|---|---|
| `'use client'` desnecessário | Server Component |
| `fetch` em `useEffect` no client | Server Component data fetch |
| Prop drilling de dados do servidor | Server Component fetch + pass as props |
| Inline `<script>` | Next.js Script component |
| Server-only code no client bundle | `'use server'` boundary |
| `loading.tsx`/`error.tsx` ausente | Criar em todo route segment |
| `any` type | `unknown` + type guards |
| Business logic em Route Handlers | Server Actions |

## Checklist

- [ ] App Router convention (layout, page, loading, error, not-found)
- [ ] Server Components maximizados (>80% do code)
- [ ] Server Actions com Zod validation
- [ ] Auth configurado (next-auth ou custom)
- [ ] Metadata API para SEO
- [ ] Testes unitários (Vitest) + E2E (Playwright)
- [ ] ESLint + Prettier
- [ ] Bundle analysis (next build — experimental bundlePagesRouterDependencies)
- [ ] Core Web Vitals (LCP, FID, CLS) dentro do budget
- [ ] Error boundaries em route boundaries

## Quando Pedir Ajuda

- SSR vs SSG vs ISR strategy por página
- Auth provider (next-auth vs custom)
- Database ORM (Prisma vs Drizzle — **perguntar ao usuário**)
- Deployment target (Vercel vs Docker vs AWS)
- Caching strategy (revalidate, no-store, ISR)

## Boas Práticas (referência)

- `reference/best-practices.md` — deep reference Next.js 15: convenção App Router/layouts, Server vs Client, as 4 camadas de cache (Next 15 `fetch` default `no-store`), revalidação, rendering (static/dynamic/streaming/PPR), Server Actions, Route Handlers vs Actions, middleware, metadata/imagens, segurança. Leia ao gerar/revisar código.
- `reference/core.md` — código canônico (Server Component com cache, Server Action com Zod, `useActionState`, NextAuth v5, Route Handler, Dockerfile).
- `reference/stack.md` — versões, `next.config.ts` e `tsconfig.json` canônicos.

## Seguranca (Baseline Compartilhado)
- `reference/security-baseline.md` — Regras universais de seguranca (OWASP, secrets, headers, TLS)

