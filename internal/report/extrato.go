package report

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Filtros struct {
	DataInicial  *time.Time
	DataFinal    *time.Time
	CedenteID    *int64
	Moeda        string
	TipoRecebivel string
	Status       string
	Pagina       int
	Tamanho      int
}

type Linha struct {
	ID              string `json:"id"`
	DataOperacao    string `json:"data_operacao"`
	DataVencimento  string `json:"data_vencimento"`
	CedenteNome     string `json:"cedente_nome"`
	CedenteDocumento string `json:"cedente_documento"`
	TipoRecebivel   string `json:"tipo_recebivel"`
	MoedaTitulo     string `json:"moeda_titulo"`
	MoedaPagamento  string `json:"moeda_pagamento"`
	ValorFace       string `json:"valor_face"`
	ValorPresente   string `json:"valor_presente"`
	ValorLiquido    string `json:"valor_liquido"`
	Desagio         string `json:"desagio"`
	SpreadAplicado  string `json:"spread_aplicado"`
	TaxaBaseAplicada string `json:"taxa_base_aplicada"`
	CotacaoAplicada *string `json:"cotacao_aplicada,omitempty"`
	Status          string `json:"status"`
	Versao          int32  `json:"versao"`
	LiquidadaEm     *time.Time `json:"liquidada_em,omitempty"`
}

type Resposta struct {
	Total       int64  `json:"total"`
	Pagina      int    `json:"pagina"`
	Tamanho     int    `json:"tamanho"`
	TotalPaginas int   `json:"total_paginas"`
	Items       []Linha `json:"items"`
}

const tamanhoMaximo = 100

type Extrato struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NovoExtrato(pool *pgxpool.Pool, logger *slog.Logger) *Extrato {
	return &Extrato{pool: pool, logger: logger}
}

func (e *Extrato) Handler(c *gin.Context) {
	filtros, err := parseFiltros(c.Request.URL.Query())
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"erro": err.Error()})
		return
	}
	res, err := e.Consultar(c.Request.Context(), filtros)
	if err != nil {
		e.logger.Error("extrato", slog.String("erro", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao consultar extrato"})
		return
	}
	c.Header("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(c.Writer).Encode(res)
}

func parseFiltros(q map[string][]string) (Filtros, error) {
	get := func(k string) string {
		if v, ok := q[k]; ok && len(v) > 0 {
			return v[0]
		}
		return ""
	}
	f := Filtros{
		Moeda:         get("moeda"),
		TipoRecebivel: get("tipo_recebivel"),
		Status:        get("status"),
		Pagina:        1,
		Tamanho:       20,
	}
	if s := get("pagina"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			return f, errors.New("pagina inválida")
		}
		f.Pagina = n
	}
	if s := get("tamanho"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			return f, errors.New("tamanho inválido")
		}
		if n > tamanhoMaximo {
			return f, fmt.Errorf("tamanho máximo permitido é %d", tamanhoMaximo)
		}
		f.Tamanho = n
	}
	if s := get("data_inicial"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return f, errors.New("data_inicial inválida")
		}
		f.DataInicial = &t
	}
	if s := get("data_final"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return f, errors.New("data_final inválida")
		}
		f.DataFinal = &t
	}
	if f.DataInicial != nil && f.DataFinal != nil && f.DataInicial.After(*f.DataFinal) {
		return f, errors.New("data_inicial não pode ser posterior a data_final")
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

func (e *Extrato) Consultar(ctx context.Context, f Filtros) (Resposta, error) {
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
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
		Limit(uint64(f.Tamanho)).
		Offset(uint64((f.Pagina - 1) * f.Tamanho))

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
	sql, args, err := sb.ToSql()
	if err != nil {
		return Resposta{}, fmt.Errorf("montar sql: %w", err)
	}
	rows, err := e.pool.Query(ctx, sql, args...)
	if err != nil {
		return Resposta{}, fmt.Errorf("executar consulta: %w", err)
	}
	defer rows.Close()
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
			return Resposta{}, fmt.Errorf("scan: %w", err)
		}
		total = totalLinha
		itens = append(itens, l)
	}
	if err := rows.Err(); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Resposta{}, fmt.Errorf("iterar: %w", err)
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
