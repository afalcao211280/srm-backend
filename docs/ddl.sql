-- SRM Credit Engine — DDL consolidado
-- Gerado a partir de migrations/000001_initial_schema.up.sql.
-- Reflete o esquema criado pelas migrations; mantido em docs/ para consulta
-- sem precisar executar o pipeline de migrations.

CREATE TABLE moedas (
    id SMALLSERIAL PRIMARY KEY,
    codigo CHAR(3) NOT NULL,
    nome VARCHAR(60) NOT NULL,
    casas_decimais SMALLINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_moedas_codigo UNIQUE (codigo),
    CONSTRAINT ck_moedas_casas CHECK (casas_decimais >= 0 AND casas_decimais <= 18)
);

CREATE TABLE tipos_recebivel (
    id SMALLSERIAL PRIMARY KEY,
    codigo VARCHAR(40) NOT NULL,
    nome VARCHAR(80) NOT NULL,
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_tipos_recebivel_codigo UNIQUE (codigo)
);

CREATE TABLE cedentes (
    id BIGSERIAL PRIMARY KEY,
    nome VARCHAR(120) NOT NULL,
    documento VARCHAR(20) NOT NULL,
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_cedentes_documento UNIQUE (documento)
);

CREATE TABLE cotacoes_cambio (
    id BIGSERIAL PRIMARY KEY,
    moeda_base_id SMALLINT NOT NULL REFERENCES moedas(id),
    moeda_cotacao_id SMALLINT NOT NULL REFERENCES moedas(id),
    taxa NUMERIC(18, 10) NOT NULL,
    vigencia_inicio TIMESTAMPTZ NOT NULL,
    vigencia_fim TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ck_cotacoes_pares CHECK (moeda_base_id <> moeda_cotacao_id),
    CONSTRAINT ck_cotacoes_taxa_positiva CHECK (taxa > 0),
    CONSTRAINT ck_cotacoes_vigencia CHECK (vigencia_fim IS NULL OR vigencia_fim > vigencia_inicio)
);
CREATE INDEX ix_cotacoes_cambio_pares ON cotacoes_cambio (moeda_base_id, moeda_cotacao_id, vigencia_inicio);

CREATE TABLE taxas_base (
    id BIGSERIAL PRIMARY KEY,
    moeda_id SMALLINT NOT NULL REFERENCES moedas(id),
    taxa_mensal NUMERIC(18, 10) NOT NULL,
    vigencia_inicio TIMESTAMPTZ NOT NULL,
    vigencia_fim TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ck_taxas_base_positiva CHECK (taxa_mensal > 0),
    CONSTRAINT ck_taxas_base_vigencia CHECK (vigencia_fim IS NULL OR vigencia_fim > vigencia_inicio)
);
CREATE INDEX ix_taxas_base_moeda ON taxas_base (moeda_id, vigencia_inicio);

CREATE TABLE transacoes (
    id UUID PRIMARY KEY,
    cedente_id BIGINT NOT NULL REFERENCES cedentes(id),
    tipo_recebivel_id SMALLINT NOT NULL REFERENCES tipos_recebivel(id),
    moeda_titulo_id SMALLINT NOT NULL REFERENCES moedas(id),
    moeda_pagamento_id SMALLINT NOT NULL REFERENCES moedas(id),
    valor_face NUMERIC(20, 2) NOT NULL,
    valor_presente NUMERIC(24, 8) NOT NULL,
    valor_liquido NUMERIC(20, 2) NOT NULL,
    desagio NUMERIC(20, 2) NOT NULL,
    spread_aplicado NUMERIC(18, 10) NOT NULL,
    taxa_base_aplicada NUMERIC(18, 10) NOT NULL,
    cotacao_aplicada NUMERIC(18, 10),
    data_operacao DATE NOT NULL,
    data_vencimento DATE NOT NULL,
    status VARCHAR(15) NOT NULL,
    versao INTEGER NOT NULL DEFAULT 1,
    liquidada_em TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ck_transacoes_valor_face CHECK (valor_face > 0),
    CONSTRAINT ck_transacoes_vencimento CHECK (data_vencimento > data_operacao),
    CONSTRAINT ck_transacoes_status CHECK (status IN ('PENDENTE', 'LIQUIDADA', 'CANCELADA')),
    CONSTRAINT ck_transacoes_versao CHECK (versao >= 1),
    CONSTRAINT ck_transacoes_cotacao CHECK (cotacao_aplicada IS NULL OR cotacao_aplicada > 0)
);
CREATE INDEX ix_transacoes_data_cedente ON transacoes (data_operacao, cedente_id);
CREATE INDEX ix_transacoes_moeda_pag_data ON transacoes (moeda_pagamento_id, data_operacao);
CREATE INDEX ix_transacoes_status ON transacoes (status);

CREATE TABLE transacao_auditoria (
    id BIGSERIAL PRIMARY KEY,
    transacao_id UUID NOT NULL REFERENCES transacoes(id),
    evento VARCHAR(40) NOT NULL,
    versao INTEGER NOT NULL,
    instante TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ck_transacao_auditoria_evento CHECK (evento IN ('CRIADA', 'LIQUIDADA', 'CANCELADA'))
);
CREATE INDEX ix_transacao_auditoria_transacao ON transacao_auditoria (transacao_id, instante);
