# ADR-003 — Prazo fracionário em meses

**Decisão:** `prazo = (vencimento − operação) em dias corridos ÷ 30`. Expoente fracionário, capitalização composta.

**Ambiguidade do enunciado:** os spreads são declarados "a.m." mas a fórmula não define unidade do prazo. Decidido por 30 dias corridos: convenção de mercado para desconto de duplicata, validada com o usuário.

**Consequência:** vencimento e data de operação são datas puras (`DATE` no banco). Cálculo do prazo é determinístico e independe do relógio.
