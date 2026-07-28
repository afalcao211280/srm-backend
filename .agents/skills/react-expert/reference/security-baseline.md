# Security Baseline — Regras Universais

Aplicável a TODAS as skills de Backend e Infra. Cada skill de domínio adiciona apenas ameaças específicas.

## Princípios

- **Security by Design** — segurança em cada camada, desde o design.
- **Deny-by-default** — negar por padrão, permitir explícito.
- **Zero trust** — nunca confiar, sempre verificar.
- **Least privilege** — mínimo acesso necessário.

## Mapa OWASP Top 10 (2021)

| # | Categoria | Controle nesta baseline |
|---|---|---|
| A01 | Broken Access Control | Authorization + IDOR + Mass Assignment |
| A02 | Cryptographic Failures | Crypto (hashing, at-rest, TLS) |
| A03 | Injection | SQL Injection + XSS + Command/Path |
| A04 | Insecure Design | Princípios + threat modeling por skill |
| A05 | Security Misconfiguration | Headers + CORS + erros genéricos |
| A06 | Vulnerable Components | Dependências auditadas (CVEs) |
| A07 | Auth Failures | Authentication + Session + Rate limiting |
| A08 | Data Integrity Failures | Deserialização + assinatura (JWT verify) |
| A09 | Logging/Monitoring Failures | Logging sem dados sensíveis |
| A10 | SSRF | Allowlist de URLs/hosts |

## Input Validation

- Validar input no boundary (handler/API). Nunca confiar em dado externo.
- Sanitizar antes de usar em SQL, shell, HTML, path.

## Secrets

- NUNCA hardcoded em código. NUNCA em logs/tracebacks.
- Env vars ou secret manager (Vault/AWS SM/Azure KV). Rotação automática.
- `SecretStr`/mascaramento em configs.
- NUNCA versionar `kind: Secret` / `**/secret.yaml` com `data`/`stringData` (Sonar `secrets:S6694`) — External Secrets, Sealed Secrets ou variable group.

## Injection (A03)

- **SQL**: 100% queries parametrizadas. Zero concatenação. sqlc/ORM faz isso — não contornar.
- **XSS**: escapar output. Nunca `innerHTML`/`dangerouslySetInnerHTML` com dado de usuário. CSP como defesa em profundidade.
- **Command/Path**: nunca passar input direto pra shell/filesystem. Allowlist + validação de path (sem `../`).

## Authentication (A07)

- JWT com expiração curta (5–15 min) + refresh tokens. **Verificar a assinatura** — nunca só decodificar o payload.
- OAuth2/OIDC para apps web. API Keys para serviço-serviço (rotação).
- MFA para admin. Senha: mínimo 12 chars + complexidade.
- Hashing de senha: **bcrypt (cost ≥12) ou argon2id**. NUNCA MD5/SHA1.

## Authorization (A01)

- RBAC com least privilege. Verificar permissão em CADA endpoint. Deny-by-default.
- **IDOR**: verificar ownership do recurso (o usuário pode acessar ESTE id?), não só autenticação.
- **Mass assignment**: aceitar apenas campos permitidos explicitamente (allowlist de campos), nunca bind direto do body.

## Cryptography (A02)

- TLS 1.3+ em produção. HSTS habilitado. Nunca HTTP para dados sensíveis.
- **Encryption at rest** para dados sensíveis (DB/storage). Gerenciar chaves em KMS/Vault, com rotação.
- Nunca roll your own crypto. Usar libs padrão auditadas.

## CSRF (A05/A01)

- Cookies `SameSite=Strict|Lax` + `Secure` + `HttpOnly`.
- Token anti-CSRF para operações state-changing em apps com sessão por cookie.

## SSRF (A10)

- Validar URLs fornecidas por usuário contra **allowlist** de hosts.
- Bloquear acesso a rede interna/metadata (169.254.169.254, RFC1918) a partir de fetch server-side.

## File Upload

- Validar MIME type E extensão contra allowlist. Limitar tamanho.
- Armazenar fora do webroot, nome gerado (não confiar no nome do cliente).

## Open Redirect

- Validar destino de redirect contra allowlist de rotas/hosts. Nunca redirecionar para URL crua de input.

## Logging (A09)

- Nunca logar: senhas, tokens, API keys, PII (CPF, cartão).
- Logar: user_id (anonimizado), ação, resultado, timestamp, IP. Detectar/alertar eventos de segurança.

## Headers de Segurança (A05)

- `Content-Security-Policy`
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `Strict-Transport-Security`
- `Referrer-Policy`

## Rate Limiting (A07)

- Rate limit em login e endpoints sensíveis (anti brute-force). Resposta 429 + backoff.

## Dependências (A06)

- Auditar dependências (CVEs) no CI. Atualizar pacotes vulneráveis. Lockfile versionado.

## Anti-padrões de Segurança

- ❌ Secrets em código, `secret.yaml` plaintext no Git, ou env sem vault
- ❌ HTTP para dados sensíveis
- ❌ Roll your own crypto
- ❌ MD5/SHA1 para senha
- ❌ JWT sem verificar assinatura / sem expiração
- ❌ CORS `*` com credentials em produção
- ❌ Bind direto do request body (mass assignment)
- ❌ Fetch de URL de usuário sem allowlist (SSRF)
- ❌ Redirect para URL crua de input
- ❌ Shared service account
- ❌ Brute-force sem rate limit

## Checklist de Segurança

- [ ] Input validado/sanitizado no boundary
- [ ] Queries parametrizadas (zero concatenação)
- [ ] Output escapado (anti-XSS) + CSP
- [ ] Auth em todos os endpoints + assinatura JWT verificada
- [ ] Authorization + ownership (anti-IDOR) verificados
- [ ] Mass assignment bloqueado (allowlist de campos)
- [ ] Senhas com bcrypt/argon2 (nunca MD5/SHA1)
- [ ] HTTPS only + HSTS; dados sensíveis encrypted at rest
- [ ] CSRF (SameSite + token) onde aplicável
- [ ] SSRF mitigado (allowlist de hosts)
- [ ] Upload valida MIME+extensão+tamanho
- [ ] Redirects validados (anti open-redirect)
- [ ] Secrets em vault/env (nunca código nem `secret.yaml` com valores)
- [ ] Logs sem dados sensíveis
- [ ] Rate limiting em login/endpoints sensíveis
- [ ] Headers de segurança presentes
- [ ] Dependências auditadas (CVEs)
