# AGENTS.md — css-expert

Skill **portável e self-contained**, pronta para múltiplos agentes de IA. Fonte única: `SKILL.md` (entrada) + `reference/*.md` (sob demanda). Qualquer agente que leia Markdown consome direto — **sem ferramenta de geração**.

## Uso por agente

| Agente | Como usar |
|---|---|
| Claude Code | Copiar a pasta para `.claude/skills/css-expert/` (projeto) ou `~/.claude/skills/css-expert/` (global). Aciona pela `description` do frontmatter. |
| Cursor | Referenciar `SKILL.md` em `.cursor/rules/*.mdc` ou @-mencionar no chat. |
| GitHub Copilot | Colar o conteúdo de `SKILL.md` em `.github/copilot-instructions.md`. |
| Windsurf / Cline | Copiar `SKILL.md` para o arquivo de regras do agente. |
| Opencode / genérico | Manter `AGENTS.md` + `SKILL.md` na raiz; apontar o agente para a pasta. |

## Regra de ouro

`SKILL.md` é a fonte única de verdade. `reference/` carrega sob demanda (progressive disclosure). Nada aqui depende de binário externo.

## Escopo

Especialista em CSS e design systems seguindo padrões de produção (cross-cutting, ADR-008). Stack: Tailwind CSS 3.4+, CSS Custom Properties (design tokens), CSS Modules, PostCSS, shadcn/ui + Radix UI, cva + cn() (clsx+tailwind-merge), Lucide/Heroicons. Gera componentes estilizados, tokens, layouts responsivos, dark mode e a11y prontos pra produção. Acionar SEMPRE que mencionar CSS, Tailwind, estilo, responsivo, layout, dark mode, design system, design token, animação CSS, shadcn, ou estilização de UI.
