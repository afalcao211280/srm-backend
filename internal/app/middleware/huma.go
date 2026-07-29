package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/srm-asset/srm-backend/internal/app/dto"
)

// ErroHuma é o formato de erro que a API expõe via Huma: o mesmo
// dto.ErroCorpo que os handlers Gin usam via RespondError — uma única
// forma de erro em toda a API, documentada no OpenAPI gerado pelo Huma.
type ErroHuma struct {
	status        int
	Codigo        string          `json:"codigo"`
	Mensagem      string          `json:"mensagem"`
	CorrelationID string          `json:"correlation_id"`
	Campos        []dto.ErroCampo `json:"campos,omitempty"`
}

func (e *ErroHuma) Error() string  { return e.Codigo + ": " + e.Mensagem }
func (e *ErroHuma) GetStatus() int { return e.status }

// HumaErr converte um erro de domínio/persistência — o mesmo que
// RespondError usa nos handlers Gin — para o formato de erro do Huma,
// preservando o correlation_id da requisição.
func HumaErr(ctx context.Context, err error) error {
	mapped := ClassificarErro(err)
	return &ErroHuma{
		status:        mapped.Status,
		Codigo:        mapped.Codigo,
		Mensagem:      mapped.Mensagem,
		CorrelationID: CorrelationIDFromContext(ctx),
		Campos:        mapped.Campos,
	}
}

// StatusInfo agrupa status HTTP e código de erro (S107: agrupa parâmetros
// para HumaErrStatus caber em ≤3 incluindo context.Context).
type StatusInfo struct {
	Status int
	Codigo string
}

// HumaErrStatus constrói um erro ad hoc (sem passar por ClassificarErro)
// para validações locais de uma operação Huma, no mesmo formato e com o
// mesmo correlation_id de HumaErr.
func HumaErrStatus(ctx context.Context, si StatusInfo, mensagem string) error {
	return &ErroHuma{
		status:        si.Status,
		Codigo:        si.Codigo,
		Mensagem:      mensagem,
		CorrelationID: CorrelationIDFromContext(ctx),
	}
}

// InstalarErroHuma substitui a fábrica de erro do Huma para que falhas de
// validação/parsing do próprio framework (JSON malformado, campo
// obrigatório ausente detectado pelo schema) saiam no mesmo formato
// dto.ErroCorpo, e não no formato padrão do Huma (RFC 9457). Chamar uma vez
// na montagem do servidor, antes de registrar qualquer operação.
func InstalarErroHuma() {
	huma.NewErrorWithContext = func(ctx huma.Context, status int, msg string, errs ...error) huma.StatusError {
		campos := make([]dto.ErroCampo, 0, len(errs))
		for _, e := range errs {
			var det *huma.ErrorDetail
			if errors.As(e, &det) {
				campos = append(campos, dto.ErroCampo{Campo: det.Location, Motivo: det.Message})
				continue
			}
			if e != nil {
				campos = append(campos, dto.ErroCampo{Campo: "body", Motivo: e.Error()})
			}
		}
		correlation := ""
		if ctx != nil {
			correlation = CorrelationIDFromContext(ctx.Context())
		}
		return &ErroHuma{
			status:        status,
			Codigo:        codigoParaStatusHuma(status),
			Mensagem:      msg,
			CorrelationID: correlation,
			Campos:        campos,
		}
	}
}

func codigoParaStatusHuma(status int) string {
	switch status {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return "entrada_invalida"
	case http.StatusNotFound:
		return "nao_encontrado"
	case http.StatusConflict:
		return "conflito_liquidacao"
	default:
		return "erro_interno"
	}
}
