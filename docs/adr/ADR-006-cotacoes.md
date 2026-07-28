# ADR-006 — Cotações direcionadas, conversão no final

**Decisão:** `cotacoes_cambio(moeda_base_id, moeda_cotacao_id, taxa)`. `taxa` = quantas unidades de `moeda_cotacao` valem 1 unidade de `moeda_base`. Conversão inversa por divisão a partir da linha canônica.

**Consequência:** a direção fica explícita no esquema. A conversão é aplicada depois do desconto, como §3.2 determina.
