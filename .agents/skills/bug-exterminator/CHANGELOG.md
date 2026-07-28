# Changelog — bug-exterminator

## [1.1.0] - 2026-06-30

### Changed
- Consolidação em `skills-optimizadas` (sem ADR, sem compare — otimização + correções de conformidade).
- `description` reduzida de 823 → ≤600 chars (piso do schema), mantendo o estilo pushy e os gatilhos.
- `keywords` de `['bug-exterminator']` para 8 keywords reais de trigger (bug-hunting, race-condition, goroutine-leak, production-reliability, concurrency, timeout, n-plus-one, qa).
- Path do security baseline corrigido para `reference/security-baseline.md`.
- Seção de referências reformatada como progressive disclosure (ponteiro de quando ler cada uma por fase).

### Added
- Seção **Cross-references** (security-expert, testing-expert, cicd-expert).
- Seção **Segurança (Baseline Compartilhado)** padronizada, com a ponte bug-de-confiabilidade → vulnerabilidade.

### Preservado
- `reference/` próprio intacto: `silent-bug-catalog.md`, `environment-recon.md`, `verification-playbook.md`.
- Metodologia em 5 fases e escala de severidade sem alteração de substância.

## [1.0.0] - 2026-06-16
### Added
- Versao inicial da skill de caca a bugs silenciosos de producao
- Metodologia em 5 fases (recon → mapear risco → cacar por categoria → verificar adversarialmente → reportar/corrigir sob aprovacao)
- Escala de severidade (Critica/Alta/Media/Baixa) orientada a impacto em producao
- Seguranca referenciada via _shared/security/baseline.md (encaminha a security-expert)
- Referencias: silent-bug-catalog.md, environment-recon.md, verification-playbook.md
- Integrada ao workflow /pick-card na etapa de QA
