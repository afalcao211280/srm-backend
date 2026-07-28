# AGENTS.md — docker-expert

Skill **portável e self-contained**, pronta para múltiplos agentes de IA. Fonte única: `SKILL.md` (entrada) + `reference/*.md` (sob demanda). Qualquer agente que leia Markdown consome direto — **sem ferramenta de geração**.

## Uso por agente

| Agente | Como usar |
|---|---|
| Claude Code | Copiar a pasta para `.claude/skills/docker-expert/` (projeto) ou `~/.claude/skills/docker-expert/` (global). Aciona pela `description` do frontmatter. |
| Cursor | Referenciar `SKILL.md` em `.cursor/rules/*.mdc` ou @-mencionar no chat. |
| GitHub Copilot | Colar o conteúdo de `SKILL.md` em `.github/copilot-instructions.md`. |
| Windsurf / Cline | Copiar `SKILL.md` para o arquivo de regras do agente. |
| Opencode / genérico | Manter `AGENTS.md` + `SKILL.md` na raiz; apontar o agente para a pasta. |

## Regra de ouro

`SKILL.md` é a fonte única de verdade. `reference/` carrega sob demanda (progressive disclosure). Nada aqui depende de binário externo.

## Escopo

Especialista em Docker e containers seguindo padrões de produção. Stack: Docker/Podman, BuildKit + buildx (multi-arch), Docker Compose, distroless/Alpine, Trivy, Cosign, Azure Container Registry. Gera Dockerfiles multi-stage, compose,.dockerignore e pipelines de scan/assinatura production-grade. Acionar SEMPRE que mencionar Docker, Dockerfile, docker-compose, compose.yaml, container, imagem, containerização, multi-stage, buildx, registry ou hardening de imagem.
