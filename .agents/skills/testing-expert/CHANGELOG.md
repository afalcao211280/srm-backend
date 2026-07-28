# Changelog — testing-expert

## [2.2.0] - 2026-07-16

### Added
- `SKILL.md` — gate **Go → Sonar**: `go test -coverprofile=coverage.out` + `sonar.go.coverage.reportPaths` (sem relatório = coverage 0% e QG falha).
- `reference/patterns.md` — após seção Go: comando canônico de coverage para CI/Sonar.

## [2.1.0] - 2026-06-30

### Changed
- Consolidação em `skills-optimizadas` (cruzamento skills/ ⨯ skills-compare/ ⨯ ADR-021).
- Path do security baseline corrigido para `reference/security-baseline.md`.
- SKILL.md reescrito e densificado: princípios com porquê, stack canônica por linguagem (tabela), pirâmide ASCII, quality gates de CI, mutation/flaky alinhados ao ADR-021.
- `description` expandida (genérica → O QUE + STACK + QUANDO, estilo pushy).
- `keywords` de `['testing-expert']` para 8 keywords reais de trigger.
- `reference/stack.md`: pirâmide corrigida 60-70/20-30/5-10 → **70/20/10** (alinha ADR-021); adicionados Stryker/PITest/Pact/Locust.

### Added
- `reference/bdd.md` — BDD/Gherkin condensado (origem: skills-compare/cucumber-gherkin), adaptado ao ecossistema (cucumber-js/Playwright, pytest-bdd) em vez de Ruby/Capybara. Inclui Three Amigos e Example Mapping.
- `reference/core.md` reescrito — princípios profundos (AAA, test doubles, property-based), TDD red-green-refactor + escolas (origem: skills-compare/tdd-orchestrator), geração assistida de testes + análise de gaps (origem: skills-compare/unit-testing-test-generate), legacy/characterization.
- `reference/patterns.md` — adicionados snippets Go (table-driven + testify) e Java (JUnit 5 + Mockito) para cobrir as 4 linguagens do ADR-021; índice no topo.

### Removed
- Duplicata: `reference/core.md` era cópia byte-a-byte de `reference/patterns.md` (mesmos snippets). `core.md` reaproveitado para princípios; snippets ficam só em `patterns.md`.

### Decisões de divergência ADR-021
- Nenhuma contradição dura. ADR-021 era mais rico (versões por linguagem, mutation score 60%, branch coverage, fail_on_decrease, Locust, Pact) → skill alinhada ao ADR.

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
