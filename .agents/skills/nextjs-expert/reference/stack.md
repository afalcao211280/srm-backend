# Next.js Expert — Stack Canônica

## Next.js 14+ / TypeScript 5.x / Node.js 20+ — Versão Canônica

| Categoria | Lib | Versão | Link | Notas |
|-----------|-----|--------|------|-------|
| **Core** | | | | |
| Framework | Next.js | 14.x (App Router) | https://nextjs.org | App Router padrão. Pages Router: legado |
| Linguagem | TypeScript | 5.x | https://typescriptlang.org | `strict: true` obrigatório |
| Runtime | Node.js | 20+ LTS | https://nodejs.org | |
| **Estado** | | | | |
| Estado servidor | TanStack Query | 5.x | https://tanstack.com/query | Client components. Server: React cache |
| Estado global | Zustand | 4.x | https://zustand-demo.pmnd.rs | Simples, sem boilerplate |
| Forms | React Hook Form | 7.x | https://react-hook-form.com | Client components |
| Validação | Zod | 3.x | https://zod.dev | Schema + tipos. Server Actions e forms |
| **Auth** | | | | |
| Autenticação | NextAuth.js (Auth.js) | 5.x (beta) / 4.x | https://authjs.dev | `next-auth@5` para App Router nativo |
| Middleware | Next.js middleware | built-in | — | Proteger rotas no edge |
| **Estilo** | | | | |
| CSS Utilities | Tailwind CSS | 3.x | https://tailwindcss.com | Padrão |
| Componentes | shadcn/ui | latest | https://ui.shadcn.com | Radix UI + Tailwind |
| Primitivos | Radix UI | 1.x | https://www.radix-ui.com | Acessibilidade |
| Animações | Framer Motion | 11.x | https://www.framer.com/motion | Client components |
| **HTTP / Data** | | | | |
| Fetch nativo | fetch (Node 18+) | built-in | — | Server Components: fetch com cache |
| HTTP client | Axios | 1.x | https://axios-http.com | Client components quando necessário |
| ORM / DB | Prisma | 5.x | https://www.prisma.io | App Router nativo. Edge: Prisma Accelerate |
| Cache | next/cache | built-in | — | `revalidatePath`, `revalidateTag`, `unstable_cache` |
| **Roteamento / Nav** | | | | |
| Router | App Router | built-in | — | `app/` directory, layouts, error boundaries |
| Link | next/link | built-in | — | Prefetch automático |
| Navigation | next/navigation | built-in | — | `useRouter`, `usePathname`, `useSearchParams` |
| **Upload / Media** | | | | |
| Imagens | next/image | built-in | — | Otimização automática |
| Upload | uploadthing | 6.x | https://uploadthing.com | Alternativa: next-cloudinary |
| **Internacionalização** | | | | |
| i18n | next-intl | 3.x | https://next-intl-docs.vercel.app | App Router nativo. Alternativa: i18next |
| **Email** | | | | |
| Email | Resend + React Email | 3.x + 2.x | https://resend.com, https://react.email | |
| **Testes** | | | | |
| Unit/Component | Vitest | 2.x | https://vitest.dev | Vite-native, mais rápido |
| DOM testing | React Testing Library | 16.x | https://testing-library.com/react | |
| E2E | Playwright | 1.45+ | https://playwright.dev | |
| **Qualidade** | | | | |
| Linting | ESLint | 9.x | https://eslint.org | `next lint` built-in |
| Formatação | Prettier | 3.x | https://prettier.io | |
| **Deploy** | | | | |
| Hosting | Vercel | — | https://vercel.com | Plataforma nativa Next.js |
| Alternativa | Docker + Node | — | — | `output: 'standalone'` no next.config |

## next.config.ts — Template Canônico

```typescript
import type { NextConfig } from 'next';

const config: NextConfig = {
// Strict mode ativa checks duplos em dev
reactStrictMode: true,

// Output standalone para deploy Docker
// output: 'standalone',

images: {
remotePatterns: [
{
protocol: 'https',
hostname: '**.example.com',
},
],
},

experimental: {
// Para Server Actions tipadas
typedRoutes: true,
},
};

export default config;
```

## tsconfig.json — Template Canônico Next.js (App Router)

```json
{
"compilerOptions": {
"target": "ES2022",
"lib": ["dom", "dom.iterable", "esnext"],
"allowJs": false,
"skipLibCheck": true,
"strict": true,
"noUncheckedIndexedAccess": true,
"noImplicitOverride": true,
"noEmit": true,
"esModuleInterop": true,
"module": "esnext",
"moduleResolution": "bundler",
"resolveJsonModule": true,
"isolatedModules": true,
"jsx": "preserve",
"incremental": true,
"plugins": [{ "name": "next" }],
"paths": {
"@/*": ["./src/*"]
}
},
"include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
"exclude": ["node_modules"]
}
```
