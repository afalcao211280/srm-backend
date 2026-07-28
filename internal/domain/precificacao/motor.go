// Package precificacao implementa o motor de cálculo do valor presente de
// um recebível, isolado de qualquer detalhe de persistência ou HTTP. A
// fórmula é Valor Presente = Valor Face / (1 + Taxa Base + Spread)^Prazo,
// com prazo fracionário em meses (dias corridos / 30) e conversão cambial
// aplicada ao final. Aritmética exclusivamente decimal.
package precificacao

import (
	"errors"
	"fmt"
	"time"

	"github.com/srm-asset/srm-backend/internal/domain/money"
	"github.com/srm-asset/srm-backend/internal/domain/recebivel"
)

type Entrada struct {
	ValorFace         money.Decimal
	DataOperacao      time.Time
	DataVencimento    time.Time
	TipoRecebivel     string
	TaxaBase          money.Decimal
	Cotacao           money.Decimal
	TemCotacao        bool
	MoedaPagamento    string
	MoedaTitulo       string
}

type Resultado struct {
	ValorPresente         money.Decimal
	ValorPresente8Casas   money.Decimal
	ValorLiquido          money.Decimal
	Desagio               money.Decimal
	SpreadAplicado        money.Decimal
	TaxaBaseAplicada      money.Decimal
	CotacaoAplicada       *money.Decimal
	MoedaPagamento        string
	MoedaTitulo           string
	Prazo                 money.Decimal
}

func (r Resultado) PrazoQuantizado() money.Decimal {
	d, _ := r.Prazo.Quantize(1)
	return d
}

type Motor struct {
	registry recebivel.Registry
}

func NewMotor(registry recebivel.Registry) *Motor {
	return &Motor{registry: registry}
}

func (m *Motor) Precificar(entrada Entrada) (Resultado, error) {
	if err := m.validarPrazo(entrada); err != nil {
		return Resultado{}, err
	}
	strategy, err := m.registry.Resolve(entrada.TipoRecebivel)
	if err != nil {
		return Resultado{}, err
	}
	spread, err := money.FromString(strategy.Spread())
	if err != nil {
		return Resultado{}, fmt.Errorf("spread inválido na strategy %s: %w", strategy.Codigo(), err)
	}
	base, err := money.MustFinite(money.New(1).Add(entrada.TaxaBase).Add(spread), "soma de base e spread")
	if err != nil {
		return Resultado{}, err
	}
	if !base.IsPositive() {
		return Resultado{}, &BaseNaoPositivaError{Base: base.String()}
	}
	prazo, err := calcularPrazo(entrada.DataOperacao, entrada.DataVencimento)
	if err != nil {
		return Resultado{}, err
	}
	fator, err := base.Pow(prazo)
	if err != nil {
		return Resultado{}, fmt.Errorf("exponenciação falhou: %w", err)
	}
	if _, err := money.MustFinite(fator, "fator de potenciação"); err != nil {
		return Resultado{}, err
	}
	valorPresente, err := entrada.ValorFace.Div(fator)
	if err != nil {
		return Resultado{}, fmt.Errorf("divisão do valor de face pelo fator: %w", err)
	}
	valorPresente, err = money.MustFinite(valorPresente, "valor presente")
	if err != nil {
		return Resultado{}, err
	}
	valorPresente8, err := valorPresente.Quantize(8)
	if err != nil {
		return Resultado{}, err
	}
	valorNaMoedaPagamento := valorPresente
	if entrada.MoedaTitulo != entrada.MoedaPagamento {
		if !entrada.TemCotacao {
			return Resultado{}, errors.New("operação cross-currency sem cotação")
		}
		if !entrada.Cotacao.IsPositive() {
			return Resultado{}, &CotacaoInvalidaError{Valor: entrada.Cotacao.String()}
		}
		valorNaMoedaPagamento, err = valorPresente.Div(entrada.Cotacao)
		if err != nil {
			return Resultado{}, fmt.Errorf("conversão cambial falhou: %w", err)
		}
		valorNaMoedaPagamento, err = money.MustFinite(valorNaMoedaPagamento, "valor convertido")
		if err != nil {
			return Resultado{}, err
		}
	}
	valorLiquido, err := valorNaMoedaPagamento.Quantize(2)
	if err != nil {
		return Resultado{}, err
	}
	desagio := entrada.ValorFace.Sub(valorLiquido)
	desagio, err = money.MustFinite(desagio, "deságio")
	if err != nil {
		return Resultado{}, err
	}
	res := Resultado{
		ValorPresente:       valorPresente,
		ValorPresente8Casas: valorPresente8,
		ValorLiquido:        valorLiquido,
		Desagio:             desagio,
		SpreadAplicado:      spread,
		TaxaBaseAplicada:    entrada.TaxaBase,
		MoedaPagamento:      entrada.MoedaPagamento,
		MoedaTitulo:         entrada.MoedaTitulo,
		Prazo:               prazo,
	}
	if entrada.MoedaTitulo != entrada.MoedaPagamento {
		c := entrada.Cotacao
		res.CotacaoAplicada = &c
	}
	return res, nil
}

func (m *Motor) validarPrazo(entrada Entrada) error {
	op := apenasData(entrada.DataOperacao)
	vc := apenasData(entrada.DataVencimento)
	if !vc.After(op) {
		return &PrazoInvalidoError{Operacao: op, Vencimento: vc}
	}
	return nil
}

func calcularPrazo(operacao, vencimento time.Time) (money.Decimal, error) {
	op := apenasData(operacao)
	vc := apenasData(vencimento)
	dias := int64(vc.Sub(op).Hours() / 24)
	if dias <= 0 {
		return money.Decimal{}, errors.New("prazo deve ser positivo")
	}
	diasDec, err := money.FromString(fmt.Sprintf("%d", dias))
	if err != nil {
		return money.Decimal{}, err
	}
	trinta, _ := money.FromString("30")
	return diasDec.Div(trinta)
}

func apenasData(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

type PrazoInvalidoError struct {
	Operacao   time.Time
	Vencimento time.Time
}

func (e *PrazoInvalidoError) Error() string {
	return "data de vencimento deve ser estritamente posterior à data de operação"
}

type BaseNaoPositivaError struct {
	Base string
}

func (e *BaseNaoPositivaError) Error() string {
	return "base da potência deve ser estritamente maior que zero, recebido " + e.Base
}

type CotacaoInvalidaError struct {
	Valor string
}

func (e *CotacaoInvalidaError) Error() string {
	return "cotação de câmbio deve ser estritamente positiva, recebido " + e.Valor
}
