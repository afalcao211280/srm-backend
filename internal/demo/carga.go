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
	pool     *pgxpool.Pool
	logger   *slog.Logger
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
	defer tx.Rollback(ctx)
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
	Moeda   string
	Taxa    string
	Inicio  time.Time
	Fim     *time.Time
}

func (c *Carga) gerarTaxas(ctx context.Context) error {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
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

func (c *Carga) gerarCotacoes(ctx context.Context) error {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var brl, usd int16
	if err := tx.QueryRow(ctx, "SELECT id FROM moedas WHERE codigo = 'BRL'").Scan(&brl); err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, "SELECT id FROM moedas WHERE codigo = 'USD'").Scan(&usd); err != nil {
		return err
	}
	pares := []struct {
		base, cot int16
		taxa      string
		inicio    time.Time
	}{
		{usd, brl, "5.4321", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{usd, brl, "5.5000", time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)},
	}
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
	return tx.Commit(ctx)
}

func (c *Carga) gerarTransacoes(ctx context.Context) error {
	registry := recebivel.DefaultRegistry()
	motor := precificacao.NewMotor(registry)
	moedaRepo := postgres.NewMoedaRepo(c.pool)
	cotacaoRepo := postgres.NewCotacaoRepo(c.pool)
	taxaRepo := postgres.NewTaxaBaseRepo(c.pool)
	txRepo := postgres.NewTransacaoRepo(c.pool)

	moedaBRL, err := moedaRepo.PorCodigo(context.Background(), "BRL")
	if err != nil {
		return err
	}
	moedaUSD, err := moedaRepo.PorCodigo(context.Background(), "USD")
	if err != nil {
		return err
	}
	var cedentesIDs []int64
	rows, err := c.pool.Query(context.Background(), "SELECT id FROM cedentes ORDER BY id")
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		cedentesIDs = append(cedentesIDs, id)
	}
	rows.Close()

	tipos := []string{"DUPLICATA_MERCANTIL", "CHEQUE_PRE_DATADO"}
	dataBase := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	for _, tipo := range tipos {
		for i := 0; i < transacoesPorTipo; i++ {
			if err := c.criarTransacao(ctx, motor, moedaRepo, cotacaoRepo, taxaRepo, txRepo, tipo, cedentesIDs[i%len(cedentesIDs)], dataBase, moedaBRL.ID, moedaUSD.ID, i); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Carga) criarTransacao(
	ctx context.Context,
	motor *precificacao.Motor,
	moedaRepo *postgres.MoedaRepo,
	cotacaoRepo *postgres.CotacaoRepo,
	taxaRepo *postgres.TaxaBaseRepo,
	txRepo *postgres.TransacaoRepo,
	tipo string,
	cedenteID int64,
	dataOp time.Time,
	brlID, usdID int16,
	idx int,
) error {
	dias := 15 + (idx % 5) * 15
	venc := dataOp.AddDate(0, 0, dias)
	valorFace, _ := money.FromString(fmt.Sprintf("%d.00", 5000+idx*1000))
	moedaTitulo := "BRL"
	moedaPagamento := "BRL"
	if idx%5 == 0 {
		moedaPagamento = "USD"
	}
	taxa, err := taxaRepo.Vigente(ctx, moedaTitulo, dataOp)
	if err != nil {
		return fmt.Errorf("taxa base: %w", err)
	}
	entrada := precificacao.Entrada{
		ValorFace:      valorFace,
		DataOperacao:   dataOp,
		DataVencimento: venc,
		TipoRecebivel:  tipo,
		TaxaBase:       taxa.TaxaMensal,
		MoedaTitulo:    moedaTitulo,
		MoedaPagamento: moedaPagamento,
	}
	if moedaTitulo != moedaPagamento {
		cot, err := cotacaoRepo.Vigente(ctx, moedaTitulo, moedaPagamento, dataOp)
		if err != nil {
			return fmt.Errorf("cotação: %w", err)
		}
		entrada.Cotacao = cot.Taxa
		entrada.TemCotacao = true
	}
	resultado, err := motor.Precificar(entrada)
	if err != nil {
		return fmt.Errorf("precificar demo: %w", err)
	}
	uuid, err := novoUUID()
	if err != nil {
		return err
	}
	status := postgres.StatusPendente
	if idx%3 == 0 {
		status = postgres.StatusLiquidada
	} else if idx%7 == 0 {
		status = postgres.StatusCancelada
	}
	t := postgres.Transacao{
		ID:               uuid,
		CedenteID:        cedenteID,
		TipoRecebivelID:  brlID,
		MoedaTituloID:    brlID,
		MoedaPagamentoID: brlID,
		ValorFace:        valorFace,
		ValorPresente:    resultado.ValorPresente8Casas,
		ValorLiquido:     resultado.ValorLiquido,
		Desagio:          resultado.Desagio,
		SpreadAplicado:   resultado.SpreadAplicado,
		TaxaBaseAplicada: resultado.TaxaBaseAplicada,
		DataOperacao:     dataOp,
		DataVencimento:   venc,
		Status:           status,
	}
	if moedaPagamento == "USD" {
		t.MoedaPagamentoID = usdID
	}
	if tipo == "CHEQUE_PRE_DATADO" {
		t.TipoRecebivelID = 2
	}
	if resultado.CotacaoAplicada != nil {
		c := *resultado.CotacaoAplicada
		t.CotacaoAplicada = &c
	}
	if err := txRepo.Criar(ctx, t); err != nil {
		return fmt.Errorf("criar transação demo: %w", err)
	}
	if status == postgres.StatusLiquidada {
		if _, err := txRepo.Liquidar(ctx, t.ID, 1, time.Now().UTC()); err != nil {
			return fmt.Errorf("liquidar demo: %w", err)
		}
	}
	return nil
}

func tempoFim(ano int, mes time.Month, dia int) *time.Time {
	t := time.Date(ano, mes, dia, 23, 59, 59, 0, time.UTC)
	return &t
}

func novoUUID() (string, error) {
	const hex = "0123456789abcdef"
	b := make([]byte, 32)
	for i := range b {
		b[i] = hex[rand.IntN(16)]
	}
	id := fmt.Sprintf("%s-%s-%s-%s-%s", b[0:8], b[8:12], b[12:16], b[16:20], b[20:32])
	return id, nil
}

var _ = pgx.ErrNoRows
