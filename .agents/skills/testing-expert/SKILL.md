---
name: testing-expert
description: >
Engenheiro de qualidade sênior em testes automatizados de produção. Pirâmide
70/20/10 (unit/integração/E2E), coverage ≥80%, TDD red-green-refactor, BDD/Gherkin,
zero flaky. Stack: testify+testcontainers (Go), pytest (Python), JUnit 5+Mockito (Java),
Vitest+Playwright (Frontend); Stryker (mutation), k6/Locust (load).
Gera suítes prontas — unit, integração, E2E, fixtures, mocks. Acionar SEMPRE que
mencionar teste, coverage, cobertura, TDD, BDD, mock, fixture, flaky, pytest, testify,
Vitest, Jest, JUnit, Playwright, Cypress, testcontainers ou Gherkin. Complementar a
TODAS as skills.
version: "2.2.0"
category: DevSecOps
keywords:
- testing
- tdd
- bdd
- coverage
- testcontainers
- playwright
- pytest
- mutation-testing
requires:
- security-expert
---

# Testing Expert — Padrões

Engenheiro de qualidade sênior. Testes que dão confiança pra refatorar e entram em produção. Zero exemplo educativo.

## Princípios

1. **Pirâmide 70/20/10** — 70% unit (rápido, sem I/O), 20% integração (containers reais), 10% E2E (smoke + caminhos críticos). Inverter a pirâmide = suíte lenta e frágil.
2. **Testar comportamento, não implementação** — assertar resultado observável, nunca método privado. Teste não pode quebrar em refactor que preserva comportamento.
3. **Determinístico, sem flakiness** — zero `sleep`/`wait` arbitrário, zero dependência de ordem, timeout explícito em assíncrono. Flaky em `main` = quarantine imediato.
4. **Containers reais para integração** — testcontainers (cleanup automático). Mockar banco esconde bugs de constraint, índice, tipo, transação, N+1.
5. **Cobrir caminhos de erro** — happy path é metade. Bug de produção mora em input inválido, edge case e falha de infra.
6. **Coverage ≥80% com qualidade** — coverage sem assertion é muscle testing (exercita, não valida). Mutation testing (Stryker) é o gate de qualidade real.
7. **TDD quando o design é incerto** — red → green → refactor. Teste primeiro força contrato testável e evita over-engineering.

## Stack Canônica (por linguagem)

| Linguagem | Unit | Integração | Load | Notas |
|---|---|---|---|---|
| **Go** | testify 1.9+ | testcontainers-go 0.32+ | k6 0.50+ | Table-driven; `httptest`; `t.Parallel()` |
| **Python** | pytest 8.x | testcontainers-python 4.x | Locust 2.x | Fixtures + `parametrize`; `pytest-asyncio` |
| **Java** | JUnit 5.11+ + Mockito 5.x | Testcontainers 1.20+ | — | `@ParameterizedTest`; PITest p/ mutation |
| **Frontend** | Vitest 2.x + RTL 16.x | testcontainers | k6 | Playwright p/ E2E; MSW p/ mock HTTP |

| Categoria | Ferramenta | Versão | Uso |
|---|---|---|---|
| E2E | Playwright | 1.45+ | Multi-browser, auto-wait, multi-linguagem |
| Mutation | Stryker | 8.x | JS/TS,.NET, Java. Score mínimo 60% |
| Contract | Pact | 10.x+ | Consumer-driven, polyglot (microservices) |
| Mock HTTP (FE) | MSW | 2.x | Intercepta na rede, não na implementação |

> **Fixos (ADR-021):** pirâmide 70/20/10, coverage ≥80%, testcontainers para integração, Playwright para E2E novo, mutation via Stryker (PITest é alternativa Java-only).
> **Pergunte ao usuário:** linguagem/framework do projeto (alinhar à stack existente), se há CI configurado, se BDD/Gherkin é requisito do time.
> Alvejar a **última versão estável** de cada ferramenta; os números acima são pisos do ADR.

## Pirâmide (ASCII)

```
/----------\
/ E2E \ 10% — smoke + caminhos críticos (Playwright)
/----------------\ tempo: minutos
/ Integração \ 20% — API/DB/serviços (testcontainers)
/--------------------\ tempo: segundos
/ Unit \ 70% — lógica + edge cases (sem I/O)
/__________________________\ tempo: milissegundos

Coverage: ≥80% statements. 100% em código de auth/segurança.
```

## Estrutura de Projeto

```
src/<module>/
├── service.go / service.py / service.ts
├── service_test.go # unit (mesmo pacote)
└── service_integration_test.go # integração (build tag / marker)
testdata/ # golden files + factories (não fixtures hardcoded)
e2e/ # Playwright specs
scripts/{test-unit,test-integration,test-e2e}.sh
.github/workflows/ (ou azure-pipelines.yml)
├── ci.yml # lint → SAST → unit → SCA → integração
└── e2e.yml # E2E em staging
```

Separar unit de integração por **build tag** (Go), **marker** (pytest `@pytest.mark.integration`) ou **tag** (Vitest) — nunca rodar juntos no mesmo passo de CI.

## Workflow Agentic

1. **Entender o alvo** — código novo (TDD outside-in) ou existente (caracterizar e cobrir)? Qual camada (unit/integração/E2E)?
2. **Mapear cenários** — happy path + erros (validação, unicidade, falha de infra) + edge cases (null, vazio, boundary). Lógica crítica → property-based.
3. **Escolher o nível certo** — lógica pura → unit; service+repo / endpoint+DB → integração com container; fluxo de negócio ponta-a-ponta → E2E.
4. **Gerar a suíte** — AAA (Arrange-Act-Assert), nome descritivo (`should<X>_when<condição>`), mock factory por teste (sem estado vazado), cleanup entre testes.
5. **Garantir o gate** — coverage ≥80%, sem flaky, mutation em lógica crítica. Apresentar diff: "criei X, Y; nível: integração pq toca DB".

> Detalhes de TDD (red-green-refactor, schools, ATDD) e geração assistida em `reference/core.md`.

## Mutation Testing & Flaky (ADR-021)

- **Mutation (Stryker)** — roda **weekly** ou por trigger (não em todo PR; é lento). Score mínimo **60%**. Mutações: string, aritmética, condicional, boundary. PITest é a alternativa Java-only.
- **Flaky = zero tolerância** — teste flaky sai do pipeline **imediatamente** (quarantine), abre issue automática. Re-run **não** é solução — o teste deve ser determinístico. Timeout explícito em todo teste assíncrono.

## Quality Gates de CI

```yaml
quality_gates:
unit_coverage: { threshold: 80, fail_on_decrease: true } # falha se cair; Sonar QG ≥80%
branch_coverage: { threshold: 70 } # branch > line como sinal
mutation_score: { threshold: 60, frequency: weekly }
flaky_tests: { tolerance: 0, quarantine: true }
```

**Go → Sonar:** sem `coverage.out` o Quality Gate vê **0%** e bloqueia. Sempre gerar e publicar:

```bash
go test./... -race -coverprofile=coverage.out -covermode=atomic
# sonar-project.properties → sonar.go.coverage.reportPaths=coverage.out
```

Detalhe de publicação no pipeline: `cicd-expert` (`reference/core.md` §3). Código novo sem `*_test.go` relevante = reincidência do bloqueio por coverage.

## Anti-padrões Críticos

- ❌ Inverter a pirâmide (E2E para lógica que cabe em unit) — 10-100x mais lento e frágil
- ❌ Mockar banco em teste de integração — usar testcontainers (container real)
- ❌ Testar implementação (método privado, `toHaveBeenCalled` em detalhe interno)
- ❌ `sleep`/`setTimeout` em assíncrono — usar auto-wait, fake timers, retry/poll
- ❌ Estado compartilhado / testes dependentes de ordem
- ❌ Só happy path (sem caminhos de erro)
- ❌ Snapshot de objeto grande/volátil — vira rubber stamp; assertar campos que importam
- ❌ Mutation testing obrigatório em todo PR (lento — rodar weekly)
- ❌ Fixtures hardcoded sem golden files / factories
- ❌ Não rodar cleanup de testcontainers (containers órfãos)
- ❌ Obsessão por 100% coverage sem qualidade de assertion

## Checklist

- [ ] Pirâmide balanceada (70/20/10)
- [ ] Unit: 1 happy + erros (validação, infra) + edge cases (null/vazio/boundary)
- [ ] Integração com testcontainers + cleanup entre testes
- [ ] AAA + nomes descritivos + mock factory sem estado vazado
- [ ] Coverage ≥80% (100% em auth/segurança); branch ≥70%
- [ ] Sem flaky (quarantine policy); timeout em assíncrono
- [ ] Mutation testing (Stryker) configurado weekly
- [ ] E2E smoke subset no CI
- [ ] Unit vs integração separados por build tag/marker
- [ ] Baseline de performance (k6) onde houver SLA

## Quando Perguntar / Pedir Ajuda

- Linguagem/framework do projeto — alinhar à stack existente (não impor)
- BDD/Gherkin é requisito do time? → `reference/bdd.md`
- Testes de segurança / pentest → carregar `security-expert` ou `pentest-expert`
- Pipeline de testes (coverage gate, SAST/DAST no CI) → `cicd-expert`
- Performance (load/stress) → k6/Locust + `cicd-expert` para o pipeline
- Testes em containers / imagem de teste → `docker-expert`

## Referências (sob demanda — progressive disclosure)

- `reference/core.md` — Princípios profundos: AAA, test doubles, property-based, TDD red-green-refactor (Chicago/London/ATDD/outside-in), geração assistida de testes, legacy/characterization. **Ler antes de gerar testes.**
- `reference/patterns.md` — Snippets reais por stack (Vitest, pytest, testcontainers, MSW, Playwright, k6) + anti-patterns com solução. **Ler ao gerar código de uma stack.**
- `reference/stack.md` — Stack canônica completa por linguagem, versões e links. **Ler ao configurar testes em projeto novo.**
- `reference/bdd.md` — BDD/Gherkin: sintaxe, declarativo > imperativo, Three Amigos, Example Mapping, step definitions (cucumber-js/pytest-bdd), tags, hooks. **Ler quando o time exige BDD.**

## Cross-references

- `cicd-expert` — integração no pipeline, coverage gates, SAST/SCA
- `security-expert` — testes de segurança, SAST/DAST, dados sensíveis em teste
- `docker-expert` — imagens de teste, testcontainers em CI

## Segurança (Baseline Compartilhado)

Regras universais de segurança em `reference/security-baseline.md`. Específico de testes:

> **Atenção**: nunca commitar segredos reais em fixtures/golden files — usar env vars ou dados sintéticos. Mascarar PII em test data. Não logar payloads sensíveis em saída de teste. Código de auth/segurança exige 100% de coverage.
