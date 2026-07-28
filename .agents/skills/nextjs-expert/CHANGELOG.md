# Changelog — nextjs-expert

## [2.1.2] - 2026-07-01

### Changed
- `reference/best-practices.md` aprofundado (raso → deep reference com TOC): convenção App Router/layouts, Server vs Client, as 4 camadas de cache (Next 15 `fetch` default `no-store`), revalidação (`revalidateTag`/`revalidatePath`), rendering (static/dynamic/streaming/PPR/`generateStaticParams`), Server Actions, Route Handlers vs Actions, middleware, metadata/`next/image`/`next/font`, segurança, anti-padrões e checklist. Cruzamento skills-compare/react-nextjs-development ⨯ ADR-005.
- SKILL.md: seção **Contexto (ADR-005)** (Next 15 + React 19, monorepo pnpm/Turborepo, Zod 4 em `packages/shared`, tRPC 11, Route Handlers só para webhook/HTTP externo); descrição das referências detalhada.

## [2.1.1] - 2026-07-01

### Added
- `reference/best-practices.md` — enriquecimento (checklist condensado, -aligned).

## [2.1.0] - 2026-06-30

### Changed
- Consolidação em skills-optimizadas (skills/ ⨯ skills-compare/ ⨯ ADR): baseline `../../`, keywords/description ajustadas, dedup core/patterns quando idêntico.


## [2.0.0] - 2026-05-28
### Changed
- SKILL.md padronizado no formato Framework 2.0 (caveman style)
- Adicionado versionamento semantico
- Adicionado keywords para trigger matching
- Seguranca referenciada via _shared/security/baseline.md
- Referencias reorganizadas

## [1.0.0] - 2024-01-10
### Added
- Versao inicial
