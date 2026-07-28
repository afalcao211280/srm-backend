// Package cambio implementa o cliente de cotação de câmbio mockado atrás
// de uma interface, com circuit breaker e timeout (ADR do design).
package cambio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/sony/gobreaker/v2"
)

type Opcoes struct {
	URL     string
	Timeout time.Duration
}

type Cliente interface {
	Cotacao(ctx context.Context, base, cotacao string) (string, error)
}

type cliente struct {
	url     string
	http    *http.Client
	cb      *gobreaker.CircuitBreaker[string]
	logger  *slog.Logger
}

func NovoCliente(op Opcoes, logger *slog.Logger) Cliente {
	if op.Timeout <= 0 {
		op.Timeout = 2 * time.Second
	}
	cb := gobreaker.NewCircuitBreaker[string](gobreaker.Settings{
		Name: "cambio",
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		Timeout: 30 * time.Second,
		OnStateChange: func(name string, from, to gobreaker.State) {
			logger.Warn("circuit breaker cambio",
				slog.String("nome", name),
				slog.String("de", from.String()),
				slog.String("para", to.String()),
			)
		},
	})
	return &cliente{
		url:    op.URL,
		http:   &http.Client{Timeout: op.Timeout},
		cb:     cb,
		logger: logger,
	}
}

func (c *cliente) Cotacao(ctx context.Context, base, cotacao string) (string, error) {
	if c.url == "" {
		return "", errors.New("cliente de câmbio desabilitado: URL vazia")
	}
	url := fmt.Sprintf("%s/cotacao?base=%s&cotacao=%s", c.url, base, cotacao)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("criar requisição: %w", err)
	}
	res, err := c.cb.Execute(func() (string, error) {
		req2 := req.WithContext(ctx)
		resp, err := c.http.Do(req2)
		if err != nil {
			return "", fmt.Errorf("chamada externa: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("status %d", resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("ler resposta: %w", err)
		}
		var payload struct {
			Taxa string `json:"taxa"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			return "", fmt.Errorf("decodificar resposta: %w", err)
		}
		if !decimalValido(payload.Taxa) {
			return "", fmt.Errorf("taxa inválida retornada: %s", payload.Taxa)
		}
		return payload.Taxa, nil
	})
	if err != nil {
		return "", fmt.Errorf("circuit breaker cambio: %w", err)
	}
	return res, nil
}

func decimalValido(s string) bool {
	if s == "" {
		return false
	}
	_, _, err := (&apd.Decimal{}).SetString(s)
	if err != nil {
		return false
	}
	// também tem que ser positivo
	r := new(big.Rat)
	if _, ok := r.SetString(s); !ok {
		return false
	}
	return r.Sign() > 0
}
