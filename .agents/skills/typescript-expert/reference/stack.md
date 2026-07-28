# TypeScript Expert — Stack Canônica

## TypeScript 5.x — Versão Canônica

| Categoria | Lib | Versão | Link | Notas |
|-----------|-----|--------|------|-------|
| **Core** | | | | |
| Linguagem | TypeScript | 5.4+ | https://typescriptlang.org | `strict: true` sempre |
| Runtime Node | tsx / ts-node | tsx 4.x | https://github.com/privatenumber/tsx | Para scripts. `tsx script.ts` |
| Build | tsc / esbuild / tsup | — | https://tsup.egoist.dev | tsup para libs. tsc para type-check only |
| **Validação em Runtime** | | | | |
| Schema / Validation | Zod | 4.x | https://zod.dev | Schema → tipo automático. Padrão (ADR-008) |
| Alternativa | Valibot | 0.31+ | https://valibot.dev | Bundle menor que Zod |
| Decode | io-ts | 2.x | https://gcanti.github.io/io-ts | FP-style, mais verboso |
| **Type Utils** | | | | |
| Utility types | type-fest | 4.x | https://github.com/sindresorhus/type-fest | `Simplify`, `SetOptional`, `ReadonlyDeep`, etc |
| ts-pattern | ts-pattern | 5.x | https://github.com/gvergnaud/ts-pattern | Pattern matching exhaustivo com tipos |
| **Testes de Tipos** | | | | |
| Type testing | tsd | 0.31+ | https://github.com/SamVerschueren/tsd | `expectType<T>()`, `expectError()` |
| Type testing | expect-type | 0.19+ | https://github.com/mmkal/expect-type | Com Vitest / Jest |
| **Qualidade** | | | | |
| Linting | ESLint 9.x + `@typescript-eslint` | 7.x | https://typescript-eslint.io | `recommendedTypeChecked` |
| Formatação | Prettier | 3.x | https://prettier.io | |
| Type check CI | `tsc --noEmit` | built-in | — | Obrigatório no CI |

## tsconfig.json — Template "Strictest" para Libs/Scripts

```json
{
"compilerOptions": {
"target": "ES2022",
"module": "NodeNext",
"moduleResolution": "NodeNext",
"lib": ["ES2022"],
"outDir": "./dist",
"declaration": true,
"declarationMap": true,
"sourceMap": true,
"strict": true,
"noUncheckedIndexedAccess": true,
"noImplicitOverride": true,
"noImplicitReturns": true,
"noFallthroughCasesInSwitch": true,
"exactOptionalPropertyTypes": true,
"useUnknownInCatchVariables": true,
"allowUnusedLabels": false,
"allowUnreachableCode": false,
"forceConsistentCasingInFileNames": true,
"skipLibCheck": false
},
"include": ["src"],
"exclude": ["node_modules", "dist"]
}
```

## Regras ESLint Typescript (flat config)

```javascript
// eslint.config.mjs
import tseslint from 'typescript-eslint';

export default tseslint.config(
...tseslint.configs.recommendedTypeChecked,
{
languageOptions: {
parserOptions: {
projectService: true,
tsconfigRootDir: import.meta.dirname,
},
},
rules: {
'@typescript-eslint/no-explicit-any': 'error',
'@typescript-eslint/no-unsafe-assignment': 'error',
'@typescript-eslint/no-unsafe-return': 'error',
'@typescript-eslint/consistent-type-imports': ['error', { prefer: 'type-imports' }],
'@typescript-eslint/no-import-type-side-effects': 'error',
'@typescript-eslint/switch-exhaustiveness-check': 'error',
},
},
);
```
