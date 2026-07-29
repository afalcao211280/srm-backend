//go:build integration

package postgres

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/srm-asset/srm-backend/internal/domain/money"
)

// fixture resolve os ids de referência (moeda, tipo de recebível) semeados
// pelas migrations e cria um cedente real — nunca hardcoda id, porque a
// ordem de inserção não é um contrato e mudar a migration não pode quebrar
// o teste silenciosamente.
type fixture struct {
	CedenteID int64
	TipoID    int16
	MoedaID   int16
}

func criarFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool) fixture {
	t.Helper()
	moeda, err := NewMoedaRepo(pool).PorCodigo(ctx, "BRL")
	if err != nil {
		t.Fatalf("resolver moeda BRL: %v", err)
	}
	tipo, err := NewTipoRecebivelRepo(pool).PorCodigo(ctx, "DUPLICATA_MERCANTIL")
	if err != nil {
		t.Fatalf("resolver tipo DUPLICATA_MERCANTIL: %v", err)
	}
	cedenteID, err := NewCedenteRepo(pool).Criar(ctx, Cedente{
		Nome:      "Cedente de Teste",
		Documento: "00000000000",
	})
	if err != nil {
		t.Fatalf("criar cedente: %v", err)
	}
	return fixture{CedenteID: cedenteID, TipoID: tipo.ID, MoedaID: moeda.ID}
}

func transacaoFixture(id string, f fixture) Transacao {
	vf, _ := money.FromString("1000.00")
	vp, _ := money.FromString("950.00")
	vl, _ := money.FromString("950.00")
	des, _ := money.FromString("50.00")
	sp, _ := money.FromString("0.015")
	tb, _ := money.FromString("0.01")
	return Transacao{
		ID:               id,
		CedenteID:        f.CedenteID,
		TipoRecebivelID:  f.TipoID,
		MoedaTituloID:    f.MoedaID,
		MoedaPagamentoID: f.MoedaID,
		ValorFace:        vf,
		ValorPresente:    vp,
		ValorLiquido:     vl,
		Desagio:          des,
		SpreadAplicado:   sp,
		TaxaBaseAplicada: tb,
		DataOperacao:     time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		DataVencimento:   time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
		Status:           StatusPendente,
	}
}

func TestLiquidacaoConcorrencia(t *testing.T) {
	pool := setupTestDB(t)
	truncate(t, pool)
	ctx := context.Background()
	f := criarFixture(t, ctx, pool)

	repo := NewTransacaoRepo(pool)
	id := "11111111-1111-1111-1111-111111111111"
	if err := repo.Criar(ctx, transacaoFixture(id, f)); err != nil {
		t.Fatalf("criar: %v", err)
	}
	const N = 10
	var sucessos, conflitos int32
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			_, err := repo.Liquidar(ctx, LiquidacaoInput{ID: id, Versao: 1, LiquidadaEm: time.Now().UTC()})
			switch {
			case err == nil:
				atomic.AddInt32(&sucessos, 1)
			case err == ErroConflitoLiquidacao:
				atomic.AddInt32(&conflitos, 1)
			}
		}()
	}
	wg.Wait()
	if sucessos != 1 {
		t.Fatalf("esperado 1 sucesso, obtido %d", sucessos)
	}
	if conflitos != N-1 {
		t.Fatalf("esperado %d conflitos, obtido %d", N-1, conflitos)
	}
	atual, err := repo.PorID(ctx, id)
	if err != nil {
		t.Fatalf("por id: %v", err)
	}
	if atual.Versao != 2 {
		t.Fatalf("versão: esperado 2, obtido %d", atual.Versao)
	}
	var totalAuditoria int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM transacao_auditoria WHERE transacao_id = $1", id,
	).Scan(&totalAuditoria); err != nil {
		t.Fatalf("contar auditoria: %v", err)
	}
	if totalAuditoria != 1 {
		t.Fatalf("esperado exatamente 1 registro de auditoria (um por liquidação bem-sucedida), obtido %d", totalAuditoria)
	}
}

func TestLiquidacaoVersaoDesatualizada(t *testing.T) {
	pool := setupTestDB(t)
	truncate(t, pool)
	ctx := context.Background()
	f := criarFixture(t, ctx, pool)

	repo := NewTransacaoRepo(pool)
	id := "22222222-2222-2222-2222-222222222222"
	if err := repo.Criar(ctx, transacaoFixture(id, f)); err != nil {
		t.Fatalf("criar: %v", err)
	}
	if _, err := repo.Liquidar(ctx, LiquidacaoInput{ID: id, Versao: 1, LiquidadaEm: time.Now().UTC()}); err != nil {
		t.Fatalf("primeira liquidação: %v", err)
	}
	_, err := repo.Liquidar(ctx, LiquidacaoInput{ID: id, Versao: 1, LiquidadaEm: time.Now().UTC()})
	if err != ErroConflitoLiquidacao {
		t.Fatalf("segunda liquidação com versão antiga: esperado conflito, obtido %v", err)
	}
}

// TestLiquidacaoConflitoNaoDeixaAuditoria prova a metade "nenhuma alteração"
// do requisito de atomicidade: quando o UPDATE condicional não afeta
// nenhuma linha (versão ou status não batem), a liquidação retorna erro
// ANTES de qualquer INSERT de auditoria — não existe estado "meio
// liquidado". A liquidação com versão/status corretos já é coberta por
// TestLiquidacaoConcorrencia (exatamente 1 linha de auditoria criada).
func TestLiquidacaoConflitoNaoDeixaAuditoria(t *testing.T) {
	pool := setupTestDB(t)
	truncate(t, pool)
	ctx := context.Background()
	f := criarFixture(t, ctx, pool)

	repo := NewTransacaoRepo(pool)
	id := "44444444-4444-4444-4444-444444444444"
	if err := repo.Criar(ctx, transacaoFixture(id, f)); err != nil {
		t.Fatalf("criar: %v", err)
	}
	// versão errada de propósito: a transação está na versão 1, não 2.
	if _, err := repo.Liquidar(ctx, LiquidacaoInput{ID: id, Versao: 2, LiquidadaEm: time.Now().UTC()}); err != ErroConflitoLiquidacao {
		t.Fatalf("esperado conflito de versão, obtido %v", err)
	}
	atual, err := repo.PorID(ctx, id)
	if err != nil {
		t.Fatalf("por id: %v", err)
	}
	if atual.Status != StatusPendente || atual.Versao != 1 {
		t.Fatalf("estado deveria permanecer inalterado, obtido status=%s versao=%d", atual.Status, atual.Versao)
	}
	var totalAuditoria int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM transacao_auditoria WHERE transacao_id = $1", id,
	).Scan(&totalAuditoria); err != nil {
		t.Fatalf("contar auditoria: %v", err)
	}
	if totalAuditoria != 0 {
		t.Fatalf("conflito não deveria deixar rastro de auditoria, obtido %d registros", totalAuditoria)
	}
}

// TestLiquidacaoFalhaNaAuditoriaRevisaStatus prova a metade "tudo ou nada"
// da atomicidade: quando o UPDATE de status já rodou dentro da transação
// mas o INSERT de auditoria falha em seguida, o rollback desfaz o UPDATE
// também — não existe estado "liquidado sem auditoria" pela metade. A falha
// no INSERT é forçada de forma determinística por uma CHECK constraint
// NOT VALID, sem depender de timing ou de dados inválidos que o INSERT já
// rejeitaria antes mesmo do rollback importar.
func TestLiquidacaoFalhaNaAuditoriaRevisaStatus(t *testing.T) {
	pool := setupTestDB(t)
	truncate(t, pool)
	ctx := context.Background()
	f := criarFixture(t, ctx, pool)

	repo := NewTransacaoRepo(pool)
	id := "77777777-7777-7777-7777-777777777777"
	if err := repo.Criar(ctx, transacaoFixture(id, f)); err != nil {
		t.Fatalf("criar: %v", err)
	}
	forcarFalhaAuditoria(t, ctx, pool)

	_, err := repo.Liquidar(ctx, LiquidacaoInput{ID: id, Versao: 1, LiquidadaEm: time.Now().UTC()})
	if err == nil || errors.Is(err, ErroConflitoLiquidacao) {
		t.Fatalf("esperado erro de constraint do postgres na auditoria, obtido %v", err)
	}

	atual, err := repo.PorID(ctx, id)
	if err != nil {
		t.Fatalf("por id: %v", err)
	}
	if atual.Status != StatusPendente || atual.Versao != 1 {
		t.Fatalf("rollback deveria reverter o UPDATE de status, obtido status=%s versao=%d", atual.Status, atual.Versao)
	}
	var totalAuditoria int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM transacao_auditoria WHERE transacao_id = $1", id,
	).Scan(&totalAuditoria); err != nil {
		t.Fatalf("contar auditoria: %v", err)
	}
	if totalAuditoria != 0 {
		t.Fatalf("insert de auditoria falhou mas deixou rastro, obtido %d registros", totalAuditoria)
	}
}

// forcarFalhaAuditoria adiciona uma CHECK constraint que faz todo INSERT
// futuro em transacao_auditoria falhar, e agenda sua remoção: os testes
// deste arquivo compartilham um único container Postgres via TestMain, e a
// constraint precisa sumir antes dos testes seguintes rodarem.
func forcarFalhaAuditoria(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	if _, err := pool.Exec(ctx,
		"ALTER TABLE transacao_auditoria ADD CONSTRAINT chk_forca_falha_teste CHECK (false) NOT VALID",
	); err != nil {
		t.Fatalf("adicionar constraint de falha forçada: %v", err)
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(),
			"ALTER TABLE transacao_auditoria DROP CONSTRAINT chk_forca_falha_teste",
		); err != nil {
			t.Fatalf("remover constraint de falha forçada: %v", err)
		}
	})
}

func TestRoundTripDecimal(t *testing.T) {
	pool := setupTestDB(t)
	truncate(t, pool)
	ctx := context.Background()
	f := criarFixture(t, ctx, pool)

	repo := NewTransacaoRepo(pool)
	id := "33333333-3333-3333-3333-333333333333"
	tx := transacaoFixture(id, f)
	vpOriginal, _ := money.FromString("9636.38630878")
	tx.ValorFace, _ = money.FromString("10000.00")
	tx.ValorPresente = vpOriginal
	tx.ValorLiquido, _ = money.FromString("9636.39")
	tx.Desagio, _ = money.FromString("363.61")
	if err := repo.Criar(ctx, tx); err != nil {
		t.Fatalf("criar: %v", err)
	}
	lida, err := repo.PorID(ctx, id)
	if err != nil {
		t.Fatalf("por id: %v", err)
	}
	if lida.ValorPresente.Cmp(vpOriginal) != 0 {
		t.Fatalf("round-trip: esperado %s, obtido %s", vpOriginal, lida.ValorPresente)
	}
}

func TestTransacaoListarPaginado(t *testing.T) {
	pool := setupTestDB(t)
	truncate(t, pool)
	ctx := context.Background()
	f := criarFixture(t, ctx, pool)
	repo := NewTransacaoRepo(pool)

	ids := []string{
		"55555555-5555-5555-5555-555555555551",
		"55555555-5555-5555-5555-555555555552",
		"55555555-5555-5555-5555-555555555553",
	}
	for _, id := range ids {
		if err := repo.Criar(ctx, transacaoFixture(id, f)); err != nil {
			t.Fatalf("criar %s: %v", id, err)
		}
	}
	pagina1, total, err := repo.Listar(ctx, 1, 2)
	if err != nil {
		t.Fatalf("listar página 1: %v", err)
	}
	if total != 3 {
		t.Fatalf("esperado total 3, obtido %d", total)
	}
	if len(pagina1) != 2 {
		t.Fatalf("esperado 2 itens na página 1, obtido %d", len(pagina1))
	}
	pagina2, _, err := repo.Listar(ctx, 2, 2)
	if err != nil {
		t.Fatalf("listar página 2: %v", err)
	}
	if len(pagina2) != 1 {
		t.Fatalf("esperado 1 item na página 2, obtido %d", len(pagina2))
	}
}
