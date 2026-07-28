# Security Expert — Padrões de Código

## 1. Autenticação Segura — JWT com jose (Node/TS)

```typescript
// src/lib/auth.ts — usando jose (Web Crypto API nativa, mais seguro que jsonwebtoken)
import { SignJWT, jwtVerify, type JWTPayload } from 'jose';

const JWT_SECRET = new TextEncoder().encode(process.env['JWT_SECRET']);
const JWT_ISSUER = 'api';
const JWT_AUDIENCE = 'app';

interface TokenPayload extends JWTPayload {
userId: string;
role: string;
}

export async function signToken(userId: string, role: string): Promise<string> {
return new SignJWT({ userId, role })
.setProtectedHeader({ alg: 'HS256' })
.setIssuer(JWT_ISSUER)
.setAudience(JWT_AUDIENCE)
.setIssuedAt()
.setExpirationTime('1h')
.sign(JWT_SECRET);
}

export async function verifyToken(token: string): Promise<TokenPayload> {
const { payload } = await jwtVerify(token, JWT_SECRET, {
algorithms: ['HS256'], // allowlist explícita — bloqueia alg:none e troca RS256->HS256
issuer: JWT_ISSUER,
audience: JWT_AUDIENCE,
});
return payload as TokenPayload;
}
```

> **Microsserviços (ADR-019)**: use **RS256** (chave pública verificável), nunca HS256. HS256 só em monolito com segredo único. Sempre fixe `algorithms:` no verify.

---

## 2. Hash de Senhas — argon2id

```typescript
// src/lib/password.ts
import { hash, verify, argon2id } from 'argon2';

// Parâmetros recomendados pelo OWASP (2024)
const ARGON2_OPTIONS = {
type: argon2id,
memoryCost: 64 * 1024, // 64 MB
timeCost: 3, // 3 iterações
parallelism: 4, // 4 threads
};

export async function hashPassword(password: string): Promise<string> {
return hash(password, ARGON2_OPTIONS);
}

export async function verifyPassword(hash: string, password: string): Promise<boolean> {
return verify(hash, password);
}
```

---

## 3. Rate Limiting + Helmet (Express/NestJS)

```typescript
// Express — segurança básica
import helmet from 'helmet';
import rateLimit from 'express-rate-limit';

// Headers de segurança
app.use(helmet({
contentSecurityPolicy: {
directives: {
defaultSrc: ["'self'"],
scriptSrc: ["'self'"],
styleSrc: ["'self'", "'unsafe-inline'"],
imgSrc: ["'self'", 'data:', 'https:'],
connectSrc: ["'self'"],
fontSrc: ["'self'"],
objectSrc: ["'none'"],
mediaSrc: ["'self'"],
frameSrc: ["'none'"],
},
},
hsts: { maxAge: 31536000, includeSubDomains: true, preload: true },
}));

// Rate limiting global
const globalLimiter = rateLimit({
windowMs: 15 * 60 * 1000, // 15 minutos
max: 100,
standardHeaders: 'draft-7',
legacyHeaders: false,
message: { error: 'Too many requests, please try again later.' },
});

// Rate limiting mais restrito para auth
const authLimiter = rateLimit({
windowMs: 15 * 60 * 1000,
max: 5, // 5 tentativas de login por 15 minutos
skipSuccessfulRequests: true,
message: { error: 'Too many login attempts.' },
});

app.use('/api/', globalLimiter);
app.use('/api/auth/login', authLimiter);
```

---

## 4. Validação de Input — Proteção Contra Injection

```typescript
// Sempre validar e sanitizar inputs externos
import { z } from 'zod';

// Zod rejeita tipos inesperados automaticamente
const searchSchema = z.object({
query: z.string().max(200).transform(s => s.trim()),
page: z.coerce.number().int().min(1).max(1000),
category: z.enum(['electronics', 'clothing', 'food']).optional(),
});

// Nunca interpolar inputs em SQL — usar prepared statements
// ERRADO (SQL injection):
// `SELECT * FROM users WHERE email = '${email}'`

// CORRETO (Prisma/TypeORM fazem automaticamente):
const user = await db.user.findFirst({ where: { email } });

// CORRETO (SQL bruto com binding):
const result = await db.$queryRaw`SELECT * FROM users WHERE email = ${email}`;

// Para regex patterns recebidos do usuário — limitar complexidade
import { JSDOM } from 'jsdom';
import DOMPurify from 'dompurify';

const window = new JSDOM('').window;
const purify = DOMPurify(window);

export function sanitizeHtml(dirty: string): string {
return purify.sanitize(dirty, { ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'p'] });
}
```

---

## 5. Segurança em Uploads de Arquivo

```typescript
// src/lib/file-upload.ts
import path from 'path';
import crypto from 'crypto';
import { z } from 'zod';

const ALLOWED_MIME_TYPES = new Set([
'image/jpeg', 'image/png', 'image/webp', 'image/gif',
'application/pdf',
]);

const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB

const fileSchema = z.object({
mimetype: z.string().refine(
(type) => ALLOWED_MIME_TYPES.has(type),
'Tipo de arquivo não permitido',
),
size: z.number().max(MAX_FILE_SIZE, 'Arquivo muito grande (máx. 10MB)'),
originalname: z.string(),
});

export function validateFile(file: Express.Multer.File): void {
fileSchema.parse({ mimetype: file.mimetype, size: file.size, originalname: file.originalname });
}

// Gerar nome aleatório — NUNCA usar nome original do usuário
export function generateSecureFilename(mimetype: string): string {
const ext = mimetype.split('/')[1]?? 'bin';
return `${crypto.randomUUID()}.${ext}`;
}

// NUNCA servir uploads do mesmo domínio da aplicação (XSS via SVG/HTML)
// Upload → storage privado (Azure Blob / S3) → URL assinada com TTL
```

---

## 6. CORS — Configuração Segura

```typescript
// NUNCA usar origin: '*' em produção
// NUNCA enviar credentials com origin: '*'
import cors from 'cors';

const ALLOWED_ORIGINS = (process.env['ALLOWED_ORIGINS']?? '')
.split(',')
.map(o => o.trim())
.filter(Boolean);

app.use(cors({
origin: (origin, callback) => {
if (!origin || ALLOWED_ORIGINS.includes(origin)) {
callback(null, true);
} else {
callback(new Error(`CORS: origin '${origin}' not allowed`));
}
},
credentials: true,
methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'],
allowedHeaders: ['Content-Type', 'Authorization'],
maxAge: 86400, // Cache preflight por 24h
}));
```

---

## 7. Pipeline de Segurança — CI/CD

```yaml
#.github/workflows/security.yml
name: Security Scan

on: [push, pull_request]

jobs:
sast:
runs-on: ubuntu-latest
steps:
- uses: actions/checkout@v4

- name: Run Semgrep
uses: semgrep/semgrep-action@v1
with:
config: >-
p/owasp-top-ten
p/nodejs
p/typescript
auditOn: push

sca:
runs-on: ubuntu-latest
steps:
- uses: actions/checkout@v4
- uses: actions/setup-node@v4
with: { node-version: '20' }
- run: npm ci
- run: npm audit --audit-level=high

secrets:
runs-on: ubuntu-latest
steps:
- uses: actions/checkout@v4
with: { fetch-depth: 0 } # histórico completo
- name: Gitleaks scan
uses: gitleaks/gitleaks-action@v2

container:
runs-on: ubuntu-latest
steps:
- uses: actions/checkout@v4
- name: Trivy vulnerability scan
uses: aquasecurity/trivy-action@master
with:
scan-type: fs
scan-ref:.
severity: CRITICAL,HIGH
exit-code: '1'
```

---

## 8. Checklist de Code Review — Segurança

```
# Verificar em TODA PR que envolva código de backend/API:

[ ] Inputs externos são validados com schema (Zod/class-validator)?
[ ] Queries SQL usam prepared statements ou ORM?
[ ] Senhas são hasheadas com argon2id ou bcrypt (custo ≥ 12)?
[ ] JWTs usam algoritmo seguro (HS256/RS256) e verificam exp/iss/aud?
[ ] Dados sensíveis (tokens, senhas) não aparecem em logs?
[ ] Headers de segurança estão configurados (Helmet)?
[ ] Rate limiting está em endpoints de auth e operações custosas?
[ ] Uploads validam tipo MIME e tamanho máximo?
[ ] Secrets não estão hardcoded (use Key Vault/Secrets Manager)?
[ ] CORS permite apenas origens explícitas?
[ ] Erros retornados ao cliente não expõem stack traces ou dados internos?
```

---

## Anti-patterns

### ❌ JWT armazenado em localStorage
**Problema:** Salvar o token JWT em `localStorage` ou `sessionStorage` no frontend.
**Por quê evitar:** Qualquer XSS tem acesso completo ao localStorage — um script injetado rouba todos os tokens de todos os usuários. XSS é a vulnerabilidade #3 do OWASP Top 10.
**Solução:**
```typescript
// Use httpOnly + Secure cookie — inacessível ao JavaScript
res.cookie('access_token', jwt, {
httpOnly: true, // JavaScript não consegue ler
secure: true, // apenas HTTPS
sameSite: 'strict', // proteção CSRF
maxAge: 15 * 60 * 1000, // 15 minutos
});
```

### ❌ Não validar o algoritmo do JWT
**Problema:** Usar `jwt.verify(token, secret)` sem especificar o algoritmo permitido.
**Por quê evitar:** Algorithm confusion attacks: um atacante forja `alg: none` (sem assinatura) ou muda de RS256 para HS256 (usando a chave pública como secret HMAC) — ambos bypassam verificação se a biblioteca não forçar o algoritmo.
**Solução:**
```typescript
// jose 5.x — especifique algoritmo explicitamente
const { payload } = await jwtVerify(token, secret, {
algorithms: ['HS256'], // whitelist explícita — nunca deixe vazio
issuer: 'auth',
audience: 'api',
});
```

### ❌ Hashar senhas com MD5, SHA1 ou bcrypt em blobs grandes
**Problema:** Usar `md5(senha)`, `sha256(senha)` ou `bcrypt` com senhas > 72 bytes.
**Por quê evitar:** MD5/SHA1 são quebráveis em segundos com rainbow tables; bcrypt trunca silenciosamente em 72 bytes — senhas longas idênticas nos primeiros 72 chars são aceitas como iguais.
**Solução:**
```typescript
import { hash, verify } from '@node-rs/argon2';

// Hash com argon2id (OWASP 2024: 64MB memória, 3 iterações)
const hashed = await hash(senha, {
algorithm: Algorithm.Argon2id,
memoryCost: 65536, // 64 MB
timeCost: 3,
parallelism: 4,
});
```

### ❌ SQL construído com concatenação de strings
**Problema:** Montar queries SQL concatenando valores de entrada do usuário.
**Por quê evitar:** SQL Injection é a vulnerabilidade #3 do OWASP — permite leitura de dados, bypass de auth e em alguns DBs execução de comandos no SO.
**Solução:**
```typescript
// ❌ NUNCA
const q = `SELECT * FROM users WHERE email = '${req.body.email}'`;

// ✅ SEMPRE: parâmetros bindados
const user = await db.query('SELECT * FROM users WHERE email = $1', [req.body.email]);
// ORMs com query builder também são seguros se usados corretamente (sem raw string interpolation)
```

### ❌ CORS wildcard com credenciais
**Problema:** Configurar `Access-Control-Allow-Origin: *` em endpoints que aceitam cookies ou headers de autorização.
**Por quê evitar:** Browsers bloqueiam `withCredentials: true` com wildcard — mas outros clients aceitam; além disso, `*` expõe APIs internas a qualquer origem maliciosa.
**Solução:**
```typescript
app.use(cors({
origin: ['https://app.example.com.br', 'https://admin.example.com.br'], // whitelist explícita
credentials: true,
methods: ['GET', 'POST', 'PUT', 'DELETE'],
}));
```

### ❌ Logar dados sensíveis
**Problema:** Incluir senhas, tokens, CPF, números de cartão ou PII em mensagens de log.
**Por quê evitar:** Logs são coletados em múltiplos sistemas (Elasticsearch, Splunk, S3) com acesso mais amplo que o banco — violação de LGPD e possível comprometimento de credenciais.
**Solução:**
```typescript
// ❌ NUNCA
logger.info('login attempt', { email, password, token });

// ✅ Mascare ou omita campos sensíveis
logger.info('login attempt', {
email: email.replace(/(?<=.).(?=[^@]*@)/, '*'), // a***@domain.com
// password: OMITIDO
// token: OMITIDO
});
```

### ❌ Dependências com CVEs conhecidos sem monitoramento
**Problema:** Não executar `npm audit`, `govulncheck` ou `pip-audit` no CI.
**Por quê evitar:** 60%+ dos ataques exploram vulnerabilidades conhecidas em dependências transitivas — sem monitoramento, o projeto fica exposto por meses sem saber.
**Solução:**
```yaml
# GitHub Actions: falha o CI se houver vuln HIGH ou CRITICAL
- name: Security audit
run: |
npm audit --audit-level=high
npx audit-ci --high
# Azure DevOps: equivalente com Snyk ou OWASP Dependency Check
```

### ❌ Uploads sem validação de tipo real (magic bytes)
**Problema:** Validar tipo de arquivo apenas pelo Content-Type header ou extensão.
**Por quê evitar:** Um atacante muda o Content-Type para `image/jpeg` e envia um arquivo PHP/JS malicioso — o servidor executa o arquivo se não validar o conteúdo real.
**Solução:**
```typescript
import { fileTypeFromBuffer } from 'file-type';

const buffer = await file.arrayBuffer();
const type = await fileTypeFromBuffer(Buffer.from(buffer));

const ALLOWED = ['image/jpeg', 'image/png', 'image/webp', 'application/pdf'];
if (!type ||!ALLOWED.includes(type.mime)) {
throw new Error('Tipo de arquivo não permitido');
}
// Renomeie o arquivo com UUID — nunca use o nome original
const safeName = `${crypto.randomUUID()}.${type.ext}`;
```
