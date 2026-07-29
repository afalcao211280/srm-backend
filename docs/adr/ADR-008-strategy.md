# ADR-008 — Strategy com spread no código

**Decisão:** `SpreadStrategy` com `DuplicataMercantil` (1,5% a.m.) e `ChequePreDatado` (2,5% a.m.), resolvidas por registry indexado pelo código. `tipos_recebivel` guarda `codigo` e `nome`, **não** guarda o spread.

**Motivo:** §3.2 do case pede o padrão Strategy explicitamente. Manter o valor em código e em tabela cria duas fontes de verdade.
