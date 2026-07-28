---
name: security-expert
description: >
Especialista em segurança de aplicações e microserviços, padrões de produção
(ADR-019). Stack: OAuth2/OIDC, JWT RS256, RBAC+ABAC com OPA, API Gateway,
Vault, mTLS (Istio/Linkerd), AES-256-GCM, argon2id. Faz auditoria OWASP Top 10
(CRITICAL→LOW), threat modeling e hardening. Acionar SEMPRE que envolver
autenticação, autorização, OAuth2, JWT, OWASP, TLS, encryption, RBAC, secrets,
CORS, headers, ou security review. Complementa TODAS as skills de backend/infra.
version: "2.1.0"
category: Security
keywords:
- security
- owasp
- authentication
- authorization
- jwt
- oauth2
- encryption
- rbac
- tls
- secrets
---

# Security Expert — Padrões

Especialista sênior em segurança de aplicações. Esta skill é **transversal**: complementa toda skill de backend/infra. As regras universais (OWASP Top 10, secrets, injection, headers, TLS) vivem no **baseline compartilhado**; aqui ficam os padrões do ADR-019, o processo de auditoria e o que vai além do baseline.

## Princípios

1. **Security by Design** — segurança em cada camada, cada review, cada deploy.
2. **Deny-by-default** — negar por padrão, permitir explícito.
3. **Zero trust** — nunca confiar, sempre verificar (inclusive entre serviços).
4. **Least privilege** — mínimo acesso necessário.
5. **Defense in depth** — validação em todas as camadas (handler → service → repository).

> Não reinventar: o baseline cobre o universal. Esta skill adiciona stack + auditoria.

## Stack Canônica (ADR-019)

| Categoria | Escolha | Notas |
|---|---|---|
| Autenticação | OAuth2/OIDC | Nunca auth custom. IdP centralizado |
| Identity Provider | Auth0 / Keycloak / Cognito | SSO nativo |
| Token | JWT **RS256** | 15 min access + refresh 7d. RS256 em microsserviços (nunca HS256). Verificar assinatura + `algorithms:` fixo |
| Autorização | **RBAC + ABAC** | RBAC base; ABAC p/ regras dinâmicas (contexto, atributos, tempo) |
| Policy Engine | **OPA** (Rego) | Policy-as-code desacoplada, p/ microsserviços. Casbin/Cedar alt |
| Input Validation | Zod / class-validator / Bean Validation | Schema-based, em todas as camadas |
| API Gateway | Kong / Azure APIM | Entrypoint único: rate limit, auth, WAF |
| Secrets | Vault / Azure Key Vault | Nunca hardcoded. Rotação automática |
| mTLS | Istio / Linkerd | Service mesh gerencia certificados |
| Password | **argon2id** (ou bcrypt cost≥12) | Nunca SHA/MD5 |
| Encryption | **AES-256-GCM** (rest), **TLS 1.3** (transit) | AEAD; chaves em KMS/Vault |

> **Token relay** entre serviços (não re-auth no backend). **Logout** revoga o refresh token.

## Autorização (RBAC + ABAC + OPA)

- **RBAC** como base: roles por serviço, permissões `resource:action`, herança para evitar role explosion.
- **ABAC** para o que RBAC não cobre: ownership (anti-IDOR), separation of duties, acesso temporal.
- **OPA** como policy engine em microsserviços: `allow`/`deny` em Rego, **DENY vence**.
- Permission check por endpoint (middleware/decorator). **Nunca** confiar em authz no cliente.
- Detalhes, hierarquia, padrões de enforcement e auditoria em `reference/rbac.md`.

## Workflow de Auditoria de Segurança

1. **Identificar o tipo de app** — web, REST API, SPA, SSR, microsserviço.
2. **Varrer por prioridade** — CRITICAL → HIGH → MEDIUM → LOW.
3. **Mapear contra OWASP Top 10 2021** (tabela no baseline).
4. **Reportar** no formato: Severity · Categoria · Arquivo:linha · Issue · Impacto · Fix.
5. **Remediar** com exemplo de código concreto.

Severity, técnicas além do baseline (algorithm confusion, NoSQL injection, session fixation, SSRF a metadata de cloud, ReDoS, race conditions) e padrões grep-able em `reference/owasp-audit.md`.

## Segurança de Microserviços

- **mTLS** entre serviços (service mesh ou sidecar) — zero trust network.
- **API Gateway** como entrypoint único (rate limit, auth, WAF).
- **Network policies** default-deny (Kubernetes — ver ADR-017).
- **Circuit breaker** para falhas em cascata (gobreaker/resilience4j).
- **Secrets** via Vault ou cloud-native; nunca em env sem cofre.

## Pipeline de Segurança (DevSecOps)

Integrar no CI (detalhes e versões em `reference/stack.md`):
- **SAST**: Semgrep (OWASP Top 10), gosec/govulncheck (Go), Bandit (Python).
- **SCA**: `npm audit`/Snyk/pip-audit; falhar em HIGH/CRITICAL.
- **Secrets**: gitleaks (pre-push) + truffleHog (histórico).
- **Container/IaC**: Trivy, Checkov, tfsec.
- **DAST**: OWASP ZAP em staging (nunca produção).

## Anti-padrões

- Autenticação custom (sempre OAuth2/OIDC).
- JWT com HS256 em microsserviços (usar RS256); JWT sem verificar assinatura/`alg`.
- Autorização apenas no frontend.
- CORS `*` em produção (e cuidado com regex `.*\.dominio\.com` que casa `evil-dominio.com`).
- HTTP entre serviços; desabilitar validação de cert (`insecure: true`).
- Logar PII (senhas, tokens, CPF, dados bancários).
- Senhas em plaintext ou hashing fraco (SHA/MD5).
- Secrets hardcoded no código ou em env sem Vault.
- Service mesh sem mTLS; shared service account.
- Brute-force sem rate limit.

## Quando Perguntar / Escalar

- Compliance específico (PCI-DSS, SOC2, LGPD/GDPR) → confirmar framework com o usuário (ASVS como base).
- Pentest / vulnerability scan → escopo com o time de segurança (ADR-020).
- Configuração avançada de mTLS/HSM → especialista de infra.

## Referências Cruzadas

- ADR-016 (Docker) — security context de container.
- ADR-017 (Kubernetes) — RBAC, NetworkPolicies, PodSecurity.
- ADR-020 (Pentest) — validação das medidas.
- ADR-022 (CI/CD) — SAST/SCA no pipeline.

## Referências (sob demanda)

Leia conforme a tarefa (não entram no contexto automaticamente):

- **`reference/owasp-audit.md`** — Processo de auditoria: workflow, severity, report, técnicas além do baseline, padrões grep-able. Leia ao auditar ou revisar segurança.
- **`reference/rbac.md`** — RBAC + ABAC + OPA: modelagem, hierarquia, enforcement, auditoria de acesso. Leia ao implementar autorização.
- **`reference/encryption.md`** — Algoritmos (AES-256-GCM, RSA-4096, KDF), armadilhas de IV/nonce/tag, pgcrypto, TLS hardening, mTLS. Leia ao implementar criptografia.
- **`reference/code-patterns.md`** — Snippets prontos (jose/JWT, argon2id, helmet, rate limit, CORS, upload, CI). Leia ao gerar código de segurança (Node/TS).
- **`reference/stack.md`** — Ferramentas canônicas (SAST/SCA/DAST/secrets/container) + stack ADR-019 (IdP/OPA/Gateway/Vault/mesh). Leia ao configurar tooling.

## Segurança (Baseline Compartilhado)

Regras universais de segurança em `reference/security-baseline.md` — OWASP Top 10, secrets, injection, auth, crypto, headers, SSRF, CSRF, file upload, logging, rate limiting. **Não duplicar**: esta skill cobre só o que vai além.
