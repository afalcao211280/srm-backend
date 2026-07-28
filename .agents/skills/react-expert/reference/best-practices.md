# React 19 — Best Practices (deep reference)

Referência aprofundada, -aligned. Leia ao gerar/revisar componentes React.
Conteúdo conceitual e de decisão — o **porquê** por trás de cada regra.
Para código canônico completo (componente padrão, hook com TanStack Query, form
com RHF+Zod, store Zustand, testes), veja `core.md`. Para versões e `tsconfig`,
veja `stack.md`.

## Índice

1. [Componentes e regras dos Hooks](#1-componentes-e-regras-dos-hooks)
2. [Server Components (RSC) vs Client Components](#2-server-components-rsc-vs-client-components)
3. [`use()` — consumir promises e context](#3-use--consumir-promises-e-context)
4. [Actions, useActionState, useOptimistic, useFormStatus](#4-actions-useactionstate-useoptimistic-useformstatus)
5. [Suspense e Error Boundaries](#5-suspense-e-error-boundaries)
6. [Efeitos: quando NÃO usar useEffect](#6-efeitos-quando-nao-usar-useeffect)
7. [Performance e React Compiler](#7-performance-e-react-compiler)
8. [Estado: colocation, server vs client](#8-estado-colocation-server-vs-client)
9. [Formulários e validação (Zod)](#9-formularios-e-validacao-zod)
10. [Anti-padrões](#10-anti-padroes)
11. [Checklist de revisão](#11-checklist-de-revisao)
12. [Cross-references](#12-cross-references)

---

## 1. Componentes e regras dos Hooks

**Princípio**: Function components + hooks. Zero class components (exceto Error
Boundaries legados sem lib — hoje prefira `react-error-boundary`).

- Hooks só no **topo** do componente/hook customizado. Nunca dentro de condição,
loop, `try/catch` ou após um `return` antecipado. O React casa cada hook pela
**ordem de chamada** entre renders; quebrar a ordem corrompe o estado.
- Um hook customizado é qualquer função `useXxx` que chama outros hooks. Extraia
lógica com estado/efeito para `useXxx()` e mantenha o componente declarativo.
- Componente é função pura durante o render: **sem** side effects, mutação de
props/estado, `Math.random()`/`Date.now()` que afete a árvore, ou I/O no corpo.
Efeitos vão em event handlers ou (último caso) `useEffect`.
- Nunca defina um componente **dentro** de outro: a cada render do pai ele vira
um tipo novo, o React desmonta/remonta a subárvore e o estado interno é perdido.
- `key` estável e única em listas. **Nunca** `index` quando a lista pode
reordenar/inserir/remover — o React reconcilia pela `key` e um índice reciclado
associa estado ao item errado. `key` também força remount quando muda de
propósito (`<Form key={userId} />` reinicia o form ao trocar de usuário).

## 2. Server Components (RSC) vs Client Components

RSC é o **default** no App Router (ver `nextjs-expert`). Regras que valem para
qualquer runtime que suporte Server Components:

- Server Component roda no servidor, **não** vai para o bundle do cliente, pode
ser `async` e acessar recursos server-side (DB, secrets, `fs`). Não tem estado,
efeito, nem acesso a browser APIs.
- `"use client"` marca a **fronteira**: o módulo e tudo que ele importa entram no
bundle do cliente. Coloque a diretiva o mais **baixo** possível na árvore (nas
folhas interativas), nunca no layout raiz — senão a página inteira vira client
e você perde o benefício do RSC.
- Server Components podem **compor** Client Components e passar Server Components
como `children`/props (`<ClientShell><ServerContent/></ClientShell>`). O inverso
não vale: um Client Component não pode importar um Server Component, só recebê-lo
via slot.
- Props que cruzam a fronteira server→client precisam ser **serializáveis** (sem
funções, classes, Dates cruas dependendo do runtime). Funções só atravessam se
forem Server Actions (`"use server"`).
- Data fetching primário no servidor (RSC) elimina waterfalls cliente↔servidor e
reduz JS. Client-side fetching (TanStack Query) fica para dados que dependem de
interação, polling, ou estado de sessão do cliente.

## 3. `use()` — consumir promises e context

`use()` é uma API de leitura que pode ser chamada **condicionalmente** e dentro
de loops (diferente dos hooks tradicionais), mas ainda só no render.

- `use(promise)` suspende o componente até a promise resolver e integra com o
Suspense boundary mais próximo. Padrão canônico: o **Server Component** cria a
promise (sem `await`) e a passa como prop; o **Client Component** faz
`const data = use(promise)` sob um `<Suspense>`. Isso permite streaming sem
bloquear o render do servidor.
- Não crie a promise **dentro** do render do client (ela seria recriada a cada
render e nunca resolveria de forma estável). Crie no servidor ou via um cache/
lib de data fetching.
- `use(Context)` substitui `useContext` e pode ser chamado após early returns ou
dentro de `if` — útil para ler context condicionalmente.

## 4. Actions, useActionState, useOptimistic, useFormStatus

"Actions" são funções (sync ou async) que atualizam estado em resposta a
interação. No cliente encapsulam pending/erro/optimistic; com Server Actions
(`"use server"`) executam mutação no servidor.

- **`useActionState(fn, initialState)`** → `[state, formAction, isPending]`.
Ligue `formAction` ao `action` de um `<form>`. Substitui o padrão manual de
`useState` de loading + `try/catch`. A `fn` recebe `(prevState, formData)`.
- **`useFormStatus()`** lê o status do `<form>` ancestral (`pending`, `data`).
Use em um botão de submit filho para desabilitar durante o envio sem prop
drilling do `isPending`. Só funciona **dentro** do `<form>`.
- **`useOptimistic(state, reducer)`** aplica um valor otimista imediato e reverte
automaticamente se a Action falhar. Ideal para likes, toggles, itens de lista —
UI instantânea sem esperar o servidor.
- Validação **sempre** com Zod no boundary da Action (ver seção 9). Retorne erros
no `state` (`{ errors }`) em vez de `throw`, para renderizar mensagens de campo.
- Progressive enhancement: `<form action={...}>` com Actions funciona mesmo antes
do JS hidratar.

## 5. Suspense e Error Boundaries

- `<Suspense fallback={...}>` para code-splitting (`lazy`) e data fetching
(`use()`/RSC streaming). Coloque boundaries **específicos** por região pesada,
com skeleton que reflita o layout final — não um spinner genérico que causa
layout shift.
- Não envolva a app inteira num único Suspense: isso serializa o streaming e
atrasa o TTFB útil. Boundaries granulares deixam o conteúdo pronto aparecer já.
- Error Boundaries capturam erros de render da subárvore. Use por route boundary
e ao redor de widgets de terceiros/features arriscadas. `react-error-boundary`
fornece `<ErrorBoundary onReset resetKeys>` + `useErrorBoundary()` para reset.
- Suspense e Error Boundary são **complementares**: Suspense trata pendência,
Error Boundary trata falha. Combine (`<ErrorBoundary><Suspense>...`).
- Erros em event handlers e código async **não** são capturados por Error
Boundaries — trate localmente (try/catch) ou via `onError` da Action/query.

## 6. Efeitos: quando NÃO usar useEffect

`useEffect` sincroniza com um **sistema externo** (subscription, timer, DOM não
gerenciado, analytics, integração de terceiros). A maioria dos usos comuns é
desnecessária e introduz renders extras e bugs de sincronização.

- **Derivar estado a partir de props/estado** → calcule no render (ou `useMemo`
se caro). Não espelhe em outro `useState` + `useEffect`.
- **Responder a evento do usuário** → coloque a lógica no **event handler**, não
em efeito disparado por mudança de estado.
- **Buscar dados** → RSC/loader ou TanStack Query, nunca `fetch` em `useEffect`
(perde cache, dedupe, retries, race handling, e cria waterfalls).
- **Resetar estado ao mudar prop** → use `key` para remontar, não efeito.
- Quando o efeito for legítimo: **cleanup** sempre (unsubscribe, clearTimeout,
abort) e array de dependências **completo e correto** (não silencie o
`exhaustive-deps`; corrija a causa).

## 7. Performance e React Compiler

Performance é passe **pós-funcionalidade**. Meça com o React DevTools Profiler
antes de otimizar.

- **React Compiler** (React 19) memoiza automaticamente componentes e valores em
build time, reduzindo drasticamente a necessidade de `memo`/`useMemo`/
`useCallback` manuais. Habilite via plugin do bundler/Babel. Requer código que
siga as Regras do React (pureza, sem mutação) — o compiler pula (bail out) o que
não puder provar seguro.
- Com o Compiler ligado, **não** adicione memoização manual preventiva; deixe o
compiler agir e memoize à mão só onde o Profiler apontar hot path que ele não
cobriu.
- Sem o Compiler: `React.memo` só em componente que re-renderiza com props iguais
e é caro; `useMemo`/`useCallback` só quando o valor alimenta uma dependência
memoizada ou um cálculo comprovadamente caro. Memoização prematura adiciona
custo (comparação + memória) sem ganho.
- **Code splitting**: `lazy()` + `<Suspense>` para rotas e componentes pesados/
raros (modais, editores, charts). Reduz o bundle inicial.
- **Listas grandes** (> ~50-100 itens): virtualize com `@tanstack/react-virtual`
para renderizar só o visível. Não virtualize listas pequenas.
- **Evite recriar objetos/arrays** passados como props em hot paths sem
necessidade; e não abstraia demais item de lista (cada componente é mais caro
que nó DOM nativo).

## 8. Estado: colocation, server vs client

- **State colocation**: mantenha o estado o mais **próximo** de quem usa. Levante
(lift) só quando realmente compartilhado. Estado global para o que é local é
anti-padrão (re-renders amplos, acoplamento).
- **Server state** (dados que vivem no backend) → **TanStack Query v5**: cache,
dedupe, `staleTime`/`gcTime`, revalidação, mutations com invalidation. Nunca
`useState` para dados do servidor.
- **Client state** (UI: filtros, toggles, wizard, seleção) → `useState`/
`useReducer` local, ou **Zustand** quando cross-tree. Evite Context para estado
que muda com frequência (todo consumidor re-renderiza) — Context é ótimo para
valores estáveis (tema, auth, i18n).
- Não duplique server state em store client: derive da query. Se precisar de
otimista, use `useOptimistic` ou o mutation cache do TanStack Query.

## 9. Formulários e validação (Zod)

- Formulários complexos: **React Hook Form + `zodResolver`** (uncontrolled,
menos re-renders). Formulários simples/Server Actions: inputs controlados ou
`<form action>` com `useActionState`.
- **Zod é a fonte única** do schema: `z.infer<typeof schema>` gera o tipo, e o
mesmo schema valida no submit. No monorepo, schemas ficam em
`packages/shared` e são reusados por frontend e backend (ver SKILL.md / ADR-005).
- Valide **no boundary**: submit do form e entrada de Server Action/API. Não
confie só em validação de UI — o servidor revalida sempre.
- Mensagens de erro por campo derivadas do resultado do Zod; foco/aria no primeiro
campo inválido (a11y).

## 10. Anti-padrões

| Proibido | Alternativa |
|---|---|
| `useEffect` para derivar estado | Calcular no render / `useMemo` |
| `fetch` em `useEffect` | RSC/loader ou TanStack Query |
| `index` como `key` em lista dinâmica | ID estável |
| Definir componente dentro de componente | Mover para escopo do módulo |
| Estado global para o que é local | State colocation |
| `"use client"` no layout raiz | Diretiva nas folhas interativas |
| Memoização preventiva (sem medir) | React Compiler / medir com Profiler |
| Server state em `useState`/Zustand | TanStack Query |
| `any` | `unknown` + type guards / `z.infer` |
| Silenciar `exhaustive-deps` | Corrigir dependências / remover efeito |
| Único `<Suspense>` na app toda | Boundaries granulares por região |
| `dangerouslySetInnerHTML` sem sanitizar | `DOMPurify.sanitize()` (ver SKILL.md) |

## 11. Checklist de revisão

- [ ] Hooks no topo, sem condicional/loop; ordem estável.
- [ ] Nenhum componente definido dentro de outro.
- [ ] `key` estável (não `index`) em listas dinâmicas.
- [ ] `"use client"` só em folhas; RSC maximizado.
- [ ] Sem `useEffect` para derivar estado ou buscar dados.
- [ ] Efeitos legítimos com cleanup e deps completas.
- [ ] Server state em TanStack Query; client state colocado/Zustand.
- [ ] Suspense + Error Boundary por região pesada / route boundary.
- [ ] Actions com Zod; erros retornados no state, não `throw`.
- [ ] Sem memoização preventiva (Compiler ligado ou Profiler-guiado).
- [ ] Zod como fonte do tipo e da validação (schemas em `packages/shared`).

## 12. Cross-references

- `nextjs-expert` — App Router, RSC no framework, Server Actions, caching.
- `typescript-expert` — tipagem strict, generics, type guards, `z.infer`.
- `css-expert` — Tailwind, design tokens, estilização de componentes.
- `core.md` — código canônico (componente, hook, form, store, testes).
- `stack.md` — versões e `tsconfig` canônico.
- ADR-005 — decisão de stack fullstack (React 19 + Next 15 + tRPC + monorepo).
