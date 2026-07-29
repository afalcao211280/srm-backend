// Package handler implementa a camada de aplicação: orquestra o motor de
// precificação, os repositórios e o cliente de câmbio, sem conter regra de
// negócio. A regra mora no domínio; a persistência mora na infra.
package handler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/srm-asset/srm-backend/internal/app/dto"
	"github.com/srm-asset/srm-backend/internal/app/middleware"
	"github.com/srm-asset/srm-backend/internal/domain/money"
	"github.com/srm-asset/srm-backend/internal/domain/precificacao"
	"github.com/srm-asset/srm-backend/internal/domain/recebivel"
	"github.com/srm-asset/srm-backend/internal/infra/cambio"
	"github.com/srm-asset/srm-backend/internal/infra/postgres"
)

type ServicoPrecificacao struct {
	motor      *precificacao.Motor
	moedas     *postgres.MoedaRepo
	tipos      *postgres.TipoRecebivelRepo
	cedentes   *postgres.CedenteRepo
	cotacoes   *postgres.CotacaoRepo
	taxas      *postgres.TaxaBaseRepo
	txs        *postgres.TransacaoRepo
	clienteCot cambio.Cliente
	logger     *slog.Logger
}

type Deps struct {
	Motor    *precificacao.Motor
	Moedas   *postgres.MoedaRepo
	Tipos    *postgres.TipoRecebivelRepo
	Cedentes *postgres.CedenteRepo
	Cotacoes *postgres.CotacaoRepo
	Taxas    *postgres.TaxaBaseRepo
	Txs      *postgres.TransacaoRepo
	Cliente  cambio.Cliente
	Logger   *slog.Logger
}

func NovoServicoPrecificacao(opts Deps) *ServicoPrecificacao {
	return &ServicoPrecificacao{
		motor:      opts.Motor,
		moedas:     opts.Moedas,
		tipos:      opts.Tipos,
		cedentes:   opts.Cedentes,
		cotacoes:   opts.Cotacoes,
		taxas:      opts.Taxas,
		txs:        opts.Txs,
		clienteCot: opts.Cliente,
		logger:     opts.Logger,
	}
}

// OpcoesSimulacao agrupa os parâmetros do use case de simulação,
// atendendo ao limite S107 (≤3 params incluindo context.Context).
type OpcoesSimulacao struct {
	Req       dto.SimulacaoRequest
	Persistir bool
}

func (s *ServicoPrecificacao) Simular(ctx context.Context, op OpcoesSimulacao) (dto.SimulacaoResponse, error) {
	entrada, err := s.prepararEntrada(ctx, op.Req)
	if err != nil {
		return dto.SimulacaoResponse{}, err
	}
	resultado, err := s.motor.Precificar(entrada)
	if err != nil {
		return dto.SimulacaoResponse{}, middleware.ClassificarErro(err)
	}
	return dto.SimulacaoResponse{
		ValorPresente:  resultado.ValorPresente8Casas.String(),
		ValorLiquido:   resultado.ValorLiquido.String(),
		Desagio:        resultado.Desagio.String(),
		MoedaTitulo:    resultado.MoedaTitulo,
		MoedaPagamento: resultado.MoedaPagamento,
		DataOperacao:   entrada.DataOperacao.Format("2006-01-02"),
		DataVencimento: entrada.DataVencimento.Format("2006-01-02"),
	}, nil
}

// OpcoesCriacao agrupa parâmetros do use case de criação de transação.
type OpcoesCriacao struct {
	Req dto.TransacaoCreateRequest
}

func (s *ServicoPrecificacao) CriarTransacao(ctx context.Context, op OpcoesCriacao) (postgres.Transacao, error) {
	entrada, err := s.prepararEntrada(ctx, dto.SimulacaoRequest(op.Req))
	if err != nil {
		return postgres.Transacao{}, err
	}
	resultado, err := s.motor.Precificar(entrada)
	if err != nil {
		return postgres.Transacao{}, middleware.ClassificarErro(err)
	}
	refs, err := s.resolverReferencias(ctx, op.Req)
	if err != nil {
		return postgres.Transacao{}, err
	}
	id, err := gerarUUID()
	if err != nil {
		return postgres.Transacao{}, err
	}
	t := montarTransacao(id, refs, calculo{Entrada: entrada, Resultado: resultado})
	if err := s.txs.Criar(ctx, t); err != nil {
		return postgres.Transacao{}, fmt.Errorf("criar transação: %w", err)
	}
	return s.txs.PorID(ctx, id)
}

// referencias são as entidades resolvidas por código/id para montar a
// transação (S107: agrupa o retorno de resolverReferencias).
type referencias struct {
	Cedente        postgres.Cedente
	Tipo           postgres.TipoRecebivel
	MoedaTitulo    postgres.Moeda
	MoedaPagamento postgres.Moeda
}

func (s *ServicoPrecificacao) resolverReferencias(ctx context.Context, req dto.TransacaoCreateRequest) (referencias, error) {
	cedente, err := s.cedentes.PorID(ctx, req.CedenteID)
	if err != nil {
		return referencias{}, fmt.Errorf("cedente: %w", err)
	}
	tipo, err := s.tipos.PorCodigo(ctx, req.TipoRecebivel)
	if err != nil {
		return referencias{}, fmt.Errorf("tipo: %w", err)
	}
	moedaTitulo, err := s.moedas.PorCodigo(ctx, req.MoedaTitulo)
	if err != nil {
		return referencias{}, fmt.Errorf("moeda título: %w", err)
	}
	moedaPagamento, err := s.moedas.PorCodigo(ctx, req.MoedaPagamento)
	if err != nil {
		return referencias{}, fmt.Errorf("moeda pagamento: %w", err)
	}
	return referencias{Cedente: cedente, Tipo: tipo, MoedaTitulo: moedaTitulo, MoedaPagamento: moedaPagamento}, nil
}

// calculo agrupa a entrada e o resultado da precificação (S107: agrupa
// parâmetros de montarTransacao).
type calculo struct {
	Entrada   precificacao.Entrada
	Resultado precificacao.Resultado
}

func montarTransacao(id string, refs referencias, c calculo) postgres.Transacao {
	t := postgres.Transacao{
		ID:               id,
		CedenteID:        refs.Cedente.ID,
		TipoRecebivelID:  refs.Tipo.ID,
		MoedaTituloID:    refs.MoedaTitulo.ID,
		MoedaPagamentoID: refs.MoedaPagamento.ID,
		ValorFace:        c.Entrada.ValorFace,
		ValorPresente:    c.Resultado.ValorPresente8Casas,
		ValorLiquido:     c.Resultado.ValorLiquido,
		Desagio:          c.Resultado.Desagio,
		SpreadAplicado:   c.Resultado.SpreadAplicado,
		TaxaBaseAplicada: c.Resultado.TaxaBaseAplicada,
		DataOperacao:     c.Entrada.DataOperacao,
		DataVencimento:   c.Entrada.DataVencimento,
		Status:           postgres.StatusPendente,
	}
	if c.Resultado.CotacaoAplicada != nil {
		cot := *c.Resultado.CotacaoAplicada
		t.CotacaoAplicada = &cot
	}
	return t
}

func (s *ServicoPrecificacao) prepararEntrada(ctx context.Context, req dto.SimulacaoRequest) (precificacao.Entrada, error) {
	valorFace, err := money.FromString(req.ValorFace)
	if err != nil {
		return precificacao.Entrada{}, middleware.Novo("entrada_invalida", "valor_face inválido", 422)
	}
	moedas, err := s.resolverMoedas(ctx, req)
	if err != nil {
		return precificacao.Entrada{}, err
	}
	vencimento, dataOp, err := parseDatas(req)
	if err != nil {
		return precificacao.Entrada{}, err
	}
	taxaBase, err := s.validarTipoEResolverTaxa(ctx, tipoEData{Tipo: req.TipoRecebivel, MoedaTitulo: moedas.Titulo.Codigo, DataOp: dataOp})
	if err != nil {
		return precificacao.Entrada{}, err
	}
	entrada := precificacao.Entrada{
		ValorFace:      valorFace,
		DataOperacao:   dataOp,
		DataVencimento: vencimento,
		TipoRecebivel:  req.TipoRecebivel,
		TaxaBase:       taxaBase.TaxaMensal,
		MoedaTitulo:    moedas.Titulo.Codigo,
		MoedaPagamento: moedas.Pagamento.Codigo,
	}
	if err := s.aplicarCotacaoSeNecessario(ctx, &entrada); err != nil {
		return precificacao.Entrada{}, err
	}
	return entrada, nil
}

// moedasResolvidas são a moeda do título e a de pagamento já validadas
// (S107: agrupa o retorno de resolverMoedas).
type moedasResolvidas struct {
	Titulo    postgres.Moeda
	Pagamento postgres.Moeda
}

func (s *ServicoPrecificacao) resolverMoedas(ctx context.Context, req dto.SimulacaoRequest) (moedasResolvidas, error) {
	titulo, err := s.moedas.PorCodigo(ctx, req.MoedaTitulo)
	if err != nil {
		return moedasResolvidas{}, middleware.Novo("moeda_invalida", "moeda_titulo desconhecida", 422)
	}
	pagamento, err := s.moedas.PorCodigo(ctx, req.MoedaPagamento)
	if err != nil {
		return moedasResolvidas{}, middleware.Novo("moeda_invalida", "moeda_pagamento desconhecida", 422)
	}
	return moedasResolvidas{Titulo: titulo, Pagamento: pagamento}, nil
}

// parseDatas resolve vencimento e data de operação a partir da requisição.
// Data de operação ausente usa a data corrente em America/Sao_Paulo
// (ADR-017), nunca o fuso do container.
func parseDatas(req dto.SimulacaoRequest) (vencimento, dataOp time.Time, err error) {
	vencimento, err = time.Parse("2006-01-02", req.DataVencimento)
	if err != nil {
		return time.Time{}, time.Time{}, middleware.Novo("entrada_invalida", "data_vencimento inválida", 422)
	}
	dataOp = precificacao.OperacaoDataCorrente()
	if req.DataOperacao != "" {
		dataOp, err = time.Parse("2006-01-02", req.DataOperacao)
		if err != nil {
			return time.Time{}, time.Time{}, middleware.Novo("entrada_invalida", "data_operacao inválida", 422)
		}
	}
	return vencimento, dataOp, nil
}

// tipoEData agrupa o código do tipo de recebível, a moeda do título e a
// data de operação (S107: agrupa parâmetros de validarTipoEResolverTaxa).
type tipoEData struct {
	Tipo        string
	MoedaTitulo string
	DataOp      time.Time
}

func (s *ServicoPrecificacao) validarTipoEResolverTaxa(ctx context.Context, td tipoEData) (postgres.TaxaBase, error) {
	if _, err := s.tipos.PorCodigo(ctx, td.Tipo); err != nil {
		return postgres.TaxaBase{}, middleware.Novo("tipo_invalido", "tipo_recebivel desconhecido", 422)
	}
	taxaBase, err := s.taxas.Vigente(ctx, td.MoedaTitulo, td.DataOp)
	if err != nil {
		return postgres.TaxaBase{}, middleware.Novo("taxa_ausente", "sem taxa base vigente para a moeda na data da operação", 422)
	}
	return taxaBase, nil
}

func (s *ServicoPrecificacao) aplicarCotacaoSeNecessario(ctx context.Context, entrada *precificacao.Entrada) error {
	if entrada.MoedaTitulo == entrada.MoedaPagamento {
		return nil
	}
	return s.aplicarCotacao(ctx, entrada)
}

// aplicarCotacao busca a cotação vigente e grava a taxa em entrada.Cotacao,
// que o motor usa via Div (valor_pagamento = valor_titulo / taxa). Por
// ADR-006 a linha cadastrada tem base=moeda de pagamento e cotação=moeda do
// título (ex.: base=USD, cotação=BRL, taxa=5.4321 → 1 USD = 5.4321 BRL);
// consultar na direção inversa (base=título) nunca encontra a linha.
func (s *ServicoPrecificacao) aplicarCotacao(ctx context.Context, entrada *precificacao.Entrada) error {
	cot, err := s.cotacoes.Vigente(ctx, postgres.ParCotacao{
		Base: entrada.MoedaPagamento, Cotacao: entrada.MoedaTitulo, Em: entrada.DataOperacao,
	})
	if err != nil {
		return middleware.Novo("cotacao_ausente", "sem cotação vigente para o par de moedas", 422)
	}
	entrada.Cotacao = cot.Taxa
	entrada.TemCotacao = true
	return nil
}

// EstrategiasDisponiveis devolve os códigos de tipo de recebível registrados
// na strategy registry. Usado pelo handler de tipos e para validação.
func EstrategiasDisponiveis(r recebivel.Registry) []string { return r.Codigos() }
