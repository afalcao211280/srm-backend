package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/srm-asset/srm-backend/internal/demo"
	"github.com/srm-asset/srm-backend/internal/infra/postgres"
	"github.com/srm-asset/srm-backend/internal/platform/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).
			Error("configuração inválida", slog.String("erro", err.Error()))
		os.Exit(1)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool, err := postgres.NewPool(ctx, postgres.Config{
		DatabaseURL: cfg.DatabaseURL,
		MaxConns:    5,
		ConnTimeout: 5 * time.Second,
	})
	if err != nil {
		logger.Error("pool", slog.String("erro", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()
	if err := demo.Nova(pool, logger).Executar(ctx); err != nil {
		logger.Error("carga", slog.String("erro", err.Error()))
		os.Exit(1)
	}
}
