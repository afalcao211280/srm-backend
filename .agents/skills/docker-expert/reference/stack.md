# Docker Expert — Stack Canônica

## Docker — Versão Canônica

| Categoria | Ferramenta/Lib | Versão | Link | Notas |
|-----------|---------------|--------|------|-------|
| **Core** | | | | |
| Runtime | Docker Engine | 26.x+ | https://docs.docker.com | `moby/moby`. Rootless mode recomendado em prod |
| Alternativa | Podman | 4.x | https://podman.io | Daemonless, rootless nativo. Drop-in substituto |
| **Build** | | | | |
| Build | Docker BuildKit | built-in 23.x+ | — | `DOCKER_BUILDKIT=1` (padrão desde 23.x) |
| Build cache | Cache mounts | `--mount=type=cache` | — | Cacheia npm/pip/go modules entre builds |
| Multi-platform | `docker buildx` | built-in | — | Builds AMD64 + ARM64 |
| SBOM | Syft / docker sbom | latest | https://github.com/anchore/syft | Gerar Software Bill of Materials |
| **Compose** | | | | |
| Compose | Docker Compose (v2) | 2.x | https://docs.docker.com/compose | `compose.yaml` (não `docker-compose.yml`) |
| **Imagens Base** | | | | |
| Node.js | `node:20-alpine` | — | — | Runtime prod. `node:20` para builder |
| Python | `python:3.12-slim` | — | — | Runtime prod. `python:3.12` para builder |
| Go | `gcr.io/distroless/static-debian12` | — | — | Scratch-like, sem shell. Máxima segurança |
| Java | `eclipse-temurin:21-jre-alpine` | — | — | JRE apenas para runtime |
|.NET | `mcr.microsoft.com/dotnet/aspnet:8.0-alpine` | — | — | ASP.NET runtime |
| **Registry** | | | | |
| Registry |: Azure Container Registry | — | — | `myacr.azurecr.io` |
| Registry | GitHub Container Registry | — | `ghcr.io` | Para OSS/CI |
| Registry | Docker Hub | — | `hub.docker.com` | Rate limiting em CI sem auth |
| **Segurança** | | | | |
| Scanner | Trivy | 0.51+ | https://github.com/aquasecurity/trivy | Vulnerabilidades em imagens. Integrar no CI |
| Scanner alt | Grype | 0.74+ | https://github.com/anchore/grype | Alternativa ao Trivy |
| Linter | Hadolint | 2.12+ | https://github.com/hadolint/hadolint | Lint do Dockerfile |
| **Runtime** | | | | |
| Orquestração local | Docker Compose | v2 | — | Dev + testes de integração |
| Orquestração prod | Kubernetes / ACA / ECS | — | — | Não usar Swarm para novos projetos |

## Dockerfile — Boas Práticas

```dockerfile
# checklist:
# ✅ Multi-stage build (builder + runtime)
# ✅ Imagem base específica (não:latest)
# ✅ Non-root user
# ✅.dockerignore configurado
# ✅ COPY arquivos de lock antes do código fonte
# ✅ Cache mounts para package manager
# ✅ No secrets em build args (usar --secret em CI)
# ✅ HEALTHCHECK definido
```

##.dockerignore — Template

```
node_modules/
.git/
.github/
*.log
*.md
.env*
coverage/
dist/ # se buildado fora do container
.DS_Store
Thumbs.db
```
