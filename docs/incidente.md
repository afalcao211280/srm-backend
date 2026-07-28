# Simulação de gestão de crise

Episódio documentado como exercício de gestão de incidente.

## Problema

Foi introduzido um bug crítico no `srm-backend` em que a conversão cambial era aplicada **antes** do desconto, violando §3.2 do case. O motor de precificação multiplicava o valor de face pela cotação e só depois aplicava o fator `(1 + Taxa Base + Spread)^Prazo`, o que produzia valores líquidos em moeda de pagamento drasticamente diferentes do esperado.

## Decisão

Aplicar `git revert` do commit que introduziu o bug, preservando o histórico do incidente, e em seguida aplicar a correção definitiva em branch de hotfix via `git cherry-pick`.

## Comandos

```bash
# Identificar o commit do bug
git -C srm-backend log --oneline

# Reverter
git -C srm-backend revert <hash-do-bug>

# Hotfix em branch separada
git -C srm-backend checkout -b hotfix/conversao-antes-depois
# aplicar a correção
git -C srm-backend commit -m "fix: aplica conversao cambial apos o desconto"

# Cherry-pick para a main
git -C srm-backend checkout main
git -C srm-backend cherry-pick hotfix/conversao-antes-depois
```

## Verificação

A suíte de testes (`go test ./...` + `go test -tags=integration ./...`) passa integralmente após o episódio.
