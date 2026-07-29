# ADR-017 — Data de operação em America/Sao_Paulo

**Decisão:** data de operação ausente é resolvida como a data corrente em `America/Sao_Paulo`. Datas de negócio são `DATE`; instantes de auditoria são `TIMESTAMPTZ` em UTC.

**Motivo:** containers rodam em UTC por padrão. Entre 21h e 00h no horário de Brasília, a data em UTC já é a do dia seguinte — o sistema registraria data de operação que o operador não reconhece, o prazo mudaria em um dia e o valor calculado junto.
