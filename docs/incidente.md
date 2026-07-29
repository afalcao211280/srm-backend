# Simulação de gestão de crise

Episódio real, executado neste repositório (não um roteiro hipotético) — os
hashes abaixo existem no histórico de `main`.

## Problema

Foi introduzido em `main` o commit `4f9527b` ("perf: converte valor de face
antes do calculo do desconto"), que aplicava a conversão cambial **antes**
do desconto, violando §3.2 do case. O motor dividia o valor de face pela
cotação e só depois calculava `(1 + Taxa Base + Spread)^Prazo` sobre o valor
já convertido — como `finalizarValores` também convertia o valor presente
resultante, o valor líquido cross-currency era convertido **duas vezes**.

O bug foi confirmado com o próprio teste de integração já existente
(`TestSimularCrossCurrencyRoundTrip`, que espera `1773.97` USD): com o
commit aplicado, o teste falhava com `obtido 326.57` — aproximadamente
`1773.97 / 5.4321`, a assinatura exata de uma dupla conversão.

## Decisão

`git revert` imediato para parar a sangria, sem investigar a causa raiz sob
pressão. Em seguida, endurecimento: um teste de regressão explícito
provando a invariante que o bug violou, desenvolvido isolado em branch de
hotfix e trazido para `main` via `git cherry-pick` — não um novo revert,
para deixar claro no histórico que se trata de um reforço, não de desfazer
algo.

## O que foi executado, de verdade

```bash
# 1. Commit do bug em main (simula que passou por revisão e foi mergeado)
git -C srm-backend commit -m "perf: converte valor de face antes do calculo do desconto"
# → 4f9527ba5f0e9bdb9e5e1f6cfe85d322a44736c4

# 2. Confirmar que quebra de verdade antes de reverter
go test -tags=integration ./internal/app/handler/... -run TestSimularCrossCurrencyRoundTrip -v
# --- FAIL: esperado 1773.97, obtido 326.57

# 3. Revert
git -C srm-backend revert --no-edit 4f9527ba5f0e9bdb9e5e1f6cfe85d322a44736c4
# → 036f1c1 "Revert "perf: converte valor de face antes do calculo do desconto""

# 4. Hotfix em branch separada: teste de regressão que prova a invariante
#    (ValorPresente8Casas não pode depender da cotação de câmbio)
git -C srm-backend checkout -b hotfix/guarda-ordem-conversao-cambial
git -C srm-backend commit -m "test: adiciona guarda de regressao contra conversao cambial antecipada"
# → c8a95c14ea431cf540b2e117698ae2ec90020c26

# Antes de commitar o teste, reaplicado 4f9527b em --no-commit sobre a
# branch de hotfix só para confirmar que o novo teste também o pega:
# FAIL — "1773.97071276 (cotação 5.4321) vs 963.64826736 (cotação 9.9999)"

# 5. Cherry-pick do hardening para main (não é outro revert — é reforço)
git -C srm-backend checkout main
git -C srm-backend cherry-pick c8a95c14ea431cf540b2e117698ae2ec90020c26
# → bd5c8e3, mesmo diff, novo hash (cherry-pick sempre cria commit novo)
```

## Verificação

`go build ./...`, `go test ./... -race` e `go test -tags=integration ./... -race`
passam integralmente em `main` depois do episódio completo (revert +
cherry-pick). O bug em si nunca chegou a ficar em `main` sem o revert — a
sequência de commits (`4f9527b` → `036f1c1` → `bd5c8e3`) fica visível e
rastreável no histórico, contando a história completa: o que quebrou, como
foi revertido, e o que foi reforçado depois.
