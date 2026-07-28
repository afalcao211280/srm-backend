# ADR-005 — Arredondamento único

**Decisão:** precisão interna de 34 dígitos. Arredondamento `RoundHalfUp` aplicado **uma única vez** sobre o valor na moeda de pagamento, em 2 casas. Valor presente na moeda do título persistido em 8 casas para trilha de auditoria.

**Consequência:** o resultado é reproduzível. A conversão cambial parte do valor presente **não arredondado** para evitar duplo arredondamento.
