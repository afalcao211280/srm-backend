# React Expert — Stack Canônica

## React 18+ / TypeScript 5.x — Versão Canônica Frontend

| Categoria | Lib | Versão | Link | Notas |
|-----------|-----|--------|------|-------|
| **Core** | | | | |
| UI Framework | React | 18.x | https://react.dev | Sempre. `react@18` |
| Linguagem | TypeScript | 5.x | https://typescriptlang.org | `strict: true` obrigatório |
| Build | Vite | 5.x | https://vitejs.dev | Para SPAs. Next.js para SSR. |
| **Estado** | | | | |
| Estado local | useState, useReducer | built-in | — | Para estado de componente |
| Estado global | Zustand | 4.x | https://zustand-demo.pmnd.rs | Simples, sem boilerplate |
| Estado servidor | TanStack Query | 5.x | https://tanstack.com/query | `useQuery`, `useMutation` |
| Forms | React Hook Form | 7.x | https://react-hook-form.com | Performance. Sem re-renders excessivos |
| Validação | Zod | 3.x | https://zod.dev | Schema + tipos TypeScript |
| **Roteamento** | | | | |
| Router | React Router | 6.x | https://reactrouter.com | SPAs. Next.js usa routing nativo |
| **Estilo** | | | | |
| Utilitários CSS | Tailwind CSS | 3.x | https://tailwindcss.com | Padrão frontend |
| Componentes | shadcn/ui | latest | https://ui.shadcn.com | Sobre Radix UI + Tailwind |
| Primitivos | Radix UI | 1.x | https://www.radix-ui.com | Acessibilidade nativa |
| Animações | Framer Motion | 11.x | https://www.framer.com/motion | Quando necessário |
| **HTTP** | | | | |
| HTTP client | Axios | 1.x | https://axios-http.com | Com TanStack Query. Interceptors |
| Alternativa | fetch nativo | built-in | — | Para casos simples |
| **Testes** | | | | |
| Unit/Component | Vitest | 2.x | https://vitest.dev | Mais rápido que Jest, Vite-native |
| DOM testing | React Testing Library | 16.x | https://testing-library.com/react | Queries accessible. Comportamento > implementação |
| E2E | Playwright | 1.45+ | https://playwright.dev | Multi-browser |
| **Qualidade** | | | | |
| Linting | ESLint | 9.x | https://eslint.org | Flat config + `@typescript-eslint` |
| Formatação | Prettier | 3.x | https://prettier.io | `.prettierrc` |
| Hooks lint | eslint-plugin-react-hooks | 5.x | — | `rules-of-hooks`, `exhaustive-deps` |

## tsconfig.json — Template Canônico React

```json
{
"compilerOptions": {
"target": "ES2022",
"lib": ["ES2022", "DOM", "DOM.Iterable"],
"module": "ESNext",
"moduleResolution": "bundler",
"jsx": "react-jsx",
"strict": true,
"noUncheckedIndexedAccess": true,
"noImplicitOverride": true,
"allowImportingTsExtensions": true,
"isolatedModules": true,
"noEmit": true,
"paths": {
"@/*": ["./src/*"]
}
}
}
```
