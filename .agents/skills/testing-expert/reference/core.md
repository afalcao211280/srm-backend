# Testing Expert — Princípios e Disciplina

Fundamentos que valem para qualquer linguagem. Ler antes de gerar testes. Snippets concretos por stack em `patterns.md`; BDD/Gherkin em `bdd.md`.

## Índice

- [Anatomia de um bom teste (AAA)](#anatomia-de-um-bom-teste-aaa)
- [Test doubles](#test-doubles)
- [Cobertura de cenários](#cobertura-de-cenários)
- [Property-based testing](#property-based-testing)
- [TDD — red-green-refactor](#tdd--red-green-refactor)
- [Escolas e estilos de TDD](#escolas-e-estilos-de-tdd)
- [Geração assistida de testes](#geração-assistida-de-testes)
- [Legacy e characterization testing](#legacy-e-characterization-testing)
- [Métricas de qualidade](#métricas-de-qualidade)

## Anatomia de um bom teste (AAA)

**Arrange-Act-Assert** — três blocos visíveis, um por seção:

- **Arrange** — monta o estado e os doubles. Use mock factory por teste (sem estado vazado entre testes).
- **Act** — uma única ação sob teste.
- **Assert** — resultado observável. Idealmente uma asserção lógica por teste.

Regras:

- **Nome descritivo** — `should<resultado>_when<condição>` / `cria_usuário_quando_email_não_existe`. O nome é a especificação.
- **Comportamento, não implementação** — assertar o retorno/efeito observável, não que um método interno foi chamado. Teste acoplado à implementação quebra em refactor que preserva comportamento (fricção falsa).
- **Determinístico** — sem relógio real, sem rede real, sem ordem. Injete relógio/UUID/random. Timeout explícito em assíncrono.
- **Rápido** — unit em milissegundos. Sem I/O. Se precisa de banco/rede, é teste de integração, não unit.
- **Isolado** — cada teste roda sozinho e em qualquer ordem. Cleanup entre testes (ou transação com rollback).

## Test doubles

| Double | O que é | Quando usar |
|---|---|---|
| **Dummy** | objeto passado mas não usado | preencher assinatura |
| **Stub** | retorna valores pré-programados | controlar entradas indiretas |
| **Spy** | stub que registra chamadas | verificar interação quando ela é o contrato |
| **Mock** | espera interações específicas | London school; verificar protocolo |
| **Fake** | implementação leve real (ex.: repo em memória) | substituir colaborador caro mantendo comportamento |

Regra: **não mockar o que você não possui** — envolva a dependência externa numa interface sua e mocke a interface. Para integração, prefira o real (testcontainers) a fakes.

## Cobertura de cenários

Para cada unidade, cobrir:

1. **Happy path** — entrada válida → resultado esperado.
2. **Erros de validação** — input inválido, vazio, formato errado.
3. **Regras de negócio** — unicidade, limites, transições de estado proibidas.
4. **Falha de infra** — colaborador (repo/HTTP) lança → erro propagado corretamente.
5. **Edge cases** — null, vazio, boundary (0, 1, max), Unicode, concorrência.

Happy path sozinho cobre metade do comportamento; bugs de produção moram nos outros quatro.

## Property-based testing

Para lógica crítica (parsing, cálculo, serialização), em vez de exemplos fixos, declare **propriedades** e deixe o runner gerar centenas de entradas:

- **fast-check** (JS/TS), **Hypothesis** (Python), **gopter/rapid** (Go), **jqwik** (Java).
- Propriedades úteis: round-trip (`decode(encode(x)) == x`), idempotência, invariantes (resultado sempre ≥0), comparação com implementação ingênua (oracle).
- O runner faz **shrinking**: ao falhar, reduz ao menor contraexemplo.

## TDD — red-green-refactor

Ciclo curto, um comportamento por vez:

1. **Red** — escreva o menor teste que falha por um motivo claro. Rode e veja falhar (confirma que o teste testa algo).
2. **Green** — escreva o **mínimo** de código para passar. Pode ser feio; o objetivo é a barra verde.
3. **Refactor** — com a rede de segurança verde, limpe duplicação e melhore nomes. Sem mudar comportamento. Rode os testes após cada passo.

Benefícios: força um contrato testável, previne over-engineering (só se escreve código exigido por um teste), e dá rede para refatorar com confiança.

Quando usar TDD: design incerto, lógica de negócio não trivial, correção de bug (escreva o teste que reproduz o bug **antes** de corrigir). Para código trivial/óbvio, use julgamento.

**Anti-padrões de TDD:** test-after disfarçado de TDD; pular o passo refactor (acumula dívida); escrever vários testes vermelhos de uma vez (perde o foco); testar getters/setters triviais.

## Escolas e estilos de TDD

| Estilo | Como | Trade-off |
|---|---|---|
| **Chicago (classicist)** | testa estado via objetos reais/fakes, mocka pouco | testes robustos a refactor; setup pode ser maior |
| **London (mockist)** | mocka colaboradores, verifica interações | isola a unidade; acopla teste ao protocolo interno |
| **Outside-in** | começa no caso de uso (aceitação) e desce | guia o design pela necessidade real; bom p/ feature nova |
| **Inside-out** | começa nas peças e compõe | bom p/ biblioteca/algoritmo |
| **ATDD/BDD** | teste de aceitação em linguagem de negócio primeiro | alinha com stakeholder (ver `bdd.md`) |

Use o estilo conforme o contexto; não há pureza obrigatória. Equilíbrio: nem over-testing (testar trivialidade) nem under-testing (lógica crítica descoberta).

## Geração assistida de testes

Ao gerar uma suíte para código existente:

1. **Analisar a unidade** — assinatura, parâmetros, retorno, dependências, complexidade ciclomática (mais ramos → mais casos).
2. **Derivar cenários** dos ramos e contratos (ver "Cobertura de cenários"), não só do happy path.
3. **Gerar mocks** para dependências externas (interface own) com `spec`/tipo, para não mascarar mudança de assinatura.
4. **Analisar gaps de coverage** — rodar com coverage, listar linhas/branches descobertos, gerar testes direcionados a eles.
5. **Revisar qualidade**, não só quantidade — todo teste gerado precisa de asserção significativa; rodar mutation testing para validar que os testes pegam regressões.

> Cuidado: testes gerados que só asseguram "não é null" inflam coverage sem valor. Cada teste deve falhar se o comportamento mudar.

## Legacy e characterization testing

Para código sem testes que você precisa mexer:

- **Characterization test** — capture o comportamento **atual** (mesmo que "errado") em testes; eles viram a rede de segurança antes de refatorar.
- **Golden master / approval testing** — para saída complexa (relatório, HTML, payload), grave a saída aprovada e compare; útil quando asserções campo-a-campo são inviáveis.
- **Quebrar dependências (seams)** — identifique pontos de injeção para tornar testável (extrair interface, parâmetro em vez de `new` interno) sem mudar comportamento.
- Adote TDD incrementalmente: cada bug corrigido e cada feature nova entram com teste.

## Métricas de qualidade

- **Line/statement coverage ≥80%**; **branch coverage ≥70%** (branch é sinal melhor que line). 100% em auth/segurança.
- **Mutation score ≥60%** (Stryker) — mede se os testes realmente pegam regressões; é o gate de qualidade real sobre coverage.
- **Cycle time** do red-green-refactor — feedback rápido mantém o fluxo.
- **Flaky rate = 0** — qualquer teste não determinístico vai para quarantine e vira issue.
