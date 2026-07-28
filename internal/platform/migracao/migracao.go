// Package migracao aplica as migrations SQL versionadas via golang-migrate
// antes do servidor começar a atender. Sem isso, o schema nunca é criado e
// a aplicação falha em toda consulta (achado de auditoria: nada aplicava
// as migrations apesar delas existirem em migrations/).
package migracao

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Aplicar roda todas as migrations pendentes em path contra databaseURL.
// Idempotente: se o schema já está na versão mais recente, não faz nada.
//
// path é sempre resolvido para absoluto antes de virar URL file://: para
// uma URL "file://<algo>", tudo até a primeira barra depois de "//" é
// interpretado como host, não como path — "file://migrations" tem host
// "migrations" e path VAZIO, que o driver silenciosamente resolve para
// "." em vez de falhar. Isso passou despercebido em teste local (onde o
// cwd por acaso fazia "." apontar para o lugar certo) e só quebrou dentro
// do container, onde "." não é a raiz esperada. Caminho absoluto elimina
// a ambiguidade: "file:///abs/caminho" não tem host.
func Aplicar(databaseURL, path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolver caminho absoluto de %q: %w", path, err)
	}
	dsn := strings.Replace(databaseURL, "postgres://", "pgx5://", 1)
	m, err := migrate.New("file://"+abs, dsn)
	if err != nil {
		return fmt.Errorf("abrir migrations: %w", err)
	}
	defer func() { _, _ = m.Close() }()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("aplicar migrations: %w", err)
	}
	return nil
}
