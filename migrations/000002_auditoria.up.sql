CREATE TABLE transacao_auditoria (
    id BIGSERIAL PRIMARY KEY,
    transacao_id UUID NOT NULL REFERENCES transacoes(id),
    evento VARCHAR(40) NOT NULL,
    versao INTEGER NOT NULL,
    instante TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ck_transacao_auditoria_evento CHECK (evento IN ('CRIADA', 'LIQUIDADA', 'CANCELADA'))
);
CREATE INDEX ix_transacao_auditoria_transacao ON transacao_auditoria (transacao_id, instante);
