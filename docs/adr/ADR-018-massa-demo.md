# ADR-018 — Massa de demonstração separada das migrations

**Decisão:** migrations criam só dados de referência (moedas, tipos). Massa de demonstração (cedentes, cotações, taxas, transações) é carregada por comando próprio, acionado explicitamente.

**Motivo:** a aplicação precisa funcionar com o banco apenas migrado. A demonstração precisa ter o que mostrar. A massa calcula os valores das transações pelo próprio motor, com as taxas vigentes na data de operação de cada registro, para que o extrato feche com a fórmula.
