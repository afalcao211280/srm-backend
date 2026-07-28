# Auditoria de Segurança — Processo OWASP

Processo repetível para auditar um codebase. O **baseline compartilhado** (`security-baseline.md`) define os controles; este arquivo define **como auditar** e cobre as técnicas que vão além do baseline.

## Workflow

1. **Identificar o tipo de app** — web app, REST API, SPA, SSR, microsserviço ou misto.
2. **Varredura por prioridade** — começar por CRITICAL, depois HIGH, MEDIUM, LOW.
3. **Mapear contra OWASP Top 10 2021** — ver tabela do baseline.
4. **Reportar achados** — usar o formato de report abaixo.
5. **Remediar** — entregar exemplo de código concreto para cada correção.

## Formato de Report

Para cada achado:

| Campo | Conteúdo |
|---|---|
| **Severity** | CRITICAL \| HIGH \| MEDIUM \| LOW |
| **Categoria** | Nome da regra / item OWASP (ex: A03 Injection) |
| **Arquivo** | Path + linha |
| **Issue** | O que está errado |
| **Impacto** | Consequência de segurança |
| **Fix** | Exemplo de código da correção |

## Severity Quick Reference

**CRITICAL (corrigir imediatamente):** SQL/XSS/Command injection; falta de autenticação em endpoint sensível; secret hardcoded; senha em plaintext; IDOR.

**HIGH (corrigir em breve):** falta de CSRF; requisito de senha fraco; headers de segurança ausentes; CORS permissivo; sessão insegura.

**MEDIUM (corrigir quando possível):** rate limiting ausente; logging incompleto; dependências desatualizadas sem exploit conhecido; validação ausente em campo não-crítico.

**LOW (melhorar):** headers opcionais ausentes; mensagens de erro verbosas (não-prod); parâmetros de crypto subótimos.

## Técnicas além do baseline

Estas vulnerabilidades NÃO estão detalhadas no baseline — verificar explicitamente:

### A01 — Access Control
- **Race condition em operação multi-step** (ex: transferência de saldo): envolver em transação (`db.$transaction()`) para evitar TOCTOU.

### A02 — Cryptographic Failures
- Detalhes de algoritmo/implementação em `encryption.md`.

### A03 — Injection (além de SQL/XSS)
- **NoSQL injection**: operadores `$where`, `$regex`, `$ne` aceitos sem sanitização (MongoDB). Rejeitar objetos onde se espera escalar; nunca passar body cru como filtro.
- **Insecure deserialization**: nunca `eval()` em input; usar `JSON.parse` e validar contra schema. Desserializar só tipos esperados.
- **ReDoS**: regex com backtracking catastrófico sobre input do usuário trava o event loop. Limitar tamanho do input e complexidade do padrão; preferir regex lineares.
- **Path traversal**: input em path sem normalizar (`../`). Usar allowlist + resolver path canônico.

### A04 — Insecure Design
- **Security by obscurity** é anti-padrão: `x-admin-secret === "admin123"` não é controle de acesso. Usar RBAC real verificado no servidor.
- **Threat modeling** por feature: mapear ator → ativo → ameaça → mitigação antes de codar fluxo sensível.

### A05 — Misconfiguration
- **Debug/stack trace** nunca para o cliente: logar server-side, retornar mensagem genérica + 500.
- **CORS regex pitfall**: `/.*\.dominio\.com/` casa `evil-dominio.com`. Ancorar o padrão (`^https://([a-z]+\.)?dominio\.com$`) ou usar allowlist literal.

### A07 — Auth/Session
- **Session fixation**: regenerar o id de sessão **após login** (e após elevação de privilégio). Não reaproveitar o id pré-autenticação.
- **Algorithm confusion (JWT)**: forçar allowlist de algoritmos no verify. Bloqueia `alg: none` e a troca RS256→HS256 (chave pública usada como secret HMAC). Nunca aceitar o `alg` do header sem validar.
- **Signed cookies**: para cookie de sessão não-JWT, assinar com HMAC para detectar adulteração.
- **Rate limiting em camadas**: limites distintos por tier (free/pro/enterprise) e por usuário, não só global por IP.

### A08 — Data Integrity
- **JWT**: verificar a assinatura, nunca só decodificar o payload (`atob(token.split('.')[1])` é bypass).
- **SRI** (Subresource Integrity) para `<script>`/`<link>` de terceiros; **commits assinados** e **imagens de container verificadas** na supply chain.

### A09 — Logging/Monitoring
- **Função de sanitização** central antes de logar: remover/mascarar `password`, `token`, `ssn`, `cpf`, cartão. Logar `email` mascarado (`a***@dominio.com`).
- Logar evento de segurança (login, falha de authz, 429) com user_id, IP e timestamp para detecção.

### A10 — SSRF
- **Bloquear metadata de cloud**: `169.254.169.254` (AWS/GCP/Azure metadata) e faixas RFC1918. Validar host resolvido (não só a string da URL) contra allowlist.

## Padrões de detecção rápida (grep-able)

Sinais para varredura automatizada em CI ou review:

| Vulnerabilidade | Padrão |
|---|---|
| Secret hardcoded | `=\s*["']sk_live`, `password\s*=\s*["']`, `API_KEY\s*=\s*["']` |
| SQL injection | template string com `${...}` dentro de `SELECT ... FROM` |
| XSS | `dangerouslySetInnerHTML`, `.innerHTML =` |
| Command injection | `execSync`, `exec(`, `spawn(... shell: true)` |
| Path traversal | `readFile(...${`, `../` em path de input |
| Crypto fraca | `createHash('md5')`, `sha1`, `sha256(...password)` |
| JWT sem verify | `atob(token.split('.')[1])` |
| Algorithm confusion | `alg: none`, verify sem `algorithms:` |
| CORS inseguro | `Allow-Origin: *` com `Allow-Credentials: true` |
| Cookie inseguro | `Set-Cookie` sem `Secure`/`HttpOnly`/`SameSite` |
| Redirect aberto | `redirect(...searchParams.get(...))` |
| Debug ligado | `DEBUG = true`, stack trace na resposta |
