# AGENTS.md — testing-expert

Skill **portável e self-contained**, pronta para múltiplos agentes de IA. Fonte única: `SKILL.md` (entrada) + `reference/*.md` (sob demanda). Qualquer agente que leia Markdown consome direto — **sem ferramenta de geração**.

## Uso por agente

| Agente | Como usar |
|---|---|
| Claude Code | Copiar a pasta para `.claude/skills/testing-expert/` (projeto) ou `~/.claude/skills/testing-expert/` (global). Aciona pela `description` do frontmatter. |
| Cursor | Referenciar `SKILL.md` em `.cursor/rules/*.mdc` ou @-mencionar no chat. |
| GitHub Copilot | Colar o conteúdo de `SKILL.md` em `.github/copilot-instructions.md`. |
| Windsurf / Cline | Copiar `SKILL.md` para o arquivo de regras do agente. |
| Opencode / genérico | Manter `AGENTS.md` + `SKILL.md` na raiz; apontar o agente para a pasta. |

## Regra de ouro

`SKILL.md` é a fonte única de verdade. `reference/` carrega sob demanda (progressive disclosure). Nada aqui depende de binário externo.

## Escopo

Engenheiro de qualidade sênior em testes automatizados de produção. Pirâmide 70/20/10 (unit/integração/E2E), coverage ≥80%, TDD red-green-refactor, BDD/Gherkin, zero flaky. Stack: testify+testcontainers (Go), pytest (Python), JUnit 5+Mockito (Java), Vitest+Playwright (Frontend); Stryker (mutation), k6/Locust (load). Gera suítes prontas — unit, integração, E2E, fixtures, mocks. Acionar SEMPRE que mencionar teste, coverage, cobertura, TDD, BDD, mock, fixture, flaky, pytest, testify, Vitest, Jest, JUnit, Playwright, Cypress, testcontainers ou Gherkin. Complementar a TODAS as skills.
