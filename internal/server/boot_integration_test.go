//go:build integration

package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/srm-asset/srm-backend/internal/app/dto"
	"github.com/srm-asset/srm-backend/internal/app/middleware"
	"github.com/srm-asset/srm-backend/internal/platform/config"
	"github.com/srm-asset/srm-backend/internal/platform/testdb"
)

// TestAplicacaoOperaComBancoApenasMigrado prova que a aplicação sobe e
// opera corretamente contra um banco apenas migrado — testdb.Subir só
// aplica as migrations, sem nenhuma carga de demonstração (sem cedentes,
// sem taxas_base, sem transações; apenas moedas e tipos_recebivel, que a
// própria migration semeia). Até aqui isso só tinha sido verificado
// manualmente.
func TestAplicacaoOperaComBancoApenasMigrado(t *testing.T) {
	ctx := context.Background()
	pool, cleanup, err := testdb.Subir(ctx)
	if err != nil {
		t.Fatalf("subir banco de teste: %v", err)
	}
	defer cleanup()

	api := novaAPIComBancoMigrado(t, pool)

	verificarTiposRecebivelSemeados(t, api)
	verificarListaTransacoesVazia(t, api)
	verificarSimulacaoSemTaxaBase(t, api)
}

// novaAPIComBancoMigrado monta a API Huma real (mesmos repos, mesmo
// serviço de precificação e mesmo registro de operações usados em
// produção — nenhum mock) sobre o pool recém-migrado.
func novaAPIComBancoMigrado(t *testing.T, pool *pgxpool.Pool) humatest.TestAPI {
	t.Helper()
	middleware.InstalarErroHuma()
	_, api := humatest.New(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	d := Deps{
		Cfg:    config.Config{HTTPAddr: ":0", APIPrefix: "/api/v1"},
		Logger: logger,
		Pool:   pool,
	}
	svc, repos := montarServico(d)
	registrarOperacoes(api, svc, repos)
	return api
}

func verificarTiposRecebivelSemeados(t *testing.T, api humatest.TestAPI) {
	t.Helper()
	resp := api.Get("/api/v1/tipos-recebivel")
	if resp.Code != http.StatusOK {
		t.Fatalf("GET tipos-recebivel: esperado 200, obtido %d: %s", resp.Code, resp.Body.String())
	}
	var corpo struct {
		Items []dto.TipoRecebivelResponse `json:"items"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("decodificar resposta: %v", err)
	}
	codigos := map[string]bool{}
	for _, item := range corpo.Items {
		codigos[item.Codigo] = true
	}
	if len(corpo.Items) != 2 || !codigos["DUPLICATA_MERCANTIL"] || !codigos["CHEQUE_PRE_DATADO"] {
		t.Fatalf("esperados os 2 tipos semeados pela migration, obtido %+v", corpo.Items)
	}
}

func verificarListaTransacoesVazia(t *testing.T, api humatest.TestAPI) {
	t.Helper()
	resp := api.Get("/api/v1/transacoes")
	if resp.Code != http.StatusOK {
		t.Fatalf("GET transacoes: esperado 200, obtido %d: %s", resp.Code, resp.Body.String())
	}
	var corpo dto.Paginada
	if err := json.Unmarshal(resp.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("decodificar resposta: %v", err)
	}
	if corpo.Total != 0 {
		t.Fatalf("esperado total 0 (sem massa de demonstração), obtido %d", corpo.Total)
	}
}

func verificarSimulacaoSemTaxaBase(t *testing.T, api humatest.TestAPI) {
	t.Helper()
	resp := api.Post("/api/v1/simulacoes", map[string]any{
		"cedente_id": 1, "tipo_recebivel": "DUPLICATA_MERCANTIL",
		"valor_face": "10000.00", "moeda_titulo": "BRL", "moeda_pagamento": "BRL",
		"data_operacao": "2026-07-01", "data_vencimento": "2026-08-15",
	})
	if resp.Code != http.StatusUnprocessableEntity {
		t.Fatalf("POST simulacoes: esperado 422, obtido %d: %s", resp.Code, resp.Body.String())
	}
	var corpo dto.ErroCorpo
	if err := json.Unmarshal(resp.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("corpo não é ErroCorpo válido: %v", err)
	}
	if corpo.Codigo != "taxa_ausente" {
		t.Fatalf("esperado código taxa_ausente, obtido %q (corpo: %+v)", corpo.Codigo, corpo)
	}
}
