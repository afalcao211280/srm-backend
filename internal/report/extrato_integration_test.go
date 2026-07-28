//go:build integration

// Este pacote é, deliberadamente, 2 camadas (ADR-007): não importa
// internal/domain, nem mesmo em teste — o lint (depguard) aplica a mesma
// regra a arquivos de teste, e a fixture abaixo por isso insere via SQL
// puro em vez de reusar o tipo internal/domain/money ou o repositório de
// domínio internal/infra/postgres.Transacao.
package report

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	infrapostgres "github.com/srm-asset/srm-backend/internal/infra/postgres"
	"github.com/srm-asset/srm-backend/internal/platform/testdb"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()
	pool, cleanup, err := testdb.Subir(ctx)
	if err != nil {
		os.Stderr.WriteString("subir banco de teste: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer cleanup()
	testPool = pool
	os.Exit(m.Run())
}

type fixtureExtrato struct {
	CedenteID int64
	TipoID    int16
	MoedaID   int16
}

func criarFixtureExtrato(t *testing.T, ctx context.Context, pool *pgxpool.Pool, nomeCedente string) fixtureExtrato {
	t.Helper()
	moeda, err := infrapostgres.NewMoedaRepo(pool).PorCodigo(ctx, "BRL")
	if err != nil {
		t.Fatalf("resolver moeda: %v", err)
	}
	tipo, err := infrapostgres.NewTipoRecebivelRepo(pool).PorCodigo(ctx, "DUPLICATA_MERCANTIL")
	if err != nil {
		t.Fatalf("resolver tipo: %v", err)
	}
	cedenteID, err := infrapostgres.NewCedenteRepo(pool).Criar(ctx, infrapostgres.Cedente{
		Nome: nomeCedente, Documento: "00000000000",
	})
	if err != nil {
		t.Fatalf("criar cedente: %v", err)
	}
	return fixtureExtrato{CedenteID: cedenteID, TipoID: tipo.ID, MoedaID: moeda.ID}
}

func criarTransacaoExtrato(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id string, f fixtureExtrato, dataOperacao time.Time) {
	t.Helper()
	_, err := pool.Exec(ctx, `
		INSERT INTO transacoes (
			id, cedente_id, tipo_recebivel_id, moeda_titulo_id, moeda_pagamento_id,
			valor_face, valor_presente, valor_liquido, desagio,
			spread_aplicado, taxa_base_aplicada,
			data_operacao, data_vencimento, status, versao
		) VALUES (
			$1, $2, $3, $4, $5,
			1000.00, 950.00, 950.00, 50.00,
			0.015, 0.01,
			$6, $7, 'PENDENTE', 1
		)`,
		id, f.CedenteID, f.TipoID, f.MoedaID, f.MoedaID,
		dataOperacao, dataOperacao.AddDate(0, 1, 15),
	)
	if err != nil {
		t.Fatalf("criar transação %s: %v", id, err)
	}
}

func TestExtratoSQLInjectionNoFiltroCedente(t *testing.T) {
	ctx := context.Background()
	testdb.Truncate(t, testPool)
	f := criarFixtureExtrato(t, ctx, testPool, "Cedente Injection")
	criarTransacaoExtrato(t, ctx, testPool, "77777777-7777-7777-7777-777777777771", f, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	extrato := NovoExtrato(testPool, logger)

	// O filtro de moeda é o único campo textual livre no parseFiltros; o
	// valor abaixo, se concatenado ao SQL, derrubaria a tabela. Como
	// passa por squirrel com placeholder, deve ser tratado como dado
	// literal — a consulta roda normal e não bate com nenhuma linha.
	res, err := extrato.Consultar(ctx, Filtros{Moeda: "BRL'; DROP TABLE transacoes; --", Pagina: 1, Tamanho: 20})
	if err != nil {
		t.Fatalf("consultar: %v", err)
	}
	if res.Total != 0 {
		t.Fatalf("esperado 0 resultados para filtro malicioso, obtido %d", res.Total)
	}

	// Prova que a tabela sobreviveu: a mesma transação ainda é encontrável
	// com o filtro correto.
	res2, err := extrato.Consultar(ctx, Filtros{Moeda: "BRL", Pagina: 1, Tamanho: 20})
	if err != nil {
		t.Fatalf("consultar após tentativa de injection: %v", err)
	}
	if res2.Total != 1 {
		t.Fatalf("tabela transacoes não sobreviveu à tentativa de injection: total=%d", res2.Total)
	}
	if !strings.Contains(res2.Items[0].CedenteNome, "Injection") {
		t.Fatalf("linha inesperada: %+v", res2.Items[0])
	}
}

func TestExtratoPaginacao(t *testing.T) {
	ctx := context.Background()
	testdb.Truncate(t, testPool)
	f := criarFixtureExtrato(t, ctx, testPool, "Cedente Paginação")
	for i := 0; i < 25; i++ {
		id := uuidDeterministico(i)
		criarTransacaoExtrato(t, ctx, testPool, id, f, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i))
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	extrato := NovoExtrato(testPool, logger)

	pagina1, err := extrato.Consultar(ctx, Filtros{Pagina: 1, Tamanho: 20})
	if err != nil {
		t.Fatalf("página 1: %v", err)
	}
	if pagina1.Total != 25 || len(pagina1.Items) != 20 || pagina1.TotalPaginas != 2 {
		t.Fatalf("página 1: total=%d itens=%d totalPaginas=%d", pagina1.Total, len(pagina1.Items), pagina1.TotalPaginas)
	}
	pagina2, err := extrato.Consultar(ctx, Filtros{Pagina: 2, Tamanho: 20})
	if err != nil {
		t.Fatalf("página 2: %v", err)
	}
	if len(pagina2.Items) != 5 {
		t.Fatalf("página 2: esperado 5 itens, obtido %d", len(pagina2.Items))
	}
	// Nenhuma linha duplicada entre as páginas (ordenação determinística).
	vistos := map[string]bool{}
	for _, l := range append(pagina1.Items, pagina2.Items...) {
		if vistos[l.ID] {
			t.Fatalf("id duplicado entre páginas: %s", l.ID)
		}
		vistos[l.ID] = true
	}
}

func TestExtratoFiltroPeriodoInvalido(t *testing.T) {
	testdb.Truncate(t, testPool)

	inicial := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	final := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	_, err := parseFiltros(map[string][]string{
		"data_inicial": {inicial.Format("2006-01-02")},
		"data_final":   {final.Format("2006-01-02")},
	})
	if err == nil {
		t.Fatal("esperado erro para período invertido")
	}
}

func TestExtratoPlanoDeExecucaoUsaIndice(t *testing.T) {
	ctx := context.Background()
	testdb.Truncate(t, testPool)
	f := criarFixtureExtrato(t, ctx, testPool, "Cedente Índice")
	for i := 0; i < 50; i++ {
		criarTransacaoExtrato(t, ctx, testPool, uuidDeterministico(100+i), f,
			time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i))
	}
	var plano string
	err := testPool.QueryRow(ctx, `
		EXPLAIN (FORMAT TEXT)
		SELECT t.id FROM transacoes t
		 WHERE t.data_operacao >= $1 AND t.data_operacao <= $2 AND t.cedente_id = $3
		 ORDER BY t.data_operacao DESC, t.id`,
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), f.CedenteID,
	).Scan(&plano)
	if err != nil {
		t.Fatalf("explain: %v", err)
	}
	if strings.Contains(plano, "Seq Scan on transacoes") {
		t.Fatalf("consulta filtrada por data_operacao+cedente_id fez varredura sequencial:\n%s", plano)
	}
}

func uuidDeterministico(n int) string {
	return fmt.Sprintf("88888888-8888-8888-8888-%012d", n)
}
