# AGENTS.md — security-expert

Skill **portável e self-contained**, pronta para múltiplos agentes de IA. Fonte única: `SKILL.md` (entrada) + `reference/*.md` (sob demanda). Qualquer agente que leia Markdown consome direto — **sem ferramenta de geração**.

## Uso por agente

| Agente | Como usar |
|---|---|
| Claude Code | Copiar a pasta para `.claude/skills/security-expert/` (projeto) ou `~/.claude/skills/security-expert/` (global). Aciona pela `description` do frontmatter. |
| Cursor | Referenciar `SKILL.md` em `.cursor/rules/*.mdc` ou @-mencionar no chat. |
| GitHub Copilot | Colar o conteúdo de `SKILL.md` em `.github/copilot-instructions.md`. |
| Windsurf / Cline | Copiar `SKILL.md` para o arquivo de regras do agente. |
| Opencode / genérico | Manter `AGENTS.md` + `SKILL.md` na raiz; apontar o agente para a pasta. |

## Regra de ouro

`SKILL.md` é a fonte única de verdade. `reference/` carrega sob demanda (progressive disclosure). Nada aqui depende de binário externo.

## Escopo

Especialista em segurança de aplicações e microserviços, padrões de produção (ADR-019). Stack: OAuth2/OIDC, JWT RS256, RBAC+ABAC com OPA, API Gateway, Vault, mTLS (Istio/Linkerd), AES-256-GCM, argon2id. Faz auditoria OWASP Top 10 (CRITICAL→LOW), threat modeling e hardening. Acionar SEMPRE que envolver autenticação, autorização, OAuth2, JWT, OWASP, TLS, encryption, RBAC, secrets, CORS, headers, ou security review. Complementa TODAS as skills de backend/infra.
