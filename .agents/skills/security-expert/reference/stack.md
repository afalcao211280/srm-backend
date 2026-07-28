# Security Expert — Stack Canônica

## Application Security — Versão Canônica

| Categoria | Ferramenta/Lib | Versão | Link | Notas |
|-----------|---------------|--------|------|-------|
| **SAST (Static Analysis)** | | | | |
| Multi-lang | Semgrep | 1.70+ | https://semgrep.dev | Regras customizáveis. OWASP Top 10 |
| Node/JS | ESLint + `eslint-plugin-security` | 3.x | https://github.com/eslint-community/eslint-plugin-security | |
| Python | Bandit | 1.7+ | https://bandit.readthedocs.io | |
| Java | SpotBugs + FindSecBugs | 4.x | https://find-sec-bugs.github.io | |
|.NET | Security Code Scan | 5.x | https://security-code-scan.github.io | |
| Go | govulncheck + gosec | latest | https://github.com/securego/gosec | |
| **SCA (Dependências)** | | | | |
| Node | `npm audit` | built-in | — | `npm audit --audit-level=high` no CI |
| Node alt | Snyk | — | https://snyk.io | Alerts contínuos. Integra com GitHub |
| Python | Safety / pip-audit | latest | https://github.com/pypa/pip-audit | |
| Multi-lang | Trivy (SCA mode) | 0.51+ | https://github.com/aquasecurity/trivy | Escaneia código fonte e containers |
| Multi-lang | OWASP Dependency-Check | 9.x | https://owasp.org/dependency-check | |
| **Secrets Detection** | | | | |
| Pre-commit | gitleaks | 8.x | https://github.com/gitleaks/gitleaks | Integrar como hook pre-push |
| CI/CD | truffleHog | 3.x | https://github.com/trufflesecurity/trufflehog | Escanear histórico git |
| GitHub | Secret scanning | built-in | — | Habilitar em todos os repositórios |
| **DAST (Dynamic Testing)** | | | | |
| Web | OWASP ZAP | 2.14+ | https://zaproxy.org | Scan ativo em staging. Não produção |
| Web alt | Burp Suite | Pro | https://portswigger.net | Para pentests manuais |
| API | Nuclei | 3.x | https://github.com/projectdiscovery/nuclei | Templates para CVEs comuns |
| **Container/IaC Security** | | | | |
| Containers | Trivy | 0.51+ | — | Imagens + Dockerfile + SBOMs |
| IaC | Checkov | 3.x | https://www.checkov.io | Terraform, k8s manifests |
| IaC | tfsec | — | https://github.com/aquasecurity/tfsec | Terraform específico |
| K8s | kubeaudit | — | https://github.com/Shopify/kubeaudit | RBAC + Pod Security |
| **Auth Libraries** | | | | |
| JWT (Node) | jose | 5.x | https://github.com/panva/jose | Web Crypto API. Mais seguro que `jsonwebtoken` |
| JWT (Node) alt | jsonwebtoken | 9.x | https://github.com/auth0/node-jsonwebtoken | Maduro. `RS256` obrigatório |
| Password | bcrypt / argon2 | — | — | argon2id: preferido. bcrypt: amplamente suportado |
| Rate Limiting | express-rate-limit / @nestjs/throttler | — | — | API throttling obrigatório |
| CSRF | csrf-csrf | 3.x | https://github.com/Psifi-Solutions/csrf-csrf | Double-submit cookie pattern |
| Headers | helmet | 7.x | https://helmetjs.github.io | CSP, HSTS, X-Frame-Options, etc. |
| **Auth/Authz/Infra (ADR-019)** | | | | |
| Identity Provider | Auth0 / Keycloak / Cognito | — | https://www.keycloak.org | OAuth2/OIDC centralizado. Token JWT RS256 |
| Policy Engine | OPA (Open Policy Agent) | 0.62+ | https://www.openpolicyagent.org | RBAC+ABAC policy-as-code (Rego) p/ microsserviços |
| API Gateway | Kong / Azure API Management | 3.6+ | https://konghq.com | Entrypoint único: rate limit, auth, WAF, logging |
| Secrets Manager | HashiCorp Vault / Azure Key Vault | 1.16+ | https://www.vaultproject.io | Multi-cloud; rotação automática |
| Service Mesh (mTLS) | Istio / Linkerd | 1.20+ / 2.14+ | https://istio.io | mTLS automático entre serviços |
| **Compliance** | | | | |
| LGPD/GDPR | OWASP ASVS | 4.0.3 | https://owasp.org/asvs | Application Security Verification Standard |
| Checklist | OWASP Top 10 | 2021 | https://owasp.org/Top10 | Review obrigatório em code review |

## OWASP Top 10 (2021) — Checklist Rápido

| # | Vulnerabilidade | Mitigação Principal |
|---|----------------|---------------------|
| A01 | Broken Access Control | RBAC + testes de autorização |
| A02 | Cryptographic Failures | TLS 1.2+, argon2id para senhas, secrets no Key Vault |
| A03 | Injection | Prepared statements, ORM, validação de input |
| A04 | Insecure Design | Threat modeling, Defense in depth |
| A05 | Security Misconfiguration | Helmet, security headers, hardened containers |
| A06 | Vulnerable Components | `npm audit`, Snyk, Dependabot alerts |
| A07 | Auth/Session Failures | MFA, secure cookies, rate limiting em login |
| A08 | Software Integrity Failures | Signed commits, verified container images |
| A09 | Logging/Monitoring Failures | Structured logs, alertas, audit logs |
| A10 | SSRF | Allowlist de URLs externas, não proxy requests |
