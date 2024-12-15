package group

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
	"snitch/snitchbe/internal/dbconfig"

	"snitch/snitchbe/pkg/ctxutil"

	_ "github.com/tursodatabase/go-libsql"
)

func NewGroupDB(ctx context.Context, token string, config dbconfig.LibSQLConfig, groupID string) (*sql.DB, error) {
	slogger, ok := ctxutil.Value[*slog.Logger](ctx)
	if !ok {
		slogger = slog.Default()
	}

	httpURL, err := config.DatabaseURL(token)
	if err != nil {
		return nil, fmt.Errorf("get http url: %w", err)
	}

	db, err := sql.Open("libsql", httpURL.String())
	if err != nil {
		slogger.ErrorContext(ctx, "Failed creating group DB", "Error", err)
		return nil, fmt.Errorf("couldnt create group DB: %w", err)
	}

	return db, nil
}
