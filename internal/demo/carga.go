// Package demo implementa a carga da massa de demonstração, separada das
// migrations (ADR-018). É acionada por comando próprio e gera cedentes,
// cotações, taxas base e transações calculadas pelo próprio motor, com as
// taxas vigentes na data de operação de cada registro — para que o extrato
// feche com a fórmula.
package demo

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/srm-asset/srm-backend/internal/domain/money"
	"github.com/srm-asset/srm-backend/internal/domain/precificacao"
	"github.com/srm-asset/srm-backend/internal/domain/recebivel"
	"github.com/srm-asset/srm-backend/internal/infra/postgres"
)

type Carga struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func Nova(pool *pgxpool.Pool, logger *slog.Logger) *Carga {
	return &Carga{pool: pool, logger: logger}
}

const (
	cedentesCount     = 8
	transacoesPorTipo = 25
)

func (c *Carga) Executar(ctx context.Context) error {
	if err := c.verificarIdempotencia(ctx); err != nil {
		return fmt.Errorf("verificar idempotência: %w", err)
	}
	if err := c.gerarCedentes(ctx); err != nil {
		return fmt.Errorf("gerar cedentes: %w", err)
	}
	if err := c.gerarTaxas(ctx); err != nil {
		return fmt.Errorf("gerar taxas: %w", err)
	}
	if err := c.gerarCotacoes(ctx); err != nil {
		return fmt.Errorf("gerar cotações: %w", err)
	}
	if err := c.gerarTransacoes(ctx); err != nil {
		return fmt.Errorf("gerar transações: %w", err)
	}
	c.logger.Info("massa de demonstração carregada com sucesso")
	return nil
}

func (c *Carga) verificarIdempotencia(ctx context.Context) error {
	var total int
	if err := c.pool.QueryRow(ctx, "SELECT count(*) FROM cedentes").Scan(&total); err != nil {
		return err
	}
	if total > 0 {
		return fmt.Errorf("cedentes já populados (%d) — carga já foi executada", total)
	}
	return nil
}

func (c *Carga) gerarCedentes(ctx context.Context) error {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	for i := 1; i <= cedentesCount; i++ {
		doc := fmt.Sprintf("%011d", 10000000000+int64(i))
		if _, err := tx.Exec(ctx,
			"INSERT INTO cedentes (nome, documento) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			fmt.Sprintf("Cedente Demonstração %02d", i), doc,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

type taxaSpec struct {
	Moeda  string
	Taxa   string
	Inicio time.Time
	Fim    *time.Time
}

func (c *Carga) gerarTaxas(ctx context.Context) error {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	fimAberto := (*time.Time)(nil)
	taxas := []taxaSpec{
		{"BRL", "0.0100", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), tempoFim(2026, 6, 30)},
		{"BRL", "0.0125", time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), fimAberto},
		{"USD", "0.0050", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), fimAberto},
	}
	for _, t := range taxas {
		var idMoeda int16
		if err := tx.QueryRow(ctx, "SELECT id FROM moedas WHERE codigo = $1", t.Moeda).Scan(&idMoeda); err != nil {
			return fmt.Errorf("moeda %s: %w", t.Moeda, err)
		}
		valorTaxa, _ := money.FromString(t.Taxa)
		if _, err := tx.Exec(ctx, `
			INSERT INTO taxas_base (moeda_id, taxa_mensal, vigencia_inicio, vigencia_fim)
			VALUES ($1, $2, $3, $4)`,
			idMoeda, valorTaxa, t.Inicio, t.Fim,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

type parCotacaoDemo struct {
	base, cot int16
	taxa      string
	inicio    time.Time
}

func (c *Carga) gerarCotacoes(ctx context.Context) error {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var brl, usd int16
	if err := tx.QueryRow(ctx, "SELECT id FROM moedas WHERE codigo = 'BRL'").Scan(&brl); err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, "SELECT id FROM moedas WHERE codigo = 'USD'").Scan(&usd); err != nil {
		return err
	}
	pares := []parCotacaoDemo{
		{usd, brl, "5.4321", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{usd, brl, "5.5000", time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)},
	}
	if err := inserirCotacoes(ctx, tx, pares); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func inserirCotacoes(ctx context.Context, tx pgx.Tx, pares []parCotacaoDemo) error {
	for _, p := range pares {
		valor, _ := money.FromString(p.taxa)
		if _, err := tx.Exec(ctx, `
			INSERT INTO cotacoes_cambio (moeda_base_id, moeda_cotacao_id, taxa, vigencia_inicio)
			VALUES ($1, $2, $3, $4)`,
			p.base, p.cot, valor, p.inicio,
		); err != nil {
			return err
		}
	}
	return nil
}

// ferramentasCarga agrupa o motor e os repositórios usados para gerar cada
// transação de demonstração (S107: agrupa parâmetros para caber em ≤3
// incluindo context.Context).
type ferramentasCarga struct {
	motor       *precificacao.Motor
	tipoRepo    *postgres.TipoRecebivelRepo
	cotacaoRepo *postgres.CotacaoRepo
	taxaRepo    *postgres.TaxaBaseRepo
	txRepo      *postgres.TransacaoRepo
}

type paramsTransacao struct {
	tipo      string
	cedenteID int64
	dataOp    time.Time
	brlID     int16
	usdID     int16
	idx       int
}

func novaFerramentasCarga(pool *pgxpool.Pool) ferramentasCarga {
	return ferramentasCarga{
		motor:       precificacao.NewMotor(recebivel.DefaultRegistry()),
		tipoRepo:    postgres.NewTipoRecebivelRepo(pool),
		cotacaoRepo: postgres.NewCotacaoRepo(pool),
		taxaRepo:    postgres.NewTaxaBaseRepo(pool),
		txRepo:      postgres.NewTransacaoRepo(pool),
	}
}

func (c *Carga) gerarTransacoes(ctx context.Context) error {
	ferramentas := novaFerramentasCarga(c.pool)
	moedaRepo := postgres.NewMoedaRepo(c.pool)
	moedaBRL, err := moedaRepo.PorCodigo(ctx, "BRL")
	if err != nil {
		return err
	}
	moedaUSD, err := moedaRepo.PorCodigo(ctx, "USD")
	if err != nil {
		return err
	}
	cedentesIDs, err := c.listarCedentesIDs(ctx)
	if err != nil {
		return err
	}
	return c.gerarParaTodosOsTipos(ctx, ferramentas, geracaoInput{CedentesIDs: cedentesIDs, BRLID: moedaBRL.ID, USDID: moedaUSD.ID})
}

// geracaoInput agrupa os dados fixos usados para gerar todas as
// transações de demonstração (S107: agrupa parâmetros de
// gerarParaTodosOsTipos).
type geracaoInput struct {
	CedentesIDs []int64
	BRLID       int16
	USDID       int16
}

func (c *Carga) gerarParaTodosOsTipos(ctx context.Context, f ferramentasCarga, in geracaoInput) error {
	tipos := []string{"DUPLICATA_MERCANTIL", "CHEQUE_PRE_DATADO"}
	dataBase := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	for _, tipo := range tipos {
		for i := 0; i < transacoesPorTipo; i++ {
			p := paramsTransacao{
				tipo: tipo, cedenteID: in.CedentesIDs[i%len(in.CedentesIDs)], dataOp: dataBase,
				brlID: in.BRLID, usdID: in.USDID, idx: i,
			}
			if err := c.criarTransacao(ctx, f, p); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Carga) listarCedentesIDs(ctx context.Context) ([]int64, error) {
	rows, err := c.pool.Query(ctx, "SELECT id FROM cedentes ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (c *Carga) criarTransacao(ctx context.Context, f ferramentasCarga, p paramsTransacao) error {
	entrada, tipoRecebivel, err := montarEntradaDemo(ctx, f, p)
	if err != nil {
		return err
	}
	resultado, err := f.motor.Precificar(entrada)
	if err != nil {
		return fmt.Errorf("precificar demo: %w", err)
	}
	uuid, err := novoUUID()
	if err != nil {
		return err
	}
	// Toda transação nasce PENDENTE — mesma regra do ciclo de vida real
	// (liquidacao-transacoes spec). Inserir já como LIQUIDADA e "liquidar"
	// de novo em seguida violaria o próprio guard de optimistic locking
	// (a segunda tentativa acha status != PENDENTE e retorna conflito).
	alvo := statusDemo(p.idx)
	t := montarTransacaoDemo(uuid, p, calculoDemo{Tipo: tipoRecebivel, Entrada: entrada, Resultado: resultado})
	if alvo == postgres.StatusCancelada {
		t.Status = postgres.StatusCancelada
	}
	if err := f.txRepo.Criar(ctx, t); err != nil {
		return fmt.Errorf("criar transação demo: %w", err)
	}
	if alvo == postgres.StatusLiquidada {
		if _, err := f.txRepo.Liquidar(ctx, postgres.LiquidacaoInput{
			ID: t.ID, Versao: 1, LiquidadaEm: time.Now().UTC(),
		}); err != nil {
			return fmt.Errorf("liquidar demo: %w", err)
		}
	}
	return nil
}

func montarEntradaDemo(ctx context.Context, f ferramentasCarga, p paramsTransacao) (precificacao.Entrada, postgres.TipoRecebivel, error) {
	venc := p.dataOp.AddDate(0, 0, 15+(p.idx%5)*15)
	valorFace, _ := money.FromString(fmt.Sprintf("%d.00", 5000+p.idx*1000))
	moedaPagamento := "BRL"
	if p.idx%5 == 0 {
		moedaPagamento = "USD"
	}
	tipoRecebivel, err := f.tipoRepo.PorCodigo(ctx, p.tipo)
	if err != nil {
		return precificacao.Entrada{}, postgres.TipoRecebivel{}, fmt.Errorf("tipo recebível: %w", err)
	}
	taxa, err := f.taxaRepo.Vigente(ctx, "BRL", p.dataOp)
	if err != nil {
		return precificacao.Entrada{}, postgres.TipoRecebivel{}, fmt.Errorf("taxa base: %w", err)
	}
	entrada := precificacao.Entrada{
		ValorFace: valorFace, DataOperacao: p.dataOp, DataVencimento: venc,
		TipoRecebivel: p.tipo, TaxaBase: taxa.TaxaMensal,
		MoedaTitulo: "BRL", MoedaPagamento: moedaPagamento,
	}
	if moedaPagamento != "BRL" {
		// ADR-006: a linha cadastrada tem base=moeda de pagamento, cotação=BRL.
		cot, err := f.cotacaoRepo.Vigente(ctx, postgres.ParCotacao{Base: moedaPagamento, Cotacao: "BRL", Em: p.dataOp})
		if err != nil {
			return precificacao.Entrada{}, postgres.TipoRecebivel{}, fmt.Errorf("cotação: %w", err)
		}
		entrada.Cotacao = cot.Taxa
		entrada.TemCotacao = true
	}
	return entrada, tipoRecebivel, nil
}

func statusDemo(idx int) postgres.TransacaoStatus {
	switch {
	case idx%3 == 0:
		return postgres.StatusLiquidada
	case idx%7 == 0:
		return postgres.StatusCancelada
	default:
		return postgres.StatusPendente
	}
}

// calculoDemo agrupa o tipo resolvido, a entrada e o resultado da
// precificação (S107: agrupa parâmetros de montarTransacaoDemo).
type calculoDemo struct {
	Tipo      postgres.TipoRecebivel
	Entrada   precificacao.Entrada
	Resultado precificacao.Resultado
}

func montarTransacaoDemo(id string, p paramsTransacao, c calculoDemo) postgres.Transacao {
	t := postgres.Transacao{
		ID:               id,
		CedenteID:        p.cedenteID,
		TipoRecebivelID:  c.Tipo.ID,
		MoedaTituloID:    p.brlID,
		MoedaPagamentoID: p.brlID,
		ValorFace:        c.Entrada.ValorFace,
		ValorPresente:    c.Resultado.ValorPresente8Casas,
		ValorLiquido:     c.Resultado.ValorLiquido,
		Desagio:          c.Resultado.Desagio,
		SpreadAplicado:   c.Resultado.SpreadAplicado,
		TaxaBaseAplicada: c.Resultado.TaxaBaseAplicada,
		DataOperacao:     c.Entrada.DataOperacao,
		DataVencimento:   c.Entrada.DataVencimento,
		// Toda transação nasce PENDENTE; o chamador decide se liquida (via
		// Liquidar, respeitando optimistic locking) ou cancela.
		Status: postgres.StatusPendente,
	}
	if c.Entrada.MoedaPagamento == "USD" {
		t.MoedaPagamentoID = p.usdID
	}
	if c.Resultado.CotacaoAplicada != nil {
		cot := *c.Resultado.CotacaoAplicada
		t.CotacaoAplicada = &cot
	}
	return t
}

func tempoFim(ano int, mes time.Month, dia int) *time.Time {
	t := time.Date(ano, mes, dia, 23, 59, 59, 0, time.UTC)
	return &t
}

func novoUUID() (string, error) {
	const hex = "0123456789abcdef"
	b := make([]byte, 32)
	for i := range b {
		b[i] = hex[rand.IntN(16)] //nolint:gosec // dado de demonstração, não é segredo nem token
	}
	id := fmt.Sprintf("%s-%s-%s-%s-%s", b[0:8], b[8:12], b[12:16], b[16:20], b[20:32])
	return id, nil
}

var _ = pgx.ErrNoRows
