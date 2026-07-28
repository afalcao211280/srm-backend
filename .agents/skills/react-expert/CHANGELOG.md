# Changelog — react-expert

## [2.1.2] - 2026-07-01

### Changed
- `reference/best-practices.md` aprofundado (raso → deep reference com TOC): regras de hooks, RSC vs Client Components, `use()`, Actions/`useActionState`/`useOptimistic`/`useFormStatus`, Suspense + Error Boundaries, efeitos (quando NÃO usar `useEffect`), performance + React Compiler, estado (colocation, server vs client), formulários + Zod, anti-padrões e checklist. Cruzamento skills-compare/react-nextjs-development ⨯ ADR-005.
- SKILL.md: seção **Contexto (ADR-005)** (React 19 base, monorepo pnpm/Turborepo, Zod 4 em `packages/shared`, tRPC 11); descrição das referências (`best-practices`/`core`/`stack`) detalhada.

## [2.1.1] - 2026-07-01

### Added
- `reference/best-practices.md` — enriquecimento (checklist condensado, -aligned).

## [2.1.0] - 2026-06-30

### Changed
- Consolidação em skills-optimizadas (skills/ ⨯ skills-compare/ ⨯ ADR): baseline `../../`, keywords/description ajustadas, dedup core/patterns quando idêntico.


## [2.0.0] - 2024-06-15
### Changed
- SKILL.md padronizado (caveman style)
- Adicionado versionamento semantico e keywords
- Seguranca movida para _shared/security/baseline.md
- Referencias reorganizadas

## [1.0.0] - 2024-01-10
### Added
- Versao inicial
- Stack: React 18+, TypeScript 5, TanStack Query, Zustand
