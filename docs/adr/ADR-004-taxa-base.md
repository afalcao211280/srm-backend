# ADR-004 — Taxa Base como tabela versionada

**Decisão:** tabela `taxas_base(moeda_id, taxa_mensal, vigencia_inicio, vigencia_fim)`. Resolução pela data de operação, não pela data corrente.

**Ambiguidade do enunciado:** `Taxa Base` aparece na fórmula e nunca é definida. O enunciado não diz valor, origem, nem se é CDI/SELIC. Modelar como dado versionado espelha o Currency Engine de §3.1 e casa com a tabela "Taxas" que §7 pede no ER.

**Consequência:** a taxa usada na precificação é congelada na própria transação (`taxa_base_aplicada`). Reprocessar uma operação antiga com a taxa vigente na data dela é determinístico mesmo que a taxa vigente atual seja outra.
