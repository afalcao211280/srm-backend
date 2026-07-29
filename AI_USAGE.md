# AI_USAGE — Backend

## Prompts estratégicos

- "Qual biblioteca Go para aritmética decimal atende General Decimal Arithmetic IEEE 754-2008, suporta `Pow` fracionário e sinaliza `Inexact`/`Rounded`?" → levou a `apd/v3` em vez de `shopspring`.
- "Como evitar SQL injection em `WHERE` dinâmico com filtros opcionais em Go?" → levou a `squirrel` com placeholders.
- "Como implementar optimistic locking em Go com pgx?" → levou a `UPDATE` condicional com `versao` e `status`.

## Alucinações corrigidas

- **`apd.Pow(0, -1)` devolve `Infinity` com `err == nil`**: a documentação não advertia. Verificado no spike antes de implementar o motor. Mitigação: helper `mustFinite` que rejeita `Form != apd.Finite`. Coberto por teste de borda.
- **`shopspring/decimal.Pow` retorna `0` silenciosamente**: alucinada como alternativa. Substituída por `apd/v3`.
- **TypeScript 7 com Next.js 16**: `typescript@latest` resolve para `7.0.2`, que removeu `lib/typescript.js` e quebra a detecção. Fixado em `5.9.3`.
- **IOF no desconto de duplicata**: o mercado real cobra, mas o enunciado não pede. Acrescentar seria escopo inventado.

## Lacunas reveladas pela auditoria do plano

- **Cross-origin entre containers**: o planejamento assistido não capturou que o browser falaria com o backend em outra porta. Solução: proxy de mesma origem no Next.
- **Massa de demonstração**: o plano original previa entregar sem dados de exemplo. A auditoria revelou que o grid com paginação server-side não pode ser avaliado sem volume. Solução: comando `cmd/carga`.
- **Fuso horário da data de operação**: o plano original não fixava. Sem regra explícita, entre 21h e 24h em Brasília a data em UTC já virou e o prazo calculava errado. Solução: `America/Sao_Paulo` em `OperacaoDataCorrente`.

## Lacunas reveladas por execução real (não por leitura de código)

Uma segunda passada de implementação, focada em *rodar* tudo em vez de só ler o código já escrito, encontrou uma classe inteira de problemas que passavam despercebidos: código que compilava, tinha teste verde e "parecia pronto", mas nunca tinha sido de fato executado contra dependências reais (banco real, container real, linter real). Cada um só apareceu ao rodar o sistema de verdade — nenhum apareceu relendo o código.

- **`swag` sugerido, nunca corrigido**: o item anterior deste arquivo já registrava "o stack canônico é Huma", mas a correção nunca foi aplicada — não havia OpenAPI, `/docs` ou `/openapi.json` algum. Migração completa para Huma v2 sobre o mesmo `*gin.Engine`, verificada ao vivo (9 rotas documentadas, `/docs` responde 200, erros no mesmo formato `dto.ErroCorpo` inclusive os gerados pelo próprio Huma).
- **Migrations nunca eram aplicadas**: os arquivos existiam em `migrations/`, mas nenhum código chamava um runner. `docker compose up` derrubava o container em crash-loop na primeira consulta. `golang-migrate` estava citado no ADR-010 como já usado — não estava.
- **`file://<caminho relativo>` é ambíguo**: numa URL `file://`, tudo antes da primeira barra depois de `//` é o *host*, não o *path*. `file://migrations` tem host `"migrations"` e path vazio, que o driver resolve silenciosamente para `"."`. Funcionava por acidente em teste local (o cwd do `go test` fazia `"."` apontar pro lugar certo) e quebrou só dentro do container. Corrigido para sempre resolver caminho absoluto antes de montar a URL.
- **Healthcheck do Docker nunca funcionava**: `docker-compose.yml` já chamava `/api --health`, mas o binário não tinha essa flag — reiniciava o servidor inteiro em vez de checar saúde. A imagem é distroless (sem shell/curl), então a checagem precisa ser o próprio binário fazendo uma requisição HTTP real.
- **`golangci-lint` nunca rodou de verdade**: `depguard` e `forbidigo` estavam configurados mas ausentes de `linters.enable` no schema v1 — suas regras, incluindo as de fronteira de camada entre `report` e `domain`, nunca eram checadas. Ao migrar para o schema v2 e habilitar os dois de verdade, apareceu que a própria regra de `depguard` estava quebrada: `allow: [$gostd]` com `files: [$all]` restringia *todo* import do módulo a só biblioteca padrão — só não quebrava o build porque o linter nunca rodava.
- **Direção da cotação de câmbio invertida**: `aplicarCotacao` consultava `(base=título, cotação=pagamento)`, mas a própria convenção documentada no ADR-006 cadastra `(base=pagamento, cotação=título)`. Nenhum teste pegou isso porque os testes do motor passam a cotação manualmente, pulando a consulta real — só apareceu rodando `docker compose run carga` contra o banco de verdade, com erro "cotação: registro não encontrado". Fechado com um teste de integração novo que exercita a consulta real, não só o cálculo.
- **Ciclo de vida da transação de demonstração quebrado**: `criarTransacao` inser/a transação já com o status final (`LIQUIDADA`) e depois tentava "liquidar" de novo — o próprio guard de optimistic locking rejeitava corretamente (`status != PENDENTE`), e a carga de demonstração falhava com "conflito de versão ou status". Corrigido para toda transação nascer `PENDENTE` e ser liquidada de verdade via `TransacaoRepo.Liquidar` quando o status alvo é `LIQUIDADA`.

## Onde a IA ajudou

- Verificação empírica de comportamentos de bibliotecas (apd, Next.js 16, TS 7, golangci-lint v2, Huma v2, testcontainers-go).
- Estruturação de ADRs no formato canônico.
- Cobertura de casos de borda para os testes unitários.
- Rodar o sistema de ponta a ponta via Docker real revelou bugs que nenhuma leitura de código ou teste isolado capturava — a lição prática é que "compila e o teste unitário passa" não é prova de que o sistema funciona.

## Onde a IA atrapalhou

- Sugeriu `chi` por familiaridade; o stack canônico da organização é Gin.
- Sugeriu `swag` para OpenAPI, registrou a correção como pendente, e não aplicou — precisou de uma segunda passada dedicada a *executar* o sistema, não só documentá-lo como corrigido.
- O primeiro `depguard` tinha configuração permissiva demais — só pegaria violação com o `linter` de fato rodando, o que só aconteceu na segunda passada.
- Declarou tarefas como concluídas (`tasks.md`) sem executar o critério de verificação escrito nelas — build e teste unitário passavam, mas o sistema nunca tinha subido de fato via Docker Compose. A lição foi metodológica: "critério de verificação" só vale alguma coisa se for de fato executado, não assumido.
