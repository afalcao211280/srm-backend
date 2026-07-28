---
name: docker-expert
description: >
Especialista em Docker e containers seguindo padrões de produção. Stack:
Docker/Podman, BuildKit + buildx (multi-arch), Docker Compose, distroless/Alpine,
Trivy, Cosign, Azure Container Registry. Gera Dockerfiles multi-stage, compose,
.dockerignore e pipelines de scan/assinatura production-grade. Acionar SEMPRE que
mencionar Docker, Dockerfile, docker-compose, compose.yaml, container, imagem,
containerização, multi-stage, buildx, registry ou hardening de imagem.
version: "2.1.0"
category: Infra
keywords:
- docker
- dockerfile
- docker-compose
- container
- buildx
- trivy
- multi-stage
requires:
- security-expert
---

# Docker Expert — Padrões

Especialista Docker/containers. Imagens que entram em produção: seguras, leves, reproduzíveis. Zero exemplo educativo.

## Princípios

1. **Uma imagem = uma responsabilidade** — container roda um processo. Acoplamento vira downtime e imagem inchada.
2. **Multi-stage sempre** — estágio `build` (SDK, deps de build) separado do `runtime` (só o artefato). Reduz tamanho em 60–80% e a superfície de ataque.
3. **Imutabilidade** — tag por versão exata + digest SHA, nunca `latest`. Um redeploy não pode trocar a imagem testada.
4. **Non-root por padrão** — `USER` com uid > 10000, `readOnlyRootFilesystem`, `CAP_DROP ALL`. Exploit no app não vira root no host.
5. **Sem secrets na imagem** — `ENV`/`ARG` vazam em `docker inspect` e nos metadados do registry. BuildKit `--mount=type=secret` no build; injeção pelo orquestrador em runtime.
6. **Cache por camada** — copiar manifestos de dependência antes do código-fonte; cache mounts (`--mount=type=cache`). `COPY..` cedo invalida tudo.
7. **Supply chain verificável** — Trivy no CI (bloqueia CRITICAL/HIGH antes do push); Cosign assina a imagem. Sem isso, imagem vulnerável chega em produção.

> Dockerfiles e compose prontos para commit.

## Stack Canônica

| Categoria | Ferramenta | Versão | Notas |
|---|---|---|---|
| Runtime | Docker Engine **ou** Podman | 26+ / 4+ | Podman: daemonless, rootless nativo, drop-in. Decisão por ambiente (sem default) |
| Build | BuildKit | built-in 23.x+ | `DOCKER_BUILDKIT=1` (padrão). Cache mounts + `--secret` |
| Multi-arch | `docker buildx` | built-in | AMD64 + ARM64, manifest list |
| Compose | Docker Compose (Compose Spec) | v2 CLI | Arquivo `compose.yaml`. Healthchecks + resource limits |
| Registry | Azure Container Registry (ACR) | Premium | Padrão. `myacr.azurecr.io`. GHCR/Docker Hub p/ OSS |
| Scanner | Trivy | 0.51+ | CI-native, CIS benchmarks. Grype: alternativa |
| Linter | Hadolint | 2.12+ | Lint do Dockerfile no CI |
| Signing | Cosign (Sigstore) | 2+ | Integridade de supply chain |
| SBOM | Syft / `docker sbom` | latest | Software Bill of Materials |

> **Multi-stage, distroless/Alpine, Trivy e ACR são fixos.** Docker vs Podman e imagem base (distroless/Alpine/slim) são decisão por caso.
> **PERGUNTAR ao usuário**: runtime (Docker/Podman)? base image (distroless/Alpine/slim)? multi-arch necessário? registry (ACR/GHCR)?

## Imagens Base (padrão por linguagem)

| Linguagem | Builder | Runtime |
|---|---|---|
| Go | `golang:1.2x-alpine` | `gcr.io/distroless/static-debian12:nonroot` |
| Node.js | `node:20` | `node:20-alpine` |
| Python | `python:3.12` | `python:3.12-slim` (C ext exigem glibc) |
| Java | build tool | `eclipse-temurin:21-jre-alpine` |
|.NET | SDK | `mcr.microsoft.com/dotnet/aspnet:8.0-alpine` |

## Estrutura de Projeto

```
repositorio/
├── docker/
│ ├── Dockerfile
│ ├── Dockerfile.dev # hot-reload (dev)
│ └──.dockerignore # exaustivo
├── compose.yaml # dev local
├── compose.override.yaml # overrides locais (gitignored)
├── k8s/ # manifests (ver kubernetes-expert)
└── scripts/docker-build.sh
```

## Workflow Agentic

1. **Entender escopo** — qual linguagem/runtime? imagem de produção ou dev? precisa multi-arch? tem dependências de sistema?
2. **Escolher base** — binário compilado → distroless/scratch; precisa shell/debug → Alpine; precisa glibc (C ext) → slim. Ver `reference/core.md`.
3. **Gerar de dentro pra fora**:
- `Dockerfile` multi-stage (deps → build → runtime), non-root, `HEALTHCHECK`, cache mounts.
- `.dockerignore` exaustivo (`node_modules`, `.git`, `.env*`, `dist/`).
- `compose.yaml` (se stack local): healthchecks, `depends_on: service_healthy`, networks isoladas, resource limits.
- Pipeline: Hadolint → Trivy scan → build/push → Cosign sign.
4. **Validar** — `hadolint Dockerfile`; `docker build`; conferir non-root, HEALTHCHECK, ausência de secrets, multi-stage.
5. **Apresentar diff** — "criei/editei: A, B, C; decisão: distroless pq binário Go".

## Anti-padrões Críticos

- ❌ Tag `latest` em produção (não é imutável; redeploy troca a imagem — usar versão exata + digest SHA).
- ❌ Rodar como root (sem `USER`) — exploit herda root no container; risco de escapar pro host.
- ❌ Secrets em `ENV`/`ARG` (vazam em `docker inspect` e no registry — usar BuildKit `--secret` no build, orquestrador em runtime).
- ❌ `COPY..` antes de instalar deps (invalida cache; `npm ci`/`go mod download` rodam toda build).
- ❌ Imagem single-stage em produção (carrega todo o toolchain de build — GBs e superfície de ataque a mais).
- ❌ Sem `HEALTHCHECK` (orquestrador não detecta app travado; sem auto-healing).
- ❌ `.dockerignore` ausente (`COPY..` inclui `.git`, `node_modules`, `.env` — contexto gigante e secrets na imagem).
- ❌ `apt-get upgrade` no Dockerfile (não reproduzível — usar imagem base atualizada).
- ❌ `ADD` para copiar arquivos locais (ambiguidade de URL/extração — usar `COPY`).
- ❌ `CMD`/`ENTRYPOINT` em shell form (`CMD npm start`) — não propaga sinais; usar exec form `["node", "app.js"]` (+ tini se PID 1 não trata SIGTERM).
- ❌ Host networking / mount de `/var/run/docker.sock` em produção.
- ❌ Push sem Trivy scan (imagem vulnerável chega ao registry).

## Checklist

- [ ] Multi-stage build (builder → runtime)
- [ ] Base image por versão exata + digest (nunca `latest`)
- [ ] Non-root (`USER` uid > 10000) + `readOnlyRootFilesystem`
- [ ] `.dockerignore` exaustivo
- [ ] Manifestos de deps antes do código + cache mounts
- [ ] `HEALTHCHECK` definido
- [ ] Sem secrets na imagem (BuildKit `--secret`)
- [ ] `CMD`/`ENTRYPOINT` em exec form
- [ ] Hadolint + Trivy no CI (bloqueia CRITICAL/HIGH)
- [ ] Cosign sign antes do deploy
- [ ] Compose: healthchecks + `depends_on: service_healthy` + resource limits + networks isoladas
- [ ] Multi-arch (buildx) quando ARM64 é alvo

## Quando Perguntar

Antes de gerar: runtime (Docker ou Podman)? imagem base (distroless/Alpine/slim — trade-off tamanho × debug × glibc)? precisa multi-arch (ARM64)? registry (ACR/GHCR/Docker Hub)? imagem de produção ou de desenvolvimento (hot-reload)?

## Referências (sob demanda — progressive disclosure)

Leia conforme a tarefa (não entram no contexto automaticamente):

- **`reference/core.md`** — Dockerfiles multi-stage prontos (Node/Go/Python), compose de desenvolvimento, buildx multi-arch, Hadolint e Trivy no CI, decision tree de base image + tabela de tamanhos, padrões de compose (rede isolada, worker+fila, logging, override dev/prod), healthchecks por serviço, anti-patterns detalhados com solução. Leia antes de gerar qualquer Dockerfile/compose.
- **`reference/stack.md`** — Stack canônica: versões, imagens base por linguagem, registries, scanner/linter, template de `.dockerignore`.

## Cross-references

- `kubernetes-expert` — orquestração, manifests, Helm, SecurityContext
- `security-expert` — container security, secrets, network policies
- `terraform-expert` — provisionar registry (ACR/ECR/GAR) via IaC
- `cicd-expert` — pipeline de build/scan/push/sign

## Segurança (Baseline Compartilhado)

Regras universais em `reference/security-baseline.md`. Aqui só ameaças específicas do domínio:

- **Non-root + `CAP_DROP ALL`** — dropar todas as capabilities Linux; `runAsNonRoot`, `readOnlyRootFilesystem`.
- **Sem secrets na imagem** — BuildKit `--mount=type=secret` no build; nunca `ENV`/`ARG` com valor sensível.
- **Trivy scan obrigatório** no CI antes do push (bloquear CRITICAL/HIGH).
- **Cosign** — assinar imagem para integridade de supply chain.
- **Isolamento de rede** — bridge networks separadas; `internal: true` para backend; nunca host networking em produção.

> **Atenção**: o padrão é **multi-stage + distroless/Alpine + Trivy + ACR**. Docker Swarm não é usado para novos projetos (orquestração de produção → Kubernetes/ACA/ECS). Imagem distroless não tem shell — troubleshooting via `kubectl debug`/ephemeral containers, não `exec`.
