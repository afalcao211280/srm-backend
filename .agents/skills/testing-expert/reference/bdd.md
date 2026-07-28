# BDD & Gherkin

Behaviour-Driven Development com especificações executáveis em linguagem de negócio. Ler quando o time exige BDD/critérios de aceitação executáveis. Conceitos são agnósticos de linguagem; exemplos de runner para o ecossistema (cucumber-js, pytest-bdd) — o Gherkin é idêntico em qualquer linguagem.

```
Feature (.feature) ──> Step Definitions (código) ──> Sistema sob teste
```

## Índice

- [Quando usar BDD](#quando-usar-bdd)
- [Sintaxe Gherkin](#sintaxe-gherkin)
- [Declarativo > imperativo](#declarativo--imperativo)
- [Step definitions](#step-definitions)
- [Tags e execução](#tags-e-execução)
- [Hooks e ordem](#hooks-e-ordem)
- [Colaboração: Three Amigos e Example Mapping](#colaboração-three-amigos-e-example-mapping)
- [Anti-padrões](#anti-padrões)

## Quando usar BDD

- Critérios de aceitação de feature precisam ser **verificáveis por stakeholders** (PO, QA, dev compartilham a mesma especificação).
- Camada de **aceitação / E2E** (outside-in), não substitui unit. BDD é a ponta da pirâmide.
- Não use Gherkin como framework de teste de integração de API interna nem para lógica que cabe em unit — overhead sem ganho de comunicação.

## Sintaxe Gherkin

```gherkin
# language: pt (opcional; default en)
@billing
Feature: Geração de fatura
Como cliente
Quero gerar minha fatura
Para acompanhar meus gastos

Background: # setup comum (≤4 linhas, contexto essencial)
Given estou autenticado como "cliente"

Rule: Cliente free não acessa conteúdo premium # agrupa por regra de negócio (Gherkin 6+)

Scenario: Free vê prompt de upgrade
Given tenho uma assinatura free
When tento acessar um artigo premium
Then vejo o prompt de upgrade

Scenario Outline: Validação de senha # mesmo cenário, várias entradas
When informo a senha "<senha>"
Then vejo "<mensagem>"

Examples:
| senha | mensagem |
| curta | Senha muito curta |
| semmaiuscula| Precisa de maiúscula |
| ValidaPass1 | Senha aceita |
```

**Keywords de step:** `Given` (contexto/setup), `When` (ação), `Then` (asserção observável), `And`/`But` (continuação), `*` (item de lista).

**Data tables** (dados tabulares para um step):

```gherkin
Given os seguintes usuários existem:
| nome | email | papel |
| Alice | alice@example.com | admin |
```

**Doc strings** (texto multilinha, ex. payload):

```gherkin
Given um corpo JSON:
"""json
{ "nome": "Teste", "ativo": true }
"""
```

**Tag inheritance:** Scenario herda tags de Feature e Rule. **Não** se pode taggear `Background` nem steps individuais.

## Declarativo > imperativo

Descreva **o que** o sistema faz, não **como** se clica. UI muda; comportamento não.

```gherkin
# Bom (declarativo) # Ruim (imperativo)
When "Bob" faz login When acesso "/login"
Then ele vê seu dashboard And preencho "#email" com "bob@x.com"
And clico em "#submit"
Then "#welcome" contém "Olá"
```

Regras:
- **Um comportamento por cenário**, 3-5 steps. Mais que ~7 → está testando coisas demais.
- **Linguagem de domínio** que o stakeholder entende; sem seletores CSS, rotas ou status HTTP no `.feature`.
- **Detalhes essenciais apenas** — nada de número de cartão, endereço completo, IDs gerados quando não são o ponto do cenário.
- **Resultado observável** no `Then` — algo que o usuário vê, não estado interno.
- **Sem lógica** no `.feature` (nada de if/else) — quebre em cenários separados.

## Step definitions

Ligam o texto Gherkin a código executável. Keyword é ignorada no matching (dois steps com mesmo texto colidem).

**cucumber-js (TS) + Playwright:**

```typescript
import { Given, When, Then } from '@cucumber/cucumber';
import { expect } from '@playwright/test';

Given('estou autenticado como {string}', async function (papel: string) {
await this.loginAs(papel); // helper de domínio no World — sem detalhe de UI no.feature
});

When('tento acessar um artigo premium', async function () {
await this.page.goto('/artigos/premium-1');
});

Then('vejo o prompt de upgrade', async function () {
await expect(this.page.getByText('Faça upgrade')).toBeVisible();
});
```

**pytest-bdd (Python):**

```python
from pytest_bdd import scenario, given, when, then, parsers

@scenario("usuario.feature", "Free vê prompt de upgrade")
def test_upgrade_prompt():...

@given(parsers.parse('tenho uma assinatura {tipo}'))
def assinatura(client, tipo):
client.set_plan(tipo)

@when("tento acessar um artigo premium")
def acessa(client):
client.get("/artigos/premium-1")

@then("vejo o prompt de upgrade")
def ve_prompt(client):
assert "Faça upgrade" in client.last_response.text
```

**Cucumber Expressions** (preferidas sobre regex): `{int}`, `{float}`, `{word}`, `{string}`, `{}` (anônimo). Texto opcional `cucumber(s)`; alternativas `home/login/dashboard`.

**Estado entre steps:** use o objeto **World** (cucumber-js) ou fixtures (pytest-bdd) com helpers de domínio (`loginAs`, `createUser`) — nunca acoplar o step ao HTML. Step de responsabilidade única; steps reutilizáveis que compõem entre cenários.

## Tags e execução

```bash
# cucumber-js
npx cucumber-js --tags "@smoke and not @wip"
npx cucumber-js --parallel 4

# pytest-bdd (markers do pytest)
pytest -m "smoke and not wip"
```

Convenção de tags: categoria (`@smoke @regression @e2e`), prioridade (`@critical @low`), status (`@wip @manual`), técnica (`@slow @database @browser`), referência (`@JIRA-123`). Expressões booleanas: `(@smoke or @critical) and not @wip`. Rode subset `@smoke` no CI para feedback rápido.

## Hooks e ordem

Hooks fazem setup/teardown técnico (fora do `.feature`, que fica só com contexto de negócio).

| Hook | Escopo | Quando |
|---|---|---|
| `BeforeAll`/`AfterAll` | suíte | uma vez |
| `Before`/`After` | cenário | a cada cenário (filtrável por tag) |
| `AfterStep` | step | após cada step |

**Ordem:** `BeforeAll` → (por cenário: `Before` → Background → steps → `After`) → `AfterAll`. `Before` filtrado por tag: `Before('@database')`. Cleanup de DB e screenshot-on-failure vivem em hooks, não em steps.

## Colaboração: Three Amigos e Example Mapping

BDD é técnica de **comunicação** antes de teste. Antes de escrever cenários:

- **Three Amigos** — PO (explica a feature e os critérios), Dev (aponta restrições técnicas), QA (levanta edge cases). Os três alinham a especificação.
- **Example Mapping** — cartões coloridos: amarelo (story), azul (regras de negócio), verde (exemplos concretos = futuros cenários), vermelho (dúvidas a resolver). Mapeia a história em ~25min e expõe ambiguidade antes do código.
- **Living documentation** — os `.feature` são documentação viva: mantidos com o sistema, em linguagem ubíqua, revisados como código.

## Anti-padrões

- ❌ **Cenário como script de UI** — steps com seletores/cliques. Use linguagem de comportamento.
- ❌ **Detalhes incidentais** — dados irrelevantes (cartão, endereço, ID) que não são o ponto do cenário.
- ❌ **Feature como suíte genérica** — uma Feature "User" com cenários não relacionados. Quebre por capacidade.
- ❌ **Cucumber como teste de API interna** — testar `GET /api/users` retorna 200 não é comportamento de usuário.
- ❌ **Lógica no `.feature`** — if/else nos steps. Separe em cenários.
- ❌ **Background longo/técnico** (>4 linhas, "seed do banco") — vai para hook.
- ❌ **Step acoplado ao HTML** — encapsule em helper de domínio no World.
- ❌ **Setup via UI quando dá pra criar dado direto** — lento e frágil; crie via factory/API no hook.
