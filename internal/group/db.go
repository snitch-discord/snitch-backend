package group

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
	"snitch/snitchbe/internal/dbconfig"
	"snitch/snitchbe/internal/jwt"

	"snitch/snitchbe/pkg/ctxutil"

	"github.com/tursodatabase/libsql-client-go/libsql"
)

func NewGroupDB(ctx context.Context, tokenCache *jwt.TokenCache, config dbconfig.LibSQLConfig, groupID string) (*sql.DB, error) {
	slogger, ok := ctxutil.Value[*slog.Logger](ctx)
	if !ok {
		slogger = slog.Default()
	}

	httpURL, err := config.HttpURL()
	if err != nil {
		return nil, fmt.Errorf("get http url: %w", err)
	}

	connector, err := libsql.NewConnector(
		fmt.Sprintf("http://%s.%s", groupID, "db"),
		libsql.WithProxy(httpURL.String()),
		libsql.WithAuthToken(tokenCache.Get()),
	)
	if err != nil {
		slogger.ErrorContext(ctx, "Failed creating group connector", "Error", err)
		return nil, fmt.Errorf("couldnt create connector: %w", err)
	}

	db := sql.OpenDB(connector)

	return db, nil
}
