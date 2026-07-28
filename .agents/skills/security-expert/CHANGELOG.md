# Changelog — security-expert

## [2.1.0] - 2026-06-30

### Changed
- Consolidação em `skills-optimizadas` (cruzamento skills/ ⨯ skills-compare/ ⨯ ADR-019).
- SKILL.md alinhado ao ADR-019: JWT **RS256** (nunca HS256 em microsserviços), **OPA** como policy engine, API Gateway (Kong/Azure APIM), Vault, mTLS (Istio/Linkerd) explícitos.
- `keywords` corrigidas (era `['security-expert']`) → 10 keywords reais de trigger.
- `requires` removido (skill transversal; antes `['security-expert']`, auto-referência inválida).
- Path do baseline corrigido para `reference/security-baseline.md`.
- `description` reescrita (stack + auditoria + gatilhos), dentro do limite.

### Added
- Seção **Workflow de Auditoria** (CRITICAL→LOW + formato de report) no SKILL.md.
- `reference/owasp-audit.md` — processo de auditoria + técnicas além do baseline (algorithm confusion, NoSQL injection, session fixation, SSRF a metadata de cloud, ReDoS, race conditions) + padrões grep-able. Origem: skills-compare/owasp-security-check.
- `reference/rbac.md` — RBAC + ABAC + OPA (modelagem, hierarquia, enforcement, auditoria). Origem: skills-compare/access-control-rbac, alinhado a OPA do ADR-019.
- `reference/encryption.md` — algoritmos (AES-256-GCM, RSA-4096, KDF), armadilhas de IV/nonce/tag, pgcrypto, TLS hardening. Origem: skills-compare/data-encryption.

### Removed / Fixed
- `reference/core.md` removido (duplicata exata de `patterns.md` no source).
- `reference/patterns.md` → renomeado `reference/code-patterns.md`; exemplo JWT corrigido (`algorithms:` fixo, nota RS256 para microsserviços).
- Stack ADR-019 (IdP/OPA/Gateway/Vault/mesh) adicionada a `reference/stack.md`.
- AGENTS.md atualizado para o formato `skills-optimizadas`.

### Descartado (já no baseline compartilhado)
- Regras 100% redundantes do owasp-security-check: authentication-failures, broken-access-control, cryptographic-failures, csrf-protection, secrets-management, security-headers, vulnerable-dependencies.
- Best practices genéricas de data-encryption e access-control-rbac já cobertas (TLS, no-MD5/SHA1, IDOR ownership, deny-by-default).

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
