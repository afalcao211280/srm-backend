# Next.js 15 (App Router) — Best Practices (deep reference)

Referência aprofundada, -aligned. App Router + React Server Components.
Conteúdo conceitual e de decisão — o **porquê** por trás de cada regra.
Para código canônico completo (Server Component com cache, Server Action com Zod,
form com `useActionState`, NextAuth, Route Handler, Dockerfile), veja `core.md`.
Para versões, `next.config.ts` e `tsconfig.json`, veja `stack.md`.

## Índice

1. [App Router: convenção de arquivos e layouts](#1-app-router-convencao-de-arquivos-e-layouts)
2. [Server vs Client Components](#2-server-vs-client-components)
3. [Data fetching e as camadas de cache](#3-data-fetching-e-as-camadas-de-cache)
4. [Revalidação e mutação de cache](#4-revalidacao-e-mutacao-de-cache)
5. [Rendering: static, dynamic, streaming, PPR](#5-rendering-static-dynamic-streaming-ppr)
6. [Server Actions](#6-server-actions)
7. [Route Handlers vs Server Actions](#7-route-handlers-vs-server-actions)
8. [Middleware](#8-middleware)
9. [Metadata, imagens e fontes](#9-metadata-imagens-e-fontes)
10. [Segurança](#10-seguranca)
11. [Anti-padrões](#11-anti-padroes)
12. [Checklist de revisão](#12-checklist-de-revisao)
13. [Cross-references](#13-cross-references)

---

## 1. App Router: convenção de arquivos e layouts

Cada segmento de rota é uma pasta; arquivos especiais definem o comportamento:

- `layout.tsx` — UI compartilhada que **preserva estado** entre navegações de
filhos (não remonta). O `app/layout.tsx` raiz é obrigatório e contém `<html>`/
`<body>`. Layouts aninhados compõem.
- `page.tsx` — UI única da rota (torna o segmento acessível publicamente).
- `loading.tsx` — Suspense boundary automático do segmento; mostra fallback
durante o streaming. **Nunca** deixe tela em branco.
- `error.tsx` — Error Boundary automático (Client Component; recebe `error` +
`reset`). `global-error.tsx` para erros do layout raiz.
- `not-found.tsx` — UI de 404 (disparada por `notFound()`).
- `route.ts` — Route Handler (API), mutuamente exclusivo com `page.tsx` no mesmo
segmento.
- `template.tsx` — como layout, mas **remonta** a cada navegação (útil para
animações de entrada / efeitos por navegação).

Recursos de organização:

- **Route groups** `(grupo)` — organizam sem afetar a URL (ex.: `(auth)/login`).
- **Dynamic segments** `[id]`, catch-all `[...slug]`, opcional `[[...slug]]`.
- **Parallel routes** `@slot` + **intercepting routes** `(.)`/`(..)` — modais com
URL própria, dashboards com seções independentes que fazem streaming separado.
- **Private folders** `_pasta` — fora do roteamento (colocation de helpers).

## 2. Server vs Client Components

- **RSC é o default**. `'use client'` só quando há event handler, hook de estado/
efeito, ou browser API. Coloque a diretiva na **folha** interativa, nunca no
layout raiz (senão toda a árvore vira client e o bundle explode).
- Server Components podem ser `async`, acessar DB/secrets e não vão para o bundle.
- Passe Server Components como `children`/props para Client Components (composição
server-in-client). Client não importa Server.
- Props que cruzam server→client devem ser serializáveis; funções só como Server
Actions.
- `server-only` e `client-only` (pacotes) travam em build módulos que vazariam
para o lado errado — use `server-only` em módulos com secret/DB.

## 3. Data fetching e as camadas de cache

Next 15 tem **quatro** camadas de cache; entender cada uma evita dados obsoletos
e revalidações erradas:

| Camada | Escopo | O que faz |
|---|---|---|
| **Request Memoization** | 1 render | Dedupe de `fetch` idêntico na mesma árvore de render |
| **Data Cache** | Persistente (servidor) | Armazena resultado de `fetch`/dados entre requests e deploys |
| **Full Route Cache** | Persistente (servidor, build) | HTML/RSC payload de rotas estáticas |
| **Router Cache** | Sessão (cliente) | Payload RSC de rotas visitadas, no navegador |

Pontos-chave (mudança importante no Next 15):

- **`fetch` NÃO é mais cacheado por padrão** no Next 15 (default `no-store`).
Opte por cache explicitamente: `fetch(url, { cache: 'force-cache' })` ou
`{ next: { revalidate: N } }`.
- Para dados não-`fetch` (ORM, SDK), use `unstable_cache(fn, keys, { tags,
revalidate })` para entrar no Data Cache.
- **Request Memoization** dedupa chamadas iguais no mesmo render — não precisa
passar dados por prop drilling só para evitar re-fetch; buscar de novo no
componente que precisa é ok.
- Ler `cookies()`, `headers()`, `searchParams` ou usar `no-store` torna o
segmento **dinâmico** (opt out do Full Route Cache). No Next 15 essas APIs são
**assíncronas** (`await cookies()`).

## 4. Revalidação e mutação de cache

- **Time-based**: `next: { revalidate: N }` no fetch, ou `export const revalidate
= N` no segmento (ISR).
- **On-demand**: em uma mutação (Server Action/Route Handler) chame
`revalidateTag('tag')` (invalida tudo marcado com `tags: ['tag']`) ou
`revalidatePath('/rota')`.
- Toda mutação que altera dados exibidos **deve** revalidar o cache
correspondente — senão a UI mostra dados velhos até o TTL. Prefira **tags**
(granular) a paths quando possível.
- `revalidatePath`/`revalidateTag` também limpam o Router Cache do cliente para as
rotas afetadas na próxima navegação.

## 5. Rendering: static, dynamic, streaming, PPR

- **Static (default)**: renderizado em build/revalidação, servido do Full Route
Cache. Melhor performance; use sempre que possível.
- **Dynamic**: renderizado por request (ao usar `cookies()`/`headers()`/
`searchParams`, `no-store`, ou `export const dynamic = 'force-dynamic'`). Use
conscientemente — dinâmico desnecessário perde cache.
- **Streaming**: `loading.tsx` (segmento) e `<Suspense>` (granular) enviam o shell
cedo e transmitem o conteúdo lento depois, melhorando TTFB/LCP percebido.
- **`generateStaticParams`** pré-gera rotas dinâmicas (`[id]`) em build (SSG).
Combine com `dynamicParams` para controlar params não pré-gerados.
- **PPR (Partial Prerendering)** — quando habilitado, serve um shell estático e
faz stream das partes dinâmicas (dentro de `<Suspense>`) no mesmo response.
Recurso experimental; use com boundaries claros.

## 6. Server Actions

- Funções `'use server'` (arquivo `actions.ts` ou inline) executam no servidor;
chamáveis via `<form action>` ou de Client Components.
- **Validação Zod obrigatória** no input — a Action é um endpoint público, trate
todo `formData` como não confiável. Retorne erros no estado (para
`useActionState`), não `throw` para erro de validação.
- **Auth verification** dentro da Action: `const session = await auth()`; não
confie em checagem só no client. Autorize por recurso antes de mutar.
- Após mutar, **revalide** (`revalidateTag`/`revalidatePath`) e/ou `redirect()`.
- **CSRF**: Server Actions têm proteção nativa (checagem de origin/POST). Route
Handlers **não** — exigem middleware.
- Rate limiting em actions sensíveis (login, envio de email, pagamento).
- Progressive enhancement: `<form action={serverAction}>` funciona sem JS.

## 7. Route Handlers vs Server Actions

- **Server Actions** para mutações originadas da própria app (forms, botões) —
type-safe, sem boilerplate de endpoint, CSRF nativo. Preferência default.
- **Route Handlers** (`app/api/.../route.ts`) para: webhooks externos, integrações
de terceiros, APIs públicas/consumidas por outros clientes, streaming/SSE, OAuth
callbacks. Exportam `GET`/`POST`/etc.
- No monorepo, a API type-safe interna é o **tRPC 11 no BFF NestJS**
(`apps/api`), não Route Handlers manuais — ver SKILL.md / ADR-005. Route
Handlers ficam para o que precisa ser HTTP puro.
- Não coloque business logic em Route Handler; delegue a serviços/actions.

## 8. Middleware

- `middleware.ts` na raiz roda no **Edge** antes da requisição casar com a rota.
Use para: auth/redirect de sessão, i18n/geo, A/B rewrite, headers de segurança,
bloqueio de bots.
- Restrinja com `config.matcher` — não rode em assets/estáticos (custo por
request). Middleware é hot path: mantenha leve, sem DB pesado; valide apenas a
presença/assinatura do token (verificação profunda fica na Action/Handler).
- Runtime Edge: sem APIs Node completas (sem `fs`, libs nativas). Use libs
edge-compatíveis (ex.: `jose` para JWT).
- Middleware **não** substitui autorização no servidor — é a primeira barreira,
não a única. Reverifique auth na Server Action/Route Handler.

## 9. Metadata, imagens e fontes

- **Metadata API**: `export const metadata` (estático) ou `generateMetadata`
(dinâmico, async) — nunca `<head>` manual. Suporta Open Graph, `alternates`,
`robots`. `opengraph-image`/`icon` por convenção de arquivo.
- **`next/image`** para todas as imagens: lazy loading, `sizes` responsivo,
formatos modernos, prevenção de CLS via `width`/`height`. Nunca `<img>` cru.
- **`next/font`** auto-hospeda fontes (sem request externo, sem layout shift).
- Core Web Vitals dentro do budget (LCP, CLS, INP) — parte do quality gate.

## 10. Segurança

- `NEXT_PUBLIC_` **apenas** para valores públicos; qualquer secret sem esse prefixo
fica só no servidor. Nunca importe módulo com secret em Client Component (use
`server-only`).
- Server Actions: Zod no input + auth + autorização por recurso + rate limit.
- Route Handlers: CSRF/origin check manual, auth explícita.
- CSP e demais headers de segurança via `next.config` `headers()` ou middleware.
- Cookies de sessão: `httpOnly`, `secure`, `sameSite`. Tokens nunca em
`localStorage`.
- Regras universais (OWASP, secrets, TLS) em `security-baseline.md`.

## 11. Anti-padrões

| Proibido | Alternativa |
|---|---|
| `'use client'` no layout raiz | Diretiva nas folhas interativas |
| `fetch` no client quando RSC resolve | Server Component data fetch |
| Assumir `fetch` cacheado (Next 15) | Opt-in explícito `force-cache`/`revalidate` |
| Mutação sem revalidar cache | `revalidateTag`/`revalidatePath` |
| `cookies()`/`headers()` sem `await` | `await cookies()` (Next 15 async) |
| Secret em `NEXT_PUBLIC_`/client | Manter no servidor + `server-only` |
| Business logic em Route Handler | Server Action / serviço |
| Rota dinâmica sem `generateStaticParams` | Pré-gerar params conhecidos |
| Waterfall de `await` sequenciais | `Promise.all` / streaming |
| `loading.tsx`/`error.tsx` ausente | Criar em cada segmento |
| `<img>` nativo | `next/image` |
| Middleware pesado (DB) sem matcher | Matcher restrito + checagem leve |

## 12. Checklist de revisão

- [ ] RSC maximizado; `'use client'` só em folhas.
- [ ] `loading.tsx` + `error.tsx` em cada route segment.
- [ ] Cache explícito por fetch (Next 15 default `no-store`).
- [ ] Toda mutação revalida a tag/path correspondente.
- [ ] `cookies()`/`headers()`/`searchParams` com `await`.
- [ ] Server Actions com Zod + `await auth()` + autorização.
- [ ] Route Handlers só para webhook/integração/HTTP externo.
- [ ] Middleware com `matcher` restrito e verificação leve.
- [ ] Metadata API para SEO; `next/image` e `next/font`.
- [ ] Sem secret em `NEXT_PUBLIC_`; `server-only` em módulos sensíveis.
- [ ] Core Web Vitals dentro do budget.

## 13. Cross-references

- `react-expert` — RSC/Client, `use()`, Actions, Suspense, hooks, performance.
- `typescript-expert` — tipagem strict, `z.infer`, type guards.
- `css-expert` — Tailwind, design tokens.
- `core.md` — código canônico (Server Component, Action, NextAuth, Route Handler).
- `stack.md` — versões, `next.config.ts`, `tsconfig.json`.
- ADR-005 — stack fullstack (Next 15 + React 19 + tRPC 11 + NestJS BFF + monorepo).
