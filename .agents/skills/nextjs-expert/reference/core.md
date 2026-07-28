# Next.js Expert — Padrões de Código (App Router)

## 1. Estrutura de Diretórios (App Router)

```
src/
├── app/
│ ├── layout.tsx # Root layout (html, body, providers)
│ ├── page.tsx # Home route
│ ├── loading.tsx # Suspense fallback global
│ ├── error.tsx # Error boundary global ('use client')
│ ├── not-found.tsx # 404 handler
│ ├── (auth)/ # Route group — não aparece na URL
│ │ ├── login/page.tsx
│ │ └── register/page.tsx
│ └── dashboard/
│ ├── layout.tsx # Layout aninhado com sidebar
│ ├── page.tsx # /dashboard
│ └── [id]/
│ ├── page.tsx # /dashboard/[id]
│ └── loading.tsx # Suspense por rota
├── components/
│ ├── ui/ # shadcn/ui — nunca editar diretamente
│ └── features/ # Componentes de domínio
├── lib/
│ ├── auth.ts # NextAuth config
│ ├── db.ts # Prisma client singleton
│ └── utils.ts # cn() e helpers
├── server/ # Código exclusivamente server-side
│ ├── actions/ # Server Actions
│ └── queries/ # Server queries (sem cache público)
└── types/
└── index.ts
```

---

## 2. Server Component — Fetch com Cache

```tsx
// src/app/dashboard/page.tsx
import { Suspense } from 'react';
import { UserList } from '@/components/features/user/UserList';
import { getUsersQuery } from '@/server/queries/users';

// Força revalidação a cada 60 segundos
export const revalidate = 60;

export default async function DashboardPage() {
// fetch direto — Next.js deduplica automaticamente por request
const users = await getUsersQuery();

return (
<main>
<h1 className="text-2xl font-bold mb-4">Dashboard</h1>
<Suspense fallback={<div>Carregando usuários...</div>}>
<UserList users={users} />
</Suspense>
</main>
);
}
```

```typescript
// src/server/queries/users.ts
import { unstable_cache } from 'next/cache';
import { db } from '@/lib/db';

export const getUsersQuery = unstable_cache(
async () => {
return db.user.findMany({
select: { id: true, name: true, email: true, active: true },
orderBy: { createdAt: 'desc' },
});
},
['users-list'],
{ revalidate: 60, tags: ['users'] },
);
```

---

## 3. Server Action com Zod

```typescript
// src/server/actions/users.ts
'use server';

import { revalidateTag } from 'next/cache';
import { redirect } from 'next/navigation';
import { z } from 'zod';
import { db } from '@/lib/db';

const createUserSchema = z.object({
name: z.string().min(2).max(100),
email: z.string().email(),
});

export type ActionState = {
error?: string;
fieldErrors?: Record<string, string[]>;
success?: boolean;
};

export async function createUserAction(
_prev: ActionState,
formData: FormData,
): Promise<ActionState> {
const raw = Object.fromEntries(formData);
const parsed = createUserSchema.safeParse(raw);

if (!parsed.success) {
return { fieldErrors: parsed.error.flatten().fieldErrors };
}

const existing = await db.user.findUnique({ where: { email: parsed.data.email } });
if (existing) {
return { error: `Email '${parsed.data.email}' já cadastrado` };
}

await db.user.create({ data: parsed.data });
revalidateTag('users');
redirect('/dashboard');
}
```

---

## 4. Form com useActionState (React 19 / Next.js 14+)

```tsx
// src/components/features/user/CreateUserForm.tsx
'use client';

import { useActionState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { createUserAction, type ActionState } from '@/server/actions/users';

const initialState: ActionState = {};

export function CreateUserForm() {
const [state, formAction, isPending] = useActionState(createUserAction, initialState);

return (
<form action={formAction} className="space-y-4">
<div>
<Label htmlFor="name">Nome</Label>
<Input id="name" name="name" placeholder="João Silva" />
{state.fieldErrors?.name && (
<p className="text-sm text-destructive mt-1">{state.fieldErrors.name[0]}</p>
)}
</div>
<div>
<Label htmlFor="email">Email</Label>
<Input id="email" name="email" type="email" placeholder="joao@example.com" />
{state.fieldErrors?.email && (
<p className="text-sm text-destructive mt-1">{state.fieldErrors.email[0]}</p>
)}
</div>
{state.error && (
<p className="text-sm text-destructive">{state.error}</p>
)}
<Button type="submit" disabled={isPending}>
{isPending? 'Criando...': 'Criar Usuário'}
</Button>
</form>
);
}
```

---

## 5. Prisma Client Singleton

```typescript
// src/lib/db.ts
import { PrismaClient } from '@prisma/client';

const globalForPrisma = globalThis as unknown as { prisma: PrismaClient };

export const db =
globalForPrisma.prisma??
new PrismaClient({
log: process.env.NODE_ENV === 'development'? ['query', 'error', 'warn']: ['error'],
});

if (process.env.NODE_ENV!== 'production') globalForPrisma.prisma = db;
```

---

## 6. NextAuth.js v5 Config (App Router)

```typescript
// src/lib/auth.ts
import NextAuth from 'next-auth';
import Credentials from 'next-auth/providers/credentials';
import { db } from './db';
import bcrypt from 'bcryptjs';

export const { handlers, auth, signIn, signOut } = NextAuth({
providers: [
Credentials({
credentials: {
email: { label: 'Email', type: 'email' },
password: { label: 'Senha', type: 'password' },
},
async authorize(credentials) {
if (!credentials?.email ||!credentials?.password) return null;
const user = await db.user.findUnique({
where: { email: credentials.email as string },
});
if (!user ||!user.passwordHash) return null;
const valid = await bcrypt.compare(credentials.password as string, user.passwordHash);
return valid? { id: user.id, name: user.name, email: user.email }: null;
},
}),
],
callbacks: {
jwt({ token, user }) {
if (user) token.id = user.id;
return token;
},
session({ session, token }) {
if (token.id) session.user.id = token.id as string;
return session;
},
},
});
```

```typescript
// src/middleware.ts
import { auth } from './lib/auth';

export default auth((req) => {
if (!req.auth && req.nextUrl.pathname.startsWith('/dashboard')) {
return Response.redirect(new URL('/login', req.url));
}
});

export const config = {
matcher: ['/dashboard/:path*'],
};
```

---

## 7. Metadata Dinâmica

```tsx
// src/app/dashboard/[id]/page.tsx
import type { Metadata } from 'next';
import { getUsersQuery } from '@/server/queries/users';

type Props = { params: Promise<{ id: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
const { id } = await params;
const users = await getUsersQuery();
const user = users.find((u) => u.id === id);

return {
title: user? `${user.name} | Dashboard`: 'Usuário não encontrado',
};
}

export default async function UserDetailPage({ params }: Props) {
const { id } = await params;
//...
}
```

---

## 8. Route Handler (API)

```typescript
// src/app/api/users/route.ts
import { NextResponse } from 'next/server';
import { auth } from '@/lib/auth';
import { db } from '@/lib/db';

export async function GET() {
const session = await auth();
if (!session) {
return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
}

const users = await db.user.findMany({
select: { id: true, name: true, email: true },
orderBy: { createdAt: 'desc' },
});

return NextResponse.json(users);
}
```

---

## 9. Providers (Client Component Wrapper)

```tsx
// src/app/providers.tsx
'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState } from 'react';

export function Providers({ children }: { children: React.ReactNode }) {
const [queryClient] = useState(
() =>
new QueryClient({
defaultOptions: {
queries: { staleTime: 60 * 1000 },
},
}),
);

return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}
```

```tsx
// src/app/layout.tsx
import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import { Providers } from './providers';
import './globals.css';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
title: ' App',
description: 'Plataforma ',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
return (
<html lang="pt-BR">
<body className={inter.className}>
<Providers>{children}</Providers>
</body>
</html>
);
}
```

---

## 10. Dockerfile (output: standalone)

```dockerfile
FROM node:20-alpine AS deps
WORKDIR /app
COPY package*.json./
RUN npm ci

FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules./node_modules
COPY..
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
RUN addgroup --system --gid 1001 nodejs && adduser --system --uid 1001 nextjs
COPY --from=builder /app/public./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static./.next/static
USER nextjs
EXPOSE 3000
ENV PORT=3000
CMD ["node", "server.js"]
```

---

## Anti-patterns

### ❌ 'use client' desnecessário em Server Components
**Problema:** O desenvolvedor adiciona `'use client'` em componentes que não usam estado, efeitos ou APIs de browser.
**Por quê evitar:** Transforma o componente em Client Component, enviando todo o JavaScript ao browser, perdendo os benefícios de SSR (sem bundle para o cliente, acesso direto ao banco/FS, zero hidratação). Qualquer componente filho do subárvore também passa a ser client-side.
**Solução:**
```tsx
// Errado — 'use client' sem necessidade
'use client';
export default async function UserList({ users }: { users: User[] }) {
return <ul>{users.map(u => <li key={u.id}>{u.name}</li>)}</ul>;
}

// Correto — Server Component puro (sem diretiva)
export default function UserList({ users }: { users: User[] }) {
return <ul>{users.map(u => <li key={u.id}>{u.name}</li>)}</ul>;
}

// 'use client' apenas onde realmente necessário (useState, useEffect, event handlers)
'use client';
export function SearchInput({ onSearch }: { onSearch: (q: string) => void }) {
const [value, setValue] = useState('');
return <input value={value} onChange={e => { setValue(e.target.value); onSearch(e.target.value); }} />;
}
```

---

### ❌ Buscar dados em Client Components quando Server Component bastaria
**Problema:** O desenvolvedor faz fetch de dados com `useEffect` ou TanStack Query em um Client Component para dados que não precisam de interatividade.
**Por quê evitar:** O Client Component aguarda hidratação, faz uma requisição adicional após o HTML inicial ser entregue (waterfall), expõe endpoints desnecessariamente ao cliente e aumenta o bundle size.
**Solução:**
```tsx
// Errado — fetch no cliente para dados estáticos
'use client';
export default function DashboardPage() {
const { data } = useQuery({ queryKey: ['users'], queryFn: fetchUsers });
return <UserTable users={data?? []} />;
}

// Correto — fetch no Server Component, zero JS enviado para dados
export default async function DashboardPage() {
const users = await getUsersQuery(); // roda no servidor
return <UserTable users={users} />; // UserTable pode ser Server Component também
}
```

---

### ❌ Não usar `unstable_cache` para fetches repetidos no servidor
**Problema:** O desenvolvedor faz fetch de dados com `db.query()` ou `fetch()` diretamente em Server Components sem nenhuma camada de cache.
**Por quê evitar:** Em rotas com múltiplos Server Components que precisam dos mesmos dados, cada um executa a query independentemente. Isso multiplica hits ao banco/API por request.
**Solução:**
```typescript
// Errado — sem cache, N queries por request
export default async function Page() {
const users = await db.user.findMany(); // executado a cada request
}

// Correto — wrappear com unstable_cache
import { unstable_cache } from 'next/cache';

export const getUsersQuery = unstable_cache(
async () => db.user.findMany({ select: { id: true, name: true } }),
['users-list'],
{ revalidate: 60, tags: ['users'] },
);
// Chamadas múltiplas na mesma request são deduplicadas automaticamente
```

---

### ❌ Importar módulos server-only em Client Components
**Problema:** O desenvolvedor importa código que usa `process.env` secretos, Prisma, `fs`, ou outras APIs Node.js dentro de um Client Component ou utilitário compartilhado.
**Por quê evitar:** O bundler do Next.js inclui o módulo no bundle do browser, expondo segredos de ambiente, credenciais de banco e lógica de negócio sensível ao cliente. Pode causar erros de runtime no browser.
**Solução:**
```typescript
// Errado — Prisma importado em arquivo sem garantia de contexto
import { db } from '@/lib/db'; // se importado por Client Component → vaza para o bundle

// Correto — adicione 'server-only' como guard no módulo
// src/lib/db.ts
import 'server-only'; // lança erro de build se importado por Client Component
import { PrismaClient } from '@prisma/client';
//...

// Correto — organize em src/server/ (convenção clara de separação)
// src/server/queries/users.ts — nunca importado por Client Components
```

---

### ❌ Não usar `loading.tsx` e `error.tsx` para streaming
**Problema:** O desenvolvedor deixa rotas sem os arquivos `loading.tsx` e `error.tsx` no App Router.
**Por quê evitar:** Sem `loading.tsx`, o Next.js bloqueia a rota até todos os dados serem resolvidos, eliminando o streaming progressivo. Sem `error.tsx`, um erro em qualquer Server Component quebra a página inteira sem recovery.
**Solução:**
```
src/app/dashboard/
├── loading.tsx ← Suspense boundary automático para a rota
├── error.tsx ← Error boundary automático ('use client')
└── page.tsx

// loading.tsx
export default function DashboardLoading() {
return <DashboardSkeleton />;
}

// error.tsx
'use client';
export default function DashboardError({ error, reset }: { error: Error; reset: () => void }) {
return (
<div>
<p>Erro: {error.message}</p>
<button onClick={reset}>Tentar novamente</button>
</div>
);
}
```

---

### ❌ Acessar cookies/headers sem `await` no Next.js 15
**Problema:** O desenvolvedor chama `cookies()` ou `headers()` diretamente sem `await` em Server Components.
**Por quê evitar:** No Next.js 15, as APIs `cookies()`, `headers()`, `params` e `searchParams` são Promises assíncronas. Acessá-las sem `await` retorna a Promise em si (não o valor), causando bugs silenciosos difíceis de rastrear.
**Solução:**
```typescript
// Errado — Next.js 15 — retorna Promise, não o valor
import { cookies } from 'next/headers';
export default async function Page() {
const cookieStore = cookies(); // ← sem await, é a Promise
const token = cookieStore.get('token'); // undefined ou erro
}

// Correto — sempre await
export default async function Page() {
const cookieStore = await cookies();
const token = cookieStore.get('token')?.value;
}

// Correto para params em rotas dinâmicas (Next.js 15)
export default async function UserPage({ params }: { params: Promise<{ id: string }> }) {
const { id } = await params;
}
```

---

### ❌ Rotas dinâmicas sem `generateStaticParams`
**Problema:** O desenvolvedor cria rotas `[id]` sem implementar `generateStaticParams` em conteúdo que poderia ser pré-renderizado.
**Por quê evitar:** Sem `generateStaticParams`, rotas dinâmicas são renderizadas sob demanda (SSR) a cada request, aumentando latência e custo de servidor. Páginas de produto, post de blog, perfis públicos se beneficiam enormemente de SSG.
**Solução:**
```typescript
// Errado — renderizado dinamicamente a cada request
export default async function ProductPage({ params }: { params: Promise<{ slug: string }> }) {
const { slug } = await params;
const product = await getProduct(slug);
return <ProductDetail product={product} />;
}

// Correto — pré-renderize as páginas conhecidas em build time
export async function generateStaticParams() {
const products = await getAllProductSlugs();
return products.map(p => ({ slug: p.slug }));
}

export const dynamicParams = true; // permite novas slugs após o build (ISR)
```

---

### ❌ Waterfall de dados com múltiplos awaits sequenciais
**Problema:** O desenvolvedor usa múltiplos `await` sequenciais para buscar dados independentes em um Server Component.
**Por quê evitar:** Cada `await` bloqueia o seguinte. Se a primeira query demora 200ms e a segunda 300ms, o total é 500ms. Com `Promise.all`, ambas rodam em paralelo e o total é ~300ms.
**Solução:**
```typescript
// Errado — waterfall: 200ms + 300ms = 500ms total
export default async function DashboardPage() {
const users = await getUsersQuery(); // 200ms
const stats = await getDashboardStats(); // 300ms (só começa após users)
return <Dashboard users={users} stats={stats} />;
}

// Correto — paralelo: max(200ms, 300ms) = 300ms total
export default async function DashboardPage() {
const [users, stats] = await Promise.all([
getUsersQuery(),
getDashboardStats(),
]);
return <Dashboard users={users} stats={stats} />;
}
```

---

### ❌ Cookies sem flags `httpOnly` e `secure`
**Problema:** O desenvolvedor armazena tokens de sessão ou dados sensíveis em cookies sem configurar as flags de segurança adequadas.
**Por quê evitar:** Cookies sem `httpOnly` são acessíveis via `document.cookie` e vulneráveis a ataques XSS. Sem `secure`, o cookie é transmitido em conexões HTTP (não apenas HTTPS), expondo tokens em redes inseguras.
**Solução:**
```typescript
// Errado — cookie acessível por JavaScript e em HTTP
cookies().set('session', token);

// Correto — flags de segurança obrigatórias para dados sensíveis
(await cookies()).set('session', token, {
httpOnly: true, // inacessível via document.cookie
secure: true, // apenas HTTPS
sameSite: 'lax', // proteção CSRF
path: '/',
maxAge: 60 * 60 * 24 * 7, // 7 dias em segundos
});
```

---

### ❌ Tags `<img>` nativas em vez de `next/image`
**Problema:** O desenvolvedor usa a tag `<img>` HTML nativa para exibir imagens na aplicação.
**Por quê evitar:** A tag `<img>` nativa não faz otimização automática de formato (WebP/AVIF), não aplica lazy loading, não previne Cumulative Layout Shift (CLS) e não redimensiona a imagem para o viewport. O impacto direto é no LCP (Largest Contentful Paint) e na nota do Core Web Vitals.
**Solução:**
```tsx
// Errado — sem otimização, sem lazy loading, CLS provável
<img src="/hero.jpg" alt="Hero" />

// Correto — next/image otimiza automaticamente
import Image from 'next/image';

// Para imagens de tamanho conhecido
<Image
src="/hero.jpg"
alt="Hero"
width={1200}
height={630}
priority // apenas para imagens above-the-fold (LCP)
/>

// Para imagens responsivas que preenchem o container
<div className="relative h-64">
<Image src={user.avatarUrl} alt={user.name} fill className="object-cover rounded" />
</div>
```
