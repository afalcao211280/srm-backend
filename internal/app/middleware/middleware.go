// Package middleware concentra a cadeia de middlewares HTTP com base em gin:
// correlation ID, recovery de panic, logger estruturado e a tradução de erros
// de domínio em respostas padronizadas. A função ClassificarErro mapeia os
// erros do domínio e da persistência para o formato ErroCorpo.
package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/srm-asset/srm-backend/internal/app/dto"
	"github.com/srm-asset/srm-backend/internal/infra/postgres"
)

type ctxKey string

const correlationIDKey ctxKey = "X-Correlation-ID"
const correlationHeader = "X-Correlation-ID"

var validate = validator.New(validator.WithRequiredStructEnabled())

// CorrelationID devolve o identificador de correlação da requisição.
func CorrelationID(c *gin.Context) string {
	if v, ok := c.Get(string(correlationIDKey)); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// CorrelationIDFromContext lê o identificador de correlação a partir de um
// context.Context puro — usado pelos operations Huma, que recebem apenas
// context.Context (não *gin.Context). Correlation() grava o mesmo valor
// nos dois lugares na mesma requisição.
func CorrelationIDFromContext(ctx context.Context) string {
	if s, ok := ctx.Value(correlationIDKey).(string); ok {
		return s
	}
	return ""
}

func Correlation() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(correlationHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(string(correlationIDKey), id)
		c.Header(correlationHeader, id)
		ctx := context.WithValue(c.Request.Context(), correlationIDKey, id)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logger.Info("request",
			slog.String("correlation_id", CorrelationID(c)),
			slog.String("metodo", c.Request.Method),
			slog.String("rota", c.FullPath()),
			slog.Int("status", c.Writer.Status()),
		)
	}
}

func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic recuperado",
					slog.Any("recuperado", rec),
					slog.String("correlation_id", CorrelationID(c)),
					slog.String("stack", string(debug.Stack())),
				)
				RespondError(c, NewAPIError("erro_interno", "falha interna do servidor", http.StatusInternalServerError))
			}
		}()
		c.Next()
	}
}

// APIError representa um erro de domínio mapeado para resposta HTTP.
type APIError struct {
	Codigo   string
	Mensagem string
	Status   int
	Campos   []dto.ErroCampo
}

func (e *APIError) Error() string { return e.Codigo + ": " + e.Mensagem }

func NewAPIError(codigo, mensagem string, status int) *APIError {
	return &APIError{Codigo: codigo, Mensagem: mensagem, Status: status}
}

func Novo(codigo, mensagem string, status int) *APIError {
	return NewAPIError(codigo, mensagem, status)
}

func Validacao(codigo, mensagem string, campos []dto.ErroCampo) *APIError {
	return &APIError{Codigo: codigo, Mensagem: mensagem, Status: http.StatusUnprocessableEntity, Campos: campos}
}

// RespondError serializa o corpo padronizado de erro.
func RespondError(c *gin.Context, err error) {
	mapped := ClassificarErro(err)
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.AbortWithStatusJSON(mapped.Status, dto.ErroCorpo{
		Codigo:        mapped.Codigo,
		Mensagem:      mapped.Mensagem,
		CorrelationID: CorrelationID(c),
		Campos:        mapped.Campos,
	})
}

// ClassificarErro mapeia erros de domínio e persistência para APIError,
// sem vazar detalhe técnico (a stack trace fica no log do middleware).
func ClassificarErro(err error) *APIError {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	if errors.Is(err, postgres.ErrNaoEncontrado) {
		return NewAPIError("nao_encontrado", "recurso não encontrado", http.StatusNotFound)
	}
	if errors.Is(err, postgres.ErroConflitoLiquidacao) {
		return NewAPIError("conflito_liquidacao", "conflito de versão ou status na liquidação", http.StatusConflict)
	}
	var verr validator.ValidationErrors
	if errors.As(err, &verr) {
		campos := make([]dto.ErroCampo, 0, len(verr))
		for _, fe := range verr {
			campos = append(campos, dto.ErroCampo{
				Campo:  fe.Field(),
				Motivo: motivoValidacao(fe),
			})
		}
		return &APIError{
			Codigo:   "entrada_invalida",
			Mensagem: "verifique os campos informados",
			Status:   http.StatusUnprocessableEntity,
			Campos:   campos,
		}
	}
	return NewAPIError("erro_interno", "falha interna do servidor", http.StatusInternalServerError)
}

func motivoValidacao(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "campo obrigatório"
	case "gt":
		return "deve ser maior que zero"
	case "len":
		return "tamanho inválido"
	default:
		return "valor inválido"
	}
}

func Validate(s any) error {
	if err := validate.Struct(s); err != nil {
		return ClassificarErro(err)
	}
	return nil
}
