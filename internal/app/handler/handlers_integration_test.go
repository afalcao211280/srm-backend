//go:build integration

// Pacote externo (handler_test), não handler: este arquivo testa via
// internal/server.Build, que importa internal/app/handler — um teste
// interno (package handler) criaria ciclo de import.
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/srm-asset/srm-backend/internal/app/dto"
	"github.com/srm-asset/srm-backend/internal/domain/money"
	"github.com/srm-asset/srm-backend/internal/infra/postgres"
	"github.com/srm-asset/srm-backend/internal/platform/config"
	"github.com/srm-asset/srm-backend/internal/platform/testdb"
	"github.com/srm-asset/srm-backend/internal/server"
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

func router(t *testing.T) http.Handler {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv, err := server.Build(context.Background(), server.Deps{
		Cfg:    config.Config{HTTPAddr: ":0", APIPrefix: "/api/v1"},
		Logger: logger,
		Pool:   testPool,
	})
	if err != nil {
		t.Fatalf("montar servidor: %v", err)
	}
	return srv.Handler
}

// fixtureHandler cria taxa base vigente e um cedente reais, necessários
// para qualquer precificação — sem eles a simulação retorna 422
// "taxa_ausente", conforme a spec de gestao-de-taxas.
func fixtureHandler(t *testing.T) (cedenteID int64) {
	t.Helper()
	ctx := context.Background()
	testdb.Truncate(t, testPool)
	taxa, _ := money.FromString("0.01")
	if _, err := postgres.NewTaxaBaseRepo(testPool).Criar(ctx, postgres.NovaTaxaBase{
		Moeda: "BRL", Taxa: taxa, Inicio: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("seed taxa base: %v", err)
	}
	id, err := postgres.NewCedenteRepo(testPool).Criar(ctx, postgres.Cedente{Nome: "Cedente Teste", Documento: "00000000000"})
	if err != nil {
		t.Fatalf("seed cedente: %v", err)
	}
	return id
}

// seedCotacaoUSDBRL semeia a cotação exatamente como o ADR-006 especifica:
// base = moeda de pagamento (USD), cotação = moeda do título (BRL).
func seedCotacaoUSDBRL(t *testing.T, taxa string) {
	t.Helper()
	tx, _ := money.FromString(taxa)
	if _, err := postgres.NewCotacaoRepo(testPool).Criar(context.Background(), postgres.NovaCotacao{
		Base: "USD", Cotacao: "BRL", Taxa: tx, Inicio: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("seed cotação USD/BRL: %v", err)
	}
}

func TestSimularSucessoRoundTrip(t *testing.T) {
	cedenteID := fixtureHandler(t)
	r := router(t)

	payload := map[string]any{
		"cedente_id":      cedenteID,
		"tipo_recebivel":  "DUPLICATA_MERCANTIL",
		"valor_face":      "10000.00",
		"moeda_titulo":    "BRL",
		"moeda_pagamento": "BRL",
		"data_operacao":   "2026-07-01",
		"data_vencimento": "2026-08-15", // 45 dias
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulacoes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtido %d: %s", w.Code, w.Body.String())
	}
	var resp dto.SimulacaoResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decodificar resposta: %v", err)
	}
	// Número conferível: motor-precificacao spec, cenário "Duplicata
	// mercantil com prazo fracionário" — VF=10000, taxa base 0.01,
	// spread 0.015, 45 dias → valor líquido 9636.39.
	if resp.ValorLiquido != "9636.39" {
		t.Fatalf("valor líquido: esperado 9636.39, obtido %s", resp.ValorLiquido)
	}
}

// TestSimularCrossCurrencyRoundTrip exercita a consulta REAL de cotação
// (CotacaoRepo.Vigente), não apenas o cálculo do motor com uma cotação
// passada manualmente — é o caminho que a carga de demonstração via
// `docker compose run carga` provou estar quebrado: a consulta buscava
// (base=título, cotação=pagamento) mas ADR-006 cadastra
// (base=pagamento, cotação=título).
func TestSimularCrossCurrencyRoundTrip(t *testing.T) {
	cedenteID := fixtureHandler(t)
	seedCotacaoUSDBRL(t, "5.4321")
	r := router(t)

	payload := map[string]any{
		"cedente_id":      cedenteID,
		"tipo_recebivel":  "DUPLICATA_MERCANTIL",
		"valor_face":      "10000.00",
		"moeda_titulo":    "BRL",
		"moeda_pagamento": "USD",
		"data_operacao":   "2026-07-01",
		"data_vencimento": "2026-08-15", // 45 dias
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulacoes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtido %d: %s", w.Code, w.Body.String())
	}
	var resp dto.SimulacaoResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decodificar resposta: %v", err)
	}
	// Número conferível: motor-precificacao spec, cenário cross-currency —
	// mesma operação da anterior, convertida a 5.4321 BRL/USD → 1773.97 USD.
	if resp.ValorLiquido != "1773.97" {
		t.Fatalf("valor líquido cross-currency: esperado 1773.97, obtido %s", resp.ValorLiquido)
	}
	if resp.MoedaPagamento != "USD" {
		t.Fatalf("moeda de pagamento: esperado USD, obtido %s", resp.MoedaPagamento)
	}
}

func TestCriarEDetalharTransacao(t *testing.T) {
	cedenteID := fixtureHandler(t)
	r := router(t)

	payload := map[string]any{
		"cedente_id":      cedenteID,
		"tipo_recebivel":  "DUPLICATA_MERCANTIL",
		"valor_face":      "10000.00",
		"moeda_titulo":    "BRL",
		"moeda_pagamento": "BRL",
		"data_operacao":   "2026-07-01",
		"data_vencimento": "2026-08-15",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transacoes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("esperado 201, obtido %d: %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Location") == "" {
		t.Fatal("esperado cabeçalho Location")
	}
	var criada dto.TransacaoResponse
	if err := json.Unmarshal(w.Body.Bytes(), &criada); err != nil {
		t.Fatalf("decodificar criação: %v", err)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/transacoes/"+criada.ID, nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("GET detalhe: esperado 200, obtido %d: %s", w2.Code, w2.Body.String())
	}
}

func TestLiquidarSucessoDepoisConflito(t *testing.T) {
	cedenteID := fixtureHandler(t)
	r := router(t)

	payload := map[string]any{
		"cedente_id": cedenteID, "tipo_recebivel": "DUPLICATA_MERCANTIL",
		"valor_face": "10000.00", "moeda_titulo": "BRL", "moeda_pagamento": "BRL",
		"data_operacao": "2026-07-01", "data_vencimento": "2026-08-15",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transacoes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var criada dto.TransacaoResponse
	_ = json.Unmarshal(w.Body.Bytes(), &criada)

	liquidar := func() int {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/transacoes/"+criada.ID+"/liquidacao", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}
	if got := liquidar(); got != http.StatusOK {
		t.Fatalf("primeira liquidação: esperado 200, obtido %d", got)
	}
	if got := liquidar(); got != http.StatusConflict {
		t.Fatalf("segunda liquidação: esperado 409, obtido %d", got)
	}
}

func TestGetTransacaoInexistente404(t *testing.T) {
	fixtureHandler(t)
	r := router(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/transacoes/00000000-0000-0000-0000-000000000000", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("esperado 404, obtido %d: %s", w.Code, w.Body.String())
	}
}

// TestPanicRecuperado prova o exception handler global: um panic dentro de
// um handler não derruba o processo, responde 500 sem vazar detalhe e o
// servidor continua atendendo as requisições seguintes. Registra uma rota
// que propositalmente entra em pânico na MESMA engine (mesma cadeia de
// middlewares) já montada por server.Build, sem alterar código de produção.
func TestPanicRecuperado(t *testing.T) {
	fixtureHandler(t)
	r := router(t)
	engine, ok := r.(*gin.Engine)
	if !ok {
		t.Fatalf("handler não é *gin.Engine: %T", r)
	}
	engine.GET("/api/v1/__panico_teste", func(c *gin.Context) {
		panic("falha proposital para provar recovery")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/__panico_teste", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("esperado 500, obtido %d", w.Code)
	}
	var corpo dto.ErroCorpo
	if err := json.Unmarshal(w.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("corpo do erro não é ErroCorpo válido: %v", err)
	}
	if corpo.Mensagem == "" || corpo.CorrelationID == "" {
		t.Fatalf("corpo de erro incompleto: %+v", corpo)
	}

	// O processo segue vivo: uma requisição normal em seguida funciona.
	req2 := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("servidor não respondeu normalmente após o panic: %d", w2.Code)
	}
}
