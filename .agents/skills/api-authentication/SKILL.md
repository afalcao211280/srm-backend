---
name: api-authentication
description: >
Autenticação e autorização de APIs: JWT (RS256), OAuth2/OIDC, API keys e sessões.
Use SEMPRE ao proteger APIs, emitir/validar tokens, implementar login, refresh, MFA ou
auth serviço-a-serviço. Alinhado ao security baseline. Acionar em "autenticação",
"login", "JWT", "OAuth", "OIDC", "API key", "token", "refresh token", "proteger API",
"auth serviço".
version: "1.0.0"
category: Crosscutting
keywords:
- authentication
- authorization
- jwt
- oauth2
- oidc
- api-keys
- tokens
requires:
- security-expert
---

# API Authentication (Cross-cutting)

Complementa o security baseline com padrões concretos de auth. Não duplica regras universais.

## Princípios

1. **JWT RS256 em microsserviços** — assinatura assimétrica (chave pública distribuída). HS256 proibido entre serviços.
2. **Expiração curta + refresh** — access token 5–15 min; refresh token rotacionado, revogável.
3. **Sempre verificar assinatura** — nunca só decodificar o payload. Validar `iss`, `aud`, `exp`.
4. **OAuth2/OIDC p/ usuários** — Authorization Code + PKCE em apps. IdP (Keycloak/Azure AD) como fonte.
5. **API keys p/ serviço-serviço** — rotação automática, escopo mínimo, nunca em código.
6. **Deny-by-default** — toda rota exige auth, exceto allowlist explícita.

## Métodos por caso

| Caso | Método | Notas |
|---|---|---|
| App web/SPA | OAuth2 Auth Code + PKCE / OIDC | IdP externo, cookie `HttpOnly`+`SameSite` ou token |
| Serviço → serviço | JWT RS256 ou mTLS | Escopos/claims; mTLS via mesh (Istio) |
| Integração externa | API key | Rotação, rate limit por key, escopo |
| Admin | OAuth2 + **MFA** | Obrigatório |

## Padrões

- **JWT claims**: `sub`, `iss`, `aud`, `exp`, `iat`, `scope`/`roles`. Curto e mínimo.
- **Refresh flow**: rotação de refresh token (detecção de reuse → revoga sessão).
- **Revogação**: blacklist curta (Redis) ou tokens de vida curta + refresh revogável.
- **Storage no cliente**: cookie `HttpOnly`+`Secure`+`SameSite` (web); nunca `localStorage` p/ token sensível.

## Anti-padrões

- ❌ JWT sem verificar assinatura / sem `exp`
- ❌ HS256 entre serviços (segredo compartilhado)
- ❌ Token em `localStorage` (XSS rouba)
- ❌ API key hardcoded / sem rotação
- ❌ Senha sem bcrypt/argon2 (ver baseline)
- ❌ Rota sem auth por esquecimento (falta deny-by-default)

## Checklist

- [ ] Assinatura JWT verificada (RS256 em microsserviços)
- [ ] Access token curto + refresh rotacionado
- [ ] `iss`/`aud`/`exp` validados
- [ ] OAuth2/OIDC com PKCE em apps
- [ ] API keys com rotação e escopo
- [ ] MFA em admin
- [ ] Deny-by-default em todas as rotas
- [ ] Rate limit em endpoints de auth

## Cross-references

- `security-expert` — RBAC/ABAC, OPA, threat model
- `golang-expert`, `nestjs-expert`, `python-expert`, `java-expert` — implementação por stack

## Seguranca (Baseline Compartilhado)

Regras universais (hashing, secrets, TLS, rate limit) em `reference/security-baseline.md`.
