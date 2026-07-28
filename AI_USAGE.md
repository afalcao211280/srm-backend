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

## Onde a IA ajudou

- Verificação empírica de comportamentos de bibliotecas (apd, Next.js 16, TS 7).
- Estruturação de ADRs no formato canônico.
- Cobertura de casos de borda para os testes unitários.

## Onde a IA atrapalhou

- Sugeriu `chi` por familiaridade; o stack canônico da organização é Gin.
- Sugeriu `swag` para OpenAPI; o stack canônico é Huma.
- O primeiro `depguard` tinha configuração permissiva demais — só pegaria violação com `linter` rodando.
