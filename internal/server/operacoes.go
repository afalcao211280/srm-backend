package server

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/srm-asset/srm-backend/internal/app/dto"
	"github.com/srm-asset/srm-backend/internal/app/handler"
	"github.com/srm-asset/srm-backend/internal/app/middleware"
	"github.com/srm-asset/srm-backend/internal/domain/money"
	"github.com/srm-asset/srm-backend/internal/domain/precificacao"
	"github.com/srm-asset/srm-backend/internal/infra/postgres"
)

// Repos agrupa os repositórios usados pelas operações de negócio (S107:
// agrupa parâmetros para caber em ≤3 por função incluindo context.Context).
type Repos struct {
	Tx      *postgres.TransacaoRepo
	Cotacao *postgres.CotacaoRepo
	Taxa    *postgres.TaxaBaseRepo
	Tipo    *postgres.TipoRecebivelRepo
}

// registrarOperacoes declara as 9 operações de negócio da API na instância
// Huma — cada uma reusa exatamente a mesma camada de aplicação
// (ServicoPrecificacao, repositórios) que os testes de integração já
// exercitam; o único código novo aqui é o bind/validação de request e o
// mapeamento de resposta, que o Huma fazia de forma manual via gin antes.
func registrarOperacoes(api huma.API, svc *handler.ServicoPrecificacao, repos Repos) {
	registrarSimulacoes(api, svc)
	registrarCriarTransacao(api, svc)
	registrarListarTransacoes(api, repos.Tx)
	registrarObterTransacao(api, repos.Tx)
	registrarLiquidarTransacao(api, repos.Tx)
	registrarCotacoes(api, repos.Cotacao)
	registrarCriarTaxaBase(api, repos.Taxa)
	registrarTaxaBaseVigente(api, repos.Taxa)
	registrarTiposRecebivel(api, repos.Tipo)
}

type bodyIn[T any] struct{ Body T }
type bodyOut[T any] struct{ Body T }

func registrarSimulacoes(api huma.API, svc *handler.ServicoPrecificacao) {
	huma.Register(api, huma.Operation{
		OperationID: "criar-simulacao",
		Method:      http.MethodPost,
		Path:        "/api/v1/simulacoes",
		Summary:     "Simula o valor líquido de um recebível",
		Description: "Calcula valor presente, deságio e valor líquido sem persistir nada — usado pelo painel do operador para simulação em tempo real.",
		Tags:        []string{"Precificação"},
	}, func(ctx context.Context, in *bodyIn[dto.SimulacaoRequest]) (*bodyOut[dto.SimulacaoResponse], error) {
		if err := middleware.Validate(in.Body); err != nil {
			return nil, middleware.HumaErr(ctx, err)
		}
		res, err := svc.Simular(ctx, handler.OpcoesSimulacao{Req: in.Body})
		if err != nil {
			return nil, middleware.HumaErr(ctx, err)
		}
		return &bodyOut[dto.SimulacaoResponse]{Body: res}, nil
	})
}

type transacaoOutput struct {
	Body     dto.TransacaoResponse
	Location string `header:"Location"`
}

func registrarCriarTransacao(api huma.API, svc *handler.ServicoPrecificacao) {
	huma.Register(api, huma.Operation{
		OperationID:   "criar-transacao",
		Method:        http.MethodPost,
		Path:          "/api/v1/transacoes",
		Summary:       "Cria uma transação de cessão",
		Description:   "Precifica e persiste a transação com status PENDENTE.",
		Tags:          []string{"Transações"},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, in *bodyIn[dto.TransacaoCreateRequest]) (*transacaoOutput, error) {
		if err := middleware.Validate(in.Body); err != nil {
			return nil, middleware.HumaErr(ctx, err)
		}
		t, err := svc.CriarTransacao(ctx, handler.OpcoesCriacao{Req: in.Body})
		if err != nil {
			return nil, middleware.HumaErr(ctx, err)
		}
		return &transacaoOutput{Body: handler.TransacaoParaDTO(t), Location: "/api/v1/transacoes/" + t.ID}, nil
	})
}

type listarTransacoesInput struct {
	Pagina  int `query:"pagina" default:"1" minimum:"1" doc:"Página, começando em 1"`
	Tamanho int `query:"tamanho" default:"20" minimum:"1" maximum:"100" doc:"Tamanho da página"`
}

func registrarListarTransacoes(api huma.API, txRepo *postgres.TransacaoRepo) {
	huma.Register(api, huma.Operation{
		OperationID: "listar-transacoes",
		Method:      http.MethodGet,
		Path:        "/api/v1/transacoes",
		Summary:     "Lista transações paginadas",
		Tags:        []string{"Transações"},
	}, func(ctx context.Context, in *listarTransacoesInput) (*bodyOut[dto.Paginada], error) {
		txs, total, err := txRepo.Listar(ctx, in.Pagina, in.Tamanho)
		if err != nil {
			return nil, middleware.HumaErr(ctx, err)
		}
		items := make([]dto.TransacaoResponse, 0, len(txs))
		for _, t := range txs {
			items = append(items, handler.TransacaoParaDTO(t))
		}
		totalPaginas := 0
		if total > 0 {
			totalPaginas = int((total + int64(in.Tamanho) - 1) / int64(in.Tamanho))
		}
		return &bodyOut[dto.Paginada]{Body: dto.Paginada{
			Total: total, Pagina: in.Pagina, Tamanho: in.Tamanho,
			TotalPaginas: totalPaginas, Items: items,
		}}, nil
	})
}

type porIDInput struct {
	ID string `path:"id"`
}

func registrarObterTransacao(api huma.API, txRepo *postgres.TransacaoRepo) {
	huma.Register(api, huma.Operation{
		OperationID: "obter-transacao",
		Method:      http.MethodGet,
		Path:        "/api/v1/transacoes/{id}",
		Summary:     "Detalha uma transação",
		Tags:        []string{"Transações"},
	}, func(ctx context.Context, in *porIDInput) (*bodyOut[dto.TransacaoResponse], error) {
		t, err := txRepo.PorID(ctx, in.ID)
		if err != nil {
			return nil, middleware.HumaErr(ctx, err)
		}
		return &bodyOut[dto.TransacaoResponse]{Body: handler.TransacaoParaDTO(t)}, nil
	})
}

func registrarLiquidarTransacao(api huma.API, txRepo *postgres.TransacaoRepo) {
	huma.Register(api, huma.Operation{
		OperationID: "liquidar-transacao",
		Method:      http.MethodPost,
		Path:        "/api/v1/transacoes/{id}/liquidacao",
		Summary:     "Liquida uma transação",
		Description: "Optimistic locking: falha com 409 se a versão ou o status não conferirem no momento do UPDATE.",
		Tags:        []string{"Transações"},
	}, func(ctx context.Context, in *porIDInput) (*bodyOut[dto.TransacaoResponse], error) {
		atual, err := txRepo.PorID(ctx, in.ID)
		if err != nil {
			return nil, middleware.HumaErr(ctx, err)
		}
		liquidada, err := txRepo.Liquidar(ctx, postgres.LiquidacaoInput{
			ID: in.ID, Versao: atual.Versao, LiquidadaEm: precificacao.OperacaoDataCorrente().UTC(),
		})
		if err != nil {
			return nil, middleware.HumaErr(ctx, err)
		}
		return &bodyOut[dto.TransacaoResponse]{Body: handler.TransacaoParaDTO(liquidada)}, nil
	})
}

func registrarCotacoes(api huma.API, repo *postgres.CotacaoRepo) {
	huma.Register(api, huma.Operation{
		OperationID:   "criar-cotacao",
		Method:        http.MethodPost,
		Path:          "/api/v1/cotacoes",
		Summary:       "Registra uma cotação de câmbio",
		Tags:          []string{"Câmbio"},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, in *bodyIn[dto.CotacaoRequest]) (*bodyOut[dto.CotacaoResponse], error) {
		if err := middleware.Validate(in.Body); err != nil {
			return nil, middleware.HumaErr(ctx, err)
		}
		taxa, err := money.FromString(in.Body.Taxa)
		if err != nil || !taxa.IsPositive() {
			return nil, middleware.HumaErr(ctx, middleware.Validacao("entrada_invalida", "verifique os campos informados",
				[]dto.ErroCampo{{Campo: "taxa", Motivo: "deve ser um decimal estritamente positivo"}}))
		}
		cot, err := repo.Criar(ctx, postgres.NovaCotacao{
			Base: in.Body.MoedaBase, Cotacao: in.Body.MoedaCotacao, Taxa: taxa, Inicio: in.Body.VigenciaInicio,
		})
		if err != nil {
			return nil, middleware.HumaErr(ctx, err)
		}
		return &bodyOut[dto.CotacaoResponse]{Body: handler.CotacaoParaDTO(cot)}, nil
	})
}

func registrarCriarTaxaBase(api huma.API, repo *postgres.TaxaBaseRepo) {
	huma.Register(api, huma.Operation{
		OperationID:   "criar-taxa-base",
		Method:        http.MethodPost,
		Path:          "/api/v1/taxas-base",
		Summary:       "Registra uma taxa base",
		Tags:          []string{"Câmbio"},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, in *bodyIn[dto.TaxaBaseRequest]) (*bodyOut[dto.TaxaBaseResponse], error) {
		if err := middleware.Validate(in.Body); err != nil {
			return nil, middleware.HumaErr(ctx, err)
		}
		taxa, err := money.FromString(in.Body.TaxaMensal)
		if err != nil || !taxa.IsPositive() {
			return nil, middleware.HumaErr(ctx, middleware.Validacao("entrada_invalida", "verifique os campos informados",
				[]dto.ErroCampo{{Campo: "taxa_mensal", Motivo: "deve ser um decimal estritamente positivo"}}))
		}
		t, err := repo.Criar(ctx, postgres.NovaTaxaBase{
			Moeda: in.Body.Moeda, Taxa: taxa, Inicio: in.Body.VigenciaInicio,
		})
		if err != nil {
			return nil, middleware.HumaErr(ctx, err)
		}
		return &bodyOut[dto.TaxaBaseResponse]{Body: handler.TaxaBaseParaDTO(t)}, nil
	})
}

type taxaBaseVigenteInput struct {
	Moeda string `query:"moeda"`
}

func registrarTaxaBaseVigente(api huma.API, repo *postgres.TaxaBaseRepo) {
	huma.Register(api, huma.Operation{
		OperationID: "taxa-base-vigente",
		Method:      http.MethodGet,
		Path:        "/api/v1/taxas-base/vigente",
		Summary:     "Consulta a taxa base vigente na data corrente",
		Tags:        []string{"Câmbio"},
	}, func(ctx context.Context, in *taxaBaseVigenteInput) (*bodyOut[dto.TaxaBaseResponse], error) {
		if in.Moeda == "" {
			return nil, middleware.HumaErrStatus(ctx, middleware.StatusInfo{
				Status: http.StatusUnprocessableEntity, Codigo: "entrada_invalida",
			}, "parâmetro moeda é obrigatório")
		}
		t, err := repo.Vigente(ctx, in.Moeda, precificacao.OperacaoDataCorrente())
		if err != nil {
			return nil, middleware.HumaErr(ctx, err)
		}
		return &bodyOut[dto.TaxaBaseResponse]{Body: handler.TaxaBaseParaDTO(t)}, nil
	})
}

func registrarTiposRecebivel(api huma.API, repo *postgres.TipoRecebivelRepo) {
	type itemsOutput struct {
		Body struct {
			Items []dto.TipoRecebivelResponse `json:"items"`
		}
	}
	huma.Register(api, huma.Operation{
		OperationID: "listar-tipos-recebivel",
		Method:      http.MethodGet,
		Path:        "/api/v1/tipos-recebivel",
		Summary:     "Lista os tipos de recebível ativos",
		Tags:        []string{"Precificação"},
	}, func(ctx context.Context, in *struct{}) (*itemsOutput, error) {
		tipos, err := repo.Listar(ctx)
		if err != nil {
			return nil, middleware.HumaErr(ctx, err)
		}
		out := &itemsOutput{}
		out.Body.Items = make([]dto.TipoRecebivelResponse, 0, len(tipos))
		for _, t := range tipos {
			out.Body.Items = append(out.Body.Items, dto.TipoRecebivelResponse{Codigo: t.Codigo, Nome: t.Nome})
		}
		return out, nil
	})
}
