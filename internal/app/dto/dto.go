// Package dto define o contrato HTTP: payloads de entrada e saída, e o
// formato padronizado de erro. Todos os valores monetários trafegam como
// string, conforme ADR-005 e o requisito da spec.
package dto

import "time"

type ErroCorpo struct {
	Codigo        string         `json:"codigo"`
	Mensagem      string         `json:"mensagem"`
	CorrelationID string         `json:"correlation_id"`
	Campos        []ErroCampo    `json:"campos,omitempty"`
}

type ErroCampo struct {
	Campo  string `json:"campo"`
	Motivo string `json:"motivo"`
}

type SimulacaoRequest struct {
	CedenteID       int64  `json:"cedente_id" validate:"required,gt=0"`
	TipoRecebivel   string `json:"tipo_recebivel" validate:"required"`
	ValorFace       string `json:"valor_face" validate:"required"`
	MoedaTitulo     string `json:"moeda_titulo" validate:"required,len=3"`
	MoedaPagamento  string `json:"moeda_pagamento" validate:"required,len=3"`
	DataVencimento  string `json:"data_vencimento" validate:"required"`
	DataOperacao    string `json:"data_operacao,omitempty"`
}

type SimulacaoResponse struct {
	ValorPresente string `json:"valor_presente"`
	ValorLiquido  string `json:"valor_liquido"`
	Desagio       string `json:"desagio"`
	MoedaTitulo   string `json:"moeda_titulo"`
	MoedaPagamento string `json:"moeda_pagamento"`
	DataOperacao  string `json:"data_operacao"`
	DataVencimento string `json:"data_vencimento"`
}

type TransacaoCreateRequest struct {
	CedenteID      int64  `json:"cedente_id" validate:"required,gt=0"`
	TipoRecebivel  string `json:"tipo_recebivel" validate:"required"`
	ValorFace      string `json:"valor_face" validate:"required"`
	MoedaTitulo    string `json:"moeda_titulo" validate:"required,len=3"`
	MoedaPagamento string `json:"moeda_pagamento" validate:"required,len=3"`
	DataVencimento string `json:"data_vencimento" validate:"required"`
	DataOperacao   string `json:"data_operacao,omitempty"`
}

type TransacaoResponse struct {
	ID                string     `json:"id"`
	Status            string     `json:"status"`
	Versao            int32      `json:"versao"`
	ValorFace         string     `json:"valor_face"`
	ValorPresente     string     `json:"valor_presente"`
	ValorLiquido      string     `json:"valor_liquido"`
	Desagio           string     `json:"desagio"`
	MoedaTitulo       string     `json:"moeda_titulo"`
	MoedaPagamento    string     `json:"moeda_pagamento"`
	DataOperacao      string     `json:"data_operacao"`
	DataVencimento    string     `json:"data_vencimento"`
	LiquidadaEm       *time.Time `json:"liquidada_em,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type CotacaoRequest struct {
	MoedaBase      string    `json:"moeda_base" validate:"required,len=3"`
	MoedaCotacao   string    `json:"moeda_cotacao" validate:"required,len=3"`
	Taxa           string    `json:"taxa" validate:"required"`
	VigenciaInicio time.Time `json:"vigencia_inicio" validate:"required"`
}

type TaxaBaseRequest struct {
	Moeda          string    `json:"moeda" validate:"required,len=3"`
	TaxaMensal     string    `json:"taxa_mensal" validate:"required"`
	VigenciaInicio time.Time `json:"vigencia_inicio" validate:"required"`
}

type TipoRecebivelResponse struct {
	Codigo string `json:"codigo"`
	Nome   string `json:"nome"`
}

type Paginada struct {
	Total      int64       `json:"total"`
	Pagina     int         `json:"pagina"`
	Tamanho    int         `json:"tamanho"`
	TotalPaginas int       `json:"total_paginas"`
	Items      interface{} `json:"items"`
}
