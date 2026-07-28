# Changelog — golang-expert

## [2.2.0] - 2026-07-16

### Added
- Princípio e checklist **limiares SonarQube** no `SKILL.md`: `go:S138` (≤30 linhas), `go:S107` (≤3 params → `Options`/`Params`), `go:S3776` (cognitive ≤15), `go:S1135` (TODO com issue), coverage ≥80%.
- `reference/patterns.md` — seção **16. Qualidade / SonarQube** com padrão Options/Params e anti-padrões que bloqueiam Quality Gate.

### Changed
- `.golangci.yml` canônico em `reference/project-scaffold.md` e `reference/stack.md`: `revive argument-limit: 3` (antes 6; alinha universal ≤3 e Sonar S107); linters `funlen` (lines: 30) e `gocognit` (min-complexity: 16 → flaga >15) para espelhar S138/S3776 localmente antes do CI.

## [2.1.0] - 2026-06-30

### Changed
- Consolidação em `skills-optimizadas` (cruzamento skills/ ⨯ skills-compare/ ⨯ ADR-001).
- Path do security baseline corrigido para `reference/security-baseline.md`.

### Decisões de divergência ADR-001 (caso a caso)
- **DI manual fixo** mantido (skill vence). ADR-001 atualizado: Google Wire removido.
- **sqlc+tern fixo** mantido (skill vence). ADR-001 atualizado: Ent removido.

### Added
- Seção **Modernização (Go 1.26+)** no SKILL.md (origem: golang-modernize; inline por ser específico).

## [2.0.0] - 2024-06-15

### Changed
- SKILL.md reduzido de 155 para ~100 linhas (caveman style)
- Conteudo granularizado em 15 arquivos de referencia
- Seguranca movida para `_shared/security/baseline.md`
- Adicionado versionamento semantico
- Adicionado keywords para trigger matching
- Removidos exemplos de codigo do SKILL.md (movidos para reference/<lib>.md)

### Added
- Arquivos de referencia sob demanda: core, stack, workflow, gin, huma, sqlc, tern, pgx, redis, rabbitmq, kafka, mongodb, minio, resty, gotenberg, tracing, metrics
- CHANGELOG.md para rastreamento de versoes

## [1.0.0] - 2024-01-10

### Added
- Versao inicial com stack 2026
- Padroes: Clean Architecture, DI manual, sqlc+tern
- Observabilidade: OpenTelemetry + Prometheus
- Testes: testify + testcontainers-go
