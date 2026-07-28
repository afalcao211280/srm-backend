package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/srm-asset/srm-backend/internal/domain/money"
)

type Moeda struct {
	ID         int16
	Codigo     string
	Nome       string
	casasDec   int16
}

func (m Moeda) CasasDecimais() int16 { return m.casasDec }

type MoedaRepo struct{ pool *pgxpool.Pool }

func NewMoedaRepo(pool *pgxpool.Pool) *MoedaRepo { return &MoedaRepo{pool: pool} }

func (r *MoedaRepo) PorCodigo(ctx context.Context, codigo string) (Moeda, error) {
	var m Moeda
	err := r.pool.QueryRow(ctx,
		`SELECT id, codigo, nome, casas_decimais FROM moedas WHERE codigo = $1`,
		codigo,
	).Scan(&m.ID, &m.Codigo, &m.Nome, &m.casasDec)
	if errors.Is(err, pgx.ErrNoRows) {
		return Moeda{}, ErrNaoEncontrado
	}
	return m, err
}

var ErrNaoEncontrado = errors.New("registro não encontrado")

type TipoRecebivel struct {
	ID     int16
	Codigo string
	Nome   string
	Ativo  bool
}

type TipoRecebivelRepo struct{ pool *pgxpool.Pool }

func NewTipoRecebivelRepo(pool *pgxpool.Pool) *TipoRecebivelRepo {
	return &TipoRecebivelRepo{pool: pool}
}

func (r *TipoRecebivelRepo) PorCodigo(ctx context.Context, codigo string) (TipoRecebivel, error) {
	var t TipoRecebivel
	err := r.pool.QueryRow(ctx,
		`SELECT id, codigo, nome, ativo FROM tipos_recebivel WHERE codigo = $1 AND ativo = TRUE`,
		codigo,
	).Scan(&t.ID, &t.Codigo, &t.Nome, &t.Ativo)
	if errors.Is(err, pgx.ErrNoRows) {
		return TipoRecebivel{}, ErrNaoEncontrado
	}
	return t, err
}

func (r *TipoRecebivelRepo) Listar(ctx context.Context) ([]TipoRecebivel, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, codigo, nome, ativo FROM tipos_recebivel WHERE ativo = TRUE ORDER BY codigo`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TipoRecebivel
	for rows.Next() {
		var t TipoRecebivel
		if err := rows.Scan(&t.ID, &t.Codigo, &t.Nome, &t.Ativo); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

type Cedente struct {
	ID        int64
	Nome      string
	Documento string
	Ativo     bool
}

type CedenteRepo struct{ pool *pgxpool.Pool }

func NewCedenteRepo(pool *pgxpool.Pool) *CedenteRepo { return &CedenteRepo{pool: pool} }

func (r *CedenteRepo) Criar(ctx context.Context, c Cedente) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx,
		`INSERT INTO cedentes (nome, documento) VALUES ($1, $2) RETURNING id`,
		c.Nome, c.Documento,
	).Scan(&id)
	return id, err
}

func (r *CedenteRepo) PorID(ctx context.Context, id int64) (Cedente, error) {
	var c Cedente
	err := r.pool.QueryRow(ctx,
		`SELECT id, nome, documento, ativo FROM cedentes WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.Nome, &c.Documento, &c.Ativo)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cedente{}, ErrNaoEncontrado
	}
	return c, err
}

type Cotacao struct {
	ID              int64
	MoedaBase       string
	MoedaCotacao    string
	Taxa            money.Decimal
	VigenciaInicio  time.Time
	VigenciaFim     *time.Time
}

type CotacaoRepo struct{ pool *pgxpool.Pool }

func NewCotacaoRepo(pool *pgxpool.Pool) *CotacaoRepo { return &CotacaoRepo{pool: pool} }

func (r *CotacaoRepo) Vigente(ctx context.Context, base, cotacao string, em time.Time) (Cotacao, error) {
	var c Cotacao
	var vigenciaFim *time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT cc.id, mb.codigo, mc.codigo, cc.taxa, cc.vigencia_inicio, cc.vigencia_fim
		FROM cotacoes_cambio cc
		JOIN moedas mb ON mb.id = cc.moeda_base_id
		JOIN moedas mc ON mc.id = cc.moeda_cotacao_id
		WHERE mb.codigo = $1 AND mc.codigo = $2
		  AND cc.vigencia_inicio <= $3
		  AND (cc.vigencia_fim IS NULL OR cc.vigencia_fim > $3)
		ORDER BY cc.vigencia_inicio DESC
		LIMIT 1`,
		base, cotacao, em,
	).Scan(&c.ID, &c.MoedaBase, &c.MoedaCotacao, &c.Taxa, &c.VigenciaInicio, &vigenciaFim)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cotacao{}, ErrNaoEncontrado
	}
	c.VigenciaFim = vigenciaFim
	return c, err
}

func (r *CotacaoRepo) Criar(ctx context.Context, base, cotacao string, taxa money.Decimal, inicio time.Time) (Cotacao, error) {
	var idMoedaBase, idMoedaCotacao int16
	if err := r.pool.QueryRow(ctx, `SELECT id FROM moedas WHERE codigo = $1`, base).Scan(&idMoedaBase); err != nil {
		return Cotacao{}, fmt.Errorf("moeda base %s: %w", base, err)
	}
	if err := r.pool.QueryRow(ctx, `SELECT id FROM moedas WHERE codigo = $1`, cotacao).Scan(&idMoedaCotacao); err != nil {
		return Cotacao{}, fmt.Errorf("moeda cotação %s: %w", cotacao, err)
	}
	var c Cotacao
	var vigenciaFim *time.Time
	err := r.pool.QueryRow(ctx, `
		INSERT INTO cotacoes_cambio (moeda_base_id, moeda_cotacao_id, taxa, vigencia_inicio)
		VALUES ($1, $2, $3, $4)
		RETURNING id, $5::text, $6::text, taxa, vigencia_inicio, NULL::timestamptz`,
		idMoedaBase, idMoedaCotacao, taxa, inicio, base, cotacao,
	).Scan(&c.ID, &c.MoedaBase, &c.MoedaCotacao, &c.Taxa, &c.VigenciaInicio, &vigenciaFim)
	c.VigenciaFim = vigenciaFim
	return c, err
}

type TaxaBase struct {
	ID             int64
	Moeda          string
	TaxaMensal     money.Decimal
	VigenciaInicio time.Time
	VigenciaFim    *time.Time
}

type TaxaBaseRepo struct{ pool *pgxpool.Pool }

func NewTaxaBaseRepo(pool *pgxpool.Pool) *TaxaBaseRepo { return &TaxaBaseRepo{pool: pool} }

func (r *TaxaBaseRepo) Vigente(ctx context.Context, moeda string, em time.Time) (TaxaBase, error) {
	var t TaxaBase
	var vigenciaFim *time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT t.id, m.codigo, t.taxa_mensal, t.vigencia_inicio, t.vigencia_fim
		FROM taxas_base t
		JOIN moedas m ON m.id = t.moeda_id
		WHERE m.codigo = $1
		  AND t.vigencia_inicio <= $2
		  AND (t.vigencia_fim IS NULL OR t.vigencia_fim > $2)
		ORDER BY t.vigencia_inicio DESC
		LIMIT 1`,
		moeda, em,
	).Scan(&t.ID, &t.Moeda, &t.TaxaMensal, &t.VigenciaInicio, &vigenciaFim)
	if errors.Is(err, pgx.ErrNoRows) {
		return TaxaBase{}, ErrNaoEncontrado
	}
	t.VigenciaFim = vigenciaFim
	return t, err
}

// Criar registra uma nova taxa base encerrando a vigência anterior da mesma
// moeda, dentro de uma única transação, para impedir duas vigências
// simultâneas para a mesma moeda.
func (r *TaxaBaseRepo) Criar(ctx context.Context, moeda string, taxa money.Decimal, inicio time.Time) (TaxaBase, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return TaxaBase{}, err
	}
	defer tx.Rollback(ctx)
	var idMoeda int16
	if err := tx.QueryRow(ctx, `SELECT id FROM moedas WHERE codigo = $1`, moeda).Scan(&idMoeda); err != nil {
		return TaxaBase{}, fmt.Errorf("moeda %s: %w", moeda, err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE taxas_base
		   SET vigencia_fim = $1
		 WHERE moeda_id = $2
		   AND vigencia_fim IS NULL`,
		inicio, idMoeda,
	); err != nil {
		return TaxaBase{}, err
	}
	var t TaxaBase
	var vigenciaFim *time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO taxas_base (moeda_id, taxa_mensal, vigencia_inicio)
		VALUES ($1, $2, $3)
		RETURNING id, $4::text, taxa_mensal, vigencia_inicio, NULL::timestamptz`,
		idMoeda, taxa, inicio, moeda,
	).Scan(&t.ID, &t.Moeda, &t.TaxaMensal, &t.VigenciaInicio, &vigenciaFim)
	if err != nil {
		return TaxaBase{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return TaxaBase{}, err
	}
	t.VigenciaFim = vigenciaFim
	return t, nil
}
