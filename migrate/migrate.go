package migrate

import (
	"context"
	"embed"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	// _ "github.com/ugent-library/muts/migrations"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func init() {
	goose.SetTableName("goose_migration")
	goose.SetBaseFS(migrationsFS)
}

func Up(ctx context.Context, conn string) error {
	db, err := goose.OpenDBWithDriver("pgx", conn)
	if err != nil {
		return err
	}
	defer db.Close()
	return goose.UpContext(ctx, db, "migrations")
}

func Down(ctx context.Context, conn string) error {
	db, err := goose.OpenDBWithDriver("pgx", conn)
	if err != nil {
		return err
	}
	defer db.Close()
	return goose.ResetContext(ctx, db, "migrations")
}
