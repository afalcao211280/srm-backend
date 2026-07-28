# Changelog — docker-expert

## [2.1.0] - 2026-06-30

### Changed
- Consolidação em `skills-optimizadas` (cruzamento skills/ ⨯ skills-compare/docker-development ⨯ ADR-016).
- Path do security baseline corrigido para `reference/security-baseline.md`.
- `keywords` corrigido de `['docker-expert']` (só o nome) para 7 termos reais de trigger (docker, dockerfile, docker-compose, container, buildx, trivy, multi-stage).
- `description` reescrita (pushy, ≤600 chars) cobrindo stack completa e gatilhos.
- SKILL.md alinhado ao ADR-016: princípios com o porquê, stack canônica com versões, imagens base por linguagem, estrutura de projeto, workflow agentic, anti-padrões detalhados, "Quando Perguntar".
- Versões alinhadas ao ADR-016/stack: Docker 26+/Podman 4+, Trivy 0.51+, Cosign 2+, ACR como registry padrão.

### Decisões de divergência ADR-016 (caso a caso)
- **Runtime**: Docker **ou** Podman documentados sem default (Podman é drop-in rootless — escolha por ambiente).
- **Compose**: reconciliado — stack usa Compose Spec via CLI v2 e arquivo `compose.yaml` (ADR menciona "v3+" referindo-se ao schema Compose Spec, não à CLI legada v1). Sem contradição dura.

### Added
- `reference/core.md` — novas seções condensadas do compare `docker-development`: decision tree de base image + tabela de tamanhos, compose (isolamento de rede, override dev/prod, worker+fila, logging), HEALTHCHECK por serviço, gatilhos proativos, troubleshooting de compose.

### Removed
- `reference/patterns.md` — duplicata byte-a-byte de `reference/core.md`.

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
