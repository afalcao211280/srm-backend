# AGENTS.md — clean-architecture

Skill **portável e self-contained**, pronta para múltiplos agentes de IA. Fonte única: `SKILL.md` (entrada) + `reference/*.md` (sob demanda). Qualquer agente que leia Markdown consome direto — **sem ferramenta de geração**.

## Uso por agente

| Agente | Como usar |
|---|---|
| Claude Code | Copiar a pasta para `.claude/skills/clean-architecture/` (projeto) ou `~/.claude/skills/clean-architecture/` (global). Aciona pela `description` do frontmatter. |
| Cursor | Referenciar `SKILL.md` em `.cursor/rules/*.mdc` ou @-mencionar no chat. |
| GitHub Copilot | Colar o conteúdo de `SKILL.md` em `.github/copilot-instructions.md`. |
| Windsurf / Cline | Copiar `SKILL.md` para o arquivo de regras do agente. |
| Opencode / genérico | Manter `AGENTS.md` + `SKILL.md` na raiz; apontar o agente para a pasta. |

## Regra de ouro

`SKILL.md` é a fonte única de verdade. `reference/` carrega sob demanda (progressive disclosure). Nada aqui depende de binário externo.

## Escopo

Princípios de Clean Architecture para desenhar sistemas manuteníveis e testáveis. Use SEMPRE ao projetar sistemas/módulos, definir camadas e fronteiras, direcionar dependências, ou refatorar acoplamento. Padrão: handler → service → repository → domain (dependências apontam pra dentro). Acionar em "arquitetura", "camadas", "boundaries", "use case", "dependency inversion", "onde colocar esta lógica".
