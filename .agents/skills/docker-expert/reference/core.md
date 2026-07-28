# Docker Expert — Padrões de Código

## 1. Dockerfile Multi-stage — Node.js

```dockerfile
# syntax=docker/dockerfile:1
FROM node:20-alpine AS deps
WORKDIR /app
COPY package*.json ./
# Cache mount: evita re-download de packages se package.json não mudou
RUN --mount=type=cache,target=/root/.npm \
    npm ci --only=production

FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci
COPY . .
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production

# Non-root user
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 --ingroup nodejs nodeapp

COPY --from=deps --chown=nodeapp:nodejs /app/node_modules ./node_modules
COPY --from=builder --chown=nodeapp:nodejs /app/dist ./dist
COPY --chown=nodeapp:nodejs package.json ./

USER nodeapp
EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD wget -qO- http://localhost:3000/health || exit 1

CMD ["node", "dist/main.js"]
```

---

## 2. Dockerfile Multi-stage — Go

```dockerfile
# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/server ./cmd/server

# Distroless: sem shell, sem package manager — superfície de ataque mínima
FROM gcr.io/distroless/static-debian12:nonroot AS runner
COPY --from=builder /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
```

---

## 3. Dockerfile Multi-stage — Python (FastAPI)

```dockerfile
# syntax=docker/dockerfile:1
FROM python:3.12-slim AS builder
WORKDIR /app
COPY requirements.txt .
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --user --no-cache-dir -r requirements.txt

FROM python:3.12-slim AS runner
WORKDIR /app
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1
ENV PATH=/root/.local/bin:$PATH

RUN groupadd --system app && useradd --system --gid app --no-create-home app

COPY --from=builder /root/.local /root/.local
COPY --chown=app:app . .

USER app
EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8000/health')"

CMD ["uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8000", "--workers", "2"]
```

---

## 4. Docker Compose — Stack de Desenvolvimento

```yaml
# compose.yaml
name: myapp

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
      target: builder  # usa estágio de dev com devDependencies
    volumes:
      - .:/app
      - /app/node_modules  # anonymous volume previne override do node_modules
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=development
    env_file:
      - .env.local
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    develop:
      watch:
        - action: sync
          path: ./src
          target: /app/src
        - action: rebuild
          path: package.json

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: dev
      POSTGRES_PASSWORD: dev
      POSTGRES_DB: myapp_dev
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql:ro
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U dev -d myapp_dev"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes --requirepass devpassword
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "--no-auth-warning", "-a", "devpassword", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
  redis_data:
```

---

## 5. Multi-platform Build (CI)

```bash
# Setup buildx para multi-platform
docker buildx create --name multiarch --driver docker-container --use
docker buildx inspect --bootstrap

# Build e push para AMD64 + ARM64
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag myacr.azurecr.io/api:$IMAGE_TAG \
  --cache-from type=registry,ref=myacr.azurecr.io/api:cache \
  --cache-to type=registry,ref=myacr.azurecr.io/api:cache,mode=max \
  --push \
  .
```

---

## 6. Hadolint — Lint no CI

```yaml
# .github/workflows/lint.yml
- name: Lint Dockerfile
  uses: hadolint/hadolint-action@v3.1.0
  with:
    dockerfile: Dockerfile
    failure-threshold: warning
    # ignore regras específicas se necessário:
    # ignore: DL3008,DL3009

# Ou localmente:
# hadolint Dockerfile
# docker run --rm -i hadolint/hadolint < Dockerfile
```

---

## 7. Trivy — Scan de Vulnerabilidades (CI)

```yaml
# .github/workflows/security.yml
- name: Scan image with Trivy
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: myacr.azurecr.io/api:${{ github.sha }}
    format: sarif
    output: trivy-results.sarif
    exit-code: '1'
    severity: CRITICAL,HIGH
    ignore-unfixed: true

- name: Upload Trivy results to GitHub Security
  uses: github/codeql-action/upload-sarif@v3
  if: always()
  with:
    sarif_file: trivy-results.sarif
```

---

## Anti-patterns

### ❌ Rodar container como root
**Problema:** Não especificar `USER` no Dockerfile — o processo executa como root (uid 0) dentro do container.
**Por quê evitar:** Se houver exploração de vulnerabilidade no app, o atacante tem root no container (e potencialmente no host se houver misconfiguration do runtime).
**Solução:**
```dockerfile
# Crie um usuário não-privilegiado
RUN addgroup --system --gid 1001 appgroup && \
    adduser --system --uid 1001 --ingroup appgroup appuser
USER appuser  # sempre no final, antes de CMD/ENTRYPOINT
```

### ❌ Secrets em instruções ENV ou ARG
**Problema:** Usar `ENV DB_PASSWORD=secret` ou `ARG API_KEY=secret` no Dockerfile.
**Por quê evitar:** Variáveis de ambiente ficam visíveis em `docker inspect`, nos metadados da imagem e em qualquer registry onde a imagem é publicada.
**Solução:**
```dockerfile
# Use --secret do BuildKit (não persiste na imagem)
RUN --mount=type=secret,id=api_key \
    API_KEY=$(cat /run/secrets/api_key) ./configure.sh

# Build com: docker buildx build --secret id=api_key,env=API_KEY .
# Em runtime: passe via env do orchestrador (Kubernetes Secret, ACA secret)
```

### ❌ COPY . . como primeira instrução de build
**Problema:** Copiar todo o código-fonte antes de instalar dependências.
**Por quê evitar:** Qualquer mudança em qualquer arquivo invalida o cache do `COPY` — `npm install` / `go mod download` são executados em toda build mesmo sem mudança de deps.
**Solução:**
```dockerfile
# Copie SOMENTE os manifestos de dependência primeiro
COPY package.json package-lock.json ./
RUN npm ci --only=production   # cache reutilizado se apenas código mudou

COPY . .                        # invalida cache aqui — npm ci já foi cacheado
RUN npm run build
```

### ❌ Usar tag :latest em produção
**Problema:** Referenciar imagens com `FROM node:latest` ou `image: myapp:latest` em manifests de produção.
**Por quê evitar:** `latest` não é imutável — um redeploy ou rebuild pode usar uma imagem diferente da testada; impossibilita rastreabilidade de versões.
**Solução:**
```dockerfile
FROM node:22.12-alpine3.21  # versão exata + alpine pinado
# Em produção, use digest SHA256 imutável
FROM node@sha256:abc123...
```

### ❌ Processo PID 1 sem signal handling
**Problema:** O processo da aplicação é PID 1 mas não trata SIGTERM.
**Por quê evitar:** `docker stop` envia SIGTERM e aguarda `--stop-timeout` (padrão 10s) — se o app não captura SIGTERM, o Docker envia SIGKILL e a aplicação termina abruptamente (conexões abertas, dados em buffer perdidos).
**Solução:**
```dockerfile
# Use tini como init process para propagar sinais corretamente
RUN apk add --no-cache tini
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["./servidor"]
# Ou implemente graceful shutdown no código (recomendado para Go/Node)
```

### ❌ .dockerignore ausente
**Problema:** Não criar `.dockerignore` no projeto.
**Por quê evitar:** `COPY . .` inclui `node_modules` (GB de dados), `.git` (histórico completo), `.env` (secrets), `dist/` — build contexto gigante e dados sensíveis na imagem.
**Solução:**
```dockerignore
# .dockerignore
node_modules/
.git/
.env*
*.local
dist/
coverage/
.DS_Store
```

### ❌ Multi-stage sem otimização de layers
**Problema:** Instalar ferramentas de build na imagem final (compilador, ferramentas de dev) sem multi-stage.
**Por quê evitar:** Imagem final carrega todo o toolchain de build — GBs desnecessários, maior superfície de ataque e tempo de pull mais lento.
**Solução:**
```dockerfile
FROM golang:1.26-alpine AS builder
RUN go build -o /app/server ./cmd/server

# Stage final: apenas o binário, sem Go runtime
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/server /server
ENTRYPOINT ["/server"]
# Imagem final: ~5MB vs ~400MB do builder
```

---

## 9. Escolha de Base Image — Decision Tree

```
Binário compilado (Go, Rust, C)?
├── Sim → distroless/static ou scratch (só CA certs)
└── Não
    ├── Precisa de shell p/ debug?
    │   ├── Sim → variante alpine (ex: node:20-alpine)
    │   └── Não → variante distroless
    ├── Precisa glibc (não musl)?
    │   ├── Sim → variante slim (ex: python:3.12-slim — C ext: numpy/pandas)
    │   └── Não → variante alpine
    └── Precisa de pacotes de SO?
        ├── Muitos → debian-slim
        └── Poucos → alpine + apk add
```

**Tamanhos aproximados** (guia de escolha):

| Base | Tamanho | Uso |
|---|---|---|
| `scratch` | 0 MB | Binário estático (Go, Rust) |
| `distroless/static` | ~2 MB | Binário estático + CA certs |
| `alpine` | ~7 MB | Linux mínimo com shell |
| `distroless/base` | ~20 MB | Binário dinâmico (C/C++) |
| `debian-slim` | ~80 MB | Quando precisa glibc + apt |
| `python:3.12-slim` | ~130 MB | Python em produção |
| `node:20-alpine` | ~130 MB | Node.js em produção |
| `golang:1.2x` / `node:20` / `python:3.12` | 800 MB–1 GB | **Só estágio de build — nunca runtime de produção** |

> Alpine usa musl (não glibc): evitar com Python que tenha C extensions; nesse caso, slim.

---

## 10. Compose — Isolamento de Rede

```yaml
# compose.yaml — backend inacessível de fora
services:
  nginx:
    image: nginx:alpine
    ports: ["80:80", "443:443"]
    networks: [frontend]
  app:
    build: .
    networks: [frontend, backend]
  db:
    image: postgres:16-alpine
    networks: [backend]   # só o app alcança
  redis:
    image: redis:7-alpine
    networks: [backend]

networks:
  frontend:
  backend:
    internal: true        # sem acesso externo — limita movimento lateral
```

---

## 11. Compose — Override Dev/Prod

```yaml
# compose.yaml (base — production-like)
services:
  app:
    build: .
    ports: ["3000:3000"]
    restart: unless-stopped

# compose.override.yaml (dev — auto-carregado por `docker compose up`)
services:
  app:
    build: { target: development }
    volumes:
      - .:/app                 # bind mount p/ hot reload
      - /app/node_modules      # preserva node_modules do container
    environment: [NODE_ENV=development, DEBUG=true]
    ports: ["9229:9229"]       # debug port
    restart: "no"
```

```bash
docker compose up                          # dev (carrega override)
docker compose -f compose.yaml up -d       # prod (ignora override)
```

---

## 12. Compose — Worker + Fila

```yaml
services:
  api:
    build: { context: ., target: runtime }
    command: uvicorn main:app --host 0.0.0.0 --port 8000
    ports: ["8000:8000"]
    depends_on:
      rabbitmq: { condition: service_healthy }
  worker:
    build: { context: ., target: runtime }
    command: celery -A tasks worker --loglevel=info
    depends_on:
      rabbitmq: { condition: service_healthy }
  rabbitmq:
    image: rabbitmq:3.13-management-alpine
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "check_running"]
      interval: 10s
      timeout: 5s
      retries: 5
```

---

## 13. Compose — Logging (evita disco cheio)

```yaml
services:
  app:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"     # default do Docker é ILIMITADO — servidor enche o disco
        max-file: "3"       # rotaciona automaticamente
```

---

## 14. HEALTHCHECK por Serviço

```dockerfile
# HTTP (com curl)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8000/health || exit 1

# HTTP sem curl (distroless/minimal — wget)
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8000/health || exit 1

# TCP
HEALTHCHECK CMD nc -z localhost 8000 || exit 1
# PostgreSQL
HEALTHCHECK CMD pg_isready -U postgres || exit 1
# Redis
HEALTHCHECK CMD redis-cli ping | grep PONG || exit 1
```

---

## 15. Gatilhos Proativos (sinalizar sem ser pedido)

- Dockerfile com `:latest` → sugerir pin de versão + digest.
- Sem `.dockerignore` → criar (mínimo: `.git`, `node_modules`, `__pycache__`, `.env`).
- `COPY . .` antes de instalar deps → reordenar (deps primeiro).
- Rodando como root → adicionar `USER` (sem exceção em produção).
- Secrets em `ENV`/`ARG` → BuildKit `--secret`.
- Imagem > 1 GB → exigir multi-stage.
- Sem `HEALTHCHECK` → adicionar (orquestrador precisa p/ ciclo de vida).
- `apt-get` sem `rm -rf /var/lib/apt/lists/*` no mesmo `RUN` → limpar na mesma camada.

---

## 16. Troubleshooting (Compose)

| Sintoma | Causa provável | Correção |
|---|---|---|
| Container sai na hora | CMD/ENTRYPOINT crasha, env var faltando | `docker compose logs <svc>` |
| Porta em uso | Outro processo na mesma porta | Trocar porta do host: `"3001:3000"` |
| Permissão negada em volume | User do container não é dono do path | Casar UID/GID ou usar named volume |
| Cache de build não funciona | `COPY . .` invalida cedo | Copiar deps antes do código |
| `depends_on` não espera | Sem `condition: service_healthy` | Adicionar healthcheck + condition |
| Container OOM killed | Sem limite de memória ou baixo demais | Ajustar `mem_limit` / `deploy.resources.limits` |
