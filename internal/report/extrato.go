package report

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/srm-asset/srm-backend/internal/app/middleware"
)

type Filtros struct {
	DataInicial   *time.Time
	DataFinal     *time.Time
	CedenteID     *int64
	Moeda         string
	TipoRecebivel string
	Status        string
	Pagina        int
	Tamanho       int
}

type Linha struct {
	ID               string     `json:"id"`
	DataOperacao     string     `json:"data_operacao"`
	DataVencimento   string     `json:"data_vencimento"`
	CedenteNome      string     `json:"cedente_nome"`
	CedenteDocumento string     `json:"cedente_documento"`
	TipoRecebivel    string     `json:"tipo_recebivel"`
	MoedaTitulo      string     `json:"moeda_titulo"`
	MoedaPagamento   string     `json:"moeda_pagamento"`
	ValorFace        string     `json:"valor_face"`
	ValorPresente    string     `json:"valor_presente"`
	ValorLiquido     string     `json:"valor_liquido"`
	Desagio          string     `json:"desagio"`
	SpreadAplicado   string     `json:"spread_aplicado"`
	TaxaBaseAplicada string     `json:"taxa_base_aplicada"`
	CotacaoAplicada  *string    `json:"cotacao_aplicada,omitempty"`
	Status           string     `json:"status"`
	Versao           int32      `json:"versao"`
	LiquidadaEm      *time.Time `json:"liquidada_em,omitempty"`
}

type Resposta struct {
	Total        int64   `json:"total"`
	Pagina       int     `json:"pagina"`
	Tamanho      int     `json:"tamanho"`
	TotalPaginas int     `json:"total_paginas"`
	Items        []Linha `json:"items"`
}

const tamanhoMaximo = 100

type Extrato struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NovoExtrato(pool *pgxpool.Pool, logger *slog.Logger) *Extrato {
	return &Extrato{pool: pool, logger: logger}
}

// extratoInput espelha os mesmos parâmetros de query que parseFiltros já
// validava quando a rota era um handler Gin — os campos ficam como string
// para que a validação de formato (pagina/tamanho/datas) continue sendo
// feita por parseFiltros, já coberta por teste, em vez de duplicada em tags
// de schema do Huma.
type extratoInput struct {
	Pagina        string `query:"pagina" doc:"Página, começando em 1" example:"1"`
	Tamanho       string `query:"tamanho" doc:"Tamanho da página, máximo 100" example:"20"`
	Moeda         string `query:"moeda" doc:"Filtra pela moeda de pagamento (ex: BRL, USD)"`
	TipoRecebivel string `query:"tipo_recebivel" doc:"Filtra pelo código do tipo de recebível"`
	Status        string `query:"status" doc:"Filtra pelo status da transação"`
	CedenteID     string `query:"cedente_id" doc:"Filtra pelo id do cedente"`
	DataInicial   string `query:"data_inicial" doc:"Data inicial do período (AAAA-MM-DD)"`
	DataFinal     string `query:"data_final" doc:"Data final do período (AAAA-MM-DD)"`
}

type extratoOutput struct {
	Body Resposta
}

// paraQuery converte os campos string do input Huma no mesmo formato
// map[string][]string que parseFiltros já validava quando a rota era um
// handler Gin (query.Values()).
func (in *extratoInput) paraQuery() map[string][]string {
	q := map[string][]string{}
	add := func(k, v string) {
		if v != "" {
			q[k] = []string{v}
		}
	}
	add("pagina", in.Pagina)
	add("tamanho", in.Tamanho)
	add("moeda", in.Moeda)
	add("tipo_recebivel", in.TipoRecebivel)
	add("status", in.Status)
	add("cedente_id", in.CedenteID)
	add("data_inicial", in.DataInicial)
	add("data_final", in.DataFinal)
	return q
}

// MontarHuma registra a operação do extrato de liquidação na API Huma. É o
// único ponto de entrada HTTP deste pacote — mantém a exceção arquitetural
// de 2 camadas (ADR-007): nenhuma dependência de internal/domain aqui.
func MontarHuma(api huma.API, pool *pgxpool.Pool, logger *slog.Logger) {
	extrato := NovoExtrato(pool, logger)
	huma.Register(api, huma.Operation{
		OperationID: "extrato-liquidacao",
		Method:      http.MethodGet,
		Path:        "/api/v1/relatorios/extrato-liquidacao",
		Summary:     "Extrato de liquidação",
		Description: "Consulta analítica de transações filtrável por período, cedente, moeda, tipo de recebível e status, com paginação server-side.",
		Tags:        []string{"Relatórios"},
	}, func(ctx context.Context, in *extratoInput) (*extratoOutput, error) {
		return executarConsulta(ctx, extrato, in)
	})
}

func executarConsulta(ctx context.Context, extrato *Extrato, in *extratoInput) (*extratoOutput, error) {
	filtros, err := parseFiltros(in.paraQuery())
	if err != nil {
		return nil, middleware.HumaErrStatus(ctx, middleware.StatusInfo{
			Status: http.StatusUnprocessableEntity, Codigo: "entrada_invalida",
		}, err.Error())
	}
	res, err := extrato.Consultar(ctx, filtros)
	if err != nil {
		extrato.logger.Error("extrato", slog.String("erro", err.Error()))
		return nil, middleware.HumaErrStatus(ctx, middleware.StatusInfo{
			Status: http.StatusInternalServerError, Codigo: "erro_interno",
		}, "falha ao consultar extrato")
	}
	return &extratoOutput{Body: res}, nil
}

type getter func(k string) string

func novoGetter(q map[string][]string) getter {
	return func(k string) string {
		if v, ok := q[k]; ok && len(v) > 0 {
			return v[0]
		}
		return ""
	}
}

func parseFiltros(q map[string][]string) (Filtros, error) {
	get := novoGetter(q)
	f := Filtros{
		Moeda:         get("moeda"),
		TipoRecebivel: get("tipo_recebivel"),
		Status:        get("status"),
		Pagina:        1,
		Tamanho:       20,
	}
	if err := parsePaginacao(get, &f); err != nil {
		return f, err
	}
	if err := parsePeriodo(get, &f); err != nil {
		return f, err
	}
	if s := get("cedente_id"); s != "" {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return f, errors.New("cedente_id inválido")
		}
		f.CedenteID = &n
	}
	return f, nil
}

func parsePaginacao(get getter, f *Filtros) error {
	if s := get("pagina"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			return errors.New("pagina inválida")
		}
		f.Pagina = n
	}
	if s := get("tamanho"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			return errors.New("tamanho inválido")
		}
		if n > tamanhoMaximo {
			return fmt.Errorf("tamanho máximo permitido é %d", tamanhoMaximo)
		}
		f.Tamanho = n
	}
	return nil
}

func parsePeriodo(get getter, f *Filtros) error {
	if s := get("data_inicial"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return errors.New("data_inicial inválida")
		}
		f.DataInicial = &t
	}
	if s := get("data_final"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return errors.New("data_final inválida")
		}
		f.DataFinal = &t
	}
	if f.DataInicial != nil && f.DataFinal != nil && f.DataInicial.After(*f.DataFinal) {
		return errors.New("data_inicial não pode ser posterior a data_final")
	}
	return nil
}

func (e *Extrato) Consultar(ctx context.Context, f Filtros) (Resposta, error) {
	sb := baseSelect(f)
	sb = aplicarFiltros(sb, f)
	sql, args, err := sb.ToSql()
	if err != nil {
		return Resposta{}, fmt.Errorf("montar sql: %w", err)
	}
	rows, err := e.pool.Query(ctx, sql, args...)
	if err != nil {
		return Resposta{}, fmt.Errorf("executar consulta: %w", err)
	}
	defer rows.Close()
	itens, total, err := escanearLinhas(rows)
	if err != nil {
		return Resposta{}, err
	}
	totalPaginas := 0
	if total > 0 {
		totalPaginas = int((total + int64(f.Tamanho) - 1) / int64(f.Tamanho))
	}
	return Resposta{
		Total:        total,
		Pagina:       f.Pagina,
		Tamanho:      f.Tamanho,
		TotalPaginas: totalPaginas,
		Items:        itens,
	}, nil
}

func baseSelect(f Filtros) squirrel.SelectBuilder {
	return squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select(
			"t.id::text", "t.data_operacao::text", "t.data_vencimento::text",
			"c.nome", "c.documento",
			"tr.codigo", "mt.codigo", "mp.codigo",
			"t.valor_face::text", "t.valor_presente::text", "t.valor_liquido::text", "t.desagio::text",
			"t.spread_aplicado::text", "t.taxa_base_aplicada::text", "t.cotacao_aplicada::text",
			"t.status", "t.versao", "t.liquidada_em",
			"COUNT(*) OVER() AS total",
		).
		From("transacoes t").
		Join("cedentes c ON c.id = t.cedente_id").
		Join("tipos_recebivel tr ON tr.id = t.tipo_recebivel_id").
		Join("moedas mt ON mt.id = t.moeda_titulo_id").
		Join("moedas mp ON mp.id = t.moeda_pagamento_id").
		OrderBy("t.data_operacao DESC, t.id").
		Limit(uint64(f.Tamanho)).                  //nolint:gosec // f.Tamanho já validado positivo (1..100) em parseFiltros
		Offset(uint64((f.Pagina - 1) * f.Tamanho)) //nolint:gosec // f.Pagina já validado >=1 em parseFiltros
}

func aplicarFiltros(sb squirrel.SelectBuilder, f Filtros) squirrel.SelectBuilder {
	if f.DataInicial != nil {
		sb = sb.Where(squirrel.GtOrEq{"t.data_operacao": *f.DataInicial})
	}
	if f.DataFinal != nil {
		sb = sb.Where(squirrel.LtOrEq{"t.data_operacao": *f.DataFinal})
	}
	if f.CedenteID != nil {
		sb = sb.Where(squirrel.Eq{"t.cedente_id": *f.CedenteID})
	}
	if f.Moeda != "" {
		sb = sb.Where(squirrel.Eq{"mp.codigo": f.Moeda})
	}
	if f.TipoRecebivel != "" {
		sb = sb.Where(squirrel.Eq{"tr.codigo": f.TipoRecebivel})
	}
	if f.Status != "" {
		sb = sb.Where(squirrel.Eq{"t.status": f.Status})
	}
	return sb
}

func escanearLinhas(rows pgx.Rows) ([]Linha, int64, error) {
	var itens []Linha
	var total int64
	for rows.Next() {
		var l Linha
		var totalLinha int64
		if err := rows.Scan(
			&l.ID, &l.DataOperacao, &l.DataVencimento,
			&l.CedenteNome, &l.CedenteDocumento,
			&l.TipoRecebivel, &l.MoedaTitulo, &l.MoedaPagamento,
			&l.ValorFace, &l.ValorPresente, &l.ValorLiquido, &l.Desagio,
			&l.SpreadAplicado, &l.TaxaBaseAplicada, &l.CotacaoAplicada,
			&l.Status, &l.Versao, &l.LiquidadaEm,
			&totalLinha,
		); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		total = totalLinha
		itens = append(itens, l)
	}
	if err := rows.Err(); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, 0, fmt.Errorf("iterar: %w", err)
	}
	return itens, total, nil
}
