# Changelog — typescript-expert

## [2.1.0] - 2026-06-30

### Changed
- Consolidação em `skills-optimizadas` (cruzamento skills/ ⨯ skills-compare/typescript-best-practices ⨯ ADR-008).
- Path do security baseline corrigido para `reference/security-baseline.md`.
- `keywords` corrigidas: de `['typescript-expert']` (nome da skill) para 9 keywords reais de trigger.
- SKILL.md reescrito no padrão Framework 2.1 (Princípios com porquê, Stack canônica, Workflow agentic, Quando Perguntar).
- Zod alinhado ao ADR-008: `3.x` → `4` (stack e SKILL).
- `reference/core.md` reconstruído: promovido o conteúdo rico de `patterns.md` (schema/parse-vs-safeParse, discriminated unions, ts-pattern, generics avançados, event emitter tipado, guards, immutability, teste de tipos, 12 anti-padrões) + seção Type Design (do core antigo) + TOC.

### Added
- Do compare (typescript-best-practices): regra explícita `parse` (trust boundary) vs `safeParse` (input de usuário); nota sobre utilitários `type-fest` (`Opaque`, `PartialDeep`, `ReadonlyDeep`, `SetRequired/Optional`, `Simplify`) no `reference/core.md`.
- Seção de segurança específica (validação no boundary, nunca `as` para dado não confiável, env vars validadas com Zod).

### Removed
- `reference/patterns.md` (conteúdo promovido ao `core.md`; o antigo `core.md` era um resumo subconjunto — consolidado num único ponto de verdade).

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
