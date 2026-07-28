package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/srm-asset/srm-backend/internal/domain/money"
)

type TransacaoStatus string

const (
	StatusPendente  TransacaoStatus = "PENDENTE"
	StatusLiquidada TransacaoStatus = "LIQUIDADA"
	StatusCancelada TransacaoStatus = "CANCELADA"
)

type Transacao struct {
	ID               string
	CedenteID        int64
	TipoRecebivelID  int16
	MoedaTituloID    int16
	MoedaPagamentoID int16
	ValorFace        money.Decimal
	ValorPresente    money.Decimal
	ValorLiquido     money.Decimal
	Desagio          money.Decimal
	SpreadAplicado   money.Decimal
	TaxaBaseAplicada money.Decimal
	CotacaoAplicada  *money.Decimal
	DataOperacao     time.Time
	DataVencimento   time.Time
	Status           TransacaoStatus
	Versao           int32
	LiquidadaEm      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type TransacaoRepo struct{ pool *pgxpool.Pool }

func NewTransacaoRepo(pool *pgxpool.Pool) *TransacaoRepo { return &TransacaoRepo{pool: pool} }

func (r *TransacaoRepo) Criar(ctx context.Context, t Transacao) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO transacoes (
			id, cedente_id, tipo_recebivel_id, moeda_titulo_id, moeda_pagamento_id,
			valor_face, valor_presente, valor_liquido, desagio,
			spread_aplicado, taxa_base_aplicada, cotacao_aplicada,
			data_operacao, data_vencimento, status, versao
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12,
			$13, $14, $15, 1
		)`,
		t.ID, t.CedenteID, t.TipoRecebivelID, t.MoedaTituloID, t.MoedaPagamentoID,
		t.ValorFace, t.ValorPresente, t.ValorLiquido, t.Desagio,
		t.SpreadAplicado, t.TaxaBaseAplicada, t.CotacaoAplicada,
		t.DataOperacao, t.DataVencimento, t.Status,
	)
	return err
}

func (r *TransacaoRepo) PorID(ctx context.Context, id string) (Transacao, error) {
	return scanTransacao(r.pool.QueryRow(ctx, `
		SELECT id, cedente_id, tipo_recebivel_id, moeda_titulo_id, moeda_pagamento_id,
		       valor_face, valor_presente, valor_liquido, desagio,
		       spread_aplicado, taxa_base_aplicada, cotacao_aplicada,
		       data_operacao, data_vencimento, status, versao, liquidada_em, created_at, updated_at
		  FROM transacoes WHERE id = $1`, id))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTransacao(row rowScanner) (Transacao, error) {
	var t Transacao
	var cotacaoAplicada *money.Decimal
	var liquidadaEm *time.Time
	err := row.Scan(
		&t.ID, &t.CedenteID, &t.TipoRecebivelID, &t.MoedaTituloID, &t.MoedaPagamentoID,
		&t.ValorFace, &t.ValorPresente, &t.ValorLiquido, &t.Desagio,
		&t.SpreadAplicado, &t.TaxaBaseAplicada, &cotacaoAplicada,
		&t.DataOperacao, &t.DataVencimento, &t.Status, &t.Versao, &liquidadaEm, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Transacao{}, ErrNaoEncontrado
	}
	if err != nil {
		return Transacao{}, err
	}
	t.CotacaoAplicada = cotacaoAplicada
	t.LiquidadaEm = liquidadaEm
	return t, nil
}

// LiquidacaoInput são os dados para liquidar uma transação.
type LiquidacaoInput struct {
	ID          string
	Versao      int32
	LiquidadaEm time.Time
}

// Liquidar executa o UPDATE condicional da liquidação com optimistic locking.
// Zero linhas afetadas → 409 Conflict. A auditoria é gravada na mesma
// transação de banco.
func (r *TransacaoRepo) Liquidar(ctx context.Context, in LiquidacaoInput) (Transacao, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Transacao{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := liquidarUpdate(ctx, tx, in); err != nil {
		return Transacao{}, err
	}
	if err := inserirAuditoriaLiquidacao(ctx, tx, in); err != nil {
		return Transacao{}, err
	}
	t, err := scanTransacao(tx.QueryRow(ctx, `
		SELECT id, cedente_id, tipo_recebivel_id, moeda_titulo_id, moeda_pagamento_id,
		       valor_face, valor_presente, valor_liquido, desagio,
		       spread_aplicado, taxa_base_aplicada, cotacao_aplicada,
		       data_operacao, data_vencimento, status, versao, liquidada_em, created_at, updated_at
		  FROM transacoes WHERE id = $1`, in.ID))
	if err != nil {
		return Transacao{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Transacao{}, err
	}
	return t, nil
}

type execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func liquidarUpdate(ctx context.Context, tx execer, in LiquidacaoInput) error {
	tag, err := tx.Exec(ctx, `
		UPDATE transacoes
		   SET status = 'LIQUIDADA',
		       versao = versao + 1,
		       liquidada_em = $1,
		       updated_at = now()
		 WHERE id = $2
		   AND versao = $3
		   AND status = 'PENDENTE'`,
		in.LiquidadaEm, in.ID, in.Versao,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErroConflitoLiquidacao
	}
	return nil
}

func inserirAuditoriaLiquidacao(ctx context.Context, tx execer, in LiquidacaoInput) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO transacao_auditoria (transacao_id, evento, versao, instante)
		VALUES ($1, 'LIQUIDADA', $2, $3)`,
		in.ID, in.Versao+1, in.LiquidadaEm,
	)
	return err
}

var ErroConflitoLiquidacao = errors.New("conflito de versão ou status na liquidação")

// Listar devolve uma página de transações sem joins, ordenada da mais
// recente para a mais antiga, com o total de registros para paginação.
// É o recurso REST genérico (GET /transacoes); o extrato com filtros e
// joins de cedente/tipo é um relatório à parte (internal/report).
func (r *TransacaoRepo) Listar(ctx context.Context, pagina, tamanho int) ([]Transacao, int64, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, cedente_id, tipo_recebivel_id, moeda_titulo_id, moeda_pagamento_id,
		       valor_face, valor_presente, valor_liquido, desagio,
		       spread_aplicado, taxa_base_aplicada, cotacao_aplicada,
		       data_operacao, data_vencimento, status, versao, liquidada_em, created_at, updated_at,
		       COUNT(*) OVER() AS total
		  FROM transacoes
		 ORDER BY data_operacao DESC, id
		 LIMIT $1 OFFSET $2`,
		tamanho, (pagina-1)*tamanho,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listar transações: %w", err)
	}
	defer rows.Close()
	out, total, err := escanearTransacoes(rows)
	if err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func escanearTransacoes(rows pgx.Rows) ([]Transacao, int64, error) {
	var out []Transacao
	var total int64
	for rows.Next() {
		var t Transacao
		var cotacaoAplicada *money.Decimal
		var liquidadaEm *time.Time
		if err := rows.Scan(
			&t.ID, &t.CedenteID, &t.TipoRecebivelID, &t.MoedaTituloID, &t.MoedaPagamentoID,
			&t.ValorFace, &t.ValorPresente, &t.ValorLiquido, &t.Desagio,
			&t.SpreadAplicado, &t.TaxaBaseAplicada, &cotacaoAplicada,
			&t.DataOperacao, &t.DataVencimento, &t.Status, &t.Versao, &liquidadaEm, &t.CreatedAt, &t.UpdatedAt,
			&total,
		); err != nil {
			return nil, 0, fmt.Errorf("scan transação: %w", err)
		}
		t.CotacaoAplicada = cotacaoAplicada
		t.LiquidadaEm = liquidadaEm
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterar transações: %w", err)
	}
	return out, total, nil
}
